package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handlePrompts() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg prompts <command>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  list               # List saved prompts")
		fmt.Println("  add <name> \"...\"   # Save a prompt")
		fmt.Println("  run <name>         # Run a saved prompt with gg ask")
		fmt.Println("  delete <name>      # Delete a prompt")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg prompts add refactor \"refactor this code for clarity\"")
		fmt.Println("  gg prompts run refactor")
		return
	}

	subCmd := os.Args[2]

	switch subCmd {
	case "list", "ls":
		listPrompts()
	case "add", "save":
		if len(os.Args) < 5 {
			fmt.Println("Usage: gg prompts add <name> \"prompt text\"")
			return
		}
		name := os.Args[3]
		promptText := strings.Join(os.Args[4:], " ")
		savePrompt(name, promptText)
	case "run", "use":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gg prompts run <name>")
			return
		}
		usePrompt(os.Args[3])
	case "delete", "rm":
		if len(os.Args) < 4 {
			fmt.Println("Usage: gg prompts delete <name>")
			return
		}
		deletePrompt(os.Args[3])
	default:
		// Try to run it as a prompt name
		usePrompt(subCmd)
	}
}

func getPromptsDir() string {
	return filepath.Join(getGGDir(), "prompts")
}

func listPrompts() {
	promptsDir := getPromptsDir()
	files, err := os.ReadDir(promptsDir)
	if err != nil || len(files) == 0 {
		fmt.Println("No saved prompts")
		fmt.Println("Save one: gg prompts save <name> \"prompt text\"")
		return
	}

	fmt.Println("Saved prompts:")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".txt") {
			name := strings.TrimSuffix(f.Name(), ".txt")
			content, _ := os.ReadFile(filepath.Join(promptsDir, f.Name()))
			preview := string(content)
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			fmt.Printf("  - %s: %s\n", name, preview)
		}
	}
}

func savePrompt(name, text string) {
	promptsDir := getPromptsDir()
	os.MkdirAll(promptsDir, 0755)

	path := filepath.Join(promptsDir, name+".txt")
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		fatalError("Failed to save prompt", err)
	}
	fmt.Printf("Saved prompt '%s'\n", name)
}

func usePrompt(name string) {
	promptsDir := getPromptsDir()
	path := filepath.Join(promptsDir, name+".txt")

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Prompt not found: %s\n", name)
		fmt.Println("Run 'gg prompts list' to see saved prompts")
		return
	}

	promptText := string(content)
	fmt.Printf("Using prompt '%s': %s\n\n", name, promptText)

	// Inject prompt into handleAsk by setting os.Args
	os.Args = []string{"gg", "ask", promptText}
	handleAsk()
}

func deletePrompt(name string) {
	promptsDir := getPromptsDir()
	path := filepath.Join(promptsDir, name+".txt")

	if err := os.Remove(path); err != nil {
		fmt.Printf("Prompt not found: %s\n", name)
		return
	}
	fmt.Printf("Deleted prompt '%s'\n", name)
}

// ============================================================================
// UTILITIES
// ============================================================================
