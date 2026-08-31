package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func printComposerUsage() {
	fmt.Println("Usage: gg composer <vendor/package>")
	fmt.Println("       gg composer --help")
}

func composerCountDeps(require map[string]interface{}) int {
	count := 0
	for key := range require {
		if key == "php" || key == "composer-runtime-api" || strings.HasPrefix(key, "ext-") {
			continue
		}
		count++
	}
	return count
}

func handleComposer() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printComposerUsage()
		return
	}

	packageName := os.Args[2]
	if !strings.Contains(packageName, "/") {
		fmt.Println("Usage: gg composer <vendor/package>")
		return
	}

	// Check cache
	escapedName := strings.ReplaceAll(packageName, "/", "__")
	cachePath := filepath.Join(getGGDir(), "cache", "composer", escapedName+".json")

	// Try to read from cache
	if data, err := os.ReadFile(cachePath); err == nil {
		var response map[string]interface{}
		if json.Unmarshal(data, &response) == nil {
			displayComposerResult(response)
			return
		}
	}

	// Fetch from API
	parts := strings.SplitN(packageName, "/", 2)
	if len(parts) != 2 {
		fmt.Println("Usage: gg composer <vendor/package>")
		return
	}

	url := fmt.Sprintf("https://repo.packagist.org/p2/%s/%s.json", parts[0], parts[1])

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching package: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Printf("  Package not found: %s\n", packageName)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Error: HTTP %d\n", resp.StatusCode)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	// Save to cache
	os.MkdirAll(filepath.Dir(cachePath), 0755)
	if data, err := json.Marshal(apiResponse); err == nil {
		os.WriteFile(cachePath, data, 0644)
	}

	displayComposerResult(apiResponse)
}

func displayComposerResult(response map[string]interface{}) {
	packages, ok := response["packages"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Invalid response format")
		return
	}

	// Find the package
	var latestVersion map[string]interface{}
	for pkgName, versions := range packages {
		if pkgName == os.Args[2] {
			versionList, ok := versions.([]interface{})
			if !ok || len(versionList) == 0 {
				fmt.Printf("  Package not found: %s\n", os.Args[2])
				return
			}
			latestVersion, ok = versionList[0].(map[string]interface{})
			if !ok {
				fmt.Println("Error: Invalid version data")
				return
			}
			break
		}
	}

	if latestVersion == nil {
		fmt.Printf("  Package not found: %s\n", os.Args[2])
		return
	}

	// Extract fields
	name, _ := latestVersion["name"].(string)
	version, _ := latestVersion["version"].(string)
	description, _ := latestVersion["description"].(string)

	// License
	licenseArr, _ := latestVersion["license"].([]interface{})
	var licenseStr string
	for i, lic := range licenseArr {
		if s, ok := lic.(string); ok {
			if i > 0 {
				licenseStr += ", "
			}
			licenseStr += s
		}
	}

	// Require
	require, _ := latestVersion["require"].(map[string]interface{})
	depCount := composerCountDeps(require)
	phpVersion, _ := require["php"].(string)

	// Display
	fmt.Printf("  %s@%s\n", name, version)
	fmt.Printf("    %s\n", description)
	fmt.Printf("    License: %s | %d deps | PHP %s\n", licenseStr, depCount, phpVersion)
	fmt.Printf("  MCP Endpoint: composer:%s\n", name)
	fmt.Printf("  Token cost: ~18\n")
}
