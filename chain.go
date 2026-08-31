package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func handleChain() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg chain <tool:pkg> [tool:pkg...]")
		fmt.Println("       gg chain --save <name> <tool:pkg> [tool:pkg...]")
		fmt.Println("       gg chain run <name>")
		fmt.Println("       gg chain <saved-name>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg chain npm:prettier npm:eslint brew:jq")
		fmt.Println("  gg chain --save webformat npm:prettier npm:eslint")
		fmt.Println("  gg chain run webformat")
		return
	}

	args := os.Args[2:]

	// Check for run subcommand
	if args[0] == "run" {
		if len(args) < 2 {
			fmt.Println("Usage: gg chain run <name>")
			return
		}
		runChain(args[1])
		return
	}

	// Check for --save flag
	if args[0] == "--save" {
		if len(args) < 3 {
			fmt.Println("Usage: gg chain --save <name> <tool:pkg>...")
			return
		}
		chainName := args[1]
		tools := args[2:]
		saveChain(chainName, tools)
		fmt.Printf("Saved chain '%s' with %d tools\n", chainName, len(tools))
		return
	}

	// Check for --list flag
	if args[0] == "--list" {
		listSavedChains()
		return
	}

	// Check if first arg is a saved chain name
	if !strings.Contains(args[0], ":") {
		tools := loadChain(args[0])
		if tools != nil {
			fmt.Printf("Running saved chain '%s'\n\n", args[0])
			args = tools
		} else {
			fmt.Printf("Unknown chain: %s\n", args[0])
			fmt.Println("Run 'gg chain --list' to see saved chains")
			return
		}
	}

	fmt.Printf("Chained %d MCPs:\n", len(args))
	totalCost := 0

	for i, tool := range args {
		parts := strings.SplitN(tool, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("   %d. Invalid format: %s (expected type:name)\n", i+1, tool)
			continue
		}

		toolType := parts[0]
		toolName := parts[1]

		cost := getTokenCost(toolType)
		fmt.Printf("   %d. %s:%s (~%d tokens)\n", i+1, toolType, toolName, cost)
		totalCost += cost
	}

	fmt.Printf("\nCombined token cost: ~%d\n", totalCost)
}

func getTokenCost(toolType string) int {
	switch toolType {
	case "npm":
		return TokenCostNPM
	case "brew":
		return TokenCostBrew
	case "git":
		return TokenCostGit
	default:
		return TokenCostBase
	}
}

func saveChain(name string, tools []string) {
	chainDir := filepath.Join(getGGDir(), "chains")
	os.MkdirAll(chainDir, 0755)

	data, _ := json.Marshal(tools)
	os.WriteFile(filepath.Join(chainDir, name+".json"), data, 0644)
}

func loadChain(name string) []string {
	chainPath := filepath.Join(getGGDir(), "chains", name+".json")
	data, err := os.ReadFile(chainPath)
	if err != nil {
		return nil
	}

	var tools []string
	json.Unmarshal(data, &tools)
	return tools
}

func listSavedChains() {
	chainDir := filepath.Join(getGGDir(), "chains")
	files, err := os.ReadDir(chainDir)
	if err != nil || len(files) == 0 {
		fmt.Println("No saved chains")
		fmt.Println("Create one: gg chain --save <name> <tool:pkg>...")
		return
	}

	fmt.Println("Saved chains:")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			name := strings.TrimSuffix(f.Name(), ".json")
			tools := loadChain(name)
			fmt.Printf("   - %s (%d tools)\n", name, len(tools))
		}
	}
}

// Curated toolbelts
var toolbelts = map[string][]string{
	"webdev": {
		"npm:eslint",
		"npm:prettier",
		"npm:typescript",
		"npm:jest",
		"npm:playwright",
	},
	"media": {
		"brew:ffmpeg",
		"brew:imagemagick",
		"brew:exiftool",
	},
	"sec": {
		"brew:semgrep",
		"npm:snyk",
		"brew:trivy",
	},
	"data": {
		"brew:duckdb",
		"brew:jq",
		"npm:csvtojson",
	},
	"devops": {
		"brew:terraform",
		"brew:kubectl",
		"brew:docker",
	},
}

