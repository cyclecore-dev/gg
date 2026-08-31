package main

import (
	"fmt"
	"os"
	"strings"
)

func handleChat() {
	if len(os.Args) < 3 || os.Args[2] == "-h" || os.Args[2] == "--help" {
		fmt.Println("Usage: gg chat \"your question\"")
		fmt.Println()
		fmt.Println("Human-friendly AI chat. Unlike 'gg a2a ask', this provides")
		fmt.Println("natural conversational responses without structured output.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg chat \"how do I install dependencies?\"")
		fmt.Println("  gg chat \"explain this error message\"")
		return
	}

	prompt := strings.Join(os.Args[2:], " ")

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("Error: not configured")
		fmt.Println("Run: gg init")
		return
	}

	// Check rate limit
	allowed, inGrace, _ := checkA2ALimit(cfg) // Reuse a2a limits
	if !allowed {
		fmt.Println("Daily limit reached. Upgrade to Pro for unlimited: https://ggdotdev.com/pro")
		return
	}

	provider, _, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider == "" || (provider != ProviderOllama && apiKey == "") {
		fmt.Println("Error: not configured")
		fmt.Println("Run: gg init")
		return
	}

	if inGrace {
		fmt.Println("(Using grace period - upgrade to Pro for unlimited)")
		fmt.Println()
	}

	// Get repo context if available
	repoName := getCurrentRepo()
	contextPrompt := prompt
	if repoName != "" {
		contextPrompt = fmt.Sprintf("[Context: working in repo %s] %s", repoName, prompt)
	}

	// Human-friendly system prompt with gg awareness
	systemPrompt := `You are a helpful coding assistant running inside gg (the 2-letter agent-native CLI).

Available gg commands:
- gg chat "..." - This conversation
- gg ask "..." - Generate code and create PR (Pro tier)
- gg pip/npm/brew <pkg> - Package lookup (~18 tokens)
- gg a2a ask/plan/code - Agent-to-agent structured output
- gg . - Current repo MCP endpoint
- gg help - Full command list

Give clear, practical answers. Be conversational but concise. If asked about gg, reference the commands above.`

	fmt.Println("Thinking...")
	fmt.Println()

	response, err := callAPIWithSystem(provider, endpoint, apiKey, systemPrompt, contextPrompt)
	if err != nil {
		fmt.Printf("Error: %v\n", sanitizeError(err))
		return
	}

	// Track usage
	incrementA2AUsage()

	fmt.Println(strings.TrimSpace(response))
}

// CLI2CLI: Agent-to-Agent modes
// Design: <100 tokens output, pipeable, parseable
