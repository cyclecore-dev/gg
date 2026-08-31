package main

import (
	"fmt"
	"os/exec"
	"strings"
)

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
