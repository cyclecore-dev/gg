package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handleReset() {
	mode := "--mixed"
	ref := "HEAD"

	// Parse args
	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "--help", "-h":
			printResetUsage()
			return
		case "--soft", "--hard", "--mixed":
			mode = os.Args[2]
			if len(os.Args) > 3 {
				ref = os.Args[3]
			}
		default:
			ref = os.Args[2]
		}
	}

	// Run git reset
	cmd := exec.Command("git", "reset", mode, ref)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ reset failed: %s\n", strings.TrimSpace(string(output)))
		return
	}

	// Check new state based on mode
	var result string
	switch mode {
	case "--soft":
		// Count staged files
		cmd = exec.Command("git", "diff", "--cached", "--stat")
		output, err = cmd.CombinedOutput()
		if err == nil {
			count := resetParseUnstaged(string(output))
			if count > 0 {
				result = fmt.Sprintf("%d files staged", count)
			} else {
				result = "working tree clean"
			}
		} else {
			result = "working tree clean"
		}
	case "--hard":
		// Check if working tree is clean
		cmd = exec.Command("git", "diff", "--stat")
		output, err = cmd.CombinedOutput()
		if err == nil && len(strings.TrimSpace(string(output))) > 0 {
			count := resetParseUnstaged(string(output))
			if count > 0 {
				result = fmt.Sprintf("%d files unstaged", count)
			} else {
				result = "working tree clean"
			}
		} else {
			result = "working tree clean"
		}
	default: // --mixed
		// Count unstaged files
		cmd = exec.Command("git", "diff", "--stat")
		output, err = cmd.CombinedOutput()
		if err == nil {
			count := resetParseUnstaged(string(output))
			if count > 0 {
				result = fmt.Sprintf("%d files unstaged", count)
			} else {
				result = "working tree clean"
			}
		} else {
			result = "working tree clean"
		}
	}

	fmt.Printf("  ✓ reset %s to %s | %s\n", mode, ref, result)
}

func resetParseUnstaged(stat string) int {
	count := 0
	lines := strings.Split(stat, "\n")
	for _, line := range lines {
		if strings.Contains(line, "|") {
			count++
		}
	}
	return count
}

func printResetUsage() {
	fmt.Println("Usage: gg reset [--soft|--hard|--mixed] [ref]")
	fmt.Println("  --soft    Reset to ref, keep staged changes")
	fmt.Println("  --hard    Reset to ref, discard all changes")
	fmt.Println("  --mixed   Reset to ref, unstage changes (default)")
	fmt.Println("  ref       Commit reference (default: HEAD)")
}
