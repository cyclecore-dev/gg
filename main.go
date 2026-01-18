// gg v0.9.1 — the 2-letter agent-native git client
// January 2026 — v1.0-ready
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/BurntSushi/toml"
)

const version = "0.9.6"

// Token costs (approximate)
const (
	TokenCostNPM  = 18
	TokenCostBrew = 22
	TokenCostGit  = 12
	TokenCostBase = 20
)

// Supported providers
const (
	ProviderAnthropic = "anthropic"
	ProviderOpenAI    = "openai"
	ProviderOllama    = "ollama"
)

// SecretsData holds encrypted API keys
type SecretsData struct {
	APIKey        string `toml:"api_key"`         // Primary API key (any provider)
	ClaudeAPIKey  string `toml:"claude_api_key"`  // Legacy: kept for backwards compat
	MaazaAPIKey   string `toml:"maaza_api_key"`
	ProLicenseKey string `toml:"pro_license_key"`
}

// Config represents the gg configuration
type Config struct {
	GG struct {
		Version string `toml:"version"`
		Tier    string `toml:"tier"`
	} `toml:"gg"`
	API struct {
		Provider    string  `toml:"provider"`    // anthropic, openai, ollama
		Model       string  `toml:"model"`       // Model name
		Temperature float64 `toml:"temperature"`
		Endpoint    string  `toml:"endpoint"`    // Custom endpoint (for Ollama)
		// Legacy fields for backwards compat
		ClaudeModel       string  `toml:"claude_model"`
		ClaudeTemperature float64 `toml:"claude_temperature"`
		MaazaModel        string  `toml:"maaza_model"`
	} `toml:"api"`
	GitHub struct {
		DefaultBranch string `toml:"default_branch"`
	} `toml:"github"`
	Secrets SecretsData `toml:"keys"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "version", "--version", "-v":
		fmt.Printf("gg v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	case "init":
		initConfig()
	case "config":
		handleConfig()
	case "maaza":
		handleMaaza()
	case ".":
		handleCurrentRepo()
	case "ask":
		handleAsk()
	case "approve":
		handleApprove()
	case "pr":
		handlePR()
	case "run":
		handleRun()
	case "stats":
		handleStats()
	case "edit":
		handleEdit()
	case "prompts":
		handlePrompts()
	case "npm":
		handleNPM()
	case "brew":
		handleBrew()
	case "chain":
		handleChain()
	case "cool":
		handleCool()
	case "cache":
		handleCache()
	case "a2a":
		handleA2A()
	default:
		if strings.Contains(cmd, "/") {
			handleRepo(cmd)
		} else {
			fmt.Printf("gg: unknown command: %s\n", cmd)
			printUsage()
		}
	}
}

func printUsage() {
	fmt.Printf("gg v%s — the 2-letter agent-native git client\n", version)
	fmt.Println()
	fmt.Println("setup:")
	fmt.Println("  gg init              Configure provider & API key")
	fmt.Println("  gg maaza             Status and setup check")
	fmt.Println()
	fmt.Println("ai tools:")
	fmt.Println("  gg ask \"...\"         Generate code → PR (Pro)")
	fmt.Println("  gg a2a [mode]        Agent-to-agent CLI modes (CLI2CLI)")
	fmt.Println("  gg edit <file>       AI-assisted file editing")
	fmt.Println("  gg prompts           Manage saved prompts")
	fmt.Println()
	fmt.Println("git:")
	fmt.Println("  gg .                 Current repo → MCP")
	fmt.Println("  gg user/repo         Any GitHub repo → MCP")
	fmt.Println("  gg pr <number>       View/manage specific PR")
	fmt.Println("  gg approve           Merge the latest PR")
	fmt.Println("  gg run <cmd>         Run command in sandbox")
	fmt.Println()
	fmt.Println("package manager:")
	fmt.Println("  gg npm <pkg>         npm package → MCP (~18 tokens)")
	fmt.Println("  gg brew [-i] <f>     Homebrew formula → MCP (~22 tokens)")
	fmt.Println("  gg chain <tools>     Chain multiple MCPs")
	fmt.Println("  gg cool <toolbelt>   Curated toolbelts (webdev, media, sec, data)")
	fmt.Println("  gg cache status      Show cache size")
	fmt.Println()
	fmt.Println("other:")
	fmt.Println("  gg stats             Usage statistics")
	fmt.Println("  gg version           Show version")
	fmt.Println("  gg help              Show this help")
	fmt.Println()
	fmt.Println("install: curl -fsSL https://raw.githubusercontent.com/cyclecore-dev/gg/main/gg.sh | sh")
	fmt.Println("docs:    github.com/cyclecore-dev/gg")
}

func handleMaaza() {
	fmt.Println("Maaza Orchestrator")
	fmt.Println("Edge-optimized Nano Language Model")
	fmt.Println()
	fmt.Println("Code-execution MCP — optimized for token efficiency")
	fmt.Println("Compatible with: Claude Desktop, Cursor, any MCP client")
	fmt.Println()
	fmt.Println("Configure: ~/.gg/config.toml")
}

func handleCurrentRepo() {
	// Get git remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Not in a git repo or no remote configured")
		return
	}

	url := strings.TrimSpace(string(output))
	repo := parseGitHubURL(url)

	if repo == "" {
		fmt.Println("Could not parse GitHub repo from:", url)
		return
	}

	fmt.Printf("Current repo: %s\n", repo)
	fmt.Println()
	fmt.Println("MCP endpoint:")
	fmt.Printf("  https://api.github.com/repos/%s\n", repo)
	fmt.Println()
	fmt.Println("Code-execution MCP active — works with Claude Desktop, Cursor")
}

func handleRepo(repo string) {
	fmt.Printf("Repo: %s\n", repo)
	fmt.Println()
	fmt.Println("MCP endpoint:")
	fmt.Printf("  https://api.github.com/repos/%s\n", repo)
	fmt.Println()
	fmt.Println("Code-execution MCP active")
}

func parseGitHubURL(url string) string {
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Handle HTTPS URLs
	if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) == 2 {
			return strings.TrimPrefix(parts[1], ":")
		}
	}

	// Handle SSH URLs (git@github.com:user/repo)
	if strings.HasPrefix(url, "git@github.com:") {
		return strings.TrimPrefix(url, "git@github.com:")
	}

	// Handle custom SSH aliases (git@github-alias:user/repo)
	if strings.HasPrefix(url, "git@") && strings.Contains(url, ":") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}

	return ""
}

// ============================================================================
// CONFIG MANAGEMENT
// ============================================================================

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

func checkGitHubAuth() (bool, error) {
	cmd := exec.Command("gh", "auth", "status")
	output, err := cmd.CombinedOutput()

	if err == nil && strings.Contains(string(output), "Logged in") {
		return true, nil
	}

	return false, nil
}

func ensureGitHubAuth() error {
	authed, _ := checkGitHubAuth()
	if authed {
		return nil
	}

	fmt.Println("GitHub authentication required")
	fmt.Println()
	fmt.Println("Run: gh auth login")
	fmt.Println("Or install gh CLI: https://cli.github.com")

	return fmt.Errorf("not authenticated")
}

// ============================================================================
// ASK COMMAND
// ============================================================================

func handleAsk() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg ask \"your prompt here\"")
		return
	}

	// Parse prompt and check for --pro flag
	args := os.Args[2:]
	proMode := false
	var promptParts []string

	for _, arg := range args {
		if arg == "--pro" {
			proMode = true
		} else {
			promptParts = append(promptParts, arg)
		}
	}

	prompt := strings.Join(promptParts, " ")
	if prompt == "" {
		fmt.Println("No prompt provided")
		return
	}

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		fatalError("Config error. Run: gg config init", err)
	}

	// Check Pro tier
	if !proMode && !checkProTier(cfg) {
		fmt.Println("Analyzing request...")
		fmt.Println()
		fmt.Println("Implementation plan:")
		fmt.Println("  1. Analyze repository structure")
		fmt.Println("  2. Generate code changes")
		fmt.Println("  3. Create pull request")
		fmt.Println()
		fmt.Println("+--------------------------------------+")
		fmt.Println("|   gg Pro required for this feature  |")
		fmt.Println("+--------------------------------------+")
		fmt.Println()
		fmt.Println("Pro features:")
		fmt.Println("  - Full AI-powered code generation")
		fmt.Println("  - Unlimited gg ask commands")
		fmt.Println("  - Priority API access")
		fmt.Println()
		fmt.Println("Upgrade: https://ggdotdev.com/pro ($15/month)")
		fmt.Println("Or use: gg ask \"...\" --pro")
		return
	}

	if proMode && !checkProTier(cfg) {
		fatalError("Pro license not found in config", nil)
	}

	// Check GitHub auth
	if err := ensureGitHubAuth(); err != nil {
		return
	}

	// Get current repo
	repoName := getCurrentRepo()
	if repoName == "" {
		fatalError("Not in a git repository", nil)
	}

	fmt.Printf("Generating code for %s...\n", repoName)
	fmt.Println()

	// Call API with streaming
	response, err := callAPIStreaming(cfg, prompt, repoName)
	if err != nil {
		fatalError("API error", sanitizeError(err))
	}

	// Track ask usage
	trackCommandUsage("ask", prompt, 0)

	// Parse code blocks
	files := parseCodeBlocks(response)
	if len(files) == 0 {
		fmt.Println("No code blocks found in response")
		fmt.Println("Response:")
		fmt.Println(response)
		return
	}

	// Create branch
	branchName := fmt.Sprintf("gg-ask-%d", time.Now().Unix())
	exec.Command("git", "checkout", "-b", branchName).Run()

	// Apply changes
	for path, content := range files {
		dir := filepath.Dir(path)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Printf("Failed to write %s: %v\n", path, err)
			continue
		}
		fmt.Printf("+ %s\n", path)
	}

	// Commit and push (only stage generated files - DOCK-030 OPSEC)
	commitMsg := fmt.Sprintf("gg ask: %s", truncate(prompt, 60))
	for path := range files {
		exec.Command("git", "add", path).Run()
	}
	exec.Command("git", "commit", "-m", commitMsg).Run()
	exec.Command("git", "push", "-u", "origin", branchName).Run()

	// Create PR
	prCmd := exec.Command("gh", "pr", "create", "--title", commitMsg, "--body", fmt.Sprintf("Generated by gg ask:\n\n%s", prompt))
	prOutput, err := prCmd.Output()
	if err != nil {
		fmt.Println("Failed to create PR. Create manually:")
		fmt.Printf("   Branch: %s\n", branchName)
		return
	}

	prURL := strings.TrimSpace(string(prOutput))
	fmt.Println()
	fmt.Printf("PR created: %s\n", prURL)
	fmt.Println()
	fmt.Println("Next: gg approve")
}

func handleApprove() {
	if err := ensureGitHubAuth(); err != nil {
		return
	}

	// Get latest PR
	cmd := exec.Command("gh", "pr", "list", "--limit", "1", "--json", "number,title,headRefName")
	output, err := cmd.Output()
	if err != nil {
		fatalError("Failed to list PRs", err)
	}

	var prs []struct {
		Number      int    `json:"number"`
		Title       string `json:"title"`
		HeadRefName string `json:"headRefName"`
	}

	if err := json.Unmarshal(output, &prs); err != nil {
		fatalError("Failed to parse PR list", err)
	}

	if len(prs) == 0 {
		fmt.Println("No open PRs found")
		return
	}

	pr := prs[0]
	fmt.Printf("PR #%d: %s\n", pr.Number, pr.Title)
	fmt.Printf("Branch: %s\n\n", pr.HeadRefName)

	// Confirm
	fmt.Print("Merge this PR? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "" && response != "y" {
		fmt.Println("Cancelled")
		return
	}

	// Merge
	mergeCmd := exec.Command("gh", "pr", "merge", fmt.Sprintf("%d", pr.Number), "--squash", "--delete-branch")
	mergeCmd.Stdout = os.Stdout
	mergeCmd.Stderr = os.Stderr

	if err := mergeCmd.Run(); err != nil {
		fatalError("Failed to merge PR", err)
	}

	fmt.Println()
	fmt.Println("PR merged successfully!")
}

// ============================================================================
// EDIT COMMAND
// ============================================================================

func handleEdit() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg edit <file> [instruction]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg edit main.go                    # Interactive edit")
		fmt.Println("  gg edit main.go \"add error handling\"")
		fmt.Println("  gg edit src/*.ts \"add types\"       # Multiple files")
		return
	}

	filePath := os.Args[2]
	var instruction string
	if len(os.Args) > 3 {
		instruction = strings.Join(os.Args[3:], " ")
	}

	// Check if file exists
	content, err := os.ReadFile(filePath)
	if err != nil {
		fatalError("Cannot read file", err)
	}

	// Get instruction if not provided
	if instruction == "" {
		fmt.Printf("Editing: %s (%d bytes)\n", filePath, len(content))
		fmt.Print("\nWhat changes do you want? > ")
		reader := bufio.NewReader(os.Stdin)
		instruction, _ = reader.ReadString('\n')
		instruction = strings.TrimSpace(instruction)
		if instruction == "" {
			fmt.Println("No instruction provided, exiting")
			return
		}
	}

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		fatalError("Config error. Run: gg config init", err)
	}

	provider, model, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider != ProviderOllama && apiKey == "" {
		fatalError("API key not configured. Run: gg config init", nil)
	}

	// Create edit prompt
	editPrompt := fmt.Sprintf("Edit this file according to the instruction.\n\n"+
		"FILE: %s\n"+
		"```\n%s\n```\n\n"+
		"INSTRUCTION: %s\n\n"+
		"Return ONLY the complete edited file content, no explanations. "+
		"Wrap in ```language:filename code block.", filePath, string(content), instruction)

	systemPrompt := "You are a code editor. Return only the edited file content in a code block."

	fmt.Printf("Editing %s with %s/%s...\n\n", filePath, provider, model)

	var response string
	switch provider {
	case ProviderOpenAI:
		response, err = callOpenAIStreaming(apiKey, model, systemPrompt, editPrompt)
	case ProviderOllama:
		response, err = callOllamaStreaming(endpoint, model, systemPrompt, editPrompt)
	default:
		response, err = callAnthropicStreaming(apiKey, model, systemPrompt, editPrompt, cfg.API.Temperature)
	}

	if err != nil {
		fatalError("API error", sanitizeError(err))
	}

	// Parse the response for code blocks
	files := parseCodeBlocks(response)
	if len(files) == 0 {
		// Try to extract content between ``` markers
		re := regexp.MustCompile("```[a-z]*\n([\\s\\S]*?)```")
		matches := re.FindStringSubmatch(response)
		if len(matches) >= 2 {
			files[filePath] = matches[1]
		}
	}

	if len(files) == 0 {
		fmt.Println("\nNo code block found in response. Raw response:")
		fmt.Println(response)
		return
	}

	// Get the edited content
	var newContent string
	for _, c := range files {
		newContent = c
		break
	}

	// Show diff summary
	oldLines := strings.Count(string(content), "\n")
	newLines := strings.Count(newContent, "\n")
	fmt.Printf("\nChanges: %d lines -> %d lines\n", oldLines, newLines)

	// Confirm
	fmt.Print("\nApply changes? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "" && confirm != "y" {
		fmt.Println("Cancelled")
		return
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		fatalError("Failed to write file", err)
	}

	fmt.Printf("Updated: %s\n", filePath)
}

