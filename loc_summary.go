package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func handleLoc() {
	root := "."
	if len(os.Args) >= 3 {
		arg := os.Args[2]
		if arg == "--help" || arg == "-h" {
			printLocUsage()
			return
		}
		if !strings.HasPrefix(arg, "-") {
			root = arg
		}
	}

	type langStats struct {
		language string
		lines    int
		files    int
	}

	stats := make(map[string]*langStats)
	totalLines := 0
	totalFiles := 0

	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true, "__pycache__": true,
		".venv": true, "venv": true, "dist": true, "build": true, ".next": true,
		"target": true, ".idea": true, ".vscode": true,
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		lang := locLang(ext)
		if lang == "" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := bytes.Count(data, []byte("\n"))

		if _, ok := stats[lang]; !ok {
			stats[lang] = &langStats{language: lang}
		}
		stats[lang].lines += lines
		stats[lang].files++

		totalLines += lines
		totalFiles++

		return nil
	})

	if err != nil {
		fatalError("failed to walk directory", err)
	}

	var sortedStats []*langStats
	for _, s := range stats {
		sortedStats = append(sortedStats, s)
	}

	sort.Slice(sortedStats, func(i, j int) bool {
		return sortedStats[i].lines > sortedStats[j].lines
	})

	fmt.Printf("  %s lines across %s files\n\n", locFmtNum(totalLines), locFmtNum(totalFiles))

	limit := 10
	if len(sortedStats) < limit {
		limit = len(sortedStats)
	}

	for i := 0; i < limit; i++ {
		s := sortedStats[i]
		percent := 0.0
		if totalLines > 0 {
			percent = float64(s.lines) / float64(totalLines) * 100
		}
		fmt.Printf("  %-12s %s  (%5.1f%%)  %s files\n",
			s.language,
			locFmtNum(s.lines),
			percent,
			locFmtNum(s.files))
	}

	fmt.Println("\n  Token cost: ~20")
}

func printLocUsage() {
	fmt.Println("Usage: gg loc [path] [--help|-h]")
	fmt.Println("  Summarize lines of code by language in the given directory.")
	fmt.Println("  If path is omitted, defaults to current directory.")
}

func locLang(ext string) string {
	langs := map[string]string{
		".go": "Go", ".rs": "Rust", ".py": "Python", ".js": "JavaScript",
		".ts": "TypeScript", ".tsx": "TSX", ".jsx": "JSX",
		".java": "Java", ".kt": "Kotlin", ".scala": "Scala",
		".c": "C", ".h": "C Header", ".cpp": "C++", ".cc": "C++",
		".cs": "C#", ".rb": "Ruby", ".php": "PHP",
		".swift": "Swift", ".m": "Obj-C",
		".html": "HTML", ".css": "CSS", ".scss": "SCSS",
		".sql": "SQL", ".sh": "Shell", ".bash": "Shell",
		".yaml": "YAML", ".yml": "YAML", ".toml": "TOML",
		".json": "JSON", ".xml": "XML",
		".md": "Markdown", ".txt": "Text",
		".ex": "Elixir", ".exs": "Elixir",
		".erl": "Erlang", ".hs": "Haskell",
		".lua": "Lua", ".r": "R", ".R": "R",
		".dart": "Dart", ".vue": "Vue", ".svelte": "Svelte",
	}
	if lang, ok := langs[ext]; ok {
		return lang
	}
	return ""
}

func locFmtNum(n int) string {
	s := strconv.Itoa(n)
	if n < 1000 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
