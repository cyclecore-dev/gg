package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func handleEdit() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg edit <file> [instruction]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg edit main.go                    # Interactive edit")
		fmt.Println("  gg edit main.go \"add error handling\"")
		fmt.Println("  gg edit src/*.ts \"add types\"       # Multiple files")
		return
	}

	filePath := os.Args[2]

	cwd, _ := os.Getwd()
	cleaned := filepath.Clean(filePath)
	if filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "..") {
		fatalError("Path outside repository: "+filePath, nil)
	}
	full := filepath.Join(cwd, cleaned)
	if !strings.HasPrefix(full, cwd) {
		fatalError("Path outside repository: "+filePath, nil)
	}
	filePath = cleaned

	var instruction string
	if len(os.Args) > 3 {
		instruction = strings.Join(os.Args[3:], " ")
	}

	// Check if file exists
	content, err := os.ReadFile(filePath)
	if err != nil {
		fatalError("Cannot read file", err)
	}

	// Get instruction if not provided
	if instruction == "" {
		fmt.Printf("Editing: %s (%d bytes)\n", filePath, len(content))
		fmt.Print("\nWhat changes do you want? > ")
		reader := bufio.NewReader(os.Stdin)
		instruction, _ = reader.ReadString('\n')
		instruction = strings.TrimSpace(instruction)
		if instruction == "" {
			fmt.Println("No instruction provided, exiting")
			return
		}
	}

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		fatalError("Config error. Run: gg config init", err)
	}

	provider, model, endpoint, apiKey := getEffectiveConfig(cfg)
	if provider != ProviderOllama && apiKey == "" {
		fatalError("API key not configured. Run: gg config init", nil)
	}

	// Create edit prompt
	editPrompt := fmt.Sprintf("Edit this file according to the instruction.\n\n"+
		"FILE: %s\n"+
		"```\n%s\n```\n\n"+
		"INSTRUCTION: %s\n\n"+
		"Return ONLY the complete edited file content, no explanations. "+
		"Wrap in ```language:filename code block.", filePath, string(content), instruction)

	systemPrompt := "You are a code editor. Return only the edited file content in a code block."

	fmt.Printf("Editing %s with %s/%s...\n\n", filePath, provider, model)

	var response string
	switch provider {
	case ProviderOpenAI:
		response, err = callOpenAIStreaming(apiKey, model, systemPrompt, editPrompt)
	case ProviderOllama:
		response, err = callOllamaStreaming(endpoint, model, systemPrompt, editPrompt)
	default:
		response, err = callAnthropicStreaming(apiKey, model, systemPrompt, editPrompt, cfg.API.Temperature)
	}

	if err != nil {
		fatalError("API error", sanitizeError(err))
	}

	// Parse the response for code blocks
	files := parseCodeBlocks(response)
	if len(files) == 0 {
		// Try to extract content between ``` markers
		re := regexp.MustCompile("```[a-z]*\n([\\s\\S]*?)```")
		matches := re.FindStringSubmatch(response)
		if len(matches) >= 2 {
			files[filePath] = matches[1]
		}
	}

	if len(files) == 0 {
		fmt.Println("\nNo code block found in response. Raw response:")
		fmt.Println(response)
		return
	}

	// Get the edited content
	var newContent string
	for _, c := range files {
		newContent = c
		break
	}

	// Show diff summary
	oldLines := strings.Count(string(content), "\n")
	newLines := strings.Count(newContent, "\n")
	fmt.Printf("\nChanges: %d lines -> %d lines\n", oldLines, newLines)

	// Confirm
	fmt.Print("\nApply changes? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "" && confirm != "y" {
		fmt.Println("Cancelled")
		return
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		fatalError("Failed to write file", err)
	}

	fmt.Printf("Updated: %s\n", filePath)
}

// ============================================================================
// PROMPTS COMMAND
// ============================================================================
