package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func printEnvUsage() {
	fmt.Println("Usage: gg env [--help|-h]")
	fmt.Println("Prints a compressed summary of the runtime environment.")
}

func envToolVersion(cmd string, args ...string) string {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	for _, token := range strings.Fields(s) {
		token = strings.TrimPrefix(token, "v")
		token = strings.TrimSuffix(token, ",")
		if len(token) > 0 && token[0] >= '0' && token[0] <= '9' && strings.Contains(token, ".") {
			return token
		}
	}
	return s
}

func handleEnv() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		printEnvUsage()
		return
	}

	user := os.Getenv("USER")
	shell := filepath.Base(os.Getenv("SHELL"))
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	fmt.Printf("  %s@%s/%s (%s)\n", user, goos, goarch, shell)

	tools := []struct {
		name string
		cmd  string
		args []string
	}{
		{"git", "git", []string{"--version"}},
		{"go", "go", []string{"version"}},
		{"node", "node", []string{"--version"}},
		{"python3", "python3", []string{"--version"}},
		{"rustc", "rustc", []string{"--version"}},
		{"docker", "docker", []string{"--version"}},
		{"java", "java", []string{"-version"}},
	}

	fmt.Println("\n  tools:")
	for _, t := range tools {
		ver := envToolVersion(t.cmd, t.args...)
		if ver != "" {
			fmt.Printf("    %-10s%s\n", t.name, ver)
		}
	}
	fmt.Println()

	fmt.Println("\n  env:")
	venv := os.Getenv("VIRTUAL_ENV")
	if venv != "" {
		fmt.Println("    python venv active")
	}
	conda := os.Getenv("CONDA_DEFAULT_ENV")
	if conda != "" {
		fmt.Printf("    conda active (%s)\n", conda)
	}
	nvm := os.Getenv("NVM_DIR")
	if nvm != "" {
		nodeVer := envToolVersion("node", "--version")
		fmt.Printf("    nvm active (node %s)\n", nodeVer)
	}
	fmt.Println()

	fmt.Println("  Token cost: ~20")
}
