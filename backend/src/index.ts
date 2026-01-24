/**
 * gg Backend - Cloudflare Worker
 * Handles Stripe subscriptions and license key management
 */

import Stripe from 'stripe';

interface Env {
  GG_KV: KVNamespace;
  STRIPE_SECRET_KEY: string;
  STRIPE_WEBHOOK_SECRET: string;
  LICENSE_SIGNING_KEY: string;
  ENVIRONMENT: string;
}

// gg Pro pricing
const GG_PRO_PRICE = 1500; // $15.00/month in cents
const GG_PRO_PRODUCT_NAME = 'gg Pro';

// License key format: gg_pro_<random>_<checksum>
function generateLicenseKey(email: string, signingKey: string): string {
  const random = crypto.randomUUID().replace(/-/g, '').slice(0, 16);
  const data = `${email}:${random}`;
  // Simple checksum (in production, use HMAC)
  const checksum = Array.from(data)
    .reduce((acc, char) => acc + char.charCodeAt(0), 0)
    .toString(16)
    .slice(-8);
  return `gg_pro_${random}_${checksum}`;
}

// Verify license key format
function isValidLicenseFormat(key: string): boolean {
  return /^gg_pro_[a-f0-9]{16}_[a-f0-9]{8}$/.test(key);
}

// CORS headers - restricted to ggdotdev.com
const corsHeaders = {
  'Access-Control-Allow-Origin': 'https://ggdotdev.com',
  'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type, Authorization',
};

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const path = url.pathname;

    // Handle CORS preflight
    if (request.method === 'OPTIONS') {
      return new Response(null, { headers: corsHeaders });
    }

    try {
      // Health check
      if (path === '/' || path === '/health') {
        return json({ status: 'ok', service: 'gg-backend', version: '1.0.0' });
      }

      // Create checkout session
      if (path === '/checkout' && request.method === 'POST') {
        return handleCheckout(request, env);
      }

      // Stripe webhook
      if (path === '/webhook' && request.method === 'POST') {
        return handleWebhook(request, env);
      }

      // Verify license
      if (path === '/verify' && request.method === 'POST') {
        return handleVerify(request, env);
      }

      // Get license for email (after purchase)
      if (path === '/license' && request.method === 'GET') {
        return handleGetLicense(request, env);
      }

      // Usage tracking (free tier)
      if (path === '/usage' && request.method === 'POST') {
        return handleUsage(request, env);
      }

      // Customer portal
      if (path === '/portal' && request.method === 'POST') {
        return handlePortal(request, env);
      }

      return json({ error: 'Not found' }, 404);
    } catch (error) {
      console.error('Error:', error);
      return json({ error: 'Internal server error' }, 500);
    }
  },
};

// Create Stripe checkout session
async function handleCheckout(request: Request, env: Env): Promise<Response> {
  const stripe = new Stripe(env.STRIPE_SECRET_KEY, { apiVersion: '2024-11-20.acacia' });

  const body = await request.json() as { email?: string };
  const email = body.email;

  if (!email) {
    return json({ error: 'Email required' }, 400);
  }

  // Check if already has active subscription
  const existingLicense = await env.GG_KV.get(`email:${email}`);
  if (existingLicense) {
    const licenseData = JSON.parse(existingLicense);
    if (licenseData.status === 'active') {
      return json({
        error: 'Already subscribed',
        license_key: licenseData.license_key
      }, 400);
    }
  }

  // Create or get customer
  let customer: Stripe.Customer;
  const customers = await stripe.customers.list({ email, limit: 1 });

  if (customers.data.length > 0) {
    customer = customers.data[0];
  } else {
    customer = await stripe.customers.create({ email });
  }

  // Create checkout session
  const session = await stripe.checkout.sessions.create({
    customer: customer.id,
    mode: 'subscription',
    line_items: [{
      price_data: {
        currency: 'usd',
        product_data: {
          name: GG_PRO_PRODUCT_NAME,
          description: 'Unlimited gg ask, priority API routing',
        },
        unit_amount: GG_PRO_PRICE,
        recurring: { interval: 'month' },
      },
      quantity: 1,
    }],
    success_url: 'https://ggdotdev.com/success?session_id={CHECKOUT_SESSION_ID}',
    cancel_url: 'https://ggdotdev.com/pricing',
    metadata: { email },
  });

  return json({
    checkout_url: session.url,
    session_id: session.id
  });
}

