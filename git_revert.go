package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handleRevert() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printRevertUsage()
		return
	}

	ref := os.Args[2]

	// Get original commit info
	logCmd := exec.Command("git", "log", "--oneline", "-1", ref)
	logOut, err := logCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ failed to get info for %s: %s\n", ref, strings.TrimSpace(string(logOut)))
		return
	}

	logLine := strings.TrimSpace(string(logOut))
	parts := strings.SplitN(logLine, " ", 2)
	if len(parts) < 2 {
		fmt.Printf("  ✗ invalid commit info for %s\n", ref)
		return
	}
	originalHash := parts[0]
	originalMsg := parts[1]

	// Run git revert --no-edit
	revertCmd := exec.Command("git", "revert", "--no-edit", ref)
	revertOut, err := revertCmd.CombinedOutput()
	revertOutput := string(revertOut)

	if err != nil || strings.Contains(revertOutput, "CONFLICT") {
		// Conflict handling
		var conflictFiles []string
		for _, line := range strings.Split(revertOutput, "\n") {
			if strings.Contains(line, "CONFLICT") {
				if idx := strings.LastIndex(line, "in "); idx != -1 {
					conflictFiles = append(conflictFiles, strings.TrimSpace(line[idx+3:]))
				}
			}
		}
		
		if len(conflictFiles) == 0 {
			fmt.Printf("  ✗ revert failed for %s %q\n", originalHash, originalMsg)
			fmt.Printf("  %s\n", strings.TrimSpace(revertOutput))
		} else {
			fmt.Printf("  ✗ revert conflict in %d file(s):\n", len(conflictFiles))
			for _, f := range conflictFiles {
				fmt.Printf("  %s\n", f)
			}
			fmt.Printf("  Resolve, then: git add . && git revert --continue\n")
		}
		fmt.Printf("  Token cost: ~10\n")
		return
	}

	// Success parsing
	newHash, branch, files, ins, del := revertParseOutput(revertOutput)
	
	fmt.Printf("  ✓ reverted %s %q → %s@%s | %d files (+%d -%d)\n", 
		originalHash, originalMsg, branch, newHash, files, ins, del)
	fmt.Printf("  Token cost: ~10\n")
}

func printRevertUsage() {
	fmt.Println("Usage: gg revert <ref>")
	fmt.Println("  Reverts the specified commit and creates a new revert commit.")
	fmt.Println("  Token cost: ~10")
}

func revertParseOutput(output string) (hash, branch string, files, ins, del int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			endBracket := strings.Index(line, "]")
			if endBracket > 0 {
				inside := line[1:endBracket]
				insideParts := strings.SplitN(inside, " ", 2)
				if len(insideParts) == 2 {
					branch = insideParts[0]
					hash = insideParts[1]
				}
			}
		} else if strings.Contains(line, "files changed") {
			fmt.Sscanf(line, "%d files changed, %d insertions(+), %d deletions(-)", &files, &ins, &del)
		}
	}
	return
}
