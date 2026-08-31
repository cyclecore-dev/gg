package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DockerRepoInfo represents the Docker Hub repository endpoint response
type DockerRepoInfo struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	StarCount   int    `json:"star_count"`
	PullCount   int64  `json:"pull_count"`
	LastUpdated string `json:"last_updated"`
	HubUser     string `json:"hub_user"`
}

// DockerTagInfo represents a single tag in the tags endpoint response
type DockerTagInfo struct {
	Name        string `json:"name"`
	FullSize    int64  `json:"full_size"`
	LastUpdated string `json:"last_updated"`
}

// DockerTagsResponse represents the Docker Hub tags endpoint response
type DockerTagsResponse struct {
	Count   int           `json:"count"`
	Results []DockerTagInfo `json:"results"`
}

// DockerCache represents the cached docker image information
type DockerCache struct {
	Name        string
	Namespace   string
	Description string
	StarCount   int
	PullCount   int64
	Tags        []string
}

func handleDocker() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printDockerUsage()
		return
	}

	imageName := os.Args[2]

	// Determine API path based on whether image is official or user
	var repoPath string
	if strings.Contains(imageName, "/") {
		repoPath = imageName
	} else {
		repoPath = "library/" + imageName
	}

	// Check cache first
	cachePath := dockerGetCachePath(imageName)
	cache, err := dockerLoadCache(cachePath)
	if err == nil && cache != nil {
		displayDockerInfo(cache)
		return
	}

	// Fetch repository info from Docker Hub API
	repoInfo, err := fetchDockerInfo(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Fetch tags from Docker Hub API
	tagsResp, err := fetchDockerTags(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching tags: %v\n", err)
		return
	}

	// Build cache object
	cache = &DockerCache{
		Name:        repoInfo.Name,
		Namespace:   repoInfo.Namespace,
		Description: repoInfo.Description,
		StarCount:   repoInfo.StarCount,
		PullCount:   repoInfo.PullCount,
		Tags:        make([]string, 0, len(tagsResp.Results)),
	}
	for _, tag := range tagsResp.Results {
		cache.Tags = append(cache.Tags, tag.Name)
	}

	// Save to cache
	dockerSaveCache(cachePath, cache)

	// Display the compressed output
	displayDockerInfo(cache)
}

func printDockerUsage() {
	fmt.Println("Usage: gg docker <image>")
	fmt.Println("  Fetches Docker Hub image info compressed to ~20 tokens")
	fmt.Println("  Examples:")
	fmt.Println("    gg docker nginx")
	fmt.Println("    gg docker grafana/grafana")
}

func dockerGetCachePath(imageName string) string {
	ggDir := getGGDir()
	// Sanitize image name for use as filename (replace / with _)
	safeName := strings.ReplaceAll(imageName, "/", "_")
	return filepath.Join(ggDir, "cache", "docker", safeName+".json")
}

func dockerLoadCache(cachePath string) (*DockerCache, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cache DockerCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func dockerSaveCache(cachePath string, cache *DockerCache) {
	dir := filepath.Dir(cachePath)
	os.MkdirAll(dir, 0755)
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(cachePath, data, 0644)
}

func fetchDockerInfo(repoPath string) (*DockerRepoInfo, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/", repoPath)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var info DockerRepoInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

func fetchDockerTags(repoPath string) (*DockerTagsResponse, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags/?page_size=5&ordering=last_updated", repoPath)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var tagsResp DockerTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, err
	}

	return &tagsResp, nil
}

func displayDockerInfo(cache *DockerCache) {
	// First line: official images show short name with full path in parens
	// User images show namespace/name with no parens (BUG 7 fix)
	if cache.Namespace == "library" {
		fmt.Printf("  %s (library/%s)\n", cache.Name, cache.Name)
	} else {
		fmt.Printf("  %s/%s\n", cache.Namespace, cache.Name)
	}

	// Description
	fmt.Printf("    %s\n", cache.Description)

	// Stars and pulls (PullCount is int64, BUG 4/5 fix)
	fmt.Printf("    ★ %d | ↓ %s pulls\n", cache.StarCount, dockerFmtPulls(cache.PullCount))

	// Tags: bounds-safe slicing (BUG 6 fix)
	n := len(cache.Tags)
	if n > 5 {
		n = 5
	}
	fmt.Printf("    Tags: %s\n", strings.Join(cache.Tags[:n], ", "))

	// MCP Endpoint
	if cache.Namespace == "library" {
		fmt.Printf("  MCP Endpoint: docker:%s\n", cache.Name)
	} else {
		fmt.Printf("  MCP Endpoint: docker:%s/%s\n", cache.Namespace, cache.Name)
	}

	// Token cost
	fmt.Println("  Token cost: ~20")
}

// dockerFmtPulls formats a pull count into a human-readable string
// Accepts int64 to handle Docker Hub pull counts exceeding 2^31 (BUG 5 fix)
func dockerFmtPulls(count int64) string {
	if count < 1000 {
		return fmt.Sprintf("%d", count)
	}
	if count < 1000000 {
		return fmt.Sprintf("%dK", count/1000)
	}
	if count < 1000000000 {
		return fmt.Sprintf("%dM", count/1000000)
	}
	// >= 1B: show one decimal place
	return fmt.Sprintf("%.1fB", float64(count)/1000000000)
}
