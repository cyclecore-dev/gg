package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func printCloneUsage() {
	fmt.Println("Usage: gg clone <url|user/repo> [target-dir]")
	fmt.Println("  url         Full git URL or GitHub user/repo shorthand")
	fmt.Println("  target-dir  Optional directory name")
	fmt.Println("  -h, --help  Show this help message")
}

func cloneShortURL(url string) string {
	url = strings.TrimSuffix(url, ".git")
	if strings.HasPrefix(url, "https://github.com/") {
		return url[19:]
	}
	if strings.HasPrefix(url, "git@github.com:") {
		return url[15:]
	}
	return url
}

func handleClone() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printCloneUsage()
		return
	}

	url := os.Args[2]
	if !strings.Contains(url, "://") && !strings.HasPrefix(url, "git@") && strings.Contains(url, "/") {
		url = "https://github.com/" + url
	}

	args := []string{"clone", url}
	var targetDir string
	if len(os.Args) > 3 {
		targetDir = os.Args[3]
		args = append(args, targetDir)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		errMsg := strings.TrimSpace(outStr)
		if strings.Contains(errMsg, "not found") {
			errMsg = "repository not found"
		} else if strings.Contains(errMsg, "already exists") {
			errMsg = "destination path already exists"
		}
		fmt.Printf("  ✗ clone failed: %s\n", errMsg)
		fmt.Println("  Token cost: ~8")
		return
	}

	if targetDir == "" {
		if idx := strings.Index(outStr, "Cloning into '"); idx != -1 {
			start := idx + 14
			end := strings.Index(outStr[start:], "'")
			if end != -1 {
				targetDir = outStr[start : start+end]
			}
		}
		if targetDir == "" {
			targetDir = filepath.Base(strings.TrimSuffix(url, ".git"))
		}
	}

	branch := "main"
	branchCmd := exec.Command("git", "-C", targetDir, "rev-parse", "--abbrev-ref", "HEAD")
	if branchOut, err := branchCmd.Output(); err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	displayURL := cloneShortURL(url)
	fmt.Printf("  ✓ cloned %s → ./%s (%s)\n", displayURL, targetDir, branch)
	fmt.Println("  Token cost: ~8")
}