// Handle Stripe webhook
async function handleWebhook(request: Request, env: Env): Promise<Response> {
  const stripe = new Stripe(env.STRIPE_SECRET_KEY, { apiVersion: '2024-11-20.acacia' });

  const signature = request.headers.get('stripe-signature');
  if (!signature) {
    return json({ error: 'Missing signature' }, 400);
  }

  const body = await request.text();

  let event: Stripe.Event;
  try {
    event = stripe.webhooks.constructEvent(body, signature, env.STRIPE_WEBHOOK_SECRET);
  } catch (err) {
    console.error('Webhook signature verification failed:', err);
    return json({ error: 'Invalid signature' }, 400);
  }

  switch (event.type) {
    case 'checkout.session.completed': {
      const session = event.data.object as Stripe.Checkout.Session;
      await provisionLicense(session, env);
      break;
    }

    case 'customer.subscription.updated': {
      const subscription = event.data.object as Stripe.Subscription;
      await updateSubscriptionStatus(subscription, env);
      break;
    }

    case 'customer.subscription.deleted': {
      const subscription = event.data.object as Stripe.Subscription;
      await revokeLicense(subscription, env);
      break;
    }

    case 'invoice.payment_failed': {
      const invoice = event.data.object as Stripe.Invoice;
      await handlePaymentFailed(invoice, env);
      break;
    }
  }

  return json({ received: true });
}

