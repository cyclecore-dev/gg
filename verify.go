package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// verify.go — one-line test/lint/build summary that complements gg ask

func handleVerify() {
	if hasFlag("--help") || hasFlag("-h") {
		printVerifyHelp()
		return
	}

	start := time.Now()
	results := []string{}
	failed := false

	// 1. Tests (reuse existing detection)
	testResult, testFail := runVerifyStep("test")
	results = append(results, testResult)
	if testFail {
		failed = true
	}

	// 2. Optional lint (best-effort)
	if !hasFlag("--no-lint") {
		lintResult, lintFail := runVerifyStep("lint")
		if lintResult != "" {
			results = append(results, lintResult)
			if lintFail {
				failed = true
			}
		}
	}

	// 3. Optional build
	if hasFlag("--build") || hasFlag("-b") {
		buildResult, buildFail := runVerifyStep("build")
		if buildResult != "" {
			results = append(results, buildResult)
			if buildFail {
				failed = true
			}
		}
	}

	icon := "✓"
	if failed {
		icon = "✗"
	}
	elapsed := time.Since(start)
	line := fmt.Sprintf("%s verify | %s | %s", icon, strings.Join(results, " · "), fmtDuration(elapsed))

	// honor --max-tokens if present
	if max := getMaxTokens(); max > 0 {
		line = truncateToTokens(line, max)
	}

	fmt.Println(line)
	if failed {
		os.Exit(1)
	}
}

func runVerifyStep(kind string) (string, bool) {
	switch kind {
	case "test":
		// lightweight wrapper around existing test detection
		for _, r := range testRunners {
			if _, err := os.Stat(r.file); err == nil {
				cmd := exec.Command(r.cmd, r.args...)
				out, err := cmd.CombinedOutput()
				passed, failed, _ := parseTestOutput(r.name, string(out))
				status := fmt.Sprintf("%s %d✓", r.name, passed)
				if failed > 0 {
					status = fmt.Sprintf("%s %d✓ %d✗", r.name, passed, failed)
				}
				return status, err != nil || failed > 0
			}
		}
		return "no-tests", false

	case "lint":
		// common linters, best-effort, don't fail the whole verify if missing
		linters := []struct{ file, cmd string; args []string }{
			{"package.json", "npx", []string{"--yes", "eslint", "."}},
			{".eslintrc", "npx", []string{"--yes", "eslint", "."}},
			{"go.mod", "golangci-lint", []string{"run", "--timeout=2m"}},
			{"pyproject.toml", "ruff", []string{"check", "."}},
		}
		for _, l := range linters {
			if _, err := os.Stat(l.file); err == nil {
				if _, err := exec.LookPath(l.cmd); err != nil && l.cmd != "npx" {
					continue
				}
				cmd := exec.Command(l.cmd, l.args...)
				err := cmd.Run()
				if err != nil {
					return "lint✗", true
				}
				return "lint✓", false
			}
		}
		return "", false

	case "build":
		builders := []struct{ file, cmd string; args []string }{
			{"go.mod", "go", []string{"build", "./..."}},
			{"package.json", "npm", []string{"run", "build", "--if-present"}},
			{"Cargo.toml", "cargo", []string{"build", "--quiet"}},
		}
		for _, b := range builders {
			if _, err := os.Stat(b.file); err == nil {
				cmd := exec.Command(b.cmd, b.args...)
				err := cmd.Run()
				if err != nil {
					return "build✗", true
				}
				return "build✓", false
			}
		}
		return "no-build", false
	}
	return "", false
}

func printVerifyHelp() {
	fmt.Println("Usage: gg verify [--no-lint] [--build|-b] [--max-tokens N]")
	fmt.Println()
	fmt.Println("One-line summary of test + optional lint/build.")
	fmt.Println("Complements 'gg ask' — run after a PR is opened.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gg verify")
	fmt.Println("  gg verify --build")
	fmt.Println("  gg verify --no-lint --max-tokens 20")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  ✓ verify | go test 12✓ · lint✓ · build✓ | 1.4s")
	fmt.Println("  ✗ verify | npm test 47✓ 3✗ · lint✗ | 2.8s")
}
