package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func printBranchUsage() {
	fmt.Println("Usage: gg branch [list|create|delete|-d|--help|-h]")
	fmt.Println("  list             List all branches (default)")
	fmt.Println("  create <name>    Create and switch to a new branch")
	fmt.Println("  delete <name>    Delete a branch")
	fmt.Println("  -d <name>        Delete a branch")
	fmt.Println("  <name>           Create and switch to a new branch")
}

func handleBranch() {
	subcmd := "list"
	if len(os.Args) > 2 {
		subcmd = os.Args[2]
	}

	switch subcmd {
	case "--help", "-h":
		printBranchUsage()
	case "list":
		branchList()
	case "create":
		if len(os.Args) < 4 {
			printBranchUsage()
			return
		}
		branchCreate(os.Args[3])
	case "delete", "-d":
		if len(os.Args) < 4 {
			printBranchUsage()
			return
		}
		branchDelete(os.Args[3])
	default:
		branchCreate(subcmd)
	}
}

func branchList() {
	cmd := exec.Command("git", "branch", "-vv",
		"--format=%(refname:short)|||%(objectname:short)|||%(upstream:short)|||%(upstream:track)|||%(subject)")
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing branches: %v\n", err)
		return
	}

	headCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	headOut, _ := headCmd.Output()
	current := strings.TrimSpace(string(headOut))

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|||", 5)
		if len(parts) < 5 {
			continue
		}
		name := parts[0]
		hash := parts[1]
		upstream := parts[2]
		track := parts[3]
		subject := parts[4]

		marker := "   "
		if name == current {
			marker = " * "
		}

		upInfo := ""
		if upstream != "" {
			upInfo = " [" + upstream
			if track != "" {
				track = strings.Trim(track, "[]")
				upInfo += ": " + track
			}
			upInfo += "]"
		}

		fmt.Printf("%s%-12s %s%s \"%s\"\n", marker, name, hash, upInfo, subject)
		count++
	}
	fmt.Printf("  %d branches | Token cost: ~%d\n", count, 4+count*3)
}

func branchCreate(name string) {
	headCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	headOut, _ := headCmd.Output()
	fromBranch := strings.TrimSpace(string(headOut))

	hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	hashOut, _ := hashCmd.Output()
	fromHash := strings.TrimSpace(string(hashOut))

	cmd := exec.Command("git", "checkout", "-b", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ branch create failed: %s\n", strings.TrimSpace(string(output)))
		return
	}
	fmt.Printf("  ✓ created branch %s from %s@%s\n", name, fromBranch, fromHash)
	fmt.Println("  Token cost: ~8")
}

func branchDelete(name string) {
	cmd := exec.Command("git", "branch", "-d", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ branch delete failed: %s\n", strings.TrimSpace(string(output)))
		return
	}
	fmt.Printf("  ✓ deleted branch %s\n", name)
	fmt.Println("  Token cost: ~5")
}
