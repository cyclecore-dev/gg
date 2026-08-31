package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var manifests = []string{
	"package.json",
	"go.mod",
	"Cargo.toml",
	"requirements.txt",
	"pyproject.toml",
	"Gemfile",
	"composer.json",
	"pubspec.yaml",
	"pom.xml",
	"build.gradle",
}

func handleDeps() {
	var manifestPath string
	if len(os.Args) > 2 {
		arg := os.Args[2]
		if arg == "--help" || arg == "-h" {
			printDepsUsage()
			return
		}
		manifestPath = arg
	}

	if manifestPath == "" {
		found := false
		for _, m := range manifests {
			if _, err := os.Stat(m); err == nil {
				manifestPath = m
				found = true
				break
			}
		}
		if !found {
			fmt.Println("No manifest file found. Supported: package.json, go.mod, Cargo.toml, requirements.txt, pyproject.toml, Gemfile, composer.json, pubspec.yaml")
			return
		}
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", manifestPath, err)
		return
	}

	content := string(data)
	filename := filepath.Base(manifestPath)

	var deps []string
	var devDeps []string

	switch {
	case strings.HasSuffix(manifestPath, "package.json"):
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err == nil {
			if d, ok := m["dependencies"].(map[string]interface{}); ok {
				for k, v := range d { deps = append(deps, fmt.Sprintf("%s@%v", k, v)) }
			}
			if d, ok := m["devDependencies"].(map[string]interface{}); ok {
				for k, v := range d { devDeps = append(devDeps, fmt.Sprintf("%s@%v", k, v)) }
			}
		}
	case strings.HasSuffix(manifestPath, "go.mod"):
		lines := strings.Split(content, "\n")
		inRequire := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "//") { continue }
			if strings.HasPrefix(line, "require (") { inRequire = true; continue }
			if inRequire {
				if line == ")" { inRequire = false; continue }
				parts := strings.Fields(line)
				if len(parts) >= 2 { deps = append(deps, fmt.Sprintf("%s@%s", parts[0], parts[1])) }
				continue
			}
			if strings.HasPrefix(line, "require ") {
				parts := strings.Fields(line)
				if len(parts) >= 3 { deps = append(deps, fmt.Sprintf("%s@%s", parts[1], parts[2])) }
			}
		}
	case strings.HasSuffix(manifestPath, "Cargo.toml"):
		lines := strings.Split(content, "\n")
		section := ""
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") { continue }
			if strings.HasPrefix(line, "[") {
				if line == "[dependencies]" || line == "[dev-dependencies]" || line == "[build-dependencies]" {
					section = strings.Trim(line, "[]")
				} else {
					section = ""
				}
				continue
			}
			if section != "" {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					name := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					version := ""
					if strings.HasPrefix(val, "\"") {
						version = strings.Trim(val, "\"")
					} else if strings.HasPrefix(val, "{") {
						if idx := strings.Index(val, "version"); idx != -1 {
							rest := val[idx:]
							if idx2 := strings.Index(rest, "\""); idx2 != -1 {
								if idx3 := strings.Index(rest[idx2+1:], "\""); idx3 != -1 {
									version = rest[idx2+1 : idx2+1+idx3]
								}
							}
						}
					}
					if version != "" {
						if section == "dev-dependencies" || section == "build-dependencies" {
							devDeps = append(devDeps, fmt.Sprintf("%s@%s", name, version))
						} else {
							deps = append(deps, fmt.Sprintf("%s@%s", name, version))
						}
					}
				}
			}
		}
	case strings.HasSuffix(manifestPath, "requirements.txt"):
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") { continue }
			ops := []string{"==", ">=", "<=", "~=", "!=", ">", "<"}
			found := false
			for _, op := range ops {
				if idx := strings.Index(line, op); idx != -1 {
					deps = append(deps, fmt.Sprintf("%s%s%s", strings.TrimSpace(line[:idx]), op, strings.TrimSpace(line[idx+len(op):])))
					found = true
					break
				}
			}
			if !found { deps = append(deps, line) }
		}
	case strings.HasSuffix(manifestPath, "pyproject.toml"):
		lines := strings.Split(content, "\n")
		inProject := false
		inDeps := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "[project]" { inProject = true; continue }
			if strings.HasPrefix(line, "[") { inProject = false; inDeps = false; continue }
			if inProject && strings.HasPrefix(line, "dependencies = [") { inDeps = true; continue }
			if inDeps {
				if strings.Contains(line, "]") { inDeps = false; continue }
				line = strings.Trim(line, "\"")
				ops := []string{">=", "<=", "!=", "~=", ">", "<", "=="}
				found := false
				for _, op := range ops {
					if idx := strings.Index(line, op); idx != -1 {
						deps = append(deps, fmt.Sprintf("%s%s%s", strings.TrimSpace(line[:idx]), op, strings.TrimSpace(line[idx+len(op):])))
						found = true
						break
					}
				}
				if !found { deps = append(deps, line) }
			}
		}
	case strings.HasSuffix(manifestPath, "Gemfile"):
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") { continue }
			if strings.HasPrefix(line, "gem ") {
				parts := strings.Split(line, "\"")
				if len(parts) >= 2 {
					name := parts[1]
					version := ""
					if len(parts) >= 4 { version = parts[3] }
					if version != "" {
						deps = append(deps, fmt.Sprintf("%s@%s", name, version))
					} else {
						deps = append(deps, name)
					}
				}
			}
		}
	case strings.HasSuffix(manifestPath, "composer.json"):
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err == nil {
			if d, ok := m["require"].(map[string]interface{}); ok {
				for k, v := range d { deps = append(deps, fmt.Sprintf("%s@%v", k, v)) }
			}
			if d, ok := m["require-dev"].(map[string]interface{}); ok {
				for k, v := range d { devDeps = append(devDeps, fmt.Sprintf("%s@%v", k, v)) }
			}
		}
	case strings.HasSuffix(manifestPath, "pubspec.yaml"):
		lines := strings.Split(content, "\n")
		inDeps := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "dependencies:" { inDeps = true; continue }
			if strings.HasPrefix(line, "  ") && inDeps {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					name := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					if len(val) > 0 && (val[0] == '^' || val[0] == '~' || (val[0] >= '0' && val[0] <= '9')) {
						deps = append(deps, fmt.Sprintf("%s@%s", name, val))
					} else {
						deps = append(deps, name)
					}
				}
			} else if !strings.HasPrefix(line, "  ") {
				inDeps = false
			}
		}
	case strings.HasSuffix(manifestPath, "pom.xml") || strings.HasSuffix(manifestPath, "build.gradle"):
		fmt.Printf("%s — not yet supported\n", filename)
		return
	}

	sort.Strings(deps)
	sort.Strings(devDeps)

	printDepsHeader(filename, deps, devDeps)
	printDepsSection("Dependencies", deps)
	if len(devDeps) > 0 { printDepsSection("Dev", devDeps) }
	printDepsTokenCost(len(deps) + len(devDeps))
}

