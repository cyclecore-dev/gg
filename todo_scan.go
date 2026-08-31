package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var sourceExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".rs": true, ".java": true,
	".rb": true, ".php": true, ".c": true, ".cpp": true, ".h": true, ".swift": true,
	".kt": true, ".ex": true, ".sh": true, ".yaml": true, ".yml": true, ".toml": true,
	".sql": true, ".html": true, ".css": true, ".md": true,
}

func handleTodo() {
	root := "."
	if len(os.Args) >= 3 {
		arg := os.Args[2]
		if arg == "--help" || arg == "-h" {
			printTodoUsage()
			return
		}
		root = arg
	}

	skippedDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true, "__pycache__": true,
		".venv": true, "venv": true, "dist": true, "build": true, ".next": true,
		"target": true, ".idea": true, ".vscode": true,
	}

	type TodoItem struct {
		File   string
		Line   int
		Marker string
		Text   string
	}

	var items []TodoItem
	uniqueFiles := make(map[string]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if skippedDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if !todoIsSource(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			marker, text := todoExtractMarker(line)
			if marker != "" {
				items = append(items, TodoItem{
					File:   path,
					Line:   lineNum,
					Marker: marker,
					Text:   text,
				})
				uniqueFiles[path] = true
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("  0 TODOs found")
		fmt.Println("  Token cost: ~3")
		return
	}

	groups := map[string][]TodoItem{}
	for _, item := range items {
		groups[item.Marker] = append(groups[item.Marker], item)
	}

	fmt.Printf("  %d TODOs in %d files\n", len(items), len(uniqueFiles))

	markerOrder := []string{"TODO", "FIXME", "HACK", "XXX"}
	totalPrinted := 0
	maxItems := 20

	for _, marker := range markerOrder {
		list := groups[marker]
		if len(list) == 0 {
			continue
		}

		fmt.Printf("\n  %s (%d):\n", marker, len(list))

		for _, item := range list {
			if totalPrinted >= maxItems {
				fmt.Printf("    ... and %d more\n", len(items)-totalPrinted)
				break
			}
			fmt.Printf("    %s:%d       %s\n", item.File, item.Line, item.Text)
			totalPrinted++
		}
	}
	
	fmt.Println("\n  Token cost: ~25")
}

func printTodoUsage() {
	fmt.Println("Usage: gg todo [path]")
	fmt.Println("  Scans source files for TODO/FIXME/HACK/XXX markers.")
	fmt.Println("  If path is omitted, scans current directory.")
}

func todoExtractMarker(line string) (string, string) {
	markers := []string{"TODO", "FIXME", "HACK", "XXX"}
	for _, m := range markers {
		idx := strings.Index(line, m)
		if idx != -1 {
			rest := line[idx+len(m):]
			rest = strings.TrimLeft(rest, ": ")
			return m, rest
		}
	}
	return "", ""
}

func todoIsSource(path string) bool {
	return sourceExtensions[filepath.Ext(path)]
}
