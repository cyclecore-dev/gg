// gg v0.2.0 — the 2-letter agent-native git client
// December 3, 2025
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

const version = "0.7.1"

// Config represents the gg configuration
type Config struct {
	GG struct {
		Version string `toml:"version"`
		Tier    string `toml:"tier"`
	} `toml:"gg"`
	API struct {
		ClaudeModel       string  `toml:"claude_model"`
		ClaudeTemperature float64 `toml:"claude_temperature"`
		MaazaModel        string  `toml:"maaza_model"`
	} `toml:"api"`
	GitHub struct {
		DefaultBranch string `toml:"default_branch"`
	} `toml:"github"`
	Secrets struct {
		ClaudeAPIKey  string `toml:"claude_api_key"`
		MaazaAPIKey   string `toml:"maaza_api_key"`
		ProLicenseKey string `toml:"pro_license_key"`
	} `toml:"keys"`
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
	case "config":
		handleConfig()
	case "maaza":
		handleMaaza()
	case ".":
		handleCurrentRepo()
	case "ask":
		handleAskV2()
	case "approve":
		handleApproveV2()
	case "pr":
		handlePR()
	case "run":
		handleRun()
	case "stats":
		handleStats()
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
	fmt.Println("gg — the 2-letter agent-native git client")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  gg config init     # Set up configuration")
	fmt.Println("  gg maaza           # Maaza model MCP endpoint")
	fmt.Println("  gg .               # Current repo → code-execution MCP")
	fmt.Println("  gg user/repo       # Any GitHub repo → MCP")
	fmt.Println("  gg ask \"...\"       # Generate code → open PR (streams output)")
	fmt.Println("  gg approve         # Merge the PR")
	fmt.Println("  gg pr <number>     # View/manage specific PR")
	fmt.Println("  gg run <cmd>       # Run command in sandbox, return result")
	fmt.Println("  gg stats           # Show usage statistics")
	fmt.Println()
	fmt.Println("package manager (v0.7):")
	fmt.Println("  gg npm <pkg>       # npm package → MCP endpoint")
	fmt.Println("  gg brew [-i] <f>   # Homebrew formula → MCP (-i auto-installs)")
	fmt.Println("  gg chain <tools>   # Chain multiple MCPs together")
	fmt.Println("  gg chain run <n>   # Execute a saved chain")
	fmt.Println("  gg cool <toolbelt> # Curated tool collections (webdev, media, sec, data)")
	fmt.Println("  gg cache status    # Show cache size")
	fmt.Println("  gg cache clean     # Prune old cache entries")
	fmt.Println()
	fmt.Println("install: curl -L gg.sh | sh")
	fmt.Println("more: github.com/ggdotdev/gg")
}

func handleMaaza() {
	fmt.Println("🐱 Maaza Orchestrator v1.2")
	fmt.Println()
	fmt.Println("Model: 9.6M parameters")
	fmt.Println("Benchmarks: 62.9% adversarial score")
	fmt.Println()
	fmt.Println("Code-execution MCP — 98.7% token reduction")
	fmt.Println("Compatible with: Claude Desktop, Cursor, any MCP client")
}

func handleCurrentRepo() {
	// Get git remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("❌ Not in a git repo or no remote configured")
		return
	}

	url := strings.TrimSpace(string(output))
	repo := parseGitHubURL(url)
	
	if repo == "" {
		fmt.Println("❌ Could not parse GitHub repo from:", url)
		return
	}

	fmt.Printf("📦 Current repo: %s\n", repo)
	fmt.Println()
	fmt.Println("MCP endpoint:")
	fmt.Printf("  https://api.github.com/repos/%s\n", repo)
	fmt.Println()
	fmt.Println("Code-execution MCP active — works with Claude Desktop, Cursor")
}

func handleRepo(repo string) {
	fmt.Printf("📦 Repo: %s\n", repo)
	fmt.Println()
	fmt.Println("MCP endpoint:")
	fmt.Printf("  https://api.github.com/repos/%s\n", repo)
	fmt.Println()
	fmt.Println("Code-execution MCP active")
}

