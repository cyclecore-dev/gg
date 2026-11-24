package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func getCurrentRepo() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	url := strings.TrimSpace(string(output))
	return parseGitHubURL(url)
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
	msg = regexp.MustCompile(`sk-ant-[a-zA-Z0-9_-]+`).ReplaceAllString(msg, "sk-***")
	msg = regexp.MustCompile(`sk-[a-zA-Z0-9_-]+`).ReplaceAllString(msg, "sk-***")
	msg = regexp.MustCompile(`mcpb_[a-zA-Z0-9_-]+`).ReplaceAllString(msg, "mcpb_***")
	msg = regexp.MustCompile(`gg_pro_[a-zA-Z0-9_-]+`).ReplaceAllString(msg, "gg_pro_***")
	msg = regexp.MustCompile(`ghp_[a-zA-Z0-9]+`).ReplaceAllString(msg, "ghp_***")
	msg = regexp.MustCompile(`ghs_[a-zA-Z0-9]+`).ReplaceAllString(msg, "ghs_***")
	msg = regexp.MustCompile(`github_pat_[a-zA-Z0-9_]+`).ReplaceAllString(msg, "github_pat_***")
	msg = regexp.MustCompile(`Bearer\s+[a-zA-Z0-9_.-]+`).ReplaceAllString(msg, "Bearer ***")
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
