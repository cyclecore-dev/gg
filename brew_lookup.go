package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func handleBrew() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg brew [-i] <formula>")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -i  Auto-install formula if not installed")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg brew ffmpeg")
		fmt.Println("  gg brew -i jq")
		return
	}

	// Parse -i flag
	autoInstall := false
	formula := ""
	for _, arg := range os.Args[2:] {
		if arg == "-i" {
			autoInstall = true
		} else {
			formula = arg
		}
	}

	if formula == "" {
		fmt.Println("No formula specified")
		return
	}

	// Check local cache first
	cacheDir := filepath.Join(getGGDir(), "cache", "brew")
	cachePath := filepath.Join(cacheDir, formula+".json")

	var info map[string]interface{}
	installed := false

	// Try local brew first
	cmd := exec.Command("brew", "info", formula, "--json=v2")
	output, err := cmd.Output()
	if err == nil {
		var brewInfo map[string]interface{}
		if json.Unmarshal(output, &brewInfo) == nil {
			if formulae, ok := brewInfo["formulae"].([]interface{}); ok && len(formulae) > 0 {
				info = formulae[0].(map[string]interface{})
				installed = true
			}
		}
	}

	// Fall back to API
	if info == nil {
		// Check cache
		if data, err := os.ReadFile(cachePath); err == nil {
			json.Unmarshal(data, &info)
			fmt.Printf("%s (cached)\n", formula)
		} else {
			fmt.Printf("Fetching %s from Homebrew...\n", formula)
			url := fmt.Sprintf("https://formulae.brew.sh/api/formula/%s.json", formula)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Failed to fetch formula: %v\n", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 404 {
				fmt.Printf("Formula not found: %s\n", formula)
				return
			}

			if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
				fmt.Printf("Failed to parse response: %v\n", err)
				return
			}

			// Cache it
			os.MkdirAll(cacheDir, 0755)
			data, _ := json.Marshal(info)
			os.WriteFile(cachePath, data, 0644)
		}
	}

	// Display info
	name, _ := info["name"].(string)
	desc, _ := info["desc"].(string)

	var formulaVersion string
	if versions, ok := info["versions"].(map[string]interface{}); ok {
		formulaVersion, _ = versions["stable"].(string)
	}

	// Auto-install if -i flag and not installed
	if !installed && autoInstall {
		fmt.Printf("Installing %s...\n", formula)
		installCmd := exec.Command("brew", "install", formula)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			fmt.Printf("Install failed: %v\n", err)
			return
		}
		fmt.Printf("%s installed\n", formula)
		installed = true
	}

	statusIcon := "installed"
	if !installed {
		statusIcon = "not installed"
	}

	fmt.Printf("\n%s@%s (%s)\n", name, formulaVersion, statusIcon)
	if desc != "" {
		fmt.Printf("   %s\n", desc)
	}

	if !installed && !autoInstall {
		fmt.Printf("\n   Install: brew install %s\n", formula)
		fmt.Println("   Or use: gg brew -i", formula)
	}

	fmt.Printf("\nMCP Endpoint: brew:%s\n", formula)
	fmt.Printf("Token cost: ~%d\n", TokenCostBrew)
}

// handleChain chains multiple MCP tools together
