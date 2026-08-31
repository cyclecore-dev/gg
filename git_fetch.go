package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func printFetchUsage() {
	fmt.Println("Usage: gg fetch [remote|--all|--help|-h]")
	fmt.Println("  remote      Fetch from specified remote (default: origin)")
	fmt.Println("  --all, -a   Fetch from all remotes")
	fmt.Println("  --help, -h  Show this help message")
}

func fetchParseOutput(output []byte) (int, int, int, int) {
	lines := strings.Split(string(output), "\n")
	newBranches := 0
	newTags := 0
	updatedBranches := 0
	forcedUpdates := 0

	for _, line := range lines {
		if strings.Contains(line, "[new branch]") {
			newBranches++
		} else if strings.Contains(line, "[new tag]") {
			newTags++
		} else if strings.Contains(line, "..") && !strings.Contains(line, "[new") {
			updatedBranches++
			if strings.Contains(line, "(forced update)") {
				forcedUpdates++
			}
		}
	}
	return newBranches, newTags, updatedBranches, forcedUpdates
}

func handleFetch() {
	remote := "origin"
	allRemotes := false

	if len(os.Args) > 2 {
		arg := os.Args[2]
		if arg == "--help" || arg == "-h" {
			printFetchUsage()
			return
		}
		if arg == "--all" || arg == "-a" {
			allRemotes = true
		} else {
			remote = arg
		}
	}

	var cmd *exec.Cmd
	if allRemotes {
		cmd = exec.Command("git", "fetch", "--all")
	} else {
		cmd = exec.Command("git", "fetch", remote)
	}

	output, _ := cmd.CombinedOutput()

	newBranches, newTags, updatedBranches, forcedUpdates := fetchParseOutput(output)

	totalUpdates := updatedBranches + forcedUpdates
	totalNew := newBranches + newTags

	if totalUpdates == 0 && totalNew == 0 {
		if allRemotes {
			fmt.Println("  ✓ fetched all remotes — up to date")
		} else {
			fmt.Printf("  ✓ fetched %s — up to date\n", remote)
		}
		return
	}

	if allRemotes {
		fmt.Printf("  ✓ fetched all remotes | %d updated, %d new branches, %d new tags\n", totalUpdates, newBranches, newTags)
	} else {
		fmt.Printf("  ✓ fetched %s | %d updated, %d new branches, %d new tags\n", remote, totalUpdates, newBranches, newTags)
	}
}
