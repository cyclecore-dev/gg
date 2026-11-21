package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type StatusFile struct {
	Status   string `json:"status"`
	Filename string `json:"path"`
}

type StatusReport struct {
	Clean  bool         `json:"clean"`
	Branch string       `json:"branch,omitempty"`
	Files  []StatusFile `json:"files"`
	Count  int          `json:"count"`
}

func handleStatus() {
	if hasFlag("--help") || hasFlag("-h") {
		printStatusUsage()
		return
	}

	// delta mode (from earlier stub)
	if hasFlag("--delta") {
		handleStatusDelta()
		return
	}

	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running git status: %v\n", err)
		return
	}

	raw := string(output)
	branch, _ := gitOutput("rev-parse", "--abbrev-ref", "HEAD")

	var files []StatusFile
	if strings.TrimSpace(raw) != "" {
		groups := statusParse(raw)
		for _, g := range []string{"staged", "modified", "untracked", "deleted", "conflicted"} {
			for _, f := range groups[g] {
				files = append(files, f)
			}
		}
	}

	report := StatusReport{
		Clean:  len(files) == 0,
		Branch: branch,
		Files:  files,
		Count:  len(files),
	}

	// plain text
	var plain strings.Builder
	if report.Clean {
		plain.WriteString("  clean — nothing to commit\n  Token cost: ~4")
	} else {
		groups := statusParse(raw)
		order := []string{"staged", "modified", "untracked", "deleted", "conflicted"}
		for _, name := range order {
			fs := groups[name]
			if len(fs) == 0 {
				continue
			}
			fmt.Fprintf(&plain, "\n  %s (%d):\n", name, len(fs))
			limit := 10
			if len(fs) < limit {
				limit = len(fs)
			}
			for i := 0; i < limit; i++ {
				fmt.Fprintf(&plain, "    %s  %s\n", fs[i].Status, fs[i].Filename)
			}
			if len(fs) > 10 {
				fmt.Fprintf(&plain, "    ... and %d more\n", len(fs)-10)
			}
		}
		fmt.Fprintf(&plain, "\n%d files | Token cost: ~%d", report.Count, report.Count)
	}

	jsonBytes, _ := json.Marshal(report)
	emit(plain.String(), string(jsonBytes))
}

func statusParse(raw string) map[string][]StatusFile {
	groups := map[string][]StatusFile{
		"staged":     {},
		"modified":   {},
		"untracked":  {},
		"deleted":    {},
		"conflicted": {},
	}

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		xy := line[:2]
		filename := strings.TrimSpace(line[3:])

		if strings.ContainsAny(xy, "U") {
			groups["conflicted"] = append(groups["conflicted"], StatusFile{Status: "U", Filename: filename})
		} else if strings.ContainsAny(string(xy[0]), "AMDR") {
			groups["staged"] = append(groups["staged"], StatusFile{Status: string(xy[0]), Filename: filename})
		} else if xy[1] == 'M' {
			groups["modified"] = append(groups["modified"], StatusFile{Status: "M", Filename: filename})
		} else if xy[1] == 'D' {
			groups["deleted"] = append(groups["deleted"], StatusFile{Status: "D", Filename: filename})
		} else if xy == "??" {
			groups["untracked"] = append(groups["untracked"], StatusFile{Status: "?", Filename: filename})
		}
	}
	return groups
}

func printStatusUsage() {
	fmt.Println("Usage: gg status [--delta] [--json] [--max-tokens N]")
	fmt.Println("  Shows file-level status grouped by state.")
	fmt.Println("  --delta       only changes since last call")
	fmt.Println("  --json        structured output")
	fmt.Println("  --max-tokens  truncate output")
}