// ============================================================================
// PROMPTS COMMAND
// ============================================================================

func handlePrompts() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg prompts <command>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  list               # List saved prompts")
		fmt.Println("  add <name> \"...\"   # Save a prompt")
		fmt.Println("  run <name>         # Run a saved prompt with gg ask")
		fmt.Println("  delete <name>      # Delete a prompt")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg prompts add refactor \"refactor this code for clarity\"")
		fmt.Println("  gg prompts run refactor")
		return
	}

	subCmd := os.Args[2]

	switch subCmd {
	case "list", "ls":
		listPrompts()
	case "add", "save":
		if len(os.Args) < 5 {
			fmt.Println("Usage: gg prompts add <name> \"prompt text\"")
			return
		}
		name := os.Args[3]
		promptText := strings.Join(os.Args[4:], " ")
		savePrompt(name, promptText)
	case "run", "use":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gg prompts run <name>")
			return
		}
		usePrompt(os.Args[3])
	case "delete", "rm":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gg prompts delete <name>")
			return
		}
		deletePrompt(os.Args[3])
	default:
		// Try to run it as a prompt name
		usePrompt(subCmd)
	}
}

func getPromptsDir() string {
	return filepath.Join(getGGDir(), "prompts")
}

func listPrompts() {
	promptsDir := getPromptsDir()
	files, err := os.ReadDir(promptsDir)
	if err != nil || len(files) == 0 {
		fmt.Println("No saved prompts")
		fmt.Println("Save one: gg prompts save <name> \"prompt text\"")
		return
	}

	fmt.Println("Saved prompts:")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".txt") {
			name := strings.TrimSuffix(f.Name(), ".txt")
			content, _ := os.ReadFile(filepath.Join(promptsDir, f.Name()))
			preview := string(content)
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			fmt.Printf("  - %s: %s\n", name, preview)
		}
	}
}

