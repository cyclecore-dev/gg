package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func handleCache() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg cache <status|clean>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  status  Show cache size and contents")
		fmt.Println("  clean   Remove old cache entries")
		return
	}

	cacheDir := filepath.Join(getGGDir(), "cache")

	switch os.Args[2] {
	case "status":
		showCacheStatus(cacheDir)
	case "clean":
		cleanCache(cacheDir)
	default:
		fmt.Printf("Unknown cache command: %s\n", os.Args[2])
	}
}

func showCacheStatus(cacheDir string) {
	total := getCacheSize(cacheDir)
	npmSize := getCacheSize(filepath.Join(cacheDir, "npm"))
	brewSize := getCacheSize(filepath.Join(cacheDir, "brew"))

	npmCount := countFiles(filepath.Join(cacheDir, "npm"))
	brewCount := countFiles(filepath.Join(cacheDir, "brew"))

	fmt.Println("Cache Status")
	fmt.Println()
	fmt.Printf("   Total: %s\n", formatSize(total))
	fmt.Printf("   npm:   %s (%d packages)\n", formatSize(npmSize), npmCount)
	fmt.Printf("   brew:  %s (%d formulas)\n", formatSize(brewSize), brewCount)
	fmt.Println()
	fmt.Printf("   Location: %s\n", cacheDir)
}

func cleanCache(cacheDir string) {
	type entry struct {
		path  string
		mtime time.Time
		size  int64
	}
	var entries []entry

	filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			entries = append(entries, entry{path, info.ModTime(), info.Size()})
		}
		return nil
	})

	if len(entries) == 0 {
		fmt.Println("Cache is empty")
		return
	}

	// Remove entries older than 7 days
	cutoff := time.Now().AddDate(0, 0, -7)
	var freed int64
	var removed int

	for _, e := range entries {
		if e.mtime.Before(cutoff) {
			os.Remove(e.path)
			freed += e.size
			removed++
		}
	}

	if removed == 0 {
		fmt.Println("No old cache entries to clean")
	} else {
		fmt.Printf("Cleaned cache: %s freed (%d entries removed)\n", formatSize(freed), removed)
	}
}

func getCacheSize(dir string) int64 {
	var size int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func countFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// handleChat provides a human-friendly AI chat experience