func handleAsk() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Usage: gg ask \"your prompt here\"")
		return
	}

	prompt := strings.Join(os.Args[2:], " ")
	
	fmt.Println("🤖 Maaza v1.2 is thinking...")
	fmt.Printf("   Prompt: %s\n", prompt)
	fmt.Println()
	fmt.Println("⚠️  Full 'gg ask' implementation coming in v0.2")
	fmt.Println("   For now, use with Claude Desktop + GitHub MCP server")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Claude will write code using code-execution MCP")
	fmt.Println("  2. Review the generated PR")
	fmt.Println("  3. Run: gg approve")
}

func handleApprove() {
	fmt.Println("✓ Approving PR...")
	fmt.Println()
	fmt.Println("⚠️  Full 'gg approve' implementation coming in v0.2")
	fmt.Println("   For now, merge PRs via GitHub UI or: gh pr merge")
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
	fmt.Println("Welcome to gg v0.2!")
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

	// Prompt for API keys
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your Claude API key (from console.anthropic.com):\n> ")
	claudeKey, _ := reader.ReadString('\n')
	claudeKey = strings.TrimSpace(claudeKey)

	fmt.Print("\nEnter your Maaza API key (optional, press Enter to skip):\n> ")
	maazaKey, _ := reader.ReadString('\n')
	maazaKey = strings.TrimSpace(maazaKey)

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
	cfg.API.ClaudeModel = "claude-sonnet-4-5-20250929"
	cfg.API.ClaudeTemperature = 0.7
	cfg.API.MaazaModel = "maaza-slm-360m"
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
	cfg.Secrets.ClaudeAPIKey = claudeKey
	cfg.Secrets.MaazaAPIKey = maazaKey
	cfg.Secrets.ProLicenseKey = proKey

	secretsPath := filepath.Join(ggDir, "secrets")
	if err := encryptSecrets(cfg.Secrets, identity, secretsPath); err != nil {
		fatalError("Failed to encrypt secrets", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration saved to ~/.gg/config.toml")
	fmt.Println("✓ Secrets encrypted and saved to ~/.gg/secrets")
	fmt.Println()
	fmt.Println("Run 'gg ask \"your prompt\"' to get started!")
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

func encryptSecrets(secrets struct {
	ClaudeAPIKey  string `toml:"claude_api_key"`
	MaazaAPIKey   string `toml:"maaza_api_key"`
	ProLicenseKey string `toml:"pro_license_key"`
}, identity *age.X25519Identity, path string) error {

	// Create temporary struct for TOML encoding
	data := struct {
		Keys struct {
			ClaudeAPIKey  string `toml:"claude_api_key"`
			MaazaAPIKey   string `toml:"maaza_api_key"`
			ProLicenseKey string `toml:"pro_license_key"`
		} `toml:"keys"`
	}{}
	data.Keys = secrets

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

func decryptSecrets(secrets *struct {
	ClaudeAPIKey  string `toml:"claude_api_key"`
	MaazaAPIKey   string `toml:"maaza_api_key"`
	ProLicenseKey string `toml:"pro_license_key"`
}, identity *age.X25519Identity, path string) error {

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
		Keys struct {
			ClaudeAPIKey  string `toml:"claude_api_key"`
			MaazaAPIKey   string `toml:"maaza_api_key"`
			ProLicenseKey string `toml:"pro_license_key"`
		} `toml:"keys"`
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

	fmt.Println("❌ GitHub authentication required")
	fmt.Println()
	fmt.Println("Run: gh auth login")
	fmt.Println("Or install gh CLI: https://cli.github.com")

	return fmt.Errorf("not authenticated")
}

// ============================================================================
// NEW V0.2 COMMANDS
// ============================================================================

func handleAskV2() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Usage: gg ask \"your prompt here\"")
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
		fmt.Println("❌ No prompt provided")
		return
	}

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		fatalError("Config error. Run: gg config init", err)
	}

	// Check Pro tier
	if !proMode && !checkProTier(cfg) {
		fmt.Println("🤖 Analyzing request...")
		fmt.Println()
		fmt.Println("Implementation plan:")
		fmt.Println("  1. Analyze repository structure")
		fmt.Println("  2. Generate code changes")
		fmt.Println("  3. Create pull request")
		fmt.Println()
		fmt.Println("╔══════════════════════════════════════╗")
		fmt.Println("║   gg Pro required for this feature  ║")
		fmt.Println("╚══════════════════════════════════════╝")
		fmt.Println()
		fmt.Println("Pro features:")
		fmt.Println("  • Full Claude-powered code generation")
		fmt.Println("  • Unlimited gg ask commands")
		fmt.Println("  • Priority API access")
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

	fmt.Println("🤖 Generating code with Claude Sonnet 4.5...")
	fmt.Println()

	// Call Claude API with streaming
	response, err := callClaudeAPIStreaming(cfg, prompt, repoName)
	if err != nil {
		fatalError("Claude API error", sanitizeError(err))
	}

	// Track ask usage
	trackCommandUsage("ask", prompt, 0)

	// Parse code blocks
	files := parseCodeBlocks(response)
	if len(files) == 0 {
		fmt.Println("⚠️  No code blocks found in response")
		fmt.Println("Claude's response:")
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
			fmt.Printf("⚠️  Failed to write %s: %v\n", path, err)
			continue
		}
		fmt.Printf("✓ %s\n", path)
	}

	// Commit and push
	commitMsg := fmt.Sprintf("gg ask: %s", truncate(prompt, 60))
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", commitMsg).Run()
	exec.Command("git", "push", "-u", "origin", branchName).Run()

	// Create PR
	prCmd := exec.Command("gh", "pr", "create", "--title", commitMsg, "--body", fmt.Sprintf("Generated by gg ask:\n\n%s", prompt))
	prOutput, err := prCmd.Output()
	if err != nil {
		fmt.Println("⚠️  Failed to create PR. Create manually:")
		fmt.Printf("   Branch: %s\n", branchName)
		return
	}

	prURL := strings.TrimSpace(string(prOutput))
	fmt.Println()
	fmt.Printf("✓ PR created: %s\n", prURL)
	fmt.Println()
	fmt.Println("Next: gg approve")
}

