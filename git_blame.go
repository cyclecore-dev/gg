package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type BlameGroup struct {
	Hash   string
	Author string
	Age    string
	Start  int
	End    int
}

func handleBlame() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		fmt.Println("Usage: gg blame <file> [lines]")
		return
	}
	if len(os.Args) < 3 {
		fmt.Println("Usage: gg blame <file> [lines]")
		return
	}

	file := os.Args[2]
	var lineRange string
	if len(os.Args) > 3 {
		lineRange = os.Args[3]
	}

	start, end, hasRange := parseBlameRange(lineRange)

	var gitArgs []string
	gitArgs = append(gitArgs, "blame", "--porcelain", file)
	if hasRange {
		gitArgs = append(gitArgs, "-L", fmt.Sprintf("%d,%d", start, end))
	}

	cmd := exec.Command("git", gitArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ blame failed: %v\n", err)
		return
	}

	groups := parsePorcelain(string(output))
	
	displayGroups := groups
	if len(groups) > 30 {
		displayGroups = groups[:30]
	}

	if hasRange {
		fmt.Printf("  %s:%d-%d\n", file, start, end)
	} else {
		fmt.Printf("  %s\n", file)
	}

	for _, g := range displayGroups {
		fmt.Printf("    %s %s %s L%d-%d\n", g.Hash, g.Author, g.Age, g.Start, g.End)
	}

	if len(groups) > 30 {
		fmt.Printf("  ... and %d more groups\n", len(groups)-30)
	}
	fmt.Printf("  Token cost: ~%d\n", 5+len(displayGroups)*2)
}

func parseBlameRange(lineRange string) (start, end int, hasRange bool) {
	if lineRange == "" {
		return 0, 0, false
	}

	sep := strings.ReplaceAll(lineRange, ",", "-")
	parts := strings.Split(sep, "-")
	
	var s, e int
	var err1, err2 error
	
	if len(parts) == 1 {
		s, err1 = strconv.Atoi(parts[0])
		e = s
	} else if len(parts) == 2 {
		s, err1 = strconv.Atoi(parts[0])
		e, err2 = strconv.Atoi(parts[1])
	} else {
		return 0, 0, false
	}
	
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}

	return s, e, true
}

func blameAge(timestamp int64) string {
	now := time.Now()
	age := now.Sub(time.Unix(timestamp, 0))
	
	days := int(age.Hours() / 24)
	if days < 1 {
		return "today"
	} else if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	}
	
	weeks := days / 7
	if weeks < 4 {
		return fmt.Sprintf("%dw ago", weeks)
	}
	
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%dmo ago", months)
	}
	
	years := days / 365
	return fmt.Sprintf("%dy ago", years)
}

func parsePorcelain(output string) []BlameGroup {
	var groups []BlameGroup
	var current *BlameGroup
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		
		if len(line) >= 40 && blameIsHex(line[:40]) {
			if current != nil {
				groups = append(groups, *current)
			}
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}
			hash := parts[0]
			if len(hash) > 7 {
				hash = hash[:7]
			}
			
			finalLine, err1 := strconv.Atoi(parts[2])
			numLines, err2 := strconv.Atoi(parts[3])
			if err1 != nil || err2 != nil {
				continue
			}
			
			current = &BlameGroup{
				Hash:  hash,
				Start: finalLine,
				End:   finalLine + numLines - 1,
			}
		} else if strings.HasPrefix(line, "author ") {
			if current != nil {
				current.Author = strings.TrimPrefix(line, "author ")
			}
		} else if strings.HasPrefix(line, "timestamp ") {
			if current != nil {
				tsStr := strings.TrimPrefix(line, "timestamp ")
				ts, err := strconv.ParseInt(tsStr, 10, 64)
				if err == nil {
					current.Age = blameAge(ts)
				}
			}
		}
	}
	
	if current != nil {
		groups = append(groups, *current)
	}
	
	return groups
}

func blameIsHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
