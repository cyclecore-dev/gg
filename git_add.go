package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handleAdd() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printAddUsage()
		return
	}

	targets := os.Args[2:]

	// Run git add
	args := append([]string{"add"}, targets...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ add failed: %s\n", strings.TrimSpace(string(output)))
		fmt.Println("  Token cost: ~8")
		return
	}

	// Get staged diff stat
	statCmd := exec.Command("git", "diff", "--cached", "--stat")
	statOut, _ := statCmd.Output()
	statStr := strings.TrimSpace(string(statOut))

	if statStr == "" {
		fmt.Println("  ✓ nothing to stage")
		fmt.Println("  Token cost: ~4")
		return
	}

	// Parse the summary line (last line of --stat output)
	// Format: " N files changed, N insertions(+), N deletions(-)"
	lines := strings.Split(statStr, "\n")
	summary := lines[len(lines)-1]

	files, ins, del := addParseStatLine(summary)
	if files == 1 {
		fmt.Printf("  ✓ staged %s (+%d -%d)\n", targets[0], ins, del)
	} else {
		fmt.Printf("  ✓ staged %d files (+%d -%d)\n", files, ins, del)
	}
	fmt.Println("  Token cost: ~6")
}

func printAddUsage() {
	fmt.Println("Usage: gg add [files...]")
	fmt.Println("  Add files to the staging area.")
	fmt.Println("  Use '.' to add all files in the current directory.")
	fmt.Println("  Use '-A' to add all files in the repository.")
}

func addParseStatLine(line string) (files, ins, del int) {
	line = strings.TrimSpace(line)
	fmt.Sscanf(line, "%d file", &files)
	if idx := strings.Index(line, "insertion"); idx != -1 {
		// Scan backwards from "insertion" to find the number
		sub := strings.TrimSpace(line[:idx])
		parts := strings.Split(sub, ",")
		last := strings.TrimSpace(parts[len(parts)-1])
		fmt.Sscanf(last, "%d", &ins)
	}
	if idx := strings.Index(line, "deletion"); idx != -1 {
		sub := strings.TrimSpace(line[:idx])
		parts := strings.Split(sub, ",")
		last := strings.TrimSpace(parts[len(parts)-1])
		fmt.Sscanf(last, "%d", &del)
	}
	return
}