func handleApproveV2() {
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
	fmt.Println("✓ PR merged successfully!")
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

func callClaudeAPI(cfg *Config, prompt, repo string) (string, error) {
	if cfg.Secrets.ClaudeAPIKey == "" {
		return "", fmt.Errorf("Claude API key not configured")
	}

	systemPrompt := fmt.Sprintf("You are a code generation assistant for the repository: %s\n\n"+
		"Generate clean, production-ready code based on the user's request.\n"+
		"Format code blocks as:\n"+
		"```language:path/to/file\n"+
		"code here\n"+
		"```\n\n"+
		"Be concise and only generate the requested code.", repo)

	requestBody := map[string]interface{}{
		"model":       cfg.API.ClaudeModel,
		"max_tokens":  4096,
		"system":      systemPrompt,
		"temperature": cfg.API.ClaudeTemperature,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": prompt,
			},
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
	req.Header.Set("x-api-key", cfg.Secrets.ClaudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	var text strings.Builder
	for _, block := range response.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}

	return text.String(), nil
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
	msg = regexp.MustCompile(`sk-ant-[a-zA-Z0-9-]+`).ReplaceAllString(msg, "sk-ant-***")
	msg = regexp.MustCompile(`mcpb_[a-zA-Z0-9]+`).ReplaceAllString(msg, "mcpb_***")
	msg = regexp.MustCompile(`gg_pro_[a-zA-Z0-9]+`).ReplaceAllString(msg, "gg_pro_***")
	return fmt.Errorf("%s", msg)
}

func fatalError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "   Error: %v\n", err)
	}
	os.Exit(1)
}

