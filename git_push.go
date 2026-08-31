package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func handlePush() {
	// 1. Parse args
	for _, arg := range os.Args[2:] {
		if arg == "--help" || arg == "-h" {
			printPushUsage()
			return
		}
	}

	force := false
	var remote, branch string
	args := os.Args[2:]
	
	var positionalArgs []string
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		} else {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	if len(positionalArgs) >= 1 {
		remote = positionalArgs[0]
	} else {
		remote = "origin"
	}

	if len(positionalArgs) >= 2 {
		branch = positionalArgs[1]
	} else {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		out, err := cmd.Output()
		if err != nil {
			fmt.Println("Error getting current branch:", err)
			return
		}
		branch = strings.TrimSpace(string(out))
	}

	// Check ahead/behind
	ahead := -1
	upstreamCmd := exec.Command("git", "rev-parse", "@{upstream}")
	if _, err := upstreamCmd.CombinedOutput(); err == nil {
		cmd := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
		out, err := cmd.CombinedOutput()
		if err == nil {
			parts := strings.Split(strings.TrimSpace(string(out)), "\t")
			if len(parts) >= 1 {
				ahead, _ = strconv.Atoi(parts[0])
			}
		}
	}

	if ahead == 0 {
		fmt.Println("Nothing to push — already up to date")
		fmt.Println("Token cost: ~5")
		return
	}

	// Build push command
	pushArgs := []string{"push"}
	if force {
		pushArgs = append(pushArgs, "--force")
	}
	pushArgs = append(pushArgs, remote, branch)

	cmd := exec.Command("git", pushArgs...)
	out, err := cmd.CombinedOutput()
	outputStr := string(out)

	// Parse output
	if err != nil {
		if strings.Contains(outputStr, "rejected") {
			fmt.Println("✗ push rejected — pull first (origin/main has diverged)")
			fmt.Println("Token cost: ~8")
		} else {
			fmt.Println("✗ push failed:", err)
			fmt.Println("Token cost: ~8")
		}
		return
	}

	if strings.Contains(outputStr, "Everything up-to-date") {
		fmt.Println("Nothing to push — already up to date")
		fmt.Println("Token cost: ~5")
		return
	}

	success, hashRange := pushParseOutput(outputStr)
	if success {
		if force {
			fmt.Printf("✓ force pushed → %s/%s (%s)\n", remote, branch, hashRange)
		} else {
			fmt.Printf("✓ pushed %d commits → %s/%s (%s)\n", ahead, remote, branch, hashRange)
		}
		fmt.Println("Token cost: ~8")
	} else {
		fmt.Println("✗ push failed: unknown error")
		fmt.Println("Token cost: ~8")
	}
}

func printPushUsage() {
	fmt.Println("Usage: gg push [remote] [branch]")
	fmt.Println("Options:")
	fmt.Println("  --force, -f    Force push")
	fmt.Println("  --help, -h     Show help")
}

func pushParseOutput(output string) (bool, string) {
	if strings.Contains(output, "..") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "..") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.Contains(part, "..") {
						return true, part
					}
				}
			}
		}
	}
	return false, ""
}
