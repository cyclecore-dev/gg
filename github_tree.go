package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func handleTree() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		fmt.Println("Usage: gg tree <user/repo> [path]")
		fmt.Println("\nExamples:")
		fmt.Println("  gg tree anthropics/anthropic-sdk-python")
		fmt.Println("  gg tree golang/go src/fmt")
		return
	}

	repoArg := os.Args[2]
	pathFilter := ""
	if len(os.Args) > 3 {
		pathFilter = os.Args[3]
	}

	parts := strings.Split(repoArg, "/")
	if len(parts) != 2 {
		fmt.Println("Error: user/repo must have exactly two segments")
		return
	}
	owner, repo := parts[0], parts[1]

	token := getGHToken()
	client := &http.Client{}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo), nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching repo info: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		fmt.Printf("Repo not found: %s/%s\n", owner, repo)
		return
	}
	if resp.StatusCode == 403 {
		fmt.Println("GitHub API rate limited. Run 'gh auth login' for higher limits.")
		return
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Error fetching repo info (HTTP %d)\n", resp.StatusCode)
		return
	}

	var repoInfo struct {
		DefaultBranch string `json:"default_branch"`
	}
	json.NewDecoder(resp.Body).Decode(&repoInfo)
	branch := repoInfo.DefaultBranch

	treeReq, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, branch), nil)
	if token != "" {
		treeReq.Header.Set("Authorization", "Bearer "+token)
	}
	treeResp, err := client.Do(treeReq)
	if err != nil {
		fmt.Printf("Error fetching tree: %v\n", err)
		return
	}
	defer treeResp.Body.Close()
	if treeResp.StatusCode != 200 {
		fmt.Printf("Error fetching tree (HTTP %d)\n", treeResp.StatusCode)
		return
	}

	var treeData struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
			Size int    `json:"size"`
		} `json:"tree"`
	}
	json.NewDecoder(treeResp.Body).Decode(&treeData)

	sizeMap := make(map[string]int)
	var filtered []struct {
		path string
		size int
	}
	for _, item := range treeData.Tree {
		if item.Type == "blob" {
			if pathFilter != "" && !strings.HasPrefix(item.Path, pathFilter+"/") && item.Path != pathFilter {
				continue
			}
			filtered = append(filtered, struct {
				path string
				size int
			}{item.Path, item.Size})
			sizeMap[item.Path] = item.Size
		}
	}

	type entry struct {
		path    string
		display string
		depth   int
		isDir   bool
		size    int
	}

	var entries []entry
	dirSet := make(map[string]bool)
	for _, f := range filtered {
		p := strings.Split(f.path, "/")
		for i := 1; i < len(p); i++ {
			dirSet[strings.Join(p[:i], "/")] = true
		}
	}

	for d := range dirSet {
		parts := strings.Split(d, "/")
		entries = append(entries, entry{
			path:    d,
			display: parts[len(parts)-1] + "/",
			depth:   len(parts),
			isDir:   true,
		})
	}
	for _, f := range filtered {
		parts := strings.Split(f.path, "/")
		entries = append(entries, entry{
			path:    f.path,
			display: parts[len(parts)-1],
			depth:   len(parts),
			isDir:   false,
			size:    f.size,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].path == entries[j].path {
			return entries[i].isDir
		}
		return entries[i].path < entries[j].path
	})

	fmt.Printf("%s/%s (%s) — %d files\n\n", owner, repo, branch, len(filtered))

	shown := 0
	for _, e := range entries {
		if shown >= 50 {
			break
		}
		indent := strings.Repeat("  ", e.depth-1)
		if e.isDir {
			fmt.Printf("%s%s\n", indent, e.display)
		} else if e.size > 0 {
			fmt.Printf("%s%s (%s)\n", indent, e.display, fmtFileSize(e.size))
		} else {
			fmt.Printf("%s%s\n", indent, e.display)
		}
		shown++
	}

	remaining := len(entries) - shown
	if remaining > 0 {
		fmt.Printf("  ... and %d more files\n", remaining)
	}

	fmt.Printf("\nToken cost: ~%d\n", len(filtered)*2)
}

func fmtFileSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fK", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1fM", float64(bytes)/(1024*1024))
}

func getGHToken() string {
	cmd := exec.Command("gh", "auth", "token")
	out, err := cmd.Output()
	if err == nil {
		t := strings.TrimSpace(string(out))
		if t != "" {
			return t
		}
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return t
	}
	return ""
}
