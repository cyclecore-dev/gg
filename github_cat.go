package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func handleCat() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		fmt.Println("Usage: gg cat <user/repo/path>")
		fmt.Println("\nExamples:")
		fmt.Println("  gg cat anthropics/anthropic-sdk-python/README.md")
		fmt.Println("  gg cat golang/go/src/fmt/print.go")
		return
	}

	fullPath := os.Args[2]
	parts := strings.Split(fullPath, "/")

	if len(parts) < 3 {
		fmt.Println("Usage: gg cat <user/repo/path>")
		return
	}

	owner := parts[0]
	repo := parts[1]
	path := strings.Join(parts[2:], "/")

	token := getGHToken()
	client := &http.Client{}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path), nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		fmt.Printf("File not found: %s in %s/%s\n", path, owner, repo)
		return
	}
	if resp.StatusCode == 403 {
		fmt.Println("GitHub API rate limited. Run 'gh auth login' for higher limits.")
		return
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Error: %s\n", resp.Status)
		return
	}

	var content struct {
		Content     string `json:"content"`
		Encoding    string `json:"encoding"`
		Name        string `json:"name"`
		Size        int    `json:"size"`
		DownloadURL string `json:"download_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	if content.Content == "" && content.DownloadURL != "" {
		fmt.Printf("File too large for API. Use: curl -L %s\n", content.DownloadURL)
		return
	}

	cleaned := strings.ReplaceAll(content.Content, "\n", "")
	decoded, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		fmt.Printf("Error decoding content: %v\n", err)
		return
	}

	fmt.Printf("%s/%s/%s (%d bytes)\n", owner, repo, path, content.Size)
	fmt.Println("---")
	fmt.Print(string(decoded))
	if len(decoded) > 0 && decoded[len(decoded)-1] != '\n' {
		fmt.Println()
	}
	fmt.Printf("\nToken cost: ~%d\n", len(decoded)/4)
}
