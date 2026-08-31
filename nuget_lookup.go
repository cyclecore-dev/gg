package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func nugetFmtDownloads(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1_000_000 {
		return fmt.Sprintf("%dK", n/1000)
	} else if n < 1_000_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
}

func printNugetUsage() {
	fmt.Println("Usage: gg nuget <package>")
	fmt.Println("  package: NuGet package name (case-insensitive)")
}

func handleNuget() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printNugetUsage()
		return
	}

	pkgName := strings.ToLower(os.Args[2])

	cacheDir := filepath.Join(getGGDir(), "cache", "nuget")
	cachePath := filepath.Join(cacheDir, pkgName+".json")

	if data, err := os.ReadFile(cachePath); err == nil {
		nugetDisplay(data)
		return
	}

	url := fmt.Sprintf("https://azuresearch-usnc.nuget.org/query?q=%s&take=1", pkgName)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching package: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	dataArr, ok := result["data"].([]interface{})
	if !ok || len(dataArr) == 0 {
		fmt.Println("Package not found")
		return
	}

	pkg, ok := dataArr[0].(map[string]interface{})
	if !ok {
		fmt.Println("Package not found")
		return
	}

	cached, err := json.Marshal(pkg)
	if err == nil {
		os.MkdirAll(cacheDir, 0755)
		os.WriteFile(cachePath, cached, 0644)
	}

	nugetDisplay(cached)
}

func nugetDisplay(data []byte) {
	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		fmt.Printf("Error parsing package info: %v\n", err)
		return
	}

	id, _ := pkg["id"].(string)
	version, _ := pkg["version"].(string)
	description, _ := pkg["description"].(string)
	totalDownloads, _ := pkg["totalDownloads"].(float64)
	authors, _ := pkg["authors"].([]interface{})

	var authorParts []string
	for _, a := range authors {
		if s, ok := a.(string); ok {
			authorParts = append(authorParts, s)
		}
	}
	authorStr := strings.Join(authorParts, ", ")

	fmt.Printf("  %s@%s\n", id, version)
	fmt.Printf("    %s\n", description)
	fmt.Printf("    By %s | ↓ %s downloads\n", authorStr, nugetFmtDownloads(int64(totalDownloads)))
	fmt.Printf("  MCP Endpoint: nuget:%s\n", id)
	fmt.Println("  Token cost: ~18")
}
