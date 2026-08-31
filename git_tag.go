package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handleTag() {
	subcmd := "list"

	if len(os.Args) >= 3 {
		arg := os.Args[2]
		switch arg {
		case "list":
			subcmd = "list"
		case "create", "new":
			subcmd = "create"
		case "delete", "rm":
			subcmd = "delete"
		case "--help", "-h":
			printTagUsage()
			return
		default:
			subcmd = "create" // shorthand: treat as tag name
		}
	}

	switch subcmd {
	case "list":
		tagList()
	case "create":
		tagCreate()
	case "delete":
		tagDelete()
	}
}

func printTagUsage() {
	fmt.Println("Usage: gg tag [list|create|delete|<name>]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list              List tags (default)")
	fmt.Println("  create <name> [msg]  Create a tag")
	fmt.Println("  delete <name>     Delete a tag")
	fmt.Println("  <name>            Shorthand for create")
	fmt.Println("  --help, -h        Show this help")
}

func tagList() {
	cmd := exec.Command("git", "tag", "--sort=-creatordate", "--format=%(refname:short)|||%(objectname:short)|||%(creatordate:relative)|||%(subject)")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ tag list failed: %v\n", sanitizeError(err))
		return
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		fmt.Println("  No tags found")
		return
	}

	lines := strings.Split(trimmed, "\n")

	// Max 15 tags
	if len(lines) > 15 {
		lines = lines[:15]
	}

	// Parse and print each tag
	for _, line := range lines {
		parts := strings.SplitN(line, "|||", 4)
		if len(parts) != 4 {
			continue
		}

		tagName := parts[0]
		hash := parts[1]
		age := tagCompressAge(parts[2])
		subject := parts[3]

		fmt.Printf("  %-10s %-8s %-10s \"%s\"\n", tagName, hash, age, subject)
	}

	fmt.Printf("\n  %d tags | Token cost: ~%d\n", len(lines), len(lines)*6)
}

func tagCompressAge(age string) string {
	age = strings.TrimSpace(age)

	if strings.Contains(age, "day") {
		parts := strings.Fields(age)
		if len(parts) >= 2 {
			return parts[0] + "d ago"
		}
	}
	if strings.Contains(age, "month") {
		parts := strings.Fields(age)
		if len(parts) >= 2 {
			return parts[0] + "mo ago"
		}
	}
	if strings.Contains(age, "year") {
		parts := strings.Fields(age)
		if len(parts) >= 2 {
			return parts[0] + "y ago"
		}
	}
	if strings.Contains(age, "week") {
		parts := strings.Fields(age)
		if len(parts) >= 2 {
			return parts[0] + "w ago"
		}
	}
	if strings.Contains(age, "hour") {
		parts := strings.Fields(age)
		if len(parts) >= 2 {
			return parts[0] + "h ago"
		}
	}
	if strings.Contains(age, "min") {
		parts := strings.Fields(age)
		if len(parts) >= 2 {
			return parts[0] + "m ago"
		}
	}

	return age
}

func tagCreate() {
	var tagName string
	var message string

	// Determine if we're in shorthand mode or subcommand mode
	if len(os.Args) >= 3 {
		arg := os.Args[2]
		if arg == "create" || arg == "new" {
			// Subcommand mode: gg tag create <name> [message]
			if len(os.Args) >= 4 {
				tagName = os.Args[3]
				if len(os.Args) >= 5 {
					message = os.Args[4]
				}
			}
		} else {
			// Shorthand mode: gg tag <name> [message]
			tagName = arg
			if len(os.Args) >= 4 {
				message = os.Args[3]
			}
		}
	}

	if tagName == "" {
		printTagUsage()
		return
	}

	var cmd *exec.Cmd
	if message != "" {
		cmd = exec.Command("git", "tag", "-a", tagName, "-m", message)
	} else {
		cmd = exec.Command("git", "tag", tagName)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ tag create failed: %v\n  %s\n", sanitizeError(err), strings.TrimSpace(string(output)))
		return
	}

	// Get the short hash of the tagged commit
	hashCmd := exec.Command("git", "rev-list", "-n", "1", tagName)
	hashOutput, err := hashCmd.Output()
	if err != nil {
		fmt.Printf("  ✓ tagged as %s\n", tagName)
		return
	}

	hash := strings.TrimSpace(string(hashOutput))
	if len(hash) > 7 {
		hash = hash[:7]
	}

	fmt.Printf("  ✓ tagged %s as %s\n", hash, tagName)
}

func tagDelete() {
	var tagName string

	// Determine if we're in shorthand mode or subcommand mode
	if len(os.Args) >= 3 {
		arg := os.Args[2]
		if arg == "delete" || arg == "rm" {
			// Subcommand mode: gg tag delete <name>
			if len(os.Args) >= 4 {
				tagName = os.Args[3]
			}
		} else {
			// This shouldn't happen since shorthand maps to create, but handle it
			tagName = arg
		}
	}

	if tagName == "" {
		printTagUsage()
		return
	}

	cmd := exec.Command("git", "tag", "-d", tagName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ tag delete failed: %v\n  %s\n", sanitizeError(err), strings.TrimSpace(string(output)))
		return
	}

	fmt.Printf("  ✓ deleted tag %s\n", tagName)
}