func savePrompt(name, text string) {
	promptsDir := getPromptsDir()
	os.MkdirAll(promptsDir, 0755)

	path := filepath.Join(promptsDir, name+".txt")
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		fatalError("Failed to save prompt", err)
	}
	fmt.Printf("Saved prompt '%s'\n", name)
}

func usePrompt(name string) {
	promptsDir := getPromptsDir()
	path := filepath.Join(promptsDir, name+".txt")

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Prompt not found: %s\n", name)
		fmt.Println("Run 'gg prompts list' to see saved prompts")
		return
	}

	promptText := string(content)
	fmt.Printf("Using prompt '%s': %s\n\n", name, promptText)

	// Inject prompt into handleAsk by setting os.Args
	os.Args = []string{"gg", "ask", promptText}
	handleAsk()
}

func deletePrompt(name string) {
	promptsDir := getPromptsDir()
	path := filepath.Join(promptsDir, name+".txt")

	if err := os.Remove(path); err != nil {
		fmt.Printf("Prompt not found: %s\n", name)
		return
	}
	fmt.Printf("Deleted prompt '%s'\n", name)
}

// ============================================================================
// UTILITIES
// ============================================================================

func getCurrentRepo() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	url := strings.TrimSpace(string(output))
	return parseGitHubURL(url)
}

func checkProTier(cfg *Config) bool {
	return cfg.Secrets.ProLicenseKey != "" &&
		strings.HasPrefix(cfg.Secrets.ProLicenseKey, "gg_pro_")
}

