package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func handleHex() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printHexUsage()
		return
	}

	packageName := os.Args[2]
	cachePath := filepath.Join(getGGDir(), "cache", "hex", packageName+".json")

	// Check cache
	if data, err := os.ReadFile(cachePath); err == nil {
		var pkgInfo map[string]interface{}
		if json.Unmarshal(data, &pkgInfo) == nil {
			displayHexInfo(pkgInfo, packageName)
			return
		}
	}

	// Fetch
	url := fmt.Sprintf("https://hex.pm/api/packages/%s", packageName)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Fprintf(os.Stderr, "  Error: package %s not found\n", packageName)
		return
	}

	var pkgInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pkgInfo); err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		return
	}

	// Save cache
	os.MkdirAll(filepath.Dir(cachePath), 0755)
	jsonData, _ := json.Marshal(pkgInfo)
	os.WriteFile(cachePath, jsonData, 0644)

	displayHexInfo(pkgInfo, packageName)
}

func printHexUsage() {
	fmt.Println("Usage: gg hex <package>")
	fmt.Println("  Fetches package info from Hex.pm")
}

func displayHexInfo(pkgInfo map[string]interface{}, name string) {
	// Parse nested structure
	meta, _ := pkgInfo["meta"].(map[string]interface{})
	description := ""
	if meta != nil {
		description, _ = meta["description"].(string)
	}

	downloads, _ := pkgInfo["downloads"].(map[string]interface{})
	allDownloads := 0.0
	if downloads != nil {
		allDownloads, _ = downloads["all"].(float64)
	}

	releases, _ := pkgInfo["releases"].([]interface{})
	latestVersion := ""
	if len(releases) > 0 {
		rel, _ := releases[0].(map[string]interface{})
		latestVersion, _ = rel["version"].(string)
	}

	licenses := []string{}
	if meta != nil {
		if licArr, ok := meta["licenses"].([]interface{}); ok {
			for _, l := range licArr {
				if s, ok := l.(string); ok {
					licenses = append(licenses, s)
				}
			}
		}
	}
	licenseStr := ""
	if len(licenses) > 0 {
		for i, l := range licenses {
			if i > 0 {
				licenseStr += ", "
			}
			licenseStr += l
		}
	}

	releaseCount := len(releases)
	downloadStr := hexFmtDownloads(int64(allDownloads))

	fmt.Printf("  %s@%s\n", name, latestVersion)
	fmt.Printf("    %s\n", description)
	fmt.Printf("    License: %s | ↓ %s downloads | %d releases\n", licenseStr, downloadStr, releaseCount)
	fmt.Printf("  MCP Endpoint: hex:%s\n", name)
	fmt.Printf("  Token cost: ~18\n")
}

func hexFmtDownloads(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%dK", n/1000)
	}
	if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000.0)
	}
	return fmt.Sprintf("%.1fB", float64(n)/1000000000.0)
}
