package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var testRunners = []struct {
	file string
	cmd  string
	args []string
	name string
}{
	{"go.mod", "go", []string{"test", "./..."}, "go test"},
	{"package.json", "npm", []string{"test"}, "npm test"},
	{"Cargo.toml", "cargo", []string{"test"}, "cargo test"},
	{"pyproject.toml", "pytest", []string{}, "pytest"},
	{"requirements.txt", "pytest", []string{}, "pytest"},
	{"Gemfile", "bundle", []string{"exec", "rspec"}, "rspec"},
	{"mix.exs", "mix", []string{"test"}, "mix test"},
}

func handleTest() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		fmt.Println("Usage: gg test [args...]")
		fmt.Println("\nAuto-detects test runner and shows compressed results (<30 tokens).")
		fmt.Println("Extra args are passed through to the test command.")
		fmt.Println("\nExamples:")
		fmt.Println("  gg test                    Auto-detect and run")
		fmt.Println("  gg test -v                 Verbose (passed to runner)")
		fmt.Println("  gg test -run TestAuth      Go: run specific test")
		fmt.Println("\nOutput:")
		fmt.Println("  ✓ go test | 12 passed | 0.8s")
		fmt.Println("  ✗ npm test | 47 passed, 3 failed | 2.1s")
		fmt.Println("    FAIL: src/auth.test.js > should validate token")
		return
	}

	var runner *struct {
		file string
		cmd  string
		args []string
		name string
	}
	for i := range testRunners {
		if _, err := os.Stat(testRunners[i].file); err == nil {
			runner = &testRunners[i]
			break
		}
	}
	if runner == nil {
		fmt.Println("No test runner detected. Supported: go, npm, cargo, pytest, rspec, mix")
		return
	}

	var passthrough []string
	for _, arg := range os.Args[2:] {
		passthrough = append(passthrough, arg)
	}

	args := append(runner.args, passthrough...)
	cmd := exec.Command(runner.cmd, args...)

	start := time.Now()
	output, err := cmd.CombinedOutput()
	elapsed := time.Since(start)

	outStr := string(output)
	passed, failed, failures := parseTestOutput(runner.name, outStr)

	icon := "✓"
	exitCode := 0
	if err != nil || failed > 0 {
		icon = "✗"
		exitCode = 1
	}

	fmt.Printf("%s %s | %d passed", icon, runner.name, passed)
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Printf(" | %s\n", fmtDuration(elapsed))

	if len(failures) > 0 {
		for i, f := range failures {
			if i >= 5 {
				fmt.Printf("  ... and %d more failures\n", len(failures)-5)
				break
			}
			fmt.Printf("  FAIL: %s\n", f)
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func parseTestOutput(runner string, output string) (passed, failed int, failures []string) {
	switch runner {
	case "go test":
		return parseGoTest(output)
	case "npm test":
		return parseNpmTest(output)
	case "cargo test":
		return parseCargoTest(output)
	case "pytest":
		return parsePytest(output)
	default:
		return parseGenericTest(output)
	}
}

func parseGoTest(output string) (passed, failed int, failures []string) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ok  \t") || strings.HasPrefix(line, "ok \t") {
			passed++
		} else if strings.HasPrefix(line, "FAIL\t") {
			failed++
		} else if strings.HasPrefix(line, "--- FAIL: ") {
			name := strings.TrimPrefix(line, "--- FAIL: ")
			if idx := strings.Index(name, " ("); idx != -1 {
				name = name[:idx]
			}
			failures = append(failures, name)
		}
	}
	return
}

func parseNpmTest(output string) (passed, failed int, failures []string) {
	re := regexp.MustCompile(`(\d+)\s+(passed|failed)`)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Tests:") {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				n, _ := strconv.Atoi(m[1])
				if m[2] == "passed" {
					passed = n
				} else if m[2] == "failed" {
					failed = n
				}
			}
		}
		if strings.Contains(line, "● ") {
			failures = append(failures, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "● ")))
		}
	}
	return
}

func parseCargoTest(output string) (passed, failed int, failures []string) {
	re := regexp.MustCompile(`(\d+)\s+(passed|failed|ignored)`)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "test result:") {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, m := range matches {
				n, _ := strconv.Atoi(m[1])
				if m[2] == "passed" {
					passed = n
				} else if m[2] == "failed" {
					failed = n
				}
			}
		}
		if strings.HasPrefix(strings.TrimSpace(line), "---- ") && strings.Contains(line, "stdout") {
			name := strings.TrimSpace(line)
			name = strings.TrimPrefix(name, "---- ")
			name = strings.TrimSuffix(name, " stdout ----")
			failures = append(failures, name)
		}
	}
	return
}

func parsePytest(output string) (passed, failed int, failures []string) {
	re := regexp.MustCompile(`(\d+)\s+(passed|failed|error|warnings?)`)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "passed") || strings.Contains(line, "failed") || strings.Contains(line, "error") {
			if strings.Contains(line, "=") {
				matches := re.FindAllStringSubmatch(line, -1)
				for _, m := range matches {
					n, _ := strconv.Atoi(m[1])
					if m[2] == "passed" {
						passed = n
					} else if m[2] == "failed" || m[2] == "error" {
						failed += n
					}
				}
			}
		}
		if strings.HasPrefix(strings.TrimSpace(line), "FAILED ") {
			name := strings.TrimPrefix(strings.TrimSpace(line), "FAILED ")
			failures = append(failures, name)
		}
	}
	return
}

func parseGenericTest(output string) (passed, failed int, failures []string) {
	for _, line := range strings.Split(output, "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "pass") && !strings.Contains(lower, "fail") {
			passed++
		}
		if strings.Contains(lower, "fail") || strings.Contains(lower, "error") {
			failed++
		}
	}
	return
}

func fmtDuration(d time.Duration) string {
	secs := d.Seconds()
	if secs < 60 {
		return fmt.Sprintf("%.1fs", secs)
	}
	mins := int(secs) / 60
	rem := int(secs) % 60
	return fmt.Sprintf("%dm%ds", mins, rem)
}