func parseCodeBlocks(response string) map[string]string {
	files := make(map[string]string)

	// Match ```language:path/to/file
	re := regexp.MustCompile("```[a-z]*:([^\n]+)\n([\\s\\S]*?)```")
	matches := re.FindAllStringSubmatch(response, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			path := strings.TrimSpace(match[1])
			content := match[2]
			files[path] = content
		}
	}

	return files
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func sanitizeError(err error) error {
	msg := err.Error()
	// Sanitize various API key formats
	msg = regexp.MustCompile(`sk-ant-[a-zA-Z0-9-]+`).ReplaceAllString(msg, "sk-***")
	msg = regexp.MustCompile(`sk-[a-zA-Z0-9]+`).ReplaceAllString(msg, "sk-***")
	msg = regexp.MustCompile(`mcpb_[a-zA-Z0-9]+`).ReplaceAllString(msg, "mcpb_***")
	msg = regexp.MustCompile(`gg_pro_[a-zA-Z0-9]+`).ReplaceAllString(msg, "gg_pro_***")
	return fmt.Errorf("%s", msg)
}

func fatalError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %v\n", err)
	}
	os.Exit(1)
}

func getHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return homeDir
}

func getGGDir() string {
	return filepath.Join(getHomeDir(), ".gg")
}

// ============================================================================
// PR, RUN, STATS
// ============================================================================

func handlePR() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg pr <number>")
		return
	}

	prNumber := os.Args[2]
	if err := ensureGitHubAuth(); err != nil {
		return
	}

	repoName := getCurrentRepo()
	if repoName == "" {
		fatalError("Not in a git repository", nil)
	}

	// Fetch PR details
	cmd := exec.Command("gh", "pr", "view", prNumber, "--json", "number,title,author,state,body,additions,deletions,changedFiles,headRefName,baseRefName,url")
	output, err := cmd.Output()
	if err != nil {
		fatalError("Failed to fetch PR", err)
	}

	var pr struct {
		Number       int    `json:"number"`
		Title        string `json:"title"`
		Author       struct{ Login string } `json:"author"`
		State        string `json:"state"`
		Body         string `json:"body"`
		Additions    int    `json:"additions"`
		Deletions    int    `json:"deletions"`
		ChangedFiles int    `json:"changedFiles"`
		HeadRefName  string `json:"headRefName"`
		BaseRefName  string `json:"baseRefName"`
		URL          string `json:"url"`
	}

	if err := json.Unmarshal(output, &pr); err != nil {
		fatalError("Failed to parse PR data", err)
	}

	fmt.Printf("PR #%d: %s\n", pr.Number, pr.Title)
	fmt.Printf("Author: %s | State: %s\n", pr.Author.Login, pr.State)
	fmt.Printf("Branch: %s -> %s\n", pr.HeadRefName, pr.BaseRefName)
	fmt.Printf("Changes: +%d -%d (%d files)\n", pr.Additions, pr.Deletions, pr.ChangedFiles)
	fmt.Println()

	if pr.Body != "" {
		fmt.Println("Description:")
		fmt.Println(truncate(pr.Body, 500))
		fmt.Println()
	}

	fmt.Printf("URL: %s\n", pr.URL)
	fmt.Println()

	// Show action menu
	if pr.State == "OPEN" {
		fmt.Println("Actions:")
		fmt.Println("  [a]pprove - Merge this PR")
		fmt.Println("  [d]iff   - Show full diff")
		fmt.Println("  [c]lose  - Close without merging")
		fmt.Println("  [q]uit   - Exit")
		fmt.Print("\nChoice: ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		switch choice {
		case "a":
			mergeCmd := exec.Command("gh", "pr", "merge", prNumber, "--squash", "--delete-branch")
			mergeCmd.Stdout = os.Stdout
			mergeCmd.Stderr = os.Stderr
			if err := mergeCmd.Run(); err != nil {
				fatalError("Failed to merge PR", err)
			}
			fmt.Println("PR merged!")
		case "d":
			diffCmd := exec.Command("gh", "pr", "diff", prNumber)
			diffCmd.Stdout = os.Stdout
			diffCmd.Stderr = os.Stderr
			diffCmd.Run()
		case "c":
			closeCmd := exec.Command("gh", "pr", "close", prNumber)
			closeCmd.Stdout = os.Stdout
			closeCmd.Stderr = os.Stderr
			if err := closeCmd.Run(); err != nil {
				fatalError("Failed to close PR", err)
			}
			fmt.Println("PR closed")
		default:
			fmt.Println("Exiting")
		}
	}
}

func handleRun() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg run <command>")
		fmt.Println("Example: gg run npm test")
		return
	}

	cmdArgs := os.Args[2:]
	cmdStr := strings.Join(cmdArgs, " ")

	fmt.Printf("Running: %s\n", cmdStr)
	fmt.Println()

	// Execute command with timeout
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	fmt.Println()
	if err != nil {
		exitCode := 1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		fmt.Printf("Exit code: %d (%.2fs)\n", exitCode, elapsed.Seconds())
	} else {
		fmt.Printf("Success (%.2fs)\n", elapsed.Seconds())
	}

	// Track usage
	trackCommandUsage("run", cmdStr, elapsed)
}

func handleStats() {
	homeDir := getHomeDir()
	statsPath := filepath.Join(homeDir, ".gg", "stats.json")
	data, err := os.ReadFile(statsPath)
	if err != nil {
		fmt.Println("Usage Statistics")
		fmt.Println()
		fmt.Println("No usage data yet. Run some commands first!")
		return
	}

	var stats UsageStats
	if err := json.Unmarshal(data, &stats); err != nil {
		fatalError("Failed to parse stats", err)
	}

	fmt.Println("Usage Statistics")
	fmt.Println()
	fmt.Printf("Month: %s\n", stats.Month)
	fmt.Printf("Total asks: %d\n", stats.AskCount)
	fmt.Printf("Total runs: %d\n", stats.RunCount)
	fmt.Printf("Total tokens: %d (input: %d, output: %d)\n", stats.TotalTokens, stats.InputTokens, stats.OutputTokens)
	fmt.Printf("Estimated cost: $%.4f\n", stats.EstimatedCost)
}

// UsageStats tracks monthly usage
type UsageStats struct {
	Month         string  `json:"month"`
	AskCount      int     `json:"ask_count"`
	RunCount      int     `json:"run_count"`
	TotalTokens   int64   `json:"total_tokens"`
	InputTokens   int64   `json:"input_tokens"`
	OutputTokens  int64   `json:"output_tokens"`
	EstimatedCost float64 `json:"estimated_cost"`
}

