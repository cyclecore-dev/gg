package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func handleSnap() {
	if _, err := gitOutput("rev-parse", "--git-dir"); err != nil {
		fmt.Println("Not a git repository")
		return
	}

	branch, err := gitOutput("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		branch = "unknown"
	}

	commit, err := gitOutput("log", "--oneline", "-1")
	var hash, msg string
	if err != nil {
		hash = "unknown"
		msg = "unknown"
	} else {
		parts := strings.SplitN(commit, " ", 2)
		hash = parts[0]
		if len(parts) > 1 {
			msg = parts[1]
		}
	}

	if len(msg) > 40 {
		msg = msg[:40] + "..."
	}

	status, err := gitOutput("status", "--porcelain")
	var modified, untracked, staged, deleted int
	if err == nil {
		lines := strings.Split(status, "\n")
		for _, line := range lines {
			if len(line) < 2 {
				continue
			}
			c0 := line[0]
			c1 := line[1]
			if c1 == 'M' {
				modified++
			}
			if c1 == '?' {
				untracked++
			}
			if c0 == 'A' {
				staged++
			}
			if c0 == 'D' {
				deleted++
			}
		}
	}

	var parts []string
	if modified > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", modified))
	}
	if untracked > 0 {
		parts = append(parts, fmt.Sprintf("%d untracked", untracked))
	}
	if staged > 0 {
		parts = append(parts, fmt.Sprintf("%d staged", staged))
	}
	if deleted > 0 {
		parts = append(parts, fmt.Sprintf("%d deleted", deleted))
	}
	var statusStr string
	if len(parts) == 0 {
		statusStr = "clean"
	} else {
		statusStr = strings.Join(parts, ", ")
	}

	upstreamStr := "no upstream"
	upstream, err := gitOutput("rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err == nil {
		parts := strings.Split(upstream, "\t")
		if len(parts) == 2 {
			ahead, _ := strconv.Atoi(parts[0])
			behind, _ := strconv.Atoi(parts[1])
			upstreamStr = fmt.Sprintf("↑%d ↓%d", ahead, behind)
		}
	}

	fmt.Printf("%s @ %s \"%s\" | %s | %s\n", branch, hash, msg, statusStr, upstreamStr)
}
