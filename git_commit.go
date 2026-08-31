package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func printCommitUsage() {
	fmt.Println("Usage: gg commit <message> [-a|--all]")
	fmt.Println("  -a, --all  Stage all modified tracked files before committing")
}

func commitParseNumber(s string) int {
	res := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			res = res*10 + int(c-'0')
		} else {
			break
		}
	}
	return res
}

func commitParseOutput(output string) (branch, hash, msg string, files, insertions, deletions int) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return
	}

	// Parse first line: [branch hash] message
	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "[") {
		return
	}

	closeBracket := strings.Index(firstLine, "]")
	if closeBracket == -1 {
		return
	}

	branchHash := firstLine[1:closeBracket]
	parts := strings.SplitN(branchHash, " ", 2)
	if len(parts) >= 2 {
		branch = parts[0]
		hash = parts[1]
	}

	// Message is after "] "
	if closeBracket+2 < len(firstLine) {
		msg = firstLine[closeBracket+2:]
	}

	// Parse second line: N file changed, ...
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "0 files") {
			files = 0
			return
		}
		if strings.Contains(trimmed, "file changed") {
			parts := strings.Fields(trimmed)
			if len(parts) > 0 {
				files = commitParseNumber(parts[0])
			}

			for i, part := range parts {
				if strings.Contains(part, "insertion") {
					if i > 0 {
						insertions = commitParseNumber(parts[i-1])
					}
				}
				if strings.Contains(part, "deletion") {
					if i > 0 {
						deletions = commitParseNumber(parts[i-1])
					}
				}
			}
			break
		}
	}
	return
}

func handleCommit() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printCommitUsage()
		return
	}

	message := os.Args[2]
	allFlag := false

	for _, arg := range os.Args[3:] {
		if arg == "-a" || arg == "--all" {
			allFlag = true
			break
		}
	}

	cmd := exec.Command("git", "diff", "--cached", "--stat")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error checking staged changes: %v\n", err)
		return
	}

	if len(strings.TrimSpace(string(out))) == 0 {
		if !allFlag {
			fmt.Println("Nothing staged. Use 'git add' first or 'gg commit \"msg\" -a' to auto-stage.")
			fmt.Println("  Token cost: ~10")
			return
		}
	}

	var gitArgs []string
	if allFlag {
		gitArgs = []string{"commit", "-a", "-m", message}
	} else {
		gitArgs = []string{"commit", "-m", message}
	}

	cmd = exec.Command("git", gitArgs...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n", strings.TrimSpace(string(out)))
		return
	}

	branch, hash, msg, files, insertions, deletions := commitParseOutput(string(out))

	filesStr := fmt.Sprintf("%d file", files)
	if files != 1 {
		filesStr += "s"
	}

	fmt.Printf("  ✓ %s@%s \"%s\" | %s (+%d -%d)\n", branch, hash, msg, filesStr, insertions, deletions)
	fmt.Println("  Token cost: ~8")
}