func trackCommandUsage(cmdType, detail string, elapsed time.Duration) {
	homeDir := getHomeDir()
	statsPath := filepath.Join(homeDir, ".gg", "stats.json")

	var stats UsageStats
	data, err := os.ReadFile(statsPath)
	if err == nil {
		json.Unmarshal(data, &stats)
	}

	currentMonth := time.Now().Format("2006-01")
	if stats.Month != currentMonth {
		stats = UsageStats{Month: currentMonth}
	}

	switch cmdType {
	case "ask":
		stats.AskCount++
	case "run":
		stats.RunCount++
	}

	outData, _ := json.MarshalIndent(stats, "", "  ")
	os.WriteFile(statsPath, outData, 0644)
}

func trackTokenUsage(inputTokens, outputTokens int64) {
	homeDir := getHomeDir()
	statsPath := filepath.Join(homeDir, ".gg", "stats.json")

	var stats UsageStats
	data, err := os.ReadFile(statsPath)
	if err == nil {
		json.Unmarshal(data, &stats)
	}

	currentMonth := time.Now().Format("2006-01")
	if stats.Month != currentMonth {
		stats = UsageStats{Month: currentMonth}
	}

	stats.InputTokens += inputTokens
	stats.OutputTokens += outputTokens
	stats.TotalTokens = stats.InputTokens + stats.OutputTokens

	// Approximate cost (varies by provider)
	stats.EstimatedCost = float64(stats.InputTokens)/1000000*3 + float64(stats.OutputTokens)/1000000*15

	outData, _ := json.MarshalIndent(stats, "", "  ")
	os.WriteFile(statsPath, outData, 0644)
}

// ============================================================================
// MULTI-PROVIDER API
// ============================================================================

func callAPIStreaming(cfg *Config, prompt, repo string) (string, error) {
	provider, model, endpoint, apiKey := getEffectiveConfig(cfg)

	if provider != ProviderOllama && apiKey == "" {
		return "", fmt.Errorf("API key not configured. Run: gg config init")
	}

	systemPrompt := fmt.Sprintf("You are a code generation assistant for the repository: %s\n\n"+
		"Generate clean, production-ready code based on the user's request.\n"+
		"Format code blocks as:\n"+
		"```language:path/to/file\n"+
		"code here\n"+
		"```\n\n"+
		"Be concise and only generate the requested code.", repo)

	switch provider {
	case ProviderOpenAI:
		return callOpenAIStreaming(apiKey, model, systemPrompt, prompt)
	case ProviderOllama:
		return callOllamaStreaming(endpoint, model, systemPrompt, prompt)
	default:
		return callAnthropicStreaming(apiKey, model, systemPrompt, prompt, cfg.API.Temperature)
	}
}

func callAnthropicStreaming(apiKey, model, systemPrompt, prompt string, temperature float64) (string, error) {
	if temperature == 0 {
		temperature = 0.7
	}

	requestBody := map[string]interface{}{
		"model":       model,
		"max_tokens":  4096,
		"stream":      true,
		"system":      systemPrompt,
		"temperature": temperature,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse SSE stream
	var fullResponse strings.Builder
	var inputTokens, outputTokens int64
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullResponse.String(), err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
			Usage struct {
				InputTokens  int64 `json:"input_tokens"`
				OutputTokens int64 `json:"output_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			fmt.Print(event.Delta.Text)
			fullResponse.WriteString(event.Delta.Text)
		}

		if event.Type == "message_delta" && event.Usage.OutputTokens > 0 {
			outputTokens = event.Usage.OutputTokens
		}
		if event.Type == "message_start" && event.Usage.InputTokens > 0 {
			inputTokens = event.Usage.InputTokens
		}
	}

	fmt.Println()

	if inputTokens > 0 || outputTokens > 0 {
		trackTokenUsage(inputTokens, outputTokens)
	}

	return fullResponse.String(), nil
}

func callOpenAIStreaming(apiKey, model, systemPrompt, prompt string) (string, error) {
	requestBody := map[string]interface{}{
		"model":  model,
		"stream": true,
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse SSE stream
	var fullResponse strings.Builder
	var totalTokens int64
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullResponse.String(), err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Usage struct {
				TotalTokens int64 `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
			fmt.Print(event.Choices[0].Delta.Content)
			fullResponse.WriteString(event.Choices[0].Delta.Content)
		}

		if event.Usage.TotalTokens > 0 {
			totalTokens = event.Usage.TotalTokens
		}
	}

	fmt.Println()

	if totalTokens > 0 {
		// Estimate input/output split (rough 30/70)
		trackTokenUsage(totalTokens*3/10, totalTokens*7/10)
	}

	return fullResponse.String(), nil
}

func callOllamaStreaming(endpoint, model, systemPrompt, prompt string) (string, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	requestBody := map[string]interface{}{
		"model":  model,
		"stream": true,
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 600 * time.Second} // Ollama can be slow
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama connection failed: %v (is Ollama running?)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Ollama returns newline-delimited JSON
	var fullResponse strings.Builder
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fullResponse.String(), err
		}

		var event struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}

		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if event.Message.Content != "" {
			fmt.Print(event.Message.Content)
			fullResponse.WriteString(event.Message.Content)
		}

		if event.Done {
			break
		}
	}

	fmt.Println()

	return fullResponse.String(), nil
}

// ============================================================================
// PACKAGE MANAGER LAYER
// ============================================================================

// handleNPM fetches npm package info and displays MCP endpoint
func handleNPM() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg npm <package> [--fn <function>]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg npm prettier")
		fmt.Println("  gg npm lodash --fn debounce")
		return
	}

	pkg := os.Args[2]

	// Check local cache first
	cacheDir := filepath.Join(getGGDir(), "cache", "npm")
	cachePath := filepath.Join(cacheDir, pkg+".json")

	var pkgInfo map[string]interface{}

	if data, err := os.ReadFile(cachePath); err == nil {
		// Cache hit
		json.Unmarshal(data, &pkgInfo)
		fmt.Printf("%s (cached)\n", pkg)
	} else {
		// Fetch from npm registry
		fmt.Printf("Fetching %s from npm...\n", pkg)
		url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", pkg)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Failed to fetch package: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			fmt.Printf("Package not found: %s\n", pkg)
			return
		}

		if resp.StatusCode != 200 {
			fmt.Printf("npm registry error: %d\n", resp.StatusCode)
			return
		}

		if err := json.NewDecoder(resp.Body).Decode(&pkgInfo); err != nil {
			fmt.Printf("Failed to parse response: %v\n", err)
			return
		}

		// Cache it
		os.MkdirAll(cacheDir, 0755)
		data, _ := json.Marshal(pkgInfo)
		os.WriteFile(cachePath, data, 0644)
	}

	// Display MCP format
	name, _ := pkgInfo["name"].(string)
	pkgVersion, _ := pkgInfo["version"].(string)
	desc, _ := pkgInfo["description"].(string)

	fmt.Printf("\n%s@%s\n", name, pkgVersion)
	if desc != "" {
		fmt.Printf("   %s\n", desc)
	}

	// Check for --fn flag
	for i, arg := range os.Args {
		if arg == "--fn" && i+1 < len(os.Args) {
			fnName := os.Args[i+1]
			fmt.Printf("\nFunction: %s\n", fnName)
		}
	}

	fmt.Printf("\nMCP Endpoint: npm:%s\n", name)
	fmt.Printf("Token cost: ~%d\n", TokenCostNPM)
}

