package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func handleDiff() {
	var diffArgs []string
	mode := "default"

	for _, arg := range os.Args[2:] {
		switch arg {
		case "--staged":
			mode = "staged"
		case "--all":
			mode = "all"
		case "--help", "-h":
			fmt.Println("Usage: gg diff [--staged|--all|<ref>]")
			fmt.Println("")
			fmt.Println("Modes:")
			fmt.Println("  (default)    Unstaged changes")
			fmt.Println("  --staged     Staged changes only")
			fmt.Println("  --all        All uncommitted changes")
			fmt.Println("  HEAD~3       Diff against specific ref")
			fmt.Println("")
			fmt.Println("Output:")
			fmt.Println("  3 files | +47 -12")
			fmt.Println("    M src/auth.go (+30 -8)")
			fmt.Println("    A src/middleware.go (+15 -0)")
			return
		default:
			diffArgs = append(diffArgs, arg)
		}
	}

	type FileStat struct {
		Add  int
		Del  int
		Type string
		Old  string
		New  string
	}

	stats := make(map[string]*FileStat)

	runStat := func(args []string) {
		cmd := exec.Command("git", append([]string{"diff", "--numstat"}, args...)...)
		out, err := cmd.Output()
		if err != nil {
			return
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Split(line, "\t")
			if len(parts) < 3 {
				continue
			}
			add, _ := strconv.Atoi(parts[0])
			del, _ := strconv.Atoi(parts[1])
			key := parts[2]
			if _, ok := stats[key]; !ok {
				stats[key] = &FileStat{}
			}
			stats[key].Add += add
			stats[key].Del += del
			if len(parts) == 4 {
				stats[key].Old = parts[2]
				stats[key].New = parts[3]
				stats[key].Type = "R"
			}
		}
	}

	runStatus := func(args []string) {
		cmd := exec.Command("git", append([]string{"diff", "--name-status"}, args...)...)
		out, err := cmd.Output()
		if err != nil {
			return
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Split(line, "\t")
			if len(parts) < 2 {
				continue
			}
			key := parts[1]
			if _, ok := stats[key]; !ok {
				stats[key] = &FileStat{}
			}
			stats[key].Type = parts[0]
			if len(parts) == 3 {
				stats[key].Old = parts[1]
				stats[key].New = parts[2]
				stats[key].Type = parts[0]
			}
		}
	}

	if mode == "staged" {
		runStat([]string{"--cached"})
		runStatus([]string{"--cached"})
	} else if mode == "all" {
		runStat([]string{})
		runStatus([]string{})
		runStat([]string{"--cached"})
		runStatus([]string{"--cached"})
	} else {
		if len(diffArgs) > 0 {
			runStat(diffArgs)
			runStatus(diffArgs)
		} else {
			runStat([]string{})
			runStatus([]string{})
		}
	}

	if len(stats) == 0 {
		fmt.Println("No changes")
		return
	}

	totalAdd := 0
	totalDel := 0
	for _, s := range stats {
		totalAdd += s.Add
		totalDel += s.Del
	}

	type fileEntry struct {
		path string
		stat *FileStat
	}
	var files []fileEntry
	for path, s := range stats {
		files = append(files, fileEntry{path, s})
	}

	limit := 30
	shown := files
	if len(shown) > limit {
		shown = shown[:limit]
	}

	fmt.Printf("%d files | +%d -%d\n", len(stats), totalAdd, totalDel)
	for _, f := range shown {
		if f.stat.Type == "R" {
			fmt.Printf("  R %s → %s (+%d -%d)\n", f.stat.Old, f.stat.New, f.stat.Add, f.stat.Del)
		} else {
			fmt.Printf("  %s %s (+%d -%d)\n", f.stat.Type, f.path, f.stat.Add, f.stat.Del)
		}
	}
	if len(stats) > limit {
		fmt.Printf("  ... and %d more files\n", len(stats)-limit)
	}
}