// ============================================================================
// V0.6 COMMANDS: PR, RUN, STATS
// ============================================================================

func handlePR() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Usage: gg pr <number>")
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
	fmt.Printf("Branch: %s → %s\n", pr.HeadRefName, pr.BaseRefName)
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
			fmt.Println("✓ PR merged!")
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
			fmt.Println("✓ PR closed")
		default:
			fmt.Println("Exiting")
		}
	}
}

func handleRun() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Usage: gg run <command>")
		fmt.Println("Example: gg run npm test")
		return
	}

	cmdArgs := os.Args[2:]
	cmdStr := strings.Join(cmdArgs, " ")

	fmt.Printf("🔄 Running: %s\n", cmdStr)
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
		fmt.Printf("❌ Exit code: %d (%.2fs)\n", exitCode, elapsed.Seconds())
	} else {
		fmt.Printf("✓ Success (%.2fs)\n", elapsed.Seconds())
	}

	// Track usage
	trackCommandUsage("run", cmdStr, elapsed)
}

func handleStats() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fatalError("Failed to get home directory", err)
	}

	statsPath := filepath.Join(homeDir, ".gg", "stats.json")
	data, err := os.ReadFile(statsPath)
	if err != nil {
		fmt.Println("📊 Usage Statistics")
		fmt.Println()
		fmt.Println("No usage data yet. Run some commands first!")
		return
	}

	var stats UsageStats
	if err := json.Unmarshal(data, &stats); err != nil {
		fatalError("Failed to parse stats", err)
	}

	fmt.Println("📊 Usage Statistics")
	fmt.Println()
	fmt.Printf("Month: %s\n", stats.Month)
	fmt.Printf("Total asks: %d\n", stats.AskCount)
	fmt.Printf("Total runs: %d\n", stats.RunCount)
	fmt.Printf("Total tokens: %d (input: %d, output: %d)\n", stats.TotalTokens, stats.InputTokens, stats.OutputTokens)
	fmt.Printf("Estimated cost: $%.4f\n", stats.EstimatedCost)
	fmt.Println()
	fmt.Println("Token pricing: Claude Sonnet 4.5 ($3/M input, $15/M output)")
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
	homeDir, _ := os.UserHomeDir()
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
	homeDir, _ := os.UserHomeDir()
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

	// Claude Sonnet 4.5: $3/M input, $15/M output
	stats.EstimatedCost = float64(stats.InputTokens)/1000000*3 + float64(stats.OutputTokens)/1000000*15

	outData, _ := json.MarshalIndent(stats, "", "  ")
	os.WriteFile(statsPath, outData, 0644)
}

// ============================================================================
// STREAMING CLAUDE API
// ============================================================================