// handleCool displays curated toolbelts
func handleCool() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg cool <toolbelt>")
		fmt.Println("       gg cool --list")
		fmt.Println()
		fmt.Println("Available toolbelts: webdev, media, sec, data, devops")
		return
	}

	arg := os.Args[2]

	if arg == "--list" {
		fmt.Println("Available toolbelts:")
		fmt.Println()
		for name, tools := range toolbelts {
			fmt.Printf("   %s (%d tools)\n", name, len(tools))
			for _, tool := range tools {
				parts := strings.SplitN(tool, ":", 2)
				fmt.Printf("      - %s (%s)\n", parts[1], parts[0])
			}
			fmt.Println()
		}
		return
	}

	tools, ok := toolbelts[arg]
	if !ok {
		fmt.Printf("Unknown toolbelt: %s\n", arg)
		fmt.Println("Run 'gg cool --list' to see available toolbelts")
		return
	}

	fmt.Printf("Toolbelt: %s\n\n", arg)
	totalCost := 0

	for _, tool := range tools {
		parts := strings.SplitN(tool, ":", 2)
		toolType := parts[0]
		toolName := parts[1]

		cost := getTokenCost(toolType)
		fmt.Printf("   - %s (%s)\n", toolName, toolType)
		totalCost += cost
	}

	fmt.Printf("\nCombined token cost: ~%d\n", totalCost)
	fmt.Printf("\nChain all: gg chain %s\n", strings.Join(tools, " "))
}

// runChain executes all tools in a saved chain
func runChain(name string) {
	tools := loadChain(name)
	if tools == nil {
		fmt.Printf("Chain not found: %s\n", name)
		fmt.Println("Run 'gg chain --list' to see saved chains")
		return
	}

	fmt.Printf("Executing chain '%s'...\n\n", name)

	success := 0
	for i, tool := range tools {
		parts := strings.SplitN(tool, ":", 2)
		if len(parts) != 2 {
			fmt.Printf("[%d/%d] Invalid: %s\n", i+1, len(tools), tool)
			continue
		}

		toolType := parts[0]
		toolName := parts[1]

		fmt.Printf("[%d/%d] %s:%s\n", i+1, len(tools), toolType, toolName)

		switch toolType {
		case "npm":
			runNPMCheck(toolName)
		case "brew":
			runBrewCheck(toolName)
		default:
			fmt.Printf("   Unknown type: %s\n", toolType)
		}

		success++
		fmt.Println()
	}

	fmt.Printf("Chain complete: %d/%d tools ready\n", success, len(tools))
}

func runNPMCheck(pkg string) {
	cacheDir := filepath.Join(getGGDir(), "cache", "npm")
	cachePath := filepath.Join(cacheDir, pkg+".json")

	if _, err := os.ReadFile(cachePath); err == nil {
		fmt.Printf("   %s (cached)\n", pkg)
		return
	}

	url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", pkg)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("   %s (error)\n", pkg)
		return
	}
	defer resp.Body.Close()

	var info map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&info)

	os.MkdirAll(cacheDir, 0755)
	data, _ := json.Marshal(info)
	os.WriteFile(cachePath, data, 0644)

	pkgVersion, _ := info["version"].(string)
	fmt.Printf("   %s@%s\n", pkg, pkgVersion)
}

func runBrewCheck(formula string) {
	cmd := exec.Command("brew", "info", formula, "--json=v2")
	if err := cmd.Run(); err == nil {
		fmt.Printf("   %s (installed)\n", formula)
	} else {
		fmt.Printf("   %s (not installed)\n", formula)
	}
}

// handleCache manages the gg cache
