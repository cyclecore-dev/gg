package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func handleNPM() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg npm <package> [--fn <function>]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg npm prettier")
		fmt.Println("  gg npm lodash --fn debounce")
		return
	}

	pkg := os.Args[2]

	// Check local cache first
	cacheDir := filepath.Join(getGGDir(), "cache", "npm")
	cachePath := filepath.Join(cacheDir, pkg+".json")

	var pkgInfo map[string]interface{}

	if data, err := os.ReadFile(cachePath); err == nil {
		// Cache hit
		json.Unmarshal(data, &pkgInfo)
		fmt.Printf("%s (cached)\n", pkg)
	} else {
		// Fetch from npm registry
		fmt.Printf("Fetching %s from npm...\n", pkg)
		url := fmt.Sprintf("https://registry.npmjs.org/%s/latest", pkg)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Failed to fetch package: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			fmt.Printf("Package not found: %s\n", pkg)
			return
		}

		if resp.StatusCode != 200 {
			fmt.Printf("npm registry error: %d\n", resp.StatusCode)
			return
		}

		if err := json.NewDecoder(resp.Body).Decode(&pkgInfo); err != nil {
			fmt.Printf("Failed to parse response: %v\n", err)
			return
		}

		// Cache it
		os.MkdirAll(cacheDir, 0755)
		data, _ := json.Marshal(pkgInfo)
		os.WriteFile(cachePath, data, 0644)
	}

	// Display MCP format
	name, _ := pkgInfo["name"].(string)
	pkgVersion, _ := pkgInfo["version"].(string)
	desc, _ := pkgInfo["description"].(string)

	fmt.Printf("\n%s@%s\n", name, pkgVersion)
	if desc != "" {
		fmt.Printf("   %s\n", desc)
	}

	// Check for --fn flag
	for i, arg := range os.Args {
		if arg == "--fn" && i+1 < len(os.Args) {
			fnName := os.Args[i+1]
			fmt.Printf("\nFunction: %s\n", fnName)
		}
	}

	fmt.Printf("\nMCP Endpoint: npm:%s\n", name)
	fmt.Printf("Token cost: ~%d\n", TokenCostNPM)
}

// handlePip fetches PyPI package info and displays MCP endpoint
