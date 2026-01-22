#!/usr/bin/env node
/**
 * gg postinstall script
 * Downloads the correct binary for the current platform
 */

const https = require('https');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const VERSION = '0.9.6.1';
const REPO = 'cyclecore-dev/gg';

// Map Node.js platform/arch to gg binary names
const PLATFORM_MAP = {
  'linux-x64': 'gg_linux_x86_64',
  'linux-arm64': 'gg_linux_aarch64',
  'darwin-x64': 'gg_darwin_x86_64',
  'darwin-arm64': 'gg_darwin_aarch64',
};

function getPlatformBinary() {
  const platform = process.platform;
  const arch = process.arch;
  const key = `${platform}-${arch}`;

  const binary = PLATFORM_MAP[key];
  if (!binary) {
    console.error(`Unsupported platform: ${platform}-${arch}`);
    console.error('Supported platforms: linux-x64, linux-arm64, darwin-x64, darwin-arm64');
    process.exit(1);
  }

  return binary;
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    https.get(url, (response) => {
      // Handle redirects (GitHub releases use them)
      if (response.statusCode === 302 || response.statusCode === 301) {
        download(response.headers.location, dest).then(resolve).catch(reject);
        return;
      }

      if (response.statusCode !== 200) {
        reject(new Error(`Download failed: ${response.statusCode}`));
        return;
      }

      response.pipe(file);
      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {});
      reject(err);
    });
  });
}

async function main() {
  const binaryName = getPlatformBinary();
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${binaryName}`;

  const binDir = path.join(__dirname, 'bin');
  const binPath = path.join(binDir, 'gg');

  // Create bin directory
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  console.log(`Downloading gg v${VERSION} for ${process.platform}-${process.arch}...`);

  try {
    await download(url, binPath);
    fs.chmodSync(binPath, 0o755);
    console.log('gg installed successfully!');
    console.log('Run "gg init" to configure, or "gg maaza" for status');
  } catch (err) {
    console.error('Failed to download gg:', err.message);
    console.error('');
    console.error('Alternative install:');
    console.error('  curl -fsSL https://raw.githubusercontent.com/cyclecore-dev/gg/main/gg.sh | sh');
    process.exit(1);
  }
}

main();