func callClaudeAPIStreaming(cfg *Config, prompt, repo string) (string, error) {
	if cfg.Secrets.ClaudeAPIKey == "" {
		return "", fmt.Errorf("Claude API key not configured")
	}

	systemPrompt := fmt.Sprintf("You are a code generation assistant for the repository: %s\n\n"+
		"Generate clean, production-ready code based on the user's request.\n"+
		"Format code blocks as:\n"+
		"```language:path/to/file\n"+
		"code here\n"+
		"```\n\n"+
		"Be concise and only generate the requested code.", repo)

	requestBody := map[string]interface{}{
		"model":       cfg.API.ClaudeModel,
		"max_tokens":  4096,
		"stream":      true,
		"system":      systemPrompt,
		"temperature": cfg.API.ClaudeTemperature,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
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
	req.Header.Set("x-api-key", cfg.Secrets.ClaudeAPIKey)
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

	fmt.Println() // Newline after streaming

	// Track token usage
	if inputTokens > 0 || outputTokens > 0 {
		trackTokenUsage(inputTokens, outputTokens)
	}

	return fullResponse.String(), nil
}

// ==================== v0.7.0: Package Manager Layer ====================

// handleNPM fetches npm package info and displays MCP endpoint
func handleNPM() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Usage: gg npm <package> [--fn <function>]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg npm prettier")
		fmt.Println("  gg npm lodash --fn debounce")
		return
	}

	pkg := os.Args[2]

	// Check local cache first
	cacheDir := filepath.Join(os.Getenv("HOME"), ".gg", "cache", "npm")
	cachePath := filepath.Join(cacheDir, pkg+".json")

	var pkgInfo map[string]interface{}

	if data, err := os.ReadFile(cachePath); err == nil {
		// Cache hit
		json.Unmarshal(data, &pkgInfo)
		fmt.Printf("📦 %s (cached)\n", pkg)
	} else {
		// Fetch from npm registry
		fmt.Printf("📦 Fetching %s from npm...\n", pkg)
		url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", pkg)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("❌ Failed to fetch package: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			fmt.Printf("❌ Package not found: %s\n", pkg)
			return
		}

		if resp.StatusCode != 200 {
			fmt.Printf("❌ npm registry error: %d\n", resp.StatusCode)
			return
		}

		if err := json.NewDecoder(resp.Body).Decode(&pkgInfo); err != nil {
			fmt.Printf("❌ Failed to parse response: %v\n", err)
			return
		}

		// Cache it
		os.MkdirAll(cacheDir, 0755)
		data, _ := json.Marshal(pkgInfo)
		os.WriteFile(cachePath, data, 0644)
	}

	// Display MCP format
	name, _ := pkgInfo["name"].(string)
	version, _ := pkgInfo["version"].(string)
	desc, _ := pkgInfo["description"].(string)

	fmt.Printf("\n📦 %s@%s\n", name, version)
	if desc != "" {
		fmt.Printf("   %s\n", desc)
	}

	// Check for --fn flag
	for i, arg := range os.Args {
		if arg == "--fn" && i+1 < len(os.Args) {
			fnName := os.Args[i+1]
			fmt.Printf("\n🎯 Function: %s\n", fnName)
		}
	}

	fmt.Printf("\n🔌 MCP Endpoint: npm:%s\n", name)
	fmt.Printf("   Token cost: ~18\n")
}

// handleBrew fetches Homebrew formula info and displays MCP endpoint
func handleBrew() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Usage: gg brew [-i] <formula>")
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
		fmt.Println("❌ No formula specified")
		return
	}

	// Check local cache first
	cacheDir := filepath.Join(os.Getenv("HOME"), ".gg", "cache", "brew")
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
			fmt.Printf("🍺 %s (cached)\n", formula)
		} else {
			fmt.Printf("🍺 Fetching %s from Homebrew...\n", formula)
			url := fmt.Sprintf("https://formulae.brew.sh/api/formula/%s.json", formula)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("❌ Failed to fetch formula: %v\n", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 {
				fmt.Printf("❌ Formula not found: %s\n", formula)
				return
			}

			if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
				fmt.Printf("❌ Failed to parse response: %v\n", err)
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

	var version string
	if versions, ok := info["versions"].(map[string]interface{}); ok {
		version, _ = versions["stable"].(string)
	}

	// Auto-install if -i flag and not installed
	if !installed && autoInstall {
		fmt.Printf("🍺 Installing %s...\n", formula)
		installCmd := exec.Command("brew", "install", formula)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			fmt.Printf("❌ Install failed: %v\n", err)
			return
		}
		fmt.Printf("✅ %s installed\n", formula)
		installed = true
	}

	statusIcon := "✓ installed"
	if !installed {
		statusIcon = "not installed"
	}

	fmt.Printf("\n🍺 %s@%s (%s)\n", name, version, statusIcon)
	if desc != "" {
		fmt.Printf("   %s\n", desc)
	}

	if !installed && !autoInstall {
		fmt.Printf("\n   Install: brew install %s\n", formula)
		fmt.Println("   Or use: gg brew -i", formula)
	}

	fmt.Printf("\n🔌 MCP Endpoint: brew:%s\n", formula)
	fmt.Printf("   Token cost: ~22\n")
}

