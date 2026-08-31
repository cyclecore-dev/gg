package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type StatusFile struct {
	Status   string
	Filename string
}

func handleStatus() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		printStatusUsage()
		return
	}

	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running git status: %v\n", err)
		return
	}

	raw := string(output)
	if strings.TrimSpace(raw) == "" {
		fmt.Println("  clean — nothing to commit")
		fmt.Println("  Token cost: ~4")
		return
	}

	groups := statusParse(raw)
	
	totalFiles := 0
	
	// Print groups in order
	printGroup := func(name string, files []StatusFile) {
		if len(files) > 0 {
			statusPrintGroup(name, files)
			totalFiles += len(files)
		}
	}

	printGroup("staged", groups["staged"])
	printGroup("modified", groups["modified"])
	printGroup("untracked", groups["untracked"])
	printGroup("deleted", groups["deleted"])
	printGroup("conflicted", groups["conflicted"])

	fmt.Printf("\n%d files | Token cost: ~%d\n", totalFiles, totalFiles)
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

func statusPrintGroup(name string, files []StatusFile) {
	count := len(files)
	fmt.Printf("\n  %s (%d):\n", name, count)
	
	limit := 10
	if count < limit {
		limit = count
	}

	for i := 0; i < limit; i++ {
		fmt.Printf("    %s  %s\n", files[i].Status, files[i].Filename)
	}

	if count > 10 {
		fmt.Printf("    ... and %d more\n", count-10)
	}
}

func printStatusUsage() {
	fmt.Println("Usage: gg status [--help|-h]")
	fmt.Println("  Shows file-level status grouped by state.")
	fmt.Println("  Example: gg status")
}
