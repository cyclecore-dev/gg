package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type LogCommit struct {
	Hash   string `json:"hash"`
	Message string `json:"message"`
	Stat   string `json:"stat,omitempty"`
}

type LogReport struct {
	Count   int         `json:"count"`
	Commits []LogCommit `json:"commits"`
}

func handleLog() {
	if hasFlag("--help") || hasFlag("-h") {
		fmt.Println("Usage: gg log [count] [--stat] [--author=name] [--json] [--max-tokens N]")
		fmt.Println("\nOptions:")
		fmt.Println("  count        Number of commits (default: 10)")
		fmt.Println("  --stat       Include change counts per commit")
		fmt.Println("  --author=X   Filter by author name")
		fmt.Println("  --json       structured output")
		fmt.Println("  --max-tokens truncate")
		return
	}

	count := "10"
	showStat := hasFlag("--stat")
	author := getFlagValue("--author")
	if author == "" {
		for _, arg := range os.Args[2:] {
			if strings.HasPrefix(arg, "--author=") {
				author = strings.TrimPrefix(arg, "--author=")
			}
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
		emit("Not a git repository", `{"error":"not a git repository"}`)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		emit("No commits", `{"count":0,"commits":[]}`)
		return
	}

	var commits []LogCommit
	var plain strings.Builder

	if !showStat {
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, " ", 2)
			hash := parts[0]
			msg := ""
			if len(parts) > 1 {
				msg = parts[1]
			}
			if len(msg) > 60 {
				msg = msg[:60] + "..."
			}
			commits = append(commits, LogCommit{Hash: hash, Message: msg})
			fmt.Fprintf(&plain, "  %s %s\n", hash, msg)
		}
	} else {
		var current string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.Contains(trimmed, "file") && strings.Contains(trimmed, "changed") {
				stat := parseStat(trimmed)
				parts := strings.SplitN(current, " ", 2)
				hash, msg := parts[0], ""
				if len(parts) > 1 {
					msg = parts[1]
				}
				commits = append(commits, LogCommit{Hash: hash, Message: msg, Stat: stat})
				fmt.Fprintf(&plain, "  %s %s %s\n", hash, msg, stat)
				current = ""
			} else {
				current = trimmed
			}
		}
		if current != "" {
			parts := strings.SplitN(current, " ", 2)
			hash, msg := parts[0], ""
			if len(parts) > 1 {
				msg = parts[1]
			}
			commits = append(commits, LogCommit{Hash: hash, Message: msg})
			fmt.Fprintf(&plain, "  %s %s\n", hash, msg)
		}
	}

	report := LogReport{Count: len(commits), Commits: commits}
	header := fmt.Sprintf("%d commits:\n", len(commits))
	jb, _ := json.Marshal(report)
	emit(header+plain.String(), string(jb))
}

func parseStat(line string) string {
	ins, del, files := 0, 0, 0
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
