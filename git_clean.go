package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func handleClean() {
	force := false
	if len(os.Args) > 2 {
		flag := os.Args[2]
		if flag == "--force" || flag == "-f" {
			force = true
		} else if flag == "--help" || flag == "-h" {
			printCleanUsage()
			return
		}
	}

	if force {
		cleanForce()
	} else {
		cleanDryRun()
	}
}

func printCleanUsage() {
	fmt.Println("Usage: gg clean [--force|-f] [--help|-h]")
	fmt.Println("  --force, -f  Actually delete untracked files and directories")
	fmt.Println("  --help, -h   Show this help message")
}

func cleanDryRun() {
	cmd := exec.Command("git", "clean", "-dn")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running git clean: %v\n", err)
		return
	}

	content := strings.TrimSpace(string(output))
	if content == "" {
		fmt.Println("Nothing to clean.")
		return
	}

	lines := strings.Split(content, "\n")
	var items []string
	var files []string
	var dirs []string
	var totalSize int64

	for _, line := range lines {
		if strings.HasPrefix(line, "Would remove ") {
			path := strings.TrimPrefix(line, "Would remove ")
			items = append(items, path)

			info, err := os.Stat(path)
			if err == nil {
				if info.IsDir() {
					size, err := cleanGetDirSize(path)
					if err == nil {
						totalSize += size
					}
				} else {
					totalSize += info.Size()
				}
			}

			if strings.HasSuffix(path, "/") {
				dirs = append(dirs, path)
			} else {
				files = append(files, path)
			}
		}
	}

	if len(items) == 0 {
		fmt.Println("Nothing to clean.")
		return
	}

	fmt.Printf("  %d untracked items (%s total)\n", len(items), formatSize(totalSize))

	if len(items) > 10 {
		showCount := 8
		if len(items) < showCount {
			showCount = len(items)
		}
		fmt.Printf("  %s\n", strings.Join(items[:showCount], ", "))
		fmt.Printf("    ... and %d more items\n", len(items)-showCount)
	} else {
		if len(files) > 0 {
			fmt.Printf("  %d files: %s\n", len(files), strings.Join(files, ", "))
		}
		if len(dirs) > 0 {
			fmt.Printf("  %d dirs: %s\n", len(dirs), strings.Join(dirs, ", "))
		}
	}

	fmt.Println("  Use `gg clean -f` to remove | Token cost: ~15")
}

func cleanForce() {
	// Get list and sizes first to report freed space
	cmd := exec.Command("git", "clean", "-dn")
	output, err := cmd.Output()
	var totalSize int64
	var items []string
	if err == nil {
		content := strings.TrimSpace(string(output))
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Would remove ") {
				path := strings.TrimPrefix(line, "Would remove ")
				items = append(items, path)
				info, err := os.Stat(path)
				if err == nil {
					if info.IsDir() {
						size, err := cleanGetDirSize(path)
						if err == nil {
							totalSize += size
						}
					} else {
						totalSize += info.Size()
					}
				}
			}
		}
	}

	if len(items) == 0 {
		fmt.Println("  Nothing to clean.")
		return
	}

	// Now remove
	cmd = exec.Command("git", "clean", "-df")
	_, err = cmd.Output()
	if err != nil {
		fmt.Printf("  Error running git clean: %v\n", err)
		return
	}

	fmt.Printf("  ✓ cleaned %d items (%s freed)\n", len(items), formatSize(totalSize))
	fmt.Println("  Token cost: ~6")
}

func cleanGetDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
