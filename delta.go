package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func deltaStatePath() string {
	dir := filepath.Join(getGGDir(), "delta")
	os.MkdirAll(dir, 0755)
	root := "."
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		root = strings.TrimSpace(string(out))
	}
	h := sha256.Sum256([]byte(root))
	return filepath.Join(dir, hex.EncodeToString(h[:8])+".state")
}

func handleStatusDelta() {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		fmt.Println("not a git repo")
		return
	}
	current := strings.TrimSpace(string(out))
	stateFile := deltaStatePath()

	prev := ""
	if b, err := os.ReadFile(stateFile); err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(b)), "\n---\n", 2)
		if len(parts) > 0 {
			prev = parts[0]
		}
	}

	os.WriteFile(stateFile, []byte(current+"\n---\n"+time.Now().UTC().Format(time.RFC3339)), 0644)

	if prev == "" || prev == current {
		fmt.Println("delta: no changes")
		return
	}

	prevLines := map[string]bool{}
	for _, l := range strings.Split(prev, "\n") {
		if l != "" {
			prevLines[l] = true
		}
	}
	var added, removed []string
	for _, l := range strings.Split(current, "\n") {
		if l == "" {
			continue
		}
		if !prevLines[l] {
			added = append(added, l)
		}
		delete(prevLines, l)
	}
	for l := range prevLines {
		removed = append(removed, l)
	}

	if len(added) == 0 && len(removed) == 0 {
		fmt.Println("delta: no changes")
		return
	}
	fmt.Printf("delta: +%d -%d\n", len(added), len(removed))
	for _, l := range added {
		fmt.Println("+", l)
	}
	for _, l := range removed {
		fmt.Println("-", l)
	}
}
