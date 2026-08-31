package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handlePR() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg pr <number>")
		return
	}

	prNumber := os.Args[2]
	if err := ensureGitHubAuth(); err != nil {
		return
	}

	repoName := getCurrentRepo()
	if repoName == "" {
		fatalError("Not in a git repository", nil)
	}

	// Fetch PR details
	cmd := exec.Command("gh", "pr", "view", prNumber, "--json", "number,title,author,state,body,additions,deletions,changedFiles,headRefName,baseRefName,url")
	output, err := cmd.Output()
	if err != nil {
		fatalError("Failed to fetch PR", err)
	}

	var pr struct {
		Number       int    `json:"number"`
		Title        string `json:"title"`
		Author       struct{ Login string } `json:"author"`
		State        string `json:"state"`
		Body         string `json:"body"`
		Additions    int    `json:"additions"`
		Deletions    int    `json:"deletions"`
		ChangedFiles int    `json:"changedFiles"`
		HeadRefName  string `json:"headRefName"`
		BaseRefName  string `json:"baseRefName"`
		URL          string `json:"url"`
	}

	if err := json.Unmarshal(output, &pr); err != nil {
		fatalError("Failed to parse PR data", err)
	}

	fmt.Printf("PR #%d: %s\n", pr.Number, pr.Title)
	fmt.Printf("Author: %s | State: %s\n", pr.Author.Login, pr.State)
	fmt.Printf("Branch: %s -> %s\n", pr.HeadRefName, pr.BaseRefName)
	fmt.Printf("Changes: +%d -%d (%d files)\n", pr.Additions, pr.Deletions, pr.ChangedFiles)
	fmt.Println()

	if pr.Body != "" {
		fmt.Println("Description:")
		fmt.Println(truncate(pr.Body, 500))
		fmt.Println()
	}

	fmt.Printf("URL: %s\n", pr.URL)
	fmt.Println()

	// Show action menu
	if pr.State == "OPEN" {
		fmt.Println("Actions:")
		fmt.Println("  [a]pprove - Merge this PR")
		fmt.Println("  [d]iff   - Show full diff")
		fmt.Println("  [c]lose  - Close without merging")
		fmt.Println("  [q]uit   - Exit")
		fmt.Print("\nChoice: ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		switch choice {
		case "a":
			mergeCmd := exec.Command("gh", "pr", "merge", prNumber, "--squash", "--delete-branch")
			mergeCmd.Stdout = os.Stdout
			mergeCmd.Stderr = os.Stderr
			if err := mergeCmd.Run(); err != nil {
				fatalError("Failed to merge PR", err)
			}
			fmt.Println("PR merged!")
		case "d":
			diffCmd := exec.Command("gh", "pr", "diff", prNumber)
			diffCmd.Stdout = os.Stdout
			diffCmd.Stderr = os.Stderr
			diffCmd.Run()
		case "c":
			closeCmd := exec.Command("gh", "pr", "close", prNumber)
			closeCmd.Stdout = os.Stdout
			closeCmd.Stderr = os.Stderr
			if err := closeCmd.Run(); err != nil {
				fatalError("Failed to close PR", err)
			}
			fmt.Println("PR closed")
		default:
			fmt.Println("Exiting")
		}
	}
}