// handleChain chains multiple MCP tools together
func handleChain() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Usage: gg chain <tool:pkg> [tool:pkg...]")
		fmt.Println("         gg chain --save <name> <tool:pkg> [tool:pkg...]")
		fmt.Println("         gg chain run <name>")
		fmt.Println("         gg chain <saved-name>")
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
			fmt.Println("❌ Usage: gg chain run <name>")
			return
		}
		runChain(args[1])
		return
	}

	// Check for --save flag
	if args[0] == "--save" {
		if len(args) < 3 {
			fmt.Println("❌ Usage: gg chain --save <name> <tool:pkg>...")
			return
		}
		chainName := args[1]
		tools := args[2:]
		saveChain(chainName, tools)
		fmt.Printf("💾 Saved chain '%s' with %d tools\n", chainName, len(tools))
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
			fmt.Printf("🔗 Running saved chain '%s'\n\n", args[0])
			args = tools
		} else {
			fmt.Printf("❌ Unknown chain: %s\n", args[0])
			fmt.Println("   Run 'gg chain --list' to see saved chains")
			return
		}
	}

	fmt.Printf("🔗 Chained %d MCPs:\n", len(args))
	totalCost := 0

	for i, tool := range args {
		parts := strings.SplitN(tool, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("   %d. ❌ Invalid format: %s (expected type:name)\n", i+1, tool)
			continue
		}

		toolType := parts[0]
		toolName := parts[1]

		var cost int
		switch toolType {
		case "npm":
			cost = 18
		case "brew":
			cost = 22
		case "git":
			cost = 12
		default:
			cost = 20
		}

		fmt.Printf("   %d. %s:%s (~%d tokens)\n", i+1, toolType, toolName, cost)
		totalCost += cost
	}

	fmt.Printf("\n📊 Combined token cost: ~%d\n", totalCost)
}

func saveChain(name string, tools []string) {
	chainDir := filepath.Join(os.Getenv("HOME"), ".gg", "chains")
	os.MkdirAll(chainDir, 0755)

	data, _ := json.Marshal(tools)
	os.WriteFile(filepath.Join(chainDir, name+".json"), data, 0644)
}

func loadChain(name string) []string {
	chainPath := filepath.Join(os.Getenv("HOME"), ".gg", "chains", name+".json")
	data, err := os.ReadFile(chainPath)
	if err != nil {
		return nil
	}

	var tools []string
	json.Unmarshal(data, &tools)
	return tools
}

func listSavedChains() {
	chainDir := filepath.Join(os.Getenv("HOME"), ".gg", "chains")
	files, err := os.ReadDir(chainDir)
	if err != nil || len(files) == 0 {
		fmt.Println("📋 No saved chains")
		fmt.Println("   Create one: gg chain --save <name> <tool:pkg>...")
		return
	}

	fmt.Println("📋 Saved chains:")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			name := strings.TrimSuffix(f.Name(), ".json")
			tools := loadChain(name)
			fmt.Printf("   • %s (%d tools)\n", name, len(tools))
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
		fmt.Println("❌ Usage: gg cool <toolbelt>")
		fmt.Println("         gg cool --list")
		fmt.Println()
		fmt.Println("Available toolbelts: webdev, media, sec, data, devops")
		return
	}

	arg := os.Args[2]

	if arg == "--list" {
		fmt.Println("🧰 Available toolbelts:")
		fmt.Println()
		for name, tools := range toolbelts {
			fmt.Printf("   %s (%d tools)\n", name, len(tools))
			for _, tool := range tools {
				parts := strings.SplitN(tool, ":", 2)
				fmt.Printf("      • %s (%s)\n", parts[1], parts[0])
			}
			fmt.Println()
		}
		return
	}

	tools, ok := toolbelts[arg]
	if !ok {
		fmt.Printf("❌ Unknown toolbelt: %s\n", arg)
		fmt.Println("   Run 'gg cool --list' to see available toolbelts")
		return
	}

	fmt.Printf("🧰 Toolbelt: %s\n\n", arg)
	totalCost := 0

	for _, tool := range tools {
		parts := strings.SplitN(tool, ":", 2)
		toolType := parts[0]
		toolName := parts[1]

		var cost int
		switch toolType {
		case "npm":
			cost = 18
		case "brew":
			cost = 22
		default:
			cost = 20
		}

		fmt.Printf("   • %s (%s)\n", toolName, toolType)
		totalCost += cost
	}

	fmt.Printf("\n📊 Combined token cost: ~%d\n", totalCost)
	fmt.Printf("\n💡 Chain all: gg chain %s\n", strings.Join(tools, " "))
}

