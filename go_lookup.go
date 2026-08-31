package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handleGo() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg go <module>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg go github.com/gin-gonic/gin")
		fmt.Println("  gg go golang.org/x/crypto")
		return
	}

	module := os.Args[2]

	cacheDir := filepath.Join(getGGDir(), "cache", "go")
	escapedModule := strings.ReplaceAll(module, "/", "__")
	cachePath := filepath.Join(cacheDir, escapedModule+".json")

	var moduleInfo map[string]interface{}

	if data, err := os.ReadFile(cachePath); err == nil {
		json.Unmarshal(data, &moduleInfo)
		fmt.Printf("%s (cached)\n", module)
	} else {
		fmt.Printf("Fetching %s from Go Proxy...\n", module)
		url := fmt.Sprintf("https://proxy.golang.org/%s/@latest", module)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Failed to fetch module: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			fmt.Printf("Module not found: %s\n", module)
			return
		}

		if resp.StatusCode != 200 {
			fmt.Printf("Go proxy error: %d\n", resp.StatusCode)
			return
		}

		if err := json.NewDecoder(resp.Body).Decode(&moduleInfo); err != nil {
			fmt.Printf("Failed to parse response: %v\n", err)
			return
		}

		os.MkdirAll(cacheDir, 0755)
		data, _ := json.Marshal(moduleInfo)
		os.WriteFile(cachePath, data, 0644)
	}

	version, _ := moduleInfo["Version"].(string)
	timeStr, _ := moduleInfo["Time"].(string)

	fmt.Printf("\n%s@%s\n", module, version)
	if timeStr != "" {
		if len(timeStr) >= 10 {
			fmt.Printf("   %s (publish date)\n", timeStr[:10])
		} else {
			fmt.Printf("   %s (publish date)\n", timeStr)
		}
	}

	fmt.Printf("\nMCP Endpoint: go:%s\n", module)
	fmt.Printf("Token cost: ~%d\n", TokenCostNPM)
}