// handleBrew fetches Homebrew formula info and displays MCP endpoint
func handleBrew() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg brew [-i] <formula>")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -i  Auto-install formula if not installed")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg brew ffmpeg")
		fmt.Println("  gg brew -i jq")
		return
	}

	// Parse -i flag
	autoInstall := false
	formula := ""
	for _, arg := range os.Args[2:] {
		if arg == "-i" {
			autoInstall = true
		} else {
			formula = arg
		}
	}

	if formula == "" {
		fmt.Println("No formula specified")
		return
	}

	// Check local cache first
	cacheDir := filepath.Join(getGGDir(), "cache", "brew")
	cachePath := filepath.Join(cacheDir, formula+".json")

	var info map[string]interface{}
	installed := false

	// Try local brew first
	cmd := exec.Command("brew", "info", formula, "--json=v2")
	output, err := cmd.Output()
	if err == nil {
		var brewInfo map[string]interface{}
		if json.Unmarshal(output, &brewInfo) == nil {
			if formulae, ok := brewInfo["formulae"].([]interface{}); ok && len(formulae) > 0 {
				info = formulae[0].(map[string]interface{})
				installed = true
			}
		}
	}

	// Fall back to API
	if info == nil {
		// Check cache
		if data, err := os.ReadFile(cachePath); err == nil {
			json.Unmarshal(data, &info)
			fmt.Printf("%s (cached)\n", formula)
		} else {
			fmt.Printf("Fetching %s from Homebrew...\n", formula)
			url := fmt.Sprintf("https://formulae.brew.sh/api/formula/%s.json", formula)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Failed to fetch formula: %v\n", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 {
				fmt.Printf("Formula not found: %s\n", formula)
				return
			}

			if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
				fmt.Printf("Failed to parse response: %v\n", err)
				return
			}

			// Cache it
			os.MkdirAll(cacheDir, 0755)
			data, _ := json.Marshal(info)
			os.WriteFile(cachePath, data, 0644)
		}
	}

	// Display info
	name, _ := info["name"].(string)
	desc, _ := info["desc"].(string)

	var formulaVersion string
	if versions, ok := info["versions"].(map[string]interface{}); ok {
		formulaVersion, _ = versions["stable"].(string)
	}

	// Auto-install if -i flag and not installed
	if !installed && autoInstall {
		fmt.Printf("Installing %s...\n", formula)
		installCmd := exec.Command("brew", "install", formula)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			fmt.Printf("Install failed: %v\n", err)
			return
		}
		fmt.Printf("%s installed\n", formula)
		installed = true
	}

	statusIcon := "installed"
	if !installed {
		statusIcon = "not installed"
	}

	fmt.Printf("\n%s@%s (%s)\n", name, formulaVersion, statusIcon)
	if desc != "" {
		fmt.Printf("   %s\n", desc)
	}

	if !installed && !autoInstall {
		fmt.Printf("\n   Install: brew install %s\n", formula)
		fmt.Println("   Or use: gg brew -i", formula)
	}

	fmt.Printf("\nMCP Endpoint: brew:%s\n", formula)
	fmt.Printf("Token cost: ~%d\n", TokenCostBrew)
}

// handleChain chains multiple MCP tools together
func handleChain() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg chain <tool:pkg> [tool:pkg...]")
		fmt.Println("       gg chain --save <name> <tool:pkg> [tool:pkg...]")
		fmt.Println("       gg chain run <name>")
		fmt.Println("       gg chain <saved-name>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg chain npm:prettier npm:eslint brew:jq")
		fmt.Println("  gg chain --save webformat npm:prettier npm:eslint")
		fmt.Println("  gg chain run webformat")
		return
	}

	args := os.Args[2:]

	// Check for run subcommand
	if args[0] == "run" {
		if len(args) < 2 {
			fmt.Println("Usage: gg chain run <name>")
			return
		}
		runChain(args[1])
		return
	}

	// Check for --save flag
	if args[0] == "--save" {
		if len(args) < 3 {
			fmt.Println("Usage: gg chain --save <name> <tool:pkg>...")
			return
		}
		chainName := args[1]
		tools := args[2:]
		saveChain(chainName, tools)
		fmt.Printf("Saved chain '%s' with %d tools\n", chainName, len(tools))
		return
	}

	// Check for --list flag
	if args[0] == "--list" {
		listSavedChains()
		return
	}

	// Check if first arg is a saved chain name
	if !strings.Contains(args[0], ":") {
		tools := loadChain(args[0])
		if tools != nil {
			fmt.Printf("Running saved chain '%s'\n\n", args[0])
			args = tools
		} else {
			fmt.Printf("Unknown chain: %s\n", args[0])
			fmt.Println("Run 'gg chain --list' to see saved chains")
			return
		}
	}

	fmt.Printf("Chained %d MCPs:\n", len(args))
	totalCost := 0

	for i, tool := range args {
		parts := strings.SplitN(tool, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("   %d. Invalid format: %s (expected type:name)\n", i+1, tool)
			continue
		}

		toolType := parts[0]
		toolName := parts[1]

		cost := getTokenCost(toolType)
		fmt.Printf("   %d. %s:%s (~%d tokens)\n", i+1, toolType, toolName, cost)
		totalCost += cost
	}

	fmt.Printf("\nCombined token cost: ~%d\n", totalCost)
}

func getTokenCost(toolType string) int {
	switch toolType {
	case "npm":
		return TokenCostNPM
	case "brew":
		return TokenCostBrew
	case "git":
		return TokenCostGit
	default:
		return TokenCostBase
	}
}

func saveChain(name string, tools []string) {
	chainDir := filepath.Join(getGGDir(), "chains")
	os.MkdirAll(chainDir, 0755)

	data, _ := json.Marshal(tools)
	os.WriteFile(filepath.Join(chainDir, name+".json"), data, 0644)
}

