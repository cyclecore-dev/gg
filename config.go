package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/BurntSushi/toml"
)

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg config init")
		return
	}

	subCmd := os.Args[2]
	if subCmd == "init" {
		initConfig()
	} else {
		fmt.Printf("Unknown config subcommand: %s\n", subCmd)
	}
}

func initConfig() {
	fmt.Printf("Welcome to gg v%s!\n", version)
	fmt.Println()
	fmt.Println("Setting up your configuration...")
	fmt.Println()

	// Create ~/.gg directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fatalError("Failed to get home directory", err)
	}

	ggDir := filepath.Join(homeDir, ".gg")
	if err := os.MkdirAll(ggDir, 0700); err != nil {
		fatalError("Failed to create .gg directory", err)
	}

	// Generate Age identity
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		fatalError("Failed to generate encryption key", err)
	}

	keyPath := filepath.Join(ggDir, ".key")
	if err := os.WriteFile(keyPath, []byte(identity.String()), 0600); err != nil {
		fatalError("Failed to save encryption key", err)
	}

	reader := bufio.NewReader(os.Stdin)

	// Provider selection
	fmt.Println("Select your AI provider:")
	fmt.Println("  1. Anthropic (Claude)")
	fmt.Println("  2. OpenAI (GPT-4)")
	fmt.Println("  3. Ollama (local)")
	fmt.Print("\nChoice [1]: ")
	providerChoice, _ := reader.ReadString('\n')
	providerChoice = strings.TrimSpace(providerChoice)

	var provider, model, endpoint string
	switch providerChoice {
	case "2":
		provider = ProviderOpenAI
		model = "gpt-4o"
		fmt.Print("\nEnter your OpenAI API key:\n> ")
	case "3":
		provider = ProviderOllama
		model = "llama3.2"
		endpoint = "http://localhost:11434"
		fmt.Print("\nOllama endpoint [http://localhost:11434]:\n> ")
		endpointInput, _ := reader.ReadString('\n')
		endpointInput = strings.TrimSpace(endpointInput)
		if endpointInput != "" {
			endpoint = endpointInput
		}
		fmt.Println("\nOllama doesn't require an API key.")
		fmt.Print("Enter API key anyway (optional, press Enter to skip):\n> ")
	default:
		provider = ProviderAnthropic
		model = "claude-sonnet-4-20250514"
		fmt.Print("\nEnter your Anthropic API key:\n> ")
	}

	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	// Auto-detect provider from key if not Ollama
	if provider != ProviderOllama && apiKey != "" {
		detected := detectProvider(apiKey)
		if detected != "" && detected != provider {
			fmt.Printf("\nNote: API key looks like %s, using that provider.\n", detected)
			provider = detected
			if provider == ProviderOpenAI {
				model = "gpt-4o"
			} else if provider == ProviderAnthropic {
				model = "claude-sonnet-4-20250514"
			}
		}
	}

	fmt.Print("\nEnter your Pro license key (optional, press Enter to skip):\n> ")
	proKey, _ := reader.ReadString('\n')
	proKey = strings.TrimSpace(proKey)

	// Create config
	cfg := Config{}
	cfg.GG.Version = version
	cfg.GG.Tier = "free"
	if proKey != "" && strings.HasPrefix(proKey, "gg_pro_") {
		cfg.GG.Tier = "pro"
	}
	cfg.API.Provider = provider
	cfg.API.Model = model
	cfg.API.Temperature = 0.7
	cfg.API.Endpoint = endpoint
	cfg.GitHub.DefaultBranch = "main"

	// Save plain config
	configPath := filepath.Join(ggDir, "config.toml")
	f, err := os.Create(configPath)
	if err != nil {
		fatalError("Failed to create config file", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		fatalError("Failed to write config", err)
	}

	// Encrypt and save secrets
	secrets := SecretsData{
		APIKey:        apiKey,
		ProLicenseKey: proKey,
	}

	secretsPath := filepath.Join(ggDir, "secrets")
	if err := encryptSecrets(secrets, identity, secretsPath); err != nil {
		fatalError("Failed to encrypt secrets", err)
	}

	fmt.Println()
	fmt.Printf("Provider: %s\n", provider)
	fmt.Printf("Model: %s\n", model)
	fmt.Println("Configuration saved to ~/.gg/config.toml")
	fmt.Println("Secrets encrypted and saved to ~/.gg/secrets")
	fmt.Println()
	fmt.Println("Run 'gg ask \"your prompt\"' to get started!")
}

// detectProvider auto-detects the provider from API key format
func detectProvider(apiKey string) string {
	if strings.HasPrefix(apiKey, "sk-ant-") {
		return ProviderAnthropic
	}
	if strings.HasPrefix(apiKey, "sk-") {
		return ProviderOpenAI
	}
	return ""
}

// getEffectiveConfig returns provider/model/key handling legacy configs
func getEffectiveConfig(cfg *Config) (provider, model, endpoint, apiKey string) {
	// New config format takes precedence
	if cfg.API.Provider != "" {
		provider = cfg.API.Provider
		model = cfg.API.Model
		endpoint = cfg.API.Endpoint
		apiKey = cfg.Secrets.APIKey
	}

	// Fall back to legacy config
	if provider == "" {
		provider = ProviderAnthropic
	}
	if model == "" {
		if cfg.API.ClaudeModel != "" {
			model = cfg.API.ClaudeModel
		} else {
			model = "claude-sonnet-4-20250514"
		}
	}
	if apiKey == "" && cfg.Secrets.ClaudeAPIKey != "" {
		apiKey = cfg.Secrets.ClaudeAPIKey
	}

	// Auto-detect from key
	if apiKey != "" && provider == ProviderAnthropic {
		detected := detectProvider(apiKey)
		if detected == ProviderOpenAI {
			provider = ProviderOpenAI
			if model == "" || strings.HasPrefix(model, "claude") {
				model = "gpt-4o"
			}
		}
	}

	return
}

func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	ggDir := filepath.Join(homeDir, ".gg")
	configPath := filepath.Join(ggDir, "config.toml")

	// Load plain config
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("config not found. Run: gg config init")
	}

	// Load and decrypt secrets
	keyPath := filepath.Join(ggDir, ".key")
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("encryption key not found")
	}

	identity, err := age.ParseX25519Identity(string(keyData))
	if err != nil {
		return nil, err
	}

	secretsPath := filepath.Join(ggDir, "secrets")
	if err := decryptSecrets(&cfg.Secrets, identity, secretsPath); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func encryptSecrets(secrets SecretsData, identity *age.X25519Identity, path string) error {
	// Create wrapper struct for TOML encoding
	data := struct {
		Keys SecretsData `toml:"keys"`
	}{Keys: secrets}

	// Encode to TOML
	var buf strings.Builder
	if err := toml.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}

	// Encrypt with Age
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	w, err := age.Encrypt(out, identity.Recipient())
	if err != nil {
		return err
	}

	if _, err := io.WriteString(w, buf.String()); err != nil {
		return err
	}

	return w.Close()
}

func decryptSecrets(secrets *SecretsData, identity *age.X25519Identity, path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	r, err := age.Decrypt(in, identity)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Decode TOML
	var temp struct {
		Keys SecretsData `toml:"keys"`
	}

	if err := toml.Unmarshal(data, &temp); err != nil {
		return err
	}

	*secrets = temp.Keys
	return nil
}

// ============================================================================
// AUTHENTICATION
// ============================================================================