func printDepsUsage() {
	fmt.Println("Usage: gg deps [manifest-file]")
	fmt.Println()
	fmt.Println("Auto-detects manifest in current directory, or specify one:")
	fmt.Println("  gg deps")
	fmt.Println("  gg deps package.json")
	fmt.Println("  gg deps path/to/go.mod")
	fmt.Println()
	fmt.Println("Supported: package.json, go.mod, Cargo.toml, requirements.txt,")
	fmt.Println("           pyproject.toml, Gemfile, composer.json, pubspec.yaml")
}

func printDepsHeader(filename string, deps, devDeps []string) {
	s := fmt.Sprintf("%s — %d deps", filename, len(deps))
	if len(devDeps) > 0 { s += fmt.Sprintf(", %d devDeps", len(devDeps)) }
	fmt.Println(s)
}

func printDepsSection(title string, items []string) {
	fmt.Printf("\n%s:\n", title)
	cap := 30
	if len(items) > cap {
		for _, item := range items[:cap] { fmt.Printf("  %s\n", item) }
		fmt.Printf("  ... and %d more\n", len(items)-cap)
	} else {
		for _, item := range items { fmt.Printf("  %s\n", item) }
	}
}

func printDepsTokenCost(n int) {
	fmt.Printf("\nToken cost: ~%d\n", n*2)
}