func loadChain(name string) []string {
	chainPath := filepath.Join(getGGDir(), "chains", name+".json")
	data, err := os.ReadFile(chainPath)
	if err != nil {
		return nil
	}

	var tools []string
	json.Unmarshal(data, &tools)
	return tools
}

func listSavedChains() {
	chainDir := filepath.Join(getGGDir(), "chains")
	files, err := os.ReadDir(chainDir)
	if err != nil || len(files) == 0 {
		fmt.Println("No saved chains")
		fmt.Println("Create one: gg chain --save <name> <tool:pkg>...")
		return
	}

	fmt.Println("Saved chains:")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			name := strings.TrimSuffix(f.Name(), ".json")
			tools := loadChain(name)
			fmt.Printf("   - %s (%d tools)\n", name, len(tools))
		}
	}
}

// Curated toolbelts
var toolbelts = map[string][]string{
	"webdev": {
		"npm:eslint",
		"npm:prettier",
		"npm:typescript",
		"npm:jest",
		"npm:playwright",
	},
	"media": {
		"brew:ffmpeg",
		"brew:imagemagick",
		"brew:exiftool",
	},
	"sec": {
		"brew:semgrep",
		"npm:snyk",
		"brew:trivy",
	},
	"data": {
		"brew:duckdb",
		"brew:jq",
		"npm:csvtojson",
	},
	"devops": {
		"brew:terraform",
		"brew:kubectl",
		"brew:docker",
	},
}

// handleCool displays curated toolbelts
func handleCool() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg cool <toolbelt>")
		fmt.Println("       gg cool --list")
		fmt.Println()
		fmt.Println("Available toolbelts: webdev, media, sec, data, devops")
		return
	}

	arg := os.Args[2]

	if arg == "--list" {
		fmt.Println("Available toolbelts:")
		fmt.Println()
		for name, tools := range toolbelts {
			fmt.Printf("   %s (%d tools)\n", name, len(tools))
			for _, tool := range tools {
				parts := strings.SplitN(tool, ":", 2)
				fmt.Printf("      - %s (%s)\n", parts[1], parts[0])
			}
			fmt.Println()
		}
		return
	}

	tools, ok := toolbelts[arg]
	if !ok {
		fmt.Printf("Unknown toolbelt: %s\n", arg)
		fmt.Println("Run 'gg cool --list' to see available toolbelts")
		return
	}

	fmt.Printf("Toolbelt: %s\n\n", arg)
	totalCost := 0

	for _, tool := range tools {
		parts := strings.SplitN(tool, ":", 2)
		toolType := parts[0]
		toolName := parts[1]

		cost := getTokenCost(toolType)
		fmt.Printf("   - %s (%s)\n", toolName, toolType)
		totalCost += cost
	}

	fmt.Printf("\nCombined token cost: ~%d\n", totalCost)
	fmt.Printf("\nChain all: gg chain %s\n", strings.Join(tools, " "))
}

// runChain executes all tools in a saved chain
func runChain(name string) {
	tools := loadChain(name)
	if tools == nil {
		fmt.Printf("Chain not found: %s\n", name)
		fmt.Println("Run 'gg chain --list' to see saved chains")
		return
	}

	fmt.Printf("Executing chain '%s'...\n\n", name)

	success := 0
	for i, tool := range tools {
		parts := strings.SplitN(tool, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("[%d/%d] Invalid: %s\n", i+1, len(tools), tool)
			continue
		}

		toolType := parts[0]
		toolName := parts[1]

		fmt.Printf("[%d/%d] %s:%s\n", i+1, len(tools), toolType, toolName)

		switch toolType {
		case "npm":
			runNPMCheck(toolName)
		case "brew":
			runBrewCheck(toolName)
		default:
			fmt.Printf("   Unknown type: %s\n", toolType)
		}

		success++
		fmt.Println()
	}

	fmt.Printf("Chain complete: %d/%d tools ready\n", success, len(tools))
}

func runNPMCheck(pkg string) {
	cacheDir := filepath.Join(getGGDir(), "cache", "npm")
	cachePath := filepath.Join(cacheDir, pkg+".json")

	if _, err := os.ReadFile(cachePath); err == nil {
		fmt.Printf("   %s (cached)\n", pkg)
		return
	}

	url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", pkg)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("   %s (error)\n", pkg)
		return
	}
	defer resp.Body.Close()

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)

	os.MkdirAll(cacheDir, 0755)
	data, _ := json.Marshal(info)
	os.WriteFile(cachePath, data, 0644)

	pkgVersion, _ := info["version"].(string)
	fmt.Printf("   %s@%s\n", pkg, pkgVersion)
}

func runBrewCheck(formula string) {
	cmd := exec.Command("brew", "info", formula, "--json=v2")
	if err := cmd.Run(); err == nil {
		fmt.Printf("   %s (installed)\n", formula)
	} else {
		fmt.Printf("   %s (not installed)\n", formula)
	}
}

// handleCache manages the gg cache
func handleCache() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg cache <status|clean>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  status  Show cache size and contents")
		fmt.Println("  clean   Remove old cache entries")
		return
	}

	cacheDir := filepath.Join(getGGDir(), "cache")

	switch os.Args[2] {
	case "status":
		showCacheStatus(cacheDir)
	case "clean":
		cleanCache(cacheDir)
	default:
		fmt.Printf("Unknown cache command: %s\n", os.Args[2])
	}
}

func showCacheStatus(cacheDir string) {
	total := getCacheSize(cacheDir)
	npmSize := getCacheSize(filepath.Join(cacheDir, "npm"))
	brewSize := getCacheSize(filepath.Join(cacheDir, "brew"))

	npmCount := countFiles(filepath.Join(cacheDir, "npm"))
	brewCount := countFiles(filepath.Join(cacheDir, "brew"))

	fmt.Println("Cache Status")
	fmt.Println()
	fmt.Printf("   Total: %s\n", formatSize(total))
	fmt.Printf("   npm:   %s (%d packages)\n", formatSize(npmSize), npmCount)
	fmt.Printf("   brew:  %s (%d formulas)\n", formatSize(brewSize), brewCount)
	fmt.Println()
	fmt.Printf("   Location: %s\n", cacheDir)
}

