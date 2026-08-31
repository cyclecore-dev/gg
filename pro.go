package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/BurntSushi/toml"
)

func checkProTier(cfg *Config) bool {
	return cfg.Secrets.ProLicenseKey != "" &&
		strings.HasPrefix(cfg.Secrets.ProLicenseKey, "gg_pro_")
}

// handleUpgrade starts the Pro upgrade flow
func handleUpgrade() {
	fmt.Println("gg Pro — $15/month")
	fmt.Println()
	fmt.Println("Benefits:")
	fmt.Println("  - Unlimited gg ask (code generation)")
	fmt.Println("  - Unlimited CLI2CLI (a2a) calls")
	fmt.Println("  - Priority API routing")
	fmt.Println("  - Support development")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if email == "" || !strings.Contains(email, "@") {
		fmt.Println("Invalid email")
		return
	}

	// Create checkout session
	fmt.Println()
	fmt.Println("Creating checkout session...")

	reqBody, _ := json.Marshal(map[string]string{"email": email})
	resp, err := http.Post(getBackendURL()+"/checkout", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Try: https://ggdotdev.com/pro")
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if errMsg, ok := result["error"].(string); ok {
		if licenseKey, hasKey := result["license_key"].(string); hasKey {
			fmt.Printf("Already subscribed! Your license key: %s\n", licenseKey)
			fmt.Println()
			fmt.Println("Add to ~/.gg/config.toml:")
			fmt.Printf("  pro_license_key = \"%s\"\n", licenseKey)
			return
		}
		fmt.Printf("Error: %s\n", errMsg)
		return
	}

	if checkoutURL, ok := result["checkout_url"].(string); ok {
		fmt.Println()
		fmt.Println("Opening checkout...")
		fmt.Println(checkoutURL)
		fmt.Println()
		fmt.Println("After payment, run: gg pro --activate")

		// Try to open browser
		exec.Command("open", checkoutURL).Start()
		exec.Command("xdg-open", checkoutURL).Start()
	}
}

// handlePro shows Pro status or activates license
func handlePro() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("Run 'gg init' first")
		return
	}

	// Check for --activate flag
	if len(os.Args) > 2 && os.Args[2] == "--activate" {
		activateProLicense(cfg)
		return
	}

	// Check for --portal flag (manage subscription)
	if len(os.Args) > 2 && os.Args[2] == "--portal" {
		openCustomerPortal(cfg)
		return
	}

	// Show current status
	fmt.Println("gg Pro Status")
	fmt.Println()

	if checkProTier(cfg) {
		fmt.Println("Status: Active")
		fmt.Printf("License: %s...%s\n",
			cfg.Secrets.ProLicenseKey[:12],
			cfg.Secrets.ProLicenseKey[len(cfg.Secrets.ProLicenseKey)-4:])
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  gg pro --portal    Manage subscription")
	} else {
		fmt.Println("Status: Free tier")
		fmt.Println()
		allowed, inGrace, remaining := checkA2ALimit(cfg)
		if !allowed {
			fmt.Println("Daily limit: Reached")
		} else if inGrace {
			fmt.Println("Daily limit: In grace period")
		} else {
			fmt.Printf("Daily limit: %d remaining\n", remaining)
		}
		fmt.Println()
		fmt.Println("Upgrade: gg upgrade")
	}
}

// activateProLicense fetches and saves the license key after payment
func activateProLicense(cfg *Config) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter email used for purchase: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if email == "" {
		fmt.Println("Email required")
		return
	}

	fmt.Println("Checking for license...")

	resp, err := http.Get(getBackendURL() + "/license?email=" + url.QueryEscape(email))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if errMsg, ok := result["error"].(string); ok {
		fmt.Printf("Error: %s\n", errMsg)
		fmt.Println()
		fmt.Println("If you just purchased, wait a moment and try again.")
		fmt.Println("Or open an issue: github.com/cyclecore-dev/gg/issues")
		return
	}

	if licenseKey, ok := result["license_key"].(string); ok {
		// Save license key to config
		cfg.Secrets.ProLicenseKey = licenseKey
		cfg.GG.Tier = "pro"

		// Save config and secrets
		homeDir, _ := os.UserHomeDir()
		ggDir := filepath.Join(homeDir, ".gg")
		configPath := filepath.Join(ggDir, "config.toml")

		f, err := os.Create(configPath)
		if err != nil {
			fmt.Printf("Failed to save config: %v\n", err)
			fmt.Printf("Manually add to ~/.gg/config.toml:\n")
			fmt.Printf("  pro_license_key = \"%s\"\n", licenseKey)
			return
		}
		defer f.Close()

		if err := toml.NewEncoder(f).Encode(cfg); err != nil {
			fmt.Printf("Failed to write config: %v\n", err)
			return
		}

		// Re-encrypt secrets with new license key
		keyPath := filepath.Join(ggDir, ".key")
		keyData, err := os.ReadFile(keyPath)
		if err == nil {
			identity, err := age.ParseX25519Identity(string(keyData))
			if err == nil {
				secretsPath := filepath.Join(ggDir, "secrets")
				encryptSecrets(cfg.Secrets, identity, secretsPath)
			}
		}

		fmt.Println()
		fmt.Println("Pro activated!")
		fmt.Printf("License: %s\n", licenseKey)
		fmt.Println()
		fmt.Println("You now have unlimited access to:")
		fmt.Println("  - gg ask")
		fmt.Println("  - gg a2a")
		fmt.Println()
		fmt.Println("Thank you for supporting gg!")
	}
}

// openCustomerPortal opens Stripe billing portal
func openCustomerPortal(cfg *Config) {
	if !checkProTier(cfg) {
		fmt.Println("No active subscription")
		fmt.Println("Run: gg upgrade")
		return
	}

	fmt.Println("Opening billing portal...")

	reqBody, _ := json.Marshal(map[string]string{
		"license_key": cfg.Secrets.ProLicenseKey,
	})
	resp, err := http.Post(getBackendURL()+"/portal", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if portalURL, ok := result["portal_url"].(string); ok {
		fmt.Println(portalURL)
		exec.Command("open", portalURL).Start()
		exec.Command("xdg-open", portalURL).Start()
	} else if errMsg, ok := result["error"].(string); ok {
		fmt.Printf("Error: %s\n", errMsg)
	}
}

// A2A rate limiting for free tier
const a2aFreeLimit = 10 // soft limit per day
const a2aGraceLimit = 2 // extra grace calls

type A2AUsage struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
