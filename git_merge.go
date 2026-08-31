package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handleMerge() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printMergeUsage()
		return
	}

	if os.Args[2] == "--abort" {
		cmd := exec.Command("git", "merge", "--abort")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("  ✗ failed to abort merge: %s\n%s\n", sanitizeError(err), string(output))
		} else {
			fmt.Println("  ✓ merge aborted")
		}
		return
	}

	branch := os.Args[2]
	var flags []string
	for _, arg := range os.Args[3:] {
		switch arg {
		case "--no-ff", "--squash":
			flags = append(flags, arg)
		}
	}

	args := append(flags, branch)
	cmd := exec.Command("git", append([]string{"merge"}, args...)...)
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	if strings.Contains(outputStr, "Already up to date") {
		fmt.Printf("  ✓ already up to date with %s\n", branch)
		return
	}

	if err != nil || strings.Contains(outputStr, "CONFLICT") {
		conflicts := mergeExtractConflicts(outputStr)
		if len(conflicts) > 0 {
			fmt.Printf("  ✗ merge conflict in %d files:\n", len(conflicts))
			for _, f := range conflicts {
				fmt.Printf("    %s\n", f)
			}
			fmt.Println("  Resolve, then: git add . && git commit")
		} else {
			fmt.Printf("  ✗ merge failed: %s\n%s\n", sanitizeError(err), outputStr)
		}
		return
	}

	files, ins, del := mergeParseStats(outputStr)
	statsStr := ""
	if files > 0 || ins > 0 || del > 0 {
		statsStr = fmt.Sprintf(" | %d files (+%d -%d)", files, ins, del)
	}

	if strings.Contains(outputStr, "Fast-forward") {
		fmt.Printf("  ✓ merged %s (fast-forward)%s\n", branch, statsStr)
	} else if strings.Contains(outputStr, "Merge made") {
		fmt.Printf("  ✓ merged %s%s\n", branch, statsStr)
	} else if strings.Contains(outputStr, "Squash commit") {
		fmt.Printf("  ✓ squash-merged %s%s\n", branch, statsStr)
		fmt.Println("  Commit staged — run: git commit")
	} else {
		fmt.Printf("  ✓ merged %s%s\n", branch, statsStr)
	}
}

func printMergeUsage() {
	fmt.Println("Usage: gg merge <branch> [--no-ff] [--squash] [--abort]")
	fmt.Println("  Merge the specified branch into the current branch.")
	fmt.Println("  Flags:")
	fmt.Println("    --no-ff   Create a merge commit even if fast-forward is possible")
	fmt.Println("    --squash  Combine changes without creating a merge commit")
	fmt.Println("    --abort   Abort an in-progress merge")
}

func mergeParseStats(output string) (files, ins, del int) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "file changed") || strings.Contains(line, "files changed") {
			fmt.Sscanf(line, "%d file", &files)
			if idx := strings.Index(line, "insertion"); idx != -1 {
				fmt.Sscanf(line[strings.LastIndex(line[:idx], " ")+1:], "%d", &ins)
			}
			if idx := strings.Index(line, "deletion"); idx != -1 {
				fmt.Sscanf(line[strings.LastIndex(line[:idx], " ")+1:], "%d", &del)
			}
			break
		}
	}
	return
}

func mergeExtractConflicts(output string) []string {
	var conflicts []string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "CONFLICT") && strings.Contains(line, "Merge conflict in") {
			parts := strings.Split(line, "Merge conflict in ")
			if len(parts) > 1 {
				conflicts = append(conflicts, strings.TrimSpace(parts[1]))
			}
		}
	}
	return conflicts
}
