package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func printCheckoutUsage() {
	fmt.Println("Usage: gg checkout <branch|file> [-b <branch>]")
	fmt.Println("  -b <branch>  Create and switch to a new branch")
	fmt.Println("  -h, --help   Show this help message")
}

func checkoutParseOutput(output string) (string, string) {
	if strings.Contains(output, "Switched to a new branch") {
		start := strings.Index(output, "'") + 1
		end := strings.LastIndex(output, "'")
		if start > 0 && end > start {
			return "new", output[start:end]
		}
		return "new", ""
	}
	if strings.Contains(output, "Switched to branch") {
		start := strings.Index(output, "'") + 1
		end := strings.LastIndex(output, "'")
		if start > 0 && end > start {
			return "switch", output[start:end]
		}
		return "switch", ""
	}
	if strings.Contains(output, "HEAD is now at") {
		parts := strings.Fields(output)
		for i, p := range parts {
			if p == "at" && i+1 < len(parts) {
				return "detached", parts[i+1]
			}
		}
		return "detached", ""
	}
	return "error", strings.TrimSpace(output)
}

func handleCheckout() {
	if len(os.Args) < 3 || os.Args[2] == "--help" || os.Args[2] == "-h" {
		printCheckoutUsage()
		return
	}

	if os.Args[2] == "-b" {
		if len(os.Args) < 4 {
			printCheckoutUsage()
			return
		}
		cmd := exec.Command("git", "checkout", "-b", os.Args[3])
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("  ✗ checkout failed: %s\n", strings.TrimSpace(string(output)))
			fmt.Println("  Token cost: ~8")
			return
		}
		fmt.Printf("  ✓ created and switched to %s\n", os.Args[3])
		fmt.Println("  Token cost: ~6")
		return
	}

	target := os.Args[2]
	cmd := exec.Command("git", "checkout", target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  ✗ checkout failed: %s\n", strings.TrimSpace(string(output)))
		fmt.Println("  Token cost: ~8")
		return
	}

	resultType, resultValue := checkoutParseOutput(string(output))
	switch resultType {
	case "switch":
		fmt.Printf("  ✓ switched to %s\n", resultValue)
		fmt.Println("  Token cost: ~5")
	case "new":
		fmt.Printf("  ✓ created and switched to %s\n", resultValue)
		fmt.Println("  Token cost: ~6")
	case "detached":
		fmt.Printf("  ✓ detached HEAD at %s\n", resultValue)
		fmt.Println("  Token cost: ~6")
	case "error":
		fmt.Printf("  ✗ checkout failed: %s\n", resultValue)
		fmt.Println("  Token cost: ~8")
	}
}
