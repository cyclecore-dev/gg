package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func handlePip() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg pip <package>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg pip requests")
		fmt.Println("  gg pip numpy")
		return
	}

	pkg := os.Args[2]

	// Check local cache first
	cacheDir := filepath.Join(getGGDir(), "cache", "pip")
	cachePath := filepath.Join(cacheDir, pkg+".json")

	var pkgInfo map[string]interface{}

	if data, err := os.ReadFile(cachePath); err == nil {
		// Cache hit
		json.Unmarshal(data, &pkgInfo)
		fmt.Printf("%s (cached)\n", pkg)
	} else {
		// Fetch from PyPI registry
		fmt.Printf("Fetching %s from PyPI...\n", pkg)
		url := fmt.Sprintf("https://pypi.org/pypi/%s/json", pkg)
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
			fmt.Printf("PyPI error: %d\n", resp.StatusCode)
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

	// Extract info from nested structure
	info, ok := pkgInfo["info"].(map[string]interface{})
	if !ok {
		fmt.Println("Invalid response format")
		return
	}

	name, _ := info["name"].(string)
	pkgVersion, _ := info["version"].(string)
	summary, _ := info["summary"].(string)
	author, _ := info["author"].(string)

	fmt.Printf("\n%s==%s\n", name, pkgVersion)
	if summary != "" {
		fmt.Printf("   %s\n", summary)
	}
	if author != "" {
		fmt.Printf("   author: %s\n", author)
	}

	fmt.Printf("\nMCP Endpoint: pip:%s\n", name)
	fmt.Printf("Token cost: ~%d\n", TokenCostNPM) // Same token cost as npm
}

// handleBrew fetches Homebrew formula info and displays MCP endpoint