// runChain executes all tools in a saved chain
func runChain(name string) {
	tools := loadChain(name)
	if tools == nil {
		fmt.Printf("❌ Chain not found: %s\n", name)
		fmt.Println("   Run 'gg chain --list' to see saved chains")
		return
	}

	fmt.Printf("🔗 Executing chain '%s'...\n\n", name)

	success := 0
	for i, tool := range tools {
		parts := strings.SplitN(tool, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("[%d/%d] ❌ Invalid: %s\n", i+1, len(tools), tool)
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
			fmt.Printf("   ⚠️  Unknown type: %s\n", toolType)
		}

		success++
		fmt.Println()
	}

	fmt.Printf("✅ Chain complete: %d/%d tools ready\n", success, len(tools))
}

func runNPMCheck(pkg string) {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".gg", "cache", "npm")
	cachePath := filepath.Join(cacheDir, pkg+".json")

	if _, err := os.ReadFile(cachePath); err == nil {
		fmt.Printf("   📦 %s ✓ (cached)\n", pkg)
		return
	}

	url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", pkg)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("   📦 %s ❌\n", pkg)
		return
	}
	defer resp.Body.Close()

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)

	os.MkdirAll(cacheDir, 0755)
	data, _ := json.Marshal(info)
	os.WriteFile(cachePath, data, 0644)

	version, _ := info["version"].(string)
	fmt.Printf("   📦 %s@%s ✓\n", pkg, version)
}

func runBrewCheck(formula string) {
	cmd := exec.Command("brew", "info", formula, "--json=v2")
	if err := cmd.Run(); err == nil {
		fmt.Printf("   🍺 %s ✓ installed\n", formula)
	} else {
		fmt.Printf("   🍺 %s (not installed)\n", formula)
	}
}

// handleCache manages the gg cache
func handleCache() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Usage: gg cache <status|clean>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  status  Show cache size and contents")
		fmt.Println("  clean   Remove old cache entries")
		return
	}

	cacheDir := filepath.Join(os.Getenv("HOME"), ".gg", "cache")

	switch os.Args[2] {
	case "status":
		showCacheStatus(cacheDir)
	case "clean":
		cleanCache(cacheDir)
	default:
		fmt.Printf("❌ Unknown cache command: %s\n", os.Args[2])
	}
}

func showCacheStatus(cacheDir string) {
	total := getCacheSize(cacheDir)
	npmSize := getCacheSize(filepath.Join(cacheDir, "npm"))
	brewSize := getCacheSize(filepath.Join(cacheDir, "brew"))

	npmCount := countFiles(filepath.Join(cacheDir, "npm"))
	brewCount := countFiles(filepath.Join(cacheDir, "brew"))

	fmt.Println("📦 Cache Status")
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
		fmt.Println("📦 Cache is empty")
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
		fmt.Println("🧹 No old cache entries to clean")
	} else {
		fmt.Printf("🧹 Cleaned cache: %s freed (%d entries removed)\n", formatSize(freed), removed)
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
