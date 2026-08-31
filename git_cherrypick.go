package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type cherryPickResult struct {
	branch  string
	hash    string
	message string
	files   int
	ins     int
	del     int
}

func printCherryPickUsage() {
	fmt.Println("Usage: gg cherry-pick <ref>")
	fmt.Println("  Cherry-pick a commit into the current branch.")
	fmt.Println("  Ref can be a commit hash, branch name, or tag.")
	fmt.Println("  Options:")
	fmt.Println("    --help, -h  Show this help message")
}

func cherryPickParseOutput(outputStr string) (*cherryPickResult, error) {
	lines := strings.Split(outputStr, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected output format")
	}

	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "[") {
		return nil, fmt.Errorf("unexpected output format")
	}

	content := strings.TrimPrefix(firstLine, "[")
	content = strings.TrimSuffix(content, "]")

	parts := strings.SplitN(content, " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("unexpected output format")
	}

	branch := parts[0]
	hash := parts[1]
	message := strings.TrimSpace(strings.TrimPrefix(firstLine, "["+branch+" "+hash+"]"))

	statsLine := strings.TrimSpace(lines[1])
	var files, ins, del int
	if strings.Contains(statsLine, "files changed") {
		statsParts := strings.Split(statsLine, ",")
		if len(statsParts) >= 2 {
			fmt.Sscanf(strings.TrimSpace(strings.ReplaceAll(statsParts[0], " files changed", "")), "%d", &files)
			fmt.Sscanf(strings.TrimSpace(strings.ReplaceAll(statsParts[1], " insertions(+)", "")), "%d", &ins)
			if len(statsParts) >= 3 {
				fmt.Sscanf(strings.TrimSpace(strings.ReplaceAll(statsParts[2], " deletions(-)", "")), "%d", &del)
			}
		}
	}

	return &cherryPickResult{
		branch:  branch,
		hash:    hash,
		message: message,
		files:   files,
		ins:     ins,
		del:     del,
	}, nil
}

func handleCherryPick() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printCherryPickUsage()
		return
	}

	ref := os.Args[2]
	cmd := exec.Command("git", "cherry-pick", ref)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		if strings.Contains(outputStr, "The previous cherry-pick is now empty") {
			fmt.Println("  ✗ cherry-pick empty — changes already present")
			fmt.Println("  Token cost: ~6")
			return
		}

		var conflictFiles []string
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "CONFLICT") {
				if idx := strings.Index(line, "Merge conflict in "); idx != -1 {
					conflictFiles = append(conflictFiles, strings.TrimSpace(line[idx+len("Merge conflict in "):]))
				} else if idx := strings.Index(line, "merge conflict in "); idx != -1 {
					conflictFiles = append(conflictFiles, strings.TrimSpace(line[idx+len("merge conflict in "):]))
				}
			}
		}

		if len(conflictFiles) > 0 {
			fmt.Printf("  ✗ cherry-pick conflict in %d files:\n", len(conflictFiles))
			for _, f := range conflictFiles {
				fmt.Printf("    %s\n", f)
			}
			fmt.Println("  Resolve, then: git add . && git cherry-pick --continue")
			fmt.Println("  Token cost: ~12")
		} else {
			fmt.Printf("  ✗ cherry-pick failed: %s\n", strings.TrimSpace(outputStr))
			fmt.Println("  Token cost: ~12")
		}
		return
	}

	result, parseErr := cherryPickParseOutput(outputStr)
	if parseErr != nil {
		fmt.Println("  ✗ cherry-pick failed: unexpected output format")
		fmt.Println("  Token cost: ~12")
		return
	}

	fmt.Printf("  ✓ cherry-picked %s → %s@%s \"%s\" | %d files (+%d -%d)\n", ref, result.branch, result.hash, result.message, result.files, result.ins, result.del)
	fmt.Println("  Token cost: ~10")
}
