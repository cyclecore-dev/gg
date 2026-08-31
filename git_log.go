package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func handleLog() {
	count := "10"
	showStat := false
	author := ""

	for _, arg := range os.Args[2:] {
		switch {
		case arg == "--stat":
			showStat = true
		case strings.HasPrefix(arg, "--author="):
			author = strings.TrimPrefix(arg, "--author=")
		case arg == "--help" || arg == "-h":
			fmt.Println("Usage: gg log [count] [--stat] [--author=name]")
			fmt.Println("\nOptions:")
			fmt.Println("  count        Number of commits (default: 10)")
			fmt.Println("  --stat       Include change counts per commit")
			fmt.Println("  --author=X   Filter by author name")
			fmt.Println("\nOutput:")
			fmt.Println("  10 commits:")
			fmt.Println("    a1b2c3d fix auth middleware")
			fmt.Println("    b3c4d5e add login flow (+15 -0 1f)")
			return
		default:
			if _, err := strconv.Atoi(arg); err == nil {
				count = arg
			}
		}
	}

	args := []string{"log", "--oneline", "-n", count}
	if author != "" {
		args = append(args, "--author="+author)
	}
	if showStat {
		args = append(args, "--shortstat")
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Not a git repository")
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Println("No commits")
		return
	}

	if !showStat {
		commitCount := len(lines)
		fmt.Printf("%d commits:\n", commitCount)
		for _, line := range lines {
			if len(line) > 67 {
				line = line[:67] + "..."
			}
			fmt.Printf("  %s\n", line)
		}
		return
	}

	var commits []string
	var currentCommit string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "file") && strings.Contains(trimmed, "changed") {
			stat := parseStat(trimmed)
			commits = append(commits, currentCommit+" "+stat)
		} else {
			currentCommit = trimmed
		}
	}
	if currentCommit != "" && (len(commits) == 0 || !strings.HasPrefix(commits[len(commits)-1], currentCommit)) {
		commits = append(commits, currentCommit)
	}

	fmt.Printf("%d commits:\n", len(commits))
	for _, c := range commits {
		if len(c) > 80 {
			c = c[:80] + "..."
		}
		fmt.Printf("  %s\n", c)
	}
}

func parseStat(line string) string {
	ins := 0
	del := 0
	files := 0
	parts := strings.Split(line, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		fields := strings.Fields(p)
		if len(fields) >= 2 {
			n, _ := strconv.Atoi(fields[0])
			if strings.Contains(p, "insertion") {
				ins = n
			} else if strings.Contains(p, "deletion") {
				del = n
			} else if strings.Contains(p, "file") {
				files = n
			}
		}
	}
	return fmt.Sprintf("(+%d -%d %df)", ins, del, files)
}
