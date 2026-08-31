package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// PullState represents the result of a git pull operation
type PullState int

const (
	StateAlreadyUpToDate PullState = iota
	StateFastForward
	StateMerge
	StateConflict
	StateUnknown
)

// PullResult holds the parsed information from a git pull operation
type PullResult struct {
	State      PullState
	Commits    int
	Files      int
	Insertions int
	Deletions  int
	Conflicts  []string
}

func handlePull() {
	// Check for help flag
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		printPullUsage()
		return
	}

	// Determine remote and branch
	remote := "origin"
	branch := ""
	if len(os.Args) > 2 {
		remote = os.Args[2]
	}
	if len(os.Args) > 3 {
		branch = os.Args[3]
	} else {
		// Get current branch
		out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
		if err == nil {
			branch = strings.TrimSpace(string(out))
		} else {
			fmt.Println("Error: Could not determine current branch")
			return
		}
	}

	// Get old HEAD hash
	oldOut, err := exec.Command("git", "rev-parse", "--short", "HEAD").CombinedOutput()
	if err != nil {
		fmt.Println("Error: Could not get old HEAD hash")
		return
	}
	oldHash := strings.TrimSpace(string(oldOut))

	// Run git pull
	pullCmd := exec.Command("git", "pull", remote, branch)
	out, err := pullCmd.CombinedOutput()
	outStr := string(out)

	// Get new HEAD hash
	newOut, err := exec.Command("git", "rev-parse", "--short", "HEAD").CombinedOutput()
	if err != nil {
		fmt.Println("Error: Could not get new HEAD hash")
		return
	}
	newHash := strings.TrimSpace(string(newOut))

	// Parse output
	result := pullParseOutput(outStr, err, oldHash, newHash)

	// Print result based on state
	switch result.State {
	case StateAlreadyUpToDate:
		fmt.Println("  ✓ already up to date")
		fmt.Println("  Token cost: ~4")
	case StateFastForward:
		fmt.Printf("  ✓ pulled %d commits (fast-forward) | %d file%s (+%d -%d)\n",
			result.Commits, result.Files, pullPluralize(result.Files), result.Insertions, result.Deletions)
		fmt.Println("  Token cost: ~10")
	case StateMerge:
		fmt.Printf("  ✓ pulled + merged | %d file%s (+%d -%d)\n",
			result.Files, pullPluralize(result.Files), result.Insertions, result.Deletions)
		fmt.Println("  Token cost: ~8")
	case StateConflict:
		fmt.Printf("  ✗ conflict in %d files:\n", len(result.Conflicts))
		for _, f := range result.Conflicts {
			fmt.Printf("    %s\n", f)
		}
		fmt.Println("  Fix conflicts, then: git add . && git commit")
		fmt.Println("  Token cost: ~12")
	default:
		fmt.Println("  ? unknown pull result")
	}
}

func printPullUsage() {
	fmt.Println("Usage: gg pull [remote] [branch]")
	fmt.Println("  remote  : Remote name (default: origin)")
	fmt.Println("  branch  : Branch name (default: current branch)")
	fmt.Println("  Example: gg pull origin main")
}

func pullParseOutput(outStr string, err error, oldHash, newHash string) PullResult {
	result := PullResult{}

	// Check for conflict first (output contains CONFLICT or exit code != 0)
	if strings.Contains(outStr, "CONFLICT") || err != nil {
		result.State = StateConflict
		result.Conflicts = pullExtractConflicts(outStr)
		return result
	}

	// Check for already up to date
	if strings.Contains(outStr, "Already up to date") {
		result.State = StateAlreadyUpToDate
		return result
	}

	// Count commits
	if oldHash != newHash {
		countOut, err := exec.Command("git", "rev-list", "--count", oldHash+".."+newHash).CombinedOutput()
		if err == nil {
			count, parseErr := strconv.Atoi(strings.TrimSpace(string(countOut)))
			if parseErr == nil {
				result.Commits = count
			}
		}
	}

	// Check for fast-forward
	if strings.Contains(outStr, "Fast-forward") {
		result.State = StateFastForward
		result.Files, result.Insertions, result.Deletions = pullParseStats(outStr)
		return result
	}

	// Check for merge
	if strings.Contains(outStr, "Merge made") || strings.Contains(outStr, "Merge branch") {
		result.State = StateMerge
		result.Files, result.Insertions, result.Deletions = pullParseStats(outStr)
		return result
	}

	// Unknown state
	result.State = StateUnknown
	return result
}

func pullExtractConflicts(outStr string) []string {
	var conflicts []string
	lines := strings.Split(outStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "CONFLICT") {
			// Try to extract filename from "Merge conflict in filename"
			if idx := strings.Index(line, "Merge conflict in "); idx != -1 {
				filename := strings.TrimSpace(line[idx+len("Merge conflict in "):])
				if filename != "" {
					conflicts = append(conflicts, filename)
				}
			}
		}
	}
	return conflicts
}

func pullParseStats(outStr string) (files, insertions, deletions int) {
	// Look for the summary line
	// Pattern: N file(s) changed, N insertion(s)(+), N deletion(s)(-)
	// Example: 1 file changed, 10 insertions(+), 5 deletions(-)
	
	// Find the line with "file changed"
	for _, line := range strings.Split(outStr, "\n") {
		if strings.Contains(line, "file changed") {
			// Split by comma
			parts := strings.Split(line, ",")
			if len(parts) >= 3 {
				// Parse files
				files = pullParseInt(parts[0])
				// Parse insertions
				insertions = pullParseInt(parts[1])
				// Parse deletions
				deletions = pullParseInt(parts[2])
			}
			break
		}
	}
	return files, insertions, deletions
}

func pullParseInt(s string) int {
	var numStr string
	for _, char := range s {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		} else {
			break
		}
	}
	if numStr == "" {
		return 0
	}
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return num
}

func pullPluralize(count int) string {
	if count == 1 {
		return "file"
	}
	return "files"
}
