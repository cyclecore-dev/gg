package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// flags.go — shared --max-tokens and --json helpers for all commands

func hasFlag(name string) bool {
	for _, a := range os.Args {
		if a == name {
			return true
		}
	}
	return false
}

func getFlagValue(name string) string {
	for i, a := range os.Args {
		if a == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
		if strings.HasPrefix(a, name+"=") {
			return strings.TrimPrefix(a, name+"=")
		}
	}
	return ""
}

// getMaxTokens returns the value of --max-tokens N (or 0 if absent)
func getMaxTokens() int {
	v := getFlagValue("--max-tokens")
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 0
	}
	return n
}

// truncateToTokens does a rough token-aware truncation.
// Approximation: 1 token ≈ 4 characters (good enough for agent output).
func truncateToTokens(s string, maxTokens int) string {
	if maxTokens <= 0 {
		return s
	}
	maxChars := maxTokens * 4
	if utf8.RuneCountInString(s) <= maxChars {
		return s
	}
	// cut on a word boundary if possible
	runes := []rune(s)
	if len(runes) > maxChars {
		cut := maxChars
		for cut > maxChars-20 && cut > 0 {
			if runes[cut] == ' ' || runes[cut] == '\n' {
				break
			}
			cut--
		}
		if cut < maxChars/2 {
			cut = maxChars
		}
		return string(runes[:cut]) + "…"
	}
	return s
}

// applyMaxTokens wraps any output string with the global policy
func applyMaxTokens(s string) string {
	if max := getMaxTokens(); max > 0 {
		return truncateToTokens(s, max)
	}
	return s
}

// isJSONRequested returns true if --json is present
func isJSONRequested() bool {
	return hasFlag("--json")
}

// emit is the recommended output helper for new/updated commands.
// It respects --json (caller must pass already-serialized JSON) and --max-tokens.
func emit(plain string, jsonStr string) {
	if isJSONRequested() && jsonStr != "" {
		fmt.Println(applyMaxTokens(jsonStr))
		return
	}
	fmt.Println(applyMaxTokens(plain))
}
