package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func handleContext() {
	var detectors = []struct {
		file      string
		lang      string
		framework string
	}{
		{"go.mod", "Go", ""},
		{"Cargo.toml", "Rust", ""},
		{"package.json", "Node.js", ""},
		{"pyproject.toml", "Python", ""},
		{"requirements.txt", "Python", ""},
		{"pom.xml", "Java", "Maven"},
		{"build.gradle", "Java/Kotlin", "Gradle"},
		{"Gemfile", "Ruby", ""},
		{"mix.exs", "Elixir", ""},
		{"pubspec.yaml", "Dart", "Flutter"},
		{"composer.json", "PHP", ""},
	}

	lang := ""
	framework := ""
	manifestFile := ""

	for _, d := range detectors {
		if _, err := os.Stat(d.file); err == nil {
			lang = d.lang
			framework = d.framework
			manifestFile = d.file
			break
		}
	}

	if lang == "" {
		fmt.Println("Unknown project type")
		return
	}

	if manifestFile == "package.json" {
		data, err := os.ReadFile("package.json")
		if err == nil {
			content := string(data)
			if strings.Contains(content, `"next"`) {
				framework = "Next.js"
			} else if strings.Contains(content, `"react"`) {
				framework = "React"
			} else if strings.Contains(content, `"vue"`) {
				framework = "Vue"
			} else if strings.Contains(content, `"express"`) {
				framework = "Express"
			} else if strings.Contains(content, `"hono"`) {
				framework = "Hono"
			}
		}
	}

	extCounts := make(map[string]int)
	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		".venv": true, "__pycache__": true, "dist": true, "build": true,
	}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if ext != "" {
			extCounts[ext]++
		}
		return nil
	})

	type extCount struct {
		ext   string
		count int
	}
	var topExts []extCount
	for ext, count := range extCounts {
		topExts = append(topExts, extCount{ext, count})
	}
	for i := 0; i < len(topExts); i++ {
		for j := i + 1; j < len(topExts); j++ {
			if topExts[j].count > topExts[i].count {
				topExts[i], topExts[j] = topExts[j], topExts[i]
			}
		}
	}
	var extStrs []string
	for i, e := range topExts {
		if i >= 3 {
			break
		}
		extStrs = append(extStrs, fmt.Sprintf("%d %s files", e.count, e.ext))
	}
	extOutput := strings.Join(extStrs, ", ")

	entryPoints := []string{
		"main.go", "cmd/main.go",
		"src/index.ts", "src/index.js", "index.ts", "index.js",
		"app.py", "main.py", "src/main.py",
		"src/main.rs", "src/lib.rs",
		"src/App.tsx", "src/App.jsx",
	}
	var entryFile string
	var entrySize string
	for _, ep := range entryPoints {
		if info, err := os.Stat(ep); err == nil {
			entryFile = ep
			entrySize = fmt.Sprintf("%dKB", info.Size()/1024)
			break
		}
	}

	depCount := 0
	depType := ""
	if manifestFile == "go.mod" {
		data, _ := os.ReadFile("go.mod")
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "require ") {
				depCount++
			}
		}
		depType = "modules"
	} else if manifestFile == "package.json" {
		data, _ := os.ReadFile("package.json")
		content := string(data)
		lines := strings.Split(content, "\n")
		inDeps := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, `"dependencies"`) {
				inDeps = true
				continue
			}
			if strings.HasPrefix(trimmed, `"devDependencies"`) {
				inDeps = true
				continue
			}
			if strings.HasPrefix(trimmed, "}") {
				inDeps = false
				continue
			}
			if inDeps && strings.Contains(trimmed, ":") {
				depCount++
			}
		}
		depType = "dependencies"
	} else if manifestFile == "Cargo.toml" {
		data, _ := os.ReadFile("Cargo.toml")
		lines := strings.Split(string(data), "\n")
		inDeps := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "[dependencies]" {
				inDeps = true
				continue
			}
			if strings.HasPrefix(line, "[") {
				inDeps = false
				continue
			}
			if inDeps && strings.Contains(line, "=") {
				depCount++
			}
		}
		depType = "dependencies"
	} else if manifestFile == "requirements.txt" {
		data, _ := os.ReadFile("requirements.txt")
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				depCount++
			}
		}
		depType = "packages"
	}

	var lastHash, lastMsg, lastTime string
	var commitCount, contribCount int
	isGit := false

	if _, err := os.Stat(".git"); err == nil {
		isGit = true
		out, err := exec.Command("git", "log", "-1", "--format=%h").Output()
		if err == nil {
			lastHash = strings.TrimSpace(string(out))
		}
		out, err = exec.Command("git", "log", "-1", "--format=%s").Output()
		if err == nil {
			lastMsg = strings.TrimSpace(string(out))
		}
		out, err = exec.Command("git", "log", "-1", "--format=%cr").Output()
		if err == nil {
			lastTime = strings.TrimSpace(string(out))
		}

		out, err = exec.Command("git", "rev-list", "--count", "--since=30.days", "HEAD").Output()
		if err == nil {
			commitCount, _ = strconv.Atoi(strings.TrimSpace(string(out)))
		}

		out, err = exec.Command("git", "shortlog", "-sn", "--since=30.days", "HEAD").Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					contribCount++
				}
			}
		}
	}

	fmt.Printf("%s project", lang)
	if framework != "" {
		fmt.Printf(" [%s]", framework)
	}
	fmt.Printf(" | %s\n", extOutput)

	if entryFile != "" {
		fmt.Printf("Entry: %s (%s)\n", entryFile, entrySize)
	}

	if depCount > 0 {
		fmt.Printf("Deps: %s (%d %s)\n", manifestFile, depCount, depType)
	}

	if isGit {
		fmt.Printf("Last: %s %q (%s)\n", lastHash, lastMsg, lastTime)
		fmt.Printf("Activity: %d commits / 30d, %d contributors\n", commitCount, contribCount)
	}
}
