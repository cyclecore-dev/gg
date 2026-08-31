package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func handleStats() {
	homeDir := getHomeDir()
	statsPath := filepath.Join(homeDir, ".gg", "stats.json")
	data, err := os.ReadFile(statsPath)
	if err != nil {
		fmt.Println("Usage Statistics")
		fmt.Println()
		fmt.Println("No usage data yet. Run some commands first!")
		return
	}

	var stats UsageStats
	if err := json.Unmarshal(data, &stats); err != nil {
		fatalError("Failed to parse stats", err)
	}

	fmt.Println("Usage Statistics")
	fmt.Println()
	fmt.Printf("Month: %s\n", stats.Month)
	fmt.Printf("Total asks: %d\n", stats.AskCount)
	fmt.Printf("Total runs: %d\n", stats.RunCount)
	fmt.Printf("Total tokens: %d (input: %d, output: %d)\n", stats.TotalTokens, stats.InputTokens, stats.OutputTokens)
	fmt.Printf("Estimated cost: $%.4f\n", stats.EstimatedCost)
}

// UsageStats tracks monthly usage
type UsageStats struct {
	Month         string  `json:"month"`
	AskCount      int     `json:"ask_count"`
	RunCount      int     `json:"run_count"`
	TotalTokens   int64   `json:"total_tokens"`
	InputTokens   int64   `json:"input_tokens"`
	OutputTokens  int64   `json:"output_tokens"`
	EstimatedCost float64 `json:"estimated_cost"`
}

func trackCommandUsage(cmdType, detail string, elapsed time.Duration) {
	homeDir := getHomeDir()
	statsPath := filepath.Join(homeDir, ".gg", "stats.json")

	var stats UsageStats
	data, err := os.ReadFile(statsPath)
	if err == nil {
		json.Unmarshal(data, &stats)
	}

	currentMonth := time.Now().Format("2006-01")
	if stats.Month != currentMonth {
		stats = UsageStats{Month: currentMonth}
	}

	switch cmdType {
	case "ask":
		stats.AskCount++
	case "run":
		stats.RunCount++
	}

	outData, _ := json.MarshalIndent(stats, "", "  ")
	os.WriteFile(statsPath, outData, 0644)
}

func trackTokenUsage(inputTokens, outputTokens int64) {
	homeDir := getHomeDir()
	statsPath := filepath.Join(homeDir, ".gg", "stats.json")

	var stats UsageStats
	data, err := os.ReadFile(statsPath)
	if err == nil {
		json.Unmarshal(data, &stats)
	}

	currentMonth := time.Now().Format("2006-01")
	if stats.Month != currentMonth {
		stats = UsageStats{Month: currentMonth}
	}

	stats.InputTokens += inputTokens
	stats.OutputTokens += outputTokens
	stats.TotalTokens = stats.InputTokens + stats.OutputTokens

	// Approximate cost (varies by provider)
	stats.EstimatedCost = float64(stats.InputTokens)/1000000*3 + float64(stats.OutputTokens)/1000000*15

	outData, _ := json.MarshalIndent(stats, "", "  ")
	os.WriteFile(statsPath, outData, 0644)
}

// ============================================================================
// MULTI-PROVIDER API
// ============================================================================
