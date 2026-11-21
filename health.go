package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type HealthReport struct {
	Version    string `json:"version"`
	BinaryPath string `json:"binary_path"`
	GitOK      bool   `json:"git_ok"`
	GitVersion string `json:"git_version,omitempty"`
	InRepo     bool   `json:"in_repo"`
	ConfigOK   bool   `json:"config_ok"`
	Tier       string `json:"tier,omitempty"`
	Provider   string `json:"provider,omitempty"`
	UptimeOK   bool   `json:"uptime_ok"`
	Timestamp  string `json:"timestamp"`
	Message    string `json:"message"`
}

func handleHealth() {
	report := buildHealthReport()
	if isJSONRequested() {
		b, _ := json.MarshalIndent(report, "", "  ")
		fmt.Println(string(b))
		return
	}
	fmt.Printf("gg %s | git:%v | repo:%v | config:%v | %s\n",
		report.Version, report.GitOK, report.InRepo, report.ConfigOK, report.Message)
}

func handleEnsure() {
	minVer := "0.0.0"
	for i, a := range os.Args {
		if a == "--min-version" && i+1 < len(os.Args) {
			minVer = os.Args[i+1]
		}
	}
	report := buildHealthReport()
	if !report.GitOK {
		fmt.Println("FAIL: git not available")
		os.Exit(1)
	}
	if compareVersions(report.Version, minVer) < 0 {
		fmt.Printf("FAIL: gg %s < required %s\n", report.Version, minVer)
		fmt.Println("Upgrade: curl -fsSL https://raw.githubusercontent.com/cyclecore-dev/gg/main/gg.sh | sh")
		os.Exit(1)
	}
	fmt.Printf("OK: gg %s (git ready)\n", report.Version)
}

func buildHealthReport() HealthReport {
	exe, _ := os.Executable()
	if exe == "" {
		exe = "gg"
	}
	gitOK := false
	gitVer := ""
	if out, err := exec.Command("git", "--version").Output(); err == nil {
		gitOK = true
		gitVer = strings.TrimSpace(string(out))
	}
	inRepo := false
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err == nil {
		inRepo = true
	}
	cfgOK := false
	tier, provider := "", ""
	if cfg, err := loadConfig(); err == nil {
		cfgOK = true
		tier = cfg.GG.Tier
		provider = cfg.API.Provider
	}
	msg := "healthy"
	if !gitOK {
		msg = "git missing"
	} else if !inRepo {
		msg = "not in a git repo"
	}
	return HealthReport{
		Version:    version,
		BinaryPath: exe,
		GitOK:      gitOK,
		GitVersion: gitVer,
		InRepo:     inRepo,
		ConfigOK:   cfgOK,
		Tier:       tier,
		Provider:   provider,
		UptimeOK:   true,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Message:    msg,
	}
}

func compareVersions(a, b string) int {
	ap := strings.Split(strings.TrimPrefix(a, "v"), ".")
	bp := strings.Split(strings.TrimPrefix(b, "v"), ".")
	for i := 0; i < 3; i++ {
		av, bv := 0, 0
		if i < len(ap) {
			fmt.Sscanf(ap[i], "%d", &av)
		}
		if i < len(bp) {
			fmt.Sscanf(bp[i], "%d", &bv)
		}
		if av != bv {
			if av > bv {
				return 1
			}
			return -1
		}
	}
	return 0
}
