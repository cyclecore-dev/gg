package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func handleCrate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg crate <name>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  gg crate serde")
		fmt.Println("  gg crate tokio")
		fmt.Println("  gg crate clap")
		return
	}

	pkg := os.Args[2]
	cachePath := filepath.Join(getGGDir(), "cache", "crate", pkg+".json")

	if data, err := os.ReadFile(cachePath); err == nil {
		var info map[string]interface{}
		if err := json.Unmarshal(data, &info); err == nil {
			fmt.Println("(cached)")
			printCrateInfo(info)
			return
		}
	}

	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", pkg)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("User-Agent", "gg-cli/0.9 (https://github.com/cyclecore-dev/gg)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error fetching crate: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Crate not found: %s\n", pkg)
		return
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		fmt.Printf("Error creating cache directory: %v\n", err)
		return
	}
	if err := os.WriteFile(cachePath, func() []byte { b, _ := json.Marshal(response); return b }(), 0644); err != nil {
		fmt.Printf("Error writing cache: %v\n", err)
		return
	}

	printCrateInfo(response)
}

func printCrateInfo(info map[string]interface{}) {
	crate, ok := info["crate"].(map[string]interface{})
	if !ok {
		return
	}

	name, _ := crate["name"].(string)
	maxVersion, _ := crate["max_version"].(string)
	description, _ := crate["description"].(string)
	repository, _ := crate["repository"].(string)

	fmt.Printf("%s@%s\n", name, maxVersion)
	if description != "" {
		fmt.Printf("   %s\n", description)
	}
	if repository != "" {
		fmt.Printf("   repo: %s\n", repository)
	}
	fmt.Println()
	fmt.Printf("MCP Endpoint: crate:%s\n", name)
	fmt.Printf("Token cost: ~%d\n", TokenCostNPM)
}
