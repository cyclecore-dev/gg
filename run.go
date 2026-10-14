package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func handleRun() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg run <command>")
		fmt.Println("Example: gg run npm test")
		return
	}

	cmdArgs := os.Args[2:]
	cmdStr := strings.Join(cmdArgs, " ")

	fmt.Printf("Running: %s\n", cmdStr)
	fmt.Println()

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	fmt.Println()
	if err != nil {
		exitCode := 1
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		fmt.Printf("Exit code: %d (%.2fs)\n", exitCode, elapsed.Seconds())
	} else {
		fmt.Printf("Success (%.2fs)\n", elapsed.Seconds())
	}

	// Track usage
	trackCommandUsage("run", cmdStr, elapsed)
}
