package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type worktreeEntry struct {
	path   string
	hash   string
	branch string
}

func worktreeParseList(output string) []worktreeEntry {
	var entries []worktreeEntry
	var current worktreeEntry
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			if current.path != "" {
				entries = append(entries, current)
				current = worktreeEntry{}
			}
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			current.path = line[9:]
		} else if strings.HasPrefix(line, "HEAD ") {
			current.hash = line[5:]
			if len(current.hash) > 7 {
				current.hash = current.hash[:7]
			}
		} else if strings.HasPrefix(line, "branch refs/heads/") {
			current.branch = line[18:]
		} else if line == "detached" {
			current.branch = "(detached)"
		}
	}
	if current.path != "" {
		entries = append(entries, current)
	}
	return entries
}

func printWorktreeUsage() {
	fmt.Println("Usage: gg worktree [list|add|remove] [path] [branch]")
	fmt.Println("  list    List worktrees")
	fmt.Println("  add     Add a new worktree")
	fmt.Println("  remove  Remove a worktree")
}

func handleWorktree() {
	action := "list"
	if len(os.Args) > 2 {
		action = os.Args[2]
	}

	if action == "--help" || action == "-h" {
		printWorktreeUsage()
		return
	}

	switch action {
	case "list":
		cmd := exec.Command("git", "worktree", "list", "--porcelain")
		output, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		entries := worktreeParseList(string(output))
		fmt.Printf("  worktrees (%d):\n", len(entries))
		for _, e := range entries {
			fmt.Printf("    %-20s %-10s %s\n", e.path, e.branch, e.hash)
		}
		fmt.Println("  Token cost: ~15")
	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gg worktree add <path> [branch]")
			return
		}
		path := os.Args[3]
		branch := ""
		if len(os.Args) > 4 {
			branch = os.Args[4]
		}
		args := []string{"worktree", "add", path}
		if branch != "" {
			args = append(args, branch)
		}
		cmd := exec.Command("git", args...)
		_, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		if branch != "" {
			fmt.Printf("  ✓ worktree added: %s (branch: %s)\n", path, branch)
		} else {
			fmt.Printf("  ✓ worktree added: %s\n", path)
		}
		fmt.Println("  Token cost: ~8")
	case "remove":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gg worktree remove <path>")
			return
		}
		path := os.Args[3]
		cmd := exec.Command("git", "worktree", "remove", path)
		_, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Printf("  ✓ worktree removed: %s\n", path)
		fmt.Println("  Token cost: ~5")
	default:
		printWorktreeUsage()
	}
}
