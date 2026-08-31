package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func handleGem() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printGemUsage()
		return
	}

	packageName := os.Args[2]
	ggDir := getGGDir()

	cachePath := filepath.Join(ggDir, "cache", "gem", packageName+".json")

	// Check cache
	if data, err := os.ReadFile(cachePath); err == nil {
		var gemData map[string]interface{}
		if err := json.Unmarshal(data, &gemData); err == nil {
			gemDisplay(gemData)
			return
		}
	}

	// Fetch from API
	url := fmt.Sprintf("https://rubygems.org/api/v1/gems/%s.json", packageName)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Fprintf(os.Stderr, "Error: package %s not found\n", packageName)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Error: unexpected status %d\n", resp.StatusCode)
		return
	}

	var gemData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&gemData); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Cache response
	cacheDir := filepath.Join(ggDir, "cache", "gem")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	jsonData, err := json.Marshal(gemData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	if err := os.WriteFile(cachePath, jsonData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	gemDisplay(gemData)
}

func printGemUsage() {
	fmt.Println("Usage: gg gem <package>")
	fmt.Println("       gg gem --help")
}

func gemDisplay(data map[string]interface{}) {
	name, _ := data["name"].(string)
	version, _ := data["version"].(string)
	info, _ := data["info"].(string)
	licenses, _ := data["licenses"].([]interface{})
	downloads, _ := data["downloads"].(float64)

	licenseStr := ""
	for i, l := range licenses {
		if s, ok := l.(string); ok {
			if i > 0 {
				licenseStr += ", "
			}
			licenseStr += s
		}
	}

	fmt.Printf("  %s@%s\n", name, version)
	fmt.Printf("    %s\n", truncate(info, 80))
	fmt.Printf("    License: %s | ↓ %s downloads\n", licenseStr, gemFmtDownloads(downloads))
	fmt.Println("  Token cost: ~18")
}

func gemFmtDownloads(d float64) string {
	if d < 1000 {
		return fmt.Sprintf("%.0f", d)
	} else if d < 1000000 {
		return fmt.Sprintf("%.0fK", d/1000)
	} else if d < 1000000000 {
		return fmt.Sprintf("%.1fM", d/1000000)
	}
	return fmt.Sprintf("%.1fB", d/1000000000)
}
