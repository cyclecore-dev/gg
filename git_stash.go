package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func stashList() {
	cmd := exec.Command("git", "stash", "list", "--format=%gd|||%s|||%cr")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("  ✗ stash list failed:", err)
		return
	}

	content := strings.TrimSpace(string(out))
	if content == "" {
		fmt.Println("  0 stashes")
		fmt.Println("  Token cost: ~3")
		return
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Println("  " + line)
		}
	}

	fmt.Printf("  %d stashes | Token cost: ~8\n", len(lines))
}

func stashPop() {
	cmd := exec.Command("git", "stash", "pop")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ stash pop failed: %s\n", strings.TrimSpace(string(out)))
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	fileCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "M\t") || strings.HasPrefix(line, "A\t") || strings.HasPrefix(line, "D\t") {
			fileCount++
		}
	}
	if fileCount == 0 {
		fileCount = len(lines) - 1
		if fileCount < 1 {
			fileCount = 1
		}
	}

	fmt.Printf("  ✓ popped stash@{0} | %d files restored\n", fileCount)
	fmt.Println("  Token cost: ~6")
}

func stashShow() {
	cmd := exec.Command("git", "stash", "show", "--stat")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("  ✗ stash show failed:", err)
		return
	}

	content := strings.TrimSpace(string(out))
	lines := strings.Split(content, "\n")

	if len(lines) == 0 {
		fmt.Println("  ✗ empty stash")
		return
	}

	summary := lines[len(lines)-1]
	var fileCount, ins, del int
	fmt.Sscanf(strings.TrimSpace(summary), "%d file", &fileCount)
	if idx := strings.Index(summary, "insertion"); idx != -1 {
		sub := strings.TrimSpace(summary[:idx])
		parts := strings.Split(sub, ",")
		last := strings.TrimSpace(parts[len(parts)-1])
		fmt.Sscanf(last, "%d", &ins)
	}
	if idx := strings.Index(summary, "deletion"); idx != -1 {
		sub := strings.TrimSpace(summary[:idx])
		parts := strings.Split(sub, ",")
		last := strings.TrimSpace(parts[len(parts)-1])
		fmt.Sscanf(last, "%d", &del)
	}

	fmt.Printf("  stash@{0}: %d files (+%d -%d)\n", fileCount, ins, del)
	fmt.Println("  Token cost: ~6")
}

func stashDrop() {
	cmd := exec.Command("git", "stash", "drop")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ stash drop failed: %s\n", strings.TrimSpace(string(out)))
		return
	}

	fmt.Println("  ✓ dropped stash@{0}")
	fmt.Println("  Token cost: ~4")
}

func stashSave() {
	var message string
	if len(os.Args) > 3 {
		message = os.Args[3]
	} else {
		message = "WIP"
	}

	cmd := exec.Command("git", "stash", "push", "-m", message)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ stash save failed: %s\n", strings.TrimSpace(string(out)))
		return
	}

	content := strings.TrimSpace(string(out))
	if strings.Contains(content, "No local changes") {
		fmt.Println("  Nothing to stash")
		fmt.Println("  Token cost: ~3")
		return
	}

	fmt.Printf("  ✓ stashed | \"%s\"\n", message)
	fmt.Println("  Token cost: ~6")
}

func handleStash() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		stashList()
		return
	}

	subcmd := os.Args[2]
	switch subcmd {
	case "list":
		stashList()
	case "pop":
		stashPop()
	case "show":
		stashShow()
	case "drop":
		stashDrop()
	case "save", "push":
		stashSave()
	default:
		stashSave()
	}
}
