package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type SnapReport struct {
	Branch   string `json:"branch"`
	Hash     string `json:"hash"`
	Message  string `json:"message"`
	Status   string `json:"status"`
	Upstream string `json:"upstream"`
	Clean    bool   `json:"clean"`
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func handleSnap() {
	if _, err := gitOutput("rev-parse", "--git-dir"); err != nil {
		emit("Not a git repository", `{"error":"not a git repository"}`)
		return
	}

	branch, _ := gitOutput("rev-parse", "--abbrev-ref", "HEAD")
	if branch == "" {
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

	status, _ := gitOutput("status", "--porcelain")
	var modified, untracked, staged, deleted int
	lines := strings.Split(status, "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		c0, c1 := line[0], line[1]
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
	statusStr := "clean"
	clean := true
	if len(parts) > 0 {
		statusStr = strings.Join(parts, ", ")
		clean = false
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

	plain := fmt.Sprintf("%s @ %s \"%s\" | %s | %s", branch, hash, msg, statusStr, upstreamStr)
	report := SnapReport{
		Branch:   branch,
		Hash:     hash,
		Message:  msg,
		Status:   statusStr,
		Upstream: upstreamStr,
		Clean:    clean,
	}
	jb, _ := json.Marshal(report)
	emit(plain, string(jb))
}
