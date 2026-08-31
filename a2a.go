package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getA2AUsagePath() string {
	return filepath.Join(getGGDir(), "a2a_usage.json")
}

// Returns: (allowed, inGrace, remaining)
func checkA2ALimit(cfg *Config) (bool, bool, int) {
	// Pro users have unlimited access
	if checkProTier(cfg) {
		return true, false, -1
	}

	// Load usage
	usagePath := getA2AUsagePath()
	today := time.Now().Format("2006-01-02")

	var usage A2AUsage
	data, err := os.ReadFile(usagePath)
	if err == nil {
		json.Unmarshal(data, &usage)
	}

	// Reset if new day
	if usage.Date != today {
		usage = A2AUsage{Date: today, Count: 0}
	}

	hardLimit := a2aFreeLimit + a2aGraceLimit
	remaining := a2aFreeLimit - usage.Count
	inGrace := usage.Count >= a2aFreeLimit && usage.Count < hardLimit

	return usage.Count < hardLimit, inGrace, remaining
}

func incrementA2AUsage() {
	usagePath := getA2AUsagePath()
	today := time.Now().Format("2006-01-02")

	var usage A2AUsage
	data, err := os.ReadFile(usagePath)
	if err == nil {
		json.Unmarshal(data, &usage)
	}

	if usage.Date != today {
		usage = A2AUsage{Date: today, Count: 0}
	}

	usage.Count++
	newData, _ := json.Marshal(usage)
	os.WriteFile(usagePath, newData, 0644)
}


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

	// Check rate limit
	allowed, inGrace, _ := checkA2ALimit(cfg)
	if !allowed {
		fmt.Println("---")
		fmt.Println("error: rate_limit")
		fmt.Println("message: Free tier limit reached")
		fmt.Println("upgrade: https://ggdotdev.com/pro")
		fmt.Println("---")
		return
	}

	provider, model, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider == "" || (provider != ProviderOllama && apiKey == "") {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}
	_ = model // unused but available

	// Show grace warning (to stderr so it doesn't break pipes)
	if inGrace {
		fmt.Fprintln(os.Stderr, "warning: daily limit reached, using grace period")
	}

	// Get repo context if available
	repoName := getCurrentRepo()
	contextPrompt := prompt
	if repoName != "" {
		contextPrompt = fmt.Sprintf("[repo: %s] %s", repoName, prompt)
	}

	// Call API with minimal system prompt for structured output
	// Light context awareness (not roleplay) to prevent hallucinations
	systemPrompt := "Context: gg CLI (token compression for git/npm/pip/brew). Respond concisely in <100 tokens. Structured text, no markdown. Be direct."

	response, err := callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, contextPrompt)
	if err != nil {
		fmt.Printf("error: %v\n", sanitizeError(err))
		return
	}

	// Track usage (free tier)
	incrementA2AUsage()

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

	// Check rate limit
	allowed, inGrace, _ := checkA2ALimit(cfg)
	if !allowed {
		fmt.Println("---")
		fmt.Println("error: rate_limit")
		fmt.Println("message: Free tier limit reached")
		fmt.Println("upgrade: https://ggdotdev.com/pro")
		fmt.Println("---")
		return
	}

	provider, _, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider == "" || (provider != ProviderOllama && apiKey == "") {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}

	if inGrace {
		fmt.Fprintln(os.Stderr, "warning: daily limit reached, using grace period")
	}

	repoName := getCurrentRepo()
	contextTask := task
	if repoName != "" {
		contextTask = fmt.Sprintf("[repo: %s] %s", repoName, task)
	}

	// Light context awareness (not roleplay)
	systemPrompt := "Context: gg CLI. Output a numbered plan (1. 2. 3. etc). Max 7 steps. No prose, just steps. Each step <15 words."

	response, err := callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, contextTask)
	if err != nil {
		fmt.Printf("error: %v\n", sanitizeError(err))
		return
	}

	// Track usage
	incrementA2AUsage()

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

	// Check rate limit
	allowed, inGrace, _ := checkA2ALimit(cfg)
	if !allowed {
		fmt.Println("---")
		fmt.Println("error: rate_limit")
		fmt.Println("message: Free tier limit reached")
		fmt.Println("upgrade: https://ggdotdev.com/pro")
		fmt.Println("---")
		return
	}

	provider, _, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider == "" || (provider != ProviderOllama && apiKey == "") {
		fmt.Println("error: not configured")
		fmt.Println("run: gg init")
		return
	}

	if inGrace {
		fmt.Fprintln(os.Stderr, "warning: daily limit reached, using grace period")
	}

	repoName := getCurrentRepo()
	contextTask := task
	if repoName != "" {
		contextTask = fmt.Sprintf("[repo: %s] %s", repoName, task)
	}

	// Light context awareness (not roleplay)
	systemPrompt := "Context: gg CLI. Output ONLY code. No explanations, no markdown fences, just raw code. If multiple files, separate with: // FILE: filename.ext"

	response, err := callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, contextTask)
	if err != nil {
		fmt.Printf("error: %v\n", sanitizeError(err))
		return
	}

	// Track usage
	incrementA2AUsage()

	fmt.Println("---")
	fmt.Printf("mode: code\n")
	fmt.Println("---")
	fmt.Println(strings.TrimSpace(response))
}

// callAPIWithSystem - simplified API call with custom system prompt
