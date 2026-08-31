package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func handleReflog() {
	// 1. Check for help
	if len(os.Args) >= 3 {
		arg := os.Args[2]
		if arg == "--help" || arg == "-h" {
			printReflogUsage()
			return
		}
	}

	// 2. Determine count
	count := 10
	if len(os.Args) >= 3 {
		parsed, err := strconv.Atoi(os.Args[2])
		if err == nil {
			count = parsed
		}
	}

	// 3. Run git reflog
	// Format: %h|||%gd|||%gs|||%cr
	gitArgs := []string{
		"reflog",
		"--format=%h|||%gd|||%gs|||%cr",
		"-n", strconv.Itoa(count),
	}

	cmd := exec.Command("git", gitArgs...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running git reflog: %v\n", err)
		return
	}

	// 4. Parse output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var entries []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|||")
		if len(parts) != 4 {
			continue
		}

		hash := parts[0]
		selector := parts[1]
		subject := parts[2]
		age := parts[3]

		// Truncate subject
		if len(subject) > 50 {
			subject = subject[:50]
		}

		// Compress age
		compressedAge := reflogCompressAge(age)

		// Format entry
		entry := fmt.Sprintf("    %s %s %s (%s)", hash, selector, subject, compressedAge)
		entries = append(entries, entry)
	}

	// 5. Print output
	fmt.Printf("  reflog (%d entries):\n", len(entries))
	for _, entry := range entries {
		fmt.Println(entry)
	}
	fmt.Println("  Token cost: ~15")
}

func reflogCompressAge(age string) string {
	age = strings.Replace(age, " hours", "h", 1)
	age = strings.Replace(age, " hour", "h", 1)
	age = strings.Replace(age, " days", "d", 1)
	age = strings.Replace(age, " day", "d", 1)
	age = strings.Replace(age, " weeks", "w", 1)
	age = strings.Replace(age, " week", "w", 1)
	age = strings.Replace(age, " months", "mo", 1)
	age = strings.Replace(age, " month", "mo", 1)
	age = strings.Replace(age, " minutes", "min", 1)
	age = strings.Replace(age, " minute", "min", 1)
	age = strings.Replace(age, " seconds", "s", 1)
	age = strings.Replace(age, " second", "s", 1)
	age = strings.Replace(age, " years", "y", 1)
	age = strings.Replace(age, " year", "y", 1)
	return age
}

func printReflogUsage() {
	fmt.Println("Usage: gg reflog [count] [--help]")
	fmt.Println("  count: Number of entries to show (default 10)")
	fmt.Println("  --help: Show this help message")
}
