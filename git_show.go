package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func handleShow() {
	if len(os.Args) >= 3 {
		arg := os.Args[2]
		if arg == "--help" || arg == "-h" {
			printShowUsage()
			return
		}
	}

	ref := "HEAD"
	if len(os.Args) >= 3 {
		ref = os.Args[2]
	}

	showCommit(ref)
}

func printShowUsage() {
	fmt.Println("Usage: gg show [ref]")
	fmt.Println("  Show compressed commit details for a reference.")
	fmt.Println("  ref: commit hash, branch, tag, or HEAD (default)")
	fmt.Println("  Example: gg show HEAD~1")
}

func showCommit(ref string) {
	metaCmd := exec.Command("git", "show", "--no-patch", fmt.Sprintf("--format=%s", "%H|||%h|||%an|||%ae|||%cr|||%s"), ref)
	statCmd := exec.Command("git", "show", "--stat", "--format=", ref)
	numstatCmd := exec.Command("git", "show", "--numstat", "--format=", ref)

	metaOut, err := metaCmd.Output()
	if err != nil {
		fatalError("git show metadata failed", err)
	}

	statOut, err := statCmd.Output()
	if err != nil {
		fatalError("git show stat failed", err)
	}

	numstatOut, err := numstatCmd.Output()
	if err != nil {
		fatalError("git show numstat failed", err)
	}

	meta := parseShowMeta(string(metaOut))
	fileSummary := parseStatSummary(string(statOut))
	files := parseNumstat(string(numstatOut))

	// Print compressed output
	fmt.Printf("  %s %s (%s)\n", meta.ShortHash, meta.Author, meta.Age)
	subject := meta.Subject
	if len(subject) > 70 {
		subject = subject[:70] + "..."
	}
	fmt.Printf("    %s\n", subject)
	fmt.Printf("    %s\n", fileSummary)

	// Top 5 files
	if len(files) > 0 {
		limit := 5
		if len(files) < limit {
			limit = len(files)
		}
		for i := 0; i < limit; i++ {
			f := files[i]
			fmt.Printf("      %-20s +%d -%d\n", f.Name, f.Insertions, f.Deletions)
		}
		if len(files) > 5 {
			fmt.Printf("    ... and %d more files\n", len(files)-5)
		}
	}
	fmt.Println("  Token cost: ~25")
}

type commitMeta struct {
	FullHash  string
	ShortHash string
	Author    string
	Email     string
	Age       string
	Subject   string
}

func parseShowMeta(output string) commitMeta {
	parts := strings.Split(strings.TrimSpace(output), "|||")
	if len(parts) < 6 {
		fatalError("Invalid git show metadata format", nil)
	}
	return commitMeta{
		FullHash:  parts[0],
		ShortHash: parts[1],
		Author:    parts[2],
		Email:     parts[3],
		Age:       parts[4],
		Subject:   parts[5],
	}
}

type fileChange struct {
	Name        string
	Insertions  int
	Deletions   int
}

func parseNumstat(output string) []fileChange {
	var files []fileChange
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		ins, err1 := strconv.Atoi(parts[0])
		del, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue
		}
		files = append(files, fileChange{
			Name:        parts[2],
			Insertions:  ins,
			Deletions:   del,
		})
	}
	return files
}

func parseStatSummary(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "0 files"
	}
	lastLine := lines[len(lines)-1]

	parts := strings.Split(lastLine, ",")
	if len(parts) < 1 {
		return "0 files"
	}

	// Part 0: " N files changed"
	fileCountStr := strings.TrimSpace(parts[0])
	idx := strings.Index(fileCountStr, "files")
	if idx == -1 {
		return "0 files"
	}
	fileCountStr = strings.TrimSpace(fileCountStr[:idx])
	fileCount, err := strconv.Atoi(fileCountStr)
	if err != nil {
		return "0 files"
	}

	totalIns := 0
	totalDel := 0

	if len(parts) > 1 {
		insStr := strings.TrimSpace(parts[1])
		insIdx := strings.Index(insStr, "insertions")
		if insIdx != -1 {
			insNumStr := strings.TrimSpace(insStr[:insIdx])
			totalIns, _ = strconv.Atoi(insNumStr)
		}
	}

	if len(parts) > 2 {
		delStr := strings.TrimSpace(parts[2])
		delIdx := strings.Index(delStr, "deletions")
		if delIdx != -1 {
			delNumStr := strings.TrimSpace(delStr[:delIdx])
			totalDel, _ = strconv.Atoi(delNumStr)
		}
	}

	return fmt.Sprintf("%d files (+%d -%d)", fileCount, totalIns, totalDel)
}