func cleanCache(cacheDir string) {
	type entry struct {
		path  string
		mtime time.Time
		size  int64
	}
	var entries []entry

	filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			entries = append(entries, entry{path, info.ModTime(), info.Size()})
		}
		return nil
	})

	if len(entries) == 0 {
		fmt.Println("Cache is empty")
		return
	}

	// Remove entries older than 7 days
	cutoff := time.Now().AddDate(0, 0, -7)
	var freed int64
	var removed int

	for _, e := range entries {
		if e.mtime.Before(cutoff) {
			os.Remove(e.path)
			freed += e.size
			removed++
		}
	}

	if removed == 0 {
		fmt.Println("No old cache entries to clean")
	} else {
		fmt.Printf("Cleaned cache: %s freed (%d entries removed)\n", formatSize(freed), removed)
	}
}

func getCacheSize(dir string) int64 {
	var size int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func countFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CLI2CLI: Agent-to-Agent modes (DOCK-030 compliant)
// Design: <100 tokens output, pipeable, parseable
func handleA2A() {
	args := os.Args[2:]
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printA2AHelp()
		return
	}

	mode := args[0]
	modeArgs := args[1:]

	switch mode {
	case ".":
		handleA2ADot()
	case "ask":
		handleA2AAsk(modeArgs)
	case "plan":
		handleA2APlan(modeArgs)
	case "code":
		handleA2ACode(modeArgs)
	default:
		fmt.Printf("a2a: unknown mode: %s\n", mode)
		printA2AHelp()
	}
}

func printA2AHelp() {
	fmt.Println("gg a2a — Agent-to-Agent CLI modes (CLI2CLI)")
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  .              Output MCP endpoint for current repo (<20 tokens)")
	fmt.Println("  ask \"prompt\"   Structured response, no git ops (<100 tokens)")
	fmt.Println("  plan \"task\"    Numbered plan steps, pipeable")
	fmt.Println("  code \"task\"    Code blocks only, no prose")
	fmt.Println()
	fmt.Println("Pipe examples:")
	fmt.Println("  gg a2a . | gg a2a ask \"summarize this repo\"")
	fmt.Println("  gg a2a plan \"auth system\" | gg a2a code")
	fmt.Println()
	fmt.Println("Design: <100 tokens, pipeable stdout, agents see minimal text")
}

func handleA2ADot() {
	// Minimal MCP endpoint output - agents "see" just this
	repoName := getCurrentRepo()
	if repoName == "" {
		fmt.Println("error: not in a git repository")
		return
	}
	// Output format: gg://<repo> (minimal, pipeable)
	fmt.Printf("gg://%s\n", repoName)
}

func handleA2AAsk(args []string) {
	if len(args) == 0 {
		// Check for piped input
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Has piped input - read it
			input, _ := io.ReadAll(os.Stdin)
			args = []string{string(input)}
		} else {
			fmt.Println("error: no prompt provided")
			fmt.Println("usage: gg a2a ask \"your prompt\"")
			return
		}
	}

	prompt := strings.Join(args, " ")

	// Load config for API call
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}
	provider, model, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider == "" || (provider != ProviderOllama && apiKey == "") {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}
	_ = model // unused but available

	// Get repo context if available
	repoName := getCurrentRepo()
	contextPrompt := prompt
	if repoName != "" {
		contextPrompt = fmt.Sprintf("[repo: %s] %s", repoName, prompt)
	}

	// Call API with minimal system prompt for structured output
	systemPrompt := "Respond concisely in <100 tokens. Output structured text, no markdown formatting. Be direct."

	response, err := callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, contextPrompt)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	// Output structured response (YAML-ish frontmatter + body)
	fmt.Println("---")
	fmt.Printf("mode: ask\n")
	fmt.Printf("repo: %s\n", repoName)
	fmt.Println("---")
	fmt.Println(strings.TrimSpace(response))
}

func handleA2APlan(args []string) {
	if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			input, _ := io.ReadAll(os.Stdin)
			args = []string{string(input)}
		} else {
			fmt.Println("error: no task provided")
			fmt.Println("usage: gg a2a plan \"your task\"")
			return
		}
	}

	task := strings.Join(args, " ")

	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}
	provider, _, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider == "" || (provider != ProviderOllama && apiKey == "") {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}

	repoName := getCurrentRepo()
	contextTask := task
	if repoName != "" {
		contextTask = fmt.Sprintf("[repo: %s] %s", repoName, task)
	}

	systemPrompt := "Output a numbered plan (1. 2. 3. etc). Max 7 steps. No prose, just steps. Each step <15 words."

	response, err := callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, contextTask)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Println("---")
	fmt.Printf("mode: plan\n")
	fmt.Printf("task: %s\n", truncate(task, 50))
	fmt.Println("---")
	fmt.Println(strings.TrimSpace(response))
}

func handleA2ACode(args []string) {
	if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			input, _ := io.ReadAll(os.Stdin)
			args = []string{string(input)}
		} else {
			fmt.Println("error: no task provided")
			fmt.Println("usage: gg a2a code \"your task\"")
			return
		}
	}

	task := strings.Join(args, " ")

	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}
	provider, _, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider == "" || (provider != ProviderOllama && apiKey == "") {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}

	repoName := getCurrentRepo()
	contextTask := task
	if repoName != "" {
		contextTask = fmt.Sprintf("[repo: %s] %s", repoName, task)
	}

	systemPrompt := "Output ONLY code. No explanations, no markdown fences, just raw code. If multiple files, separate with: // FILE: filename.ext"

	response, err := callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, contextTask)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Println("---")
	fmt.Printf("mode: code\n")
	fmt.Println("---")
	fmt.Println(strings.TrimSpace(response))
}

// callAPIWithSystem - simplified API call with custom system prompt
func callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, userPrompt string) (string, error) {
	switch provider {
	case ProviderAnthropic:
		return callAnthropicWithSystem(apiKey, systemPrompt, userPrompt)
	case ProviderOllama:
		return callOllamaWithSystem(endpoint, "", systemPrompt, userPrompt)
	case ProviderOpenAI:
		return callOpenAIWithSystem(apiKey, systemPrompt, userPrompt)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func callAnthropicWithSystem(apiKey, systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 500,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("%s", result.Error.Message)
	}

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}
	return "", fmt.Errorf("no response content")
}

func callOllamaWithSystem(endpoint, model, systemPrompt, userPrompt string) (string, error) {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}

	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": userPrompt,
		"system": systemPrompt,
		"stream": false,
	}

	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post(endpoint+"/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("%s", result.Error)
	}

	return result.Response, nil
}

func callOpenAIWithSystem(apiKey, systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens": 500,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("%s", result.Error.Message)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response content")
}
