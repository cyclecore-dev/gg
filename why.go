package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// why.go — compressed causal summary for agents
// gg why <symbol|file:line|"error snippet">

func handleWhy() {
	if len(os.Args) < 3 || hasFlag("--help") || hasFlag("-h") {
		fmt.Println("Usage: gg why <symbol|file:line|\"error text\"> [--json] [--max-tokens N]")
		fmt.Println()
		fmt.Println("Compressed causal summary (replaces blame+log+grep+test).")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg why auth.ValidateToken")
		fmt.Println("  gg why src/auth.go:47")
		fmt.Println("  gg why \"nil pointer in handler\"")
		return
	}

	query := strings.Join(os.Args[2:], " ")
	// strip flags
	var clean []string
	for _, a := range strings.Fields(query) {
		if !strings.HasPrefix(a, "--") {
			clean = append(clean, a)
		}
	}
	query = strings.Join(clean, " ")

	report := whyAnalyze(query)
	plain := report.Plain()
	jb, _ := json.Marshal(report)
	emit(plain, string(jb))
}

type WhyReport struct {
	Query       string   `json:"query"`
	Introduced  string   `json:"introduced,omitempty"`
	LastChanged string   `json:"last_changed,omitempty"`
	Callers     int      `json:"callers"`
	Tests       string   `json:"tests,omitempty"`
	Context     string   `json:"context,omitempty"`
	Files       []string `json:"files,omitempty"`
}

func (r WhyReport) Plain() string {
	var b strings.Builder
	fmt.Fprintf(&b, "why: %s\n", r.Query)
	if r.Introduced != "" {
		fmt.Fprintf(&b, "  introduced: %s\n", r.Introduced)
	}
	if r.LastChanged != "" {
		fmt.Fprintf(&b, "  last changed: %s\n", r.LastChanged)
	}
	if r.Callers > 0 {
		fmt.Fprintf(&b, "  callers: %d\n", r.Callers)
	}
	if r.Tests != "" {
		fmt.Fprintf(&b, "  tests: %s\n", r.Tests)
	}
	if r.Context != "" {
		fmt.Fprintf(&b, "  related: %s\n", r.Context)
	}
	return strings.TrimRight(b.String(), "\n")
}

func whyAnalyze(query string) WhyReport {
	r := WhyReport{Query: query}

	// file:line form
	if strings.Contains(query, ":") && !strings.Contains(query, " ") {
		parts := strings.SplitN(query, ":", 2)
		file, line := parts[0], parts[1]
		r.Context = file + ":" + line
		// blame that line
		if out, err := exec.Command("git", "blame", "-L", line+","+line, "--porcelain", file).Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				hash := strings.Fields(lines[0])[0]
				if msg, err := exec.Command("git", "log", "-1", "--format=%s", hash).Output(); err == nil {
					r.LastChanged = hash[:7] + " " + strings.TrimSpace(string(msg))
				}
			}
		}
		return r
	}

	// symbol or free-text search
	// 1. find files containing the symbol
	grep := exec.Command("git", "grep", "-l", "-n", query)
	if out, err := grep.Output(); err == nil {
		files := strings.Split(strings.TrimSpace(string(out)), "\n")
		r.Files = files
		r.Callers = len(files)
		if len(files) > 0 && files[0] != "" {
			r.Context = files[0]
		}
	}

	// 2. last commit that touched any matching file (best-effort)
	if len(r.Files) > 0 && r.Files[0] != "" {
		if out, err := exec.Command("git", "log", "-1", "--format=%h %s", "--", r.Files[0]).Output(); err == nil {
			r.LastChanged = strings.TrimSpace(string(out))
		}
		// first introduction
		if out, err := exec.Command("git", "log", "--reverse", "-1", "--format=%h %s", "--", r.Files[0]).Output(); err == nil {
			r.Introduced = strings.TrimSpace(string(out))
		}
	}

	// 3. rough test hint
	if strings.Contains(strings.ToLower(query), "test") || true {
		// look for test files mentioning the symbol
		if out, err := exec.Command("git", "grep", "-l", query, "--", "*_test.go", "*test*", "*spec*").Output(); err == nil {
			tests := strings.Split(strings.TrimSpace(string(out)), "\n")
			if len(tests) > 0 && tests[0] != "" {
				r.Tests = fmt.Sprintf("✓ %d files", len(tests))
			}
		}
	}

	return r
}