// Provision license after successful checkout
async function provisionLicense(session: Stripe.Checkout.Session, env: Env): Promise<void> {
  const email = session.metadata?.email || session.customer_email;
  if (!email) {
    console.error('No email in session');
    return;
  }

  const licenseKey = generateLicenseKey(email, env.LICENSE_SIGNING_KEY);

  const licenseData = {
    license_key: licenseKey,
    email,
    customer_id: session.customer,
    subscription_id: session.subscription,
    status: 'active',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  // Store by email and by license key
  await env.GG_KV.put(`email:${email}`, JSON.stringify(licenseData));
  await env.GG_KV.put(`license:${licenseKey}`, JSON.stringify(licenseData));

  console.log(`Provisioned license for ${email}: ${licenseKey}`);
}

// Update subscription status
async function updateSubscriptionStatus(subscription: Stripe.Subscription, env: Env): Promise<void> {
  const customerId = subscription.customer as string;

  // Find license by customer ID
  const licenseKey = await findLicenseByCustomer(customerId, env);
  if (!licenseKey) return;

  const licenseData = await env.GG_KV.get(`license:${licenseKey}`);
  if (!licenseData) return;

  const data = JSON.parse(licenseData);
  data.status = subscription.status === 'active' ? 'active' : 'inactive';
  data.updated_at = new Date().toISOString();

  await env.GG_KV.put(`license:${licenseKey}`, JSON.stringify(data));
  await env.GG_KV.put(`email:${data.email}`, JSON.stringify(data));
}

// Revoke license on subscription cancellation
async function revokeLicense(subscription: Stripe.Subscription, env: Env): Promise<void> {
  const customerId = subscription.customer as string;

  const licenseKey = await findLicenseByCustomer(customerId, env);
  if (!licenseKey) return;

  const licenseData = await env.GG_KV.get(`license:${licenseKey}`);
  if (!licenseData) return;

  const data = JSON.parse(licenseData);
  data.status = 'revoked';
  data.revoked_at = new Date().toISOString();
  data.updated_at = new Date().toISOString();

  await env.GG_KV.put(`license:${licenseKey}`, JSON.stringify(data));
  await env.GG_KV.put(`email:${data.email}`, JSON.stringify(data));

  console.log(`Revoked license for ${data.email}`);
}

// Handle payment failure
async function handlePaymentFailed(invoice: Stripe.Invoice, env: Env): Promise<void> {
  const customerId = invoice.customer as string;

  const licenseKey = await findLicenseByCustomer(customerId, env);
  if (!licenseKey) return;

  const licenseData = await env.GG_KV.get(`license:${licenseKey}`);
  if (!licenseData) return;

  const data = JSON.parse(licenseData);
  data.status = 'payment_failed';
  data.updated_at = new Date().toISOString();

  await env.GG_KV.put(`license:${licenseKey}`, JSON.stringify(data));
  await env.GG_KV.put(`email:${data.email}`, JSON.stringify(data));

  console.log(`Payment failed for ${data.email}`);
}

// Find license by customer ID (scan - in production use secondary index)
async function findLicenseByCustomer(customerId: string, env: Env): Promise<string | null> {
  // This is inefficient - in production, store customer_id -> license_key mapping
  const list = await env.GG_KV.list({ prefix: 'license:' });

  for (const key of list.keys) {
    const data = await env.GG_KV.get(key.name);
    if (data) {
      const parsed = JSON.parse(data);
      if (parsed.customer_id === customerId) {
        return key.name.replace('license:', '');
      }
    }
  }

  return null;
}

// Verify license key
async function handleVerify(request: Request, env: Env): Promise<Response> {
  const body = await request.json() as { license_key?: string };
  const licenseKey = body.license_key;

  if (!licenseKey) {
    return json({ valid: false, error: 'License key required' }, 400);
  }

  if (!isValidLicenseFormat(licenseKey)) {
    return json({ valid: false, error: 'Invalid license format' });
  }

  const licenseData = await env.GG_KV.get(`license:${licenseKey}`);
  if (!licenseData) {
    return json({ valid: false, error: 'License not found' });
  }

  const data = JSON.parse(licenseData);

  return json({
    valid: data.status === 'active',
    status: data.status,
    email: data.email,
    created_at: data.created_at,
  });
}

// Get license for email
async function handleGetLicense(request: Request, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const email = url.searchParams.get('email');
  const sessionId = url.searchParams.get('session_id');

  if (!email && !sessionId) {
    return json({ error: 'Email or session_id required' }, 400);
  }

  if (sessionId) {
    // Verify session and get email
    const stripe = new Stripe(env.STRIPE_SECRET_KEY, { apiVersion: '2024-11-20.acacia' });
    try {
      const session = await stripe.checkout.sessions.retrieve(sessionId);
      if (session.payment_status !== 'paid') {
        return json({ error: 'Payment not completed' }, 400);
      }
      const sessionEmail = session.metadata?.email || session.customer_email;
      if (sessionEmail) {
        const licenseData = await env.GG_KV.get(`email:${sessionEmail}`);
        if (licenseData) {
          const data = JSON.parse(licenseData);
          return json({ license_key: data.license_key, email: data.email });
        }
      }
    } catch (err) {
      return json({ error: 'Invalid session' }, 400);
    }
  }

  if (email) {
    const licenseData = await env.GG_KV.get(`email:${email}`);
    if (licenseData) {
      const data = JSON.parse(licenseData);
      if (data.status === 'active') {
        return json({ license_key: data.license_key, email: data.email });
      }
    }
  }

  return json({ error: 'No active license found' }, 404);
}

// Track free tier usage
async function handleUsage(request: Request, env: Env): Promise<Response> {
  const body = await request.json() as {
    machine_id?: string;
    command?: string;
    license_key?: string;
  };

  const machineId = body.machine_id;
  const command = body.command || 'unknown';
  const licenseKey = body.license_key;

  // If has valid license, unlimited usage
  if (licenseKey) {
    const licenseData = await env.GG_KV.get(`license:${licenseKey}`);
    if (licenseData) {
      const data = JSON.parse(licenseData);
      if (data.status === 'active') {
        return json({ allowed: true, tier: 'pro', remaining: -1 });
      }
    }
  }

  // Free tier rate limiting
  if (!machineId) {
    return json({ error: 'Machine ID required for free tier' }, 400);
  }

  const today = new Date().toISOString().slice(0, 10);
  const usageKey = `usage:${machineId}:${today}`;

  const currentUsage = await env.GG_KV.get(usageKey);
  const count = currentUsage ? parseInt(currentUsage) : 0;

  const FREE_LIMIT = 10;
  const GRACE = 2;
  const totalLimit = FREE_LIMIT + GRACE;

  if (count >= totalLimit) {
    return json({
      allowed: false,
      tier: 'free',
      used: count,
      limit: FREE_LIMIT,
      message: 'Daily limit reached. Upgrade to Pro for unlimited: https://ggdotdev.com/pro'
    });
  }

  // Increment usage
  await env.GG_KV.put(usageKey, (count + 1).toString(), { expirationTtl: 86400 * 2 });

  const inGrace = count >= FREE_LIMIT;

  return json({
    allowed: true,
    tier: 'free',
    used: count + 1,
    limit: FREE_LIMIT,
    remaining: Math.max(0, FREE_LIMIT - count - 1),
    grace: inGrace,
    message: inGrace ? `Grace period: ${GRACE - (count - FREE_LIMIT) - 1} calls remaining` : undefined,
  });
}

// Create customer portal session
async function handlePortal(request: Request, env: Env): Promise<Response> {
  const stripe = new Stripe(env.STRIPE_SECRET_KEY, { apiVersion: '2024-11-20.acacia' });

  const body = await request.json() as { email?: string; license_key?: string };
  const email = body.email;
  const licenseKey = body.license_key;

  let customerId: string | null = null;

  if (licenseKey) {
    const licenseData = await env.GG_KV.get(`license:${licenseKey}`);
    if (licenseData) {
      const data = JSON.parse(licenseData);
      customerId = data.customer_id;
    }
  } else if (email) {
    const licenseData = await env.GG_KV.get(`email:${email}`);
    if (licenseData) {
      const data = JSON.parse(licenseData);
      customerId = data.customer_id;
    }
  }

  if (!customerId) {
    return json({ error: 'No subscription found' }, 404);
  }

  const session = await stripe.billingPortal.sessions.create({
    customer: customerId,
    return_url: 'https://ggdotdev.com/account',
  });

  return json({ portal_url: session.url });
}

// Helper to return JSON response
function json(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      'Content-Type': 'application/json',
      ...corsHeaders,
    },
  });
}
