package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handleRemote() {
	subcmd := "list"
	if len(os.Args) > 2 {
		subcmd = os.Args[2]
	}

	switch subcmd {
	case "add":
		remoteAdd()
		return
	case "rm":
		remoteRemove()
		return
	case "--help", "-h":
		printRemoteUsage()
		return
	default:
		remoteList()
		return
	}
}

func printRemoteUsage() {
	fmt.Println("Usage: gg remote [add|rm|--help|-h]")
	fmt.Println("  add <name> <url>  Add a new remote")
	fmt.Println("  rm <name>         Remove a remote")
	fmt.Println("  --help, -h        Show this help message")
}

func cleanRemoteURL(url string) string {
	if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "git@") {
		url = strings.TrimPrefix(url, "git@")
		url = strings.Replace(url, ":", "/", 1)
	} else if strings.HasPrefix(url, "ssh://") {
		url = strings.TrimPrefix(url, "ssh://")
	}

	if strings.HasSuffix(url, ".git") {
		url = strings.TrimSuffix(url, ".git")
	}

	return url
}

func remoteList() {
	cmd := exec.Command("git", "remote", "-v")
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running git remote -v: %v\n", err)
		return
	}

	type remoteInfo struct {
		name string
		url  string
	}

	remotes := make(map[string]remoteInfo)
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		url := parts[1]
		if idx := strings.LastIndex(url, " ("); idx != -1 {
			url = url[:idx]
		}
		if _, ok := remotes[name]; ok {
			continue
		}
		remotes[name] = remoteInfo{name: name, url: url}
	}

	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current branch: %v\n", err)
		return
	}
	currentBranch := strings.TrimSpace(string(branchOut))

	count := 0
	for name, info := range remotes {
		cleanURL := cleanRemoteURL(info.url)
		fmt.Printf("  %s  %s", name, cleanURL)

		upstream := fmt.Sprintf("%s/%s", name, currentBranch)
		trackCmd := exec.Command("git", "rev-list", "--left-right", "--count", fmt.Sprintf("%s...HEAD", upstream))
		trackOut, err := trackCmd.Output()
		if err == nil {
			trackParts := strings.Split(strings.TrimSpace(string(trackOut)), "\t")
			if len(trackParts) == 2 {
				var behind, ahead int
				fmt.Sscanf(trackParts[0], "%d", &behind)
				fmt.Sscanf(trackParts[1], "%d", &ahead)
				fmt.Printf(" (↑%d ↓%d)", ahead, behind)
			}
		}
		fmt.Println()
		count++
	}

	fmt.Printf("%d remotes | Token cost: ~10\n", count)
}

func remoteAdd() {
	if len(os.Args) < 5 {
		printRemoteUsage()
		return
	}
	name := os.Args[3]
	url := os.Args[4]

	cmd := exec.Command("git", "remote", "add", name, url)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding remote: %v\n", err)
		return
	}

	cleanURL := cleanRemoteURL(url)
	fmt.Printf("✓ added remote %s → %s\n", name, cleanURL)
}

func remoteRemove() {
	if len(os.Args) < 4 {
		printRemoteUsage()
		return
	}
	name := os.Args[3]

	cmd := exec.Command("git", "remote", "remove", name)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing remote: %v\n", err)
		return
	}

	fmt.Printf("✓ removed remote %s\n", name)
}
