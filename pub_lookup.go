package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handlePub() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printPubUsage()
		return
	}

	packageName := os.Args[2]
	cacheDir := filepath.Join(getGGDir(), "cache", "pub")
	cachePath := filepath.Join(cacheDir, packageName+".json")

	// Check cache
	if data, err := os.ReadFile(cachePath); err == nil {
		var cached map[string]interface{}
		if json.Unmarshal(data, &cached) == nil {
			displayPubInfo(cached, packageName)
			return
		}
	}

	// Fetch from pub.dev
	url := "https://pub.dev/api/packages/" + packageName
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching package info:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Printf("Package '%s' not found on pub.dev\n", packageName)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Error: HTTP %d fetching package info\n", resp.StatusCode)
		return
	}

	var pkgInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pkgInfo); err != nil {
		fmt.Println("Error parsing response:", err)
		return
	}

	// Cache the result
	os.MkdirAll(cacheDir, 0755)
	if data, err := json.Marshal(pkgInfo); err == nil {
		os.WriteFile(cachePath, data, 0644)
	}

	displayPubInfo(pkgInfo, packageName)
}

func displayPubInfo(pkgInfo map[string]interface{}, packageName string) {
	latest, _ := pkgInfo["latest"].(map[string]interface{})
	if latest == nil {
		fmt.Println("Package data unavailable")
		return
	}

	version, _ := latest["version"].(string)
	pubspec, _ := latest["pubspec"].(map[string]interface{})
	if pubspec == nil {
		fmt.Println("Package data unavailable")
		return
	}

	description, _ := pubspec["description"].(string)
	description = truncate(description, 70)

	deps, _ := pubspec["dependencies"].(map[string]interface{})
	depCount := len(deps)

	env, _ := pubspec["environment"].(map[string]interface{})
	sdkConstraint := ""
	if env != nil {
		sdkConstraint, _ = env["sdk"].(string)
		// Extract first constraint (e.g., ">=2.14.0 <4.0.0" → ">=2.14.0")
		parts := strings.Fields(sdkConstraint)
		if len(parts) > 0 {
			sdkConstraint = parts[0]
		}
	}

	fmt.Printf("  %s@%s\n", packageName, version)
	fmt.Printf("    %s\n", description)
	fmt.Printf("    Dart %s | %d deps\n", sdkConstraint, depCount)
	fmt.Printf("  MCP Endpoint: pub:%s\n", packageName)
	fmt.Println("  Token cost: ~18")
}

func printPubUsage() {
	fmt.Println("Usage: gg pub <package>")
	fmt.Println("  Fetches pub.dev package info for Dart/Flutter packages")
	fmt.Println("  Example: gg pub flutter_bloc")
}
