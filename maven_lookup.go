package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Maven API response structures
type mavenResponse struct {
	Response struct {
		NumFound int        `json:"numFound"`
		Docs     []mavenDoc `json:"docs"`
	} `json:"response"`
}

type mavenDoc struct {
	ID            string   `json:"id"`
	G             string   `json:"g"`
	A             string   `json:"a"`
	LatestVersion string   `json:"latestVersion"`
	RepositoryID  string   `json:"repositoryId"`
	P             string   `json:"p"`
	Timestamp     int64    `json:"timestamp"`
	VersionCount  int      `json:"versionCount"`
	Tags          []string `json:"tags"`
}

func handleMaven() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printMavenUsage()
		return
	}

	query := os.Args[2]
	apiQuery := mavenBuildQuery(query)

	// Check cache
	cachePath := filepath.Join(getGGDir(), "cache", "maven", mavenEscapeQuery(query)+".json")

	if data, err := os.ReadFile(cachePath); err == nil {
		var doc mavenDoc
		if err := json.Unmarshal(data, &doc); err == nil {
			displayMavenResult(&doc)
			return
		}
	}

	// Fetch from Maven Central API
	url := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=%s&rows=1&wt=json", apiQuery)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching Maven data: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var mavenResp mavenResponse
	if err := json.NewDecoder(resp.Body).Decode(&mavenResp); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing Maven response: %v\n", err)
		return
	}

	if len(mavenResp.Response.Docs) == 0 {
		fmt.Fprintf(os.Stderr, "No Maven package found for: %s\n", query)
		return
	}

	doc := mavenResp.Response.Docs[0]

	// Cache the result
	if data, err := json.Marshal(doc); err == nil {
		os.MkdirAll(filepath.Dir(cachePath), 0755)
		os.WriteFile(cachePath, data, 0644)
	}

	displayMavenResult(&doc)
}

func mavenBuildQuery(input string) string {
	if strings.Contains(input, ":") {
		parts := strings.SplitN(input, ":", 2)
		return fmt.Sprintf("g:%s+AND+a:%s", parts[0], parts[1])
	}
	return input
}

func printMavenUsage() {
	fmt.Println("Usage: gg maven <group:artifact> or gg maven <query>")
	fmt.Println("  Examples:")
	fmt.Println("    gg maven com.google.guava:guava  # exact lookup")
	fmt.Println("    gg maven guava                   # search query")
}

func mavenEscapeQuery(query string) string {
	return strings.ReplaceAll(query, ":", "__")
}

func displayMavenResult(doc *mavenDoc) {
	// Format tags or default to "Java library"
	var tagsStr string
	if len(doc.Tags) > 0 {
		tagsStr = strings.Join(doc.Tags, ", ")
	} else {
		tagsStr = "Java library"
	}

	// Format timestamp as date
	var dateStr string
	if doc.Timestamp > 0 {
		t := time.Unix(doc.Timestamp/1000, 0)
		dateStr = t.Format("2006-01-02")
	}

	// Build output string for token cost estimation
	output := fmt.Sprintf("%s@%s\n    Tags: %s\n    Type: %s | %d versions",
		doc.ID, doc.LatestVersion, tagsStr, doc.P, doc.VersionCount)
	if dateStr != "" {
		output += fmt.Sprintf("\n    Updated: %s", dateStr)
	}

	// Estimate token cost (rough: ~4 chars per token for English)
	tokenCost := len(output) / 4
	if tokenCost < 10 {
		tokenCost = 10
	}

	fmt.Printf("  %s@%s\n", doc.ID, doc.LatestVersion)
	fmt.Printf("    Tags: %s\n", tagsStr)
	fmt.Printf("    Type: %s | %d versions\n", doc.P, doc.VersionCount)
	if dateStr != "" {
		fmt.Printf("    Updated: %s\n", dateStr)
	}
	fmt.Printf("  MCP Endpoint: maven:%s\n", doc.ID)
	fmt.Printf("  Token cost: ~%d\n", tokenCost)
}
