package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func handleLicense() {
	if len(os.Args) > 2 && (os.Args[2] == "--help" || os.Args[2] == "-h") {
		printLicenseUsage()
		return
	}

	foundFile, content := licenseFindFile()
	if foundFile == "" {
		pkgData, err := os.ReadFile("package.json")
		if err == nil {
			var pkg map[string]interface{}
			if json.Unmarshal(pkgData, &pkg) == nil {
				if lic, ok := pkg["license"].(string); ok {
					fmt.Printf("  %s (from package.json)\n", lic)
					fmt.Println("  Token cost: ~5")
					return
				}
			}
		}
		fmt.Println("No license detected")
		return
	}

	lines := strings.Split(content, "\n")
	if len(lines) > 20 {
		lines = lines[:20]
	}
	first20Lines := strings.Join(lines, "\n")

	licenseType := licenseDetect(first20Lines)

	var copyrightLine string
	for _, line := range lines {
		if strings.Contains(line, "Copyright") {
			copyrightLine = line
			break
		}
	}

	var copyrightHolder string
	if copyrightLine != "" {
		idx := strings.Index(copyrightLine, "Copyright")
		if idx != -1 {
			rest := copyrightLine[idx:]
			yearIdx := -1
			for i := 0; i < len(rest)-4; i++ {
				if rest[i] >= '0' && rest[i] <= '9' && rest[i+1] >= '0' && rest[i+1] <= '9' && rest[i+2] >= '0' && rest[i+2] <= '9' && rest[i+3] >= '0' && rest[i+3] <= '9' {
					yearIdx = i
					break
				}
			}
			if yearIdx != -1 {
				copyrightHolder = rest[yearIdx+4:]
				copyrightHolder = strings.TrimLeft(copyrightHolder, " ")
				if len(copyrightHolder) > 0 && copyrightHolder[0] == '(' {
					copyrightHolder = copyrightHolder[1:]
				}
				if len(copyrightHolder) > 0 && copyrightHolder[0] == ')' {
					copyrightHolder = copyrightHolder[1:]
				}
				copyrightHolder = strings.TrimSpace(copyrightHolder)
			} else {
				copyrightHolder = strings.TrimSpace(rest)
			}
		}
	}

	if licenseType == "Unknown" {
		fmt.Printf("  Unknown license (%s file exists but type not recognized)\n", foundFile)
		fmt.Println("  Token cost: ~6")
	} else {
		fmt.Printf("  %s License\n", licenseType)
		if copyrightHolder != "" {
			fmt.Printf("    Copyright (c) %s\n", copyrightHolder)
		}
		fmt.Printf("    File: %s\n", foundFile)
		fmt.Println("  Token cost: ~8")
	}
}

func licenseFindFile() (string, string) {
	licenseFiles := []string{
		"LICENSE", "LICENSE.md", "LICENSE.txt",
		"LICENCE", "LICENCE.md", "LICENCE.txt",
		"COPYING", "COPYING.md", "COPYING.txt",
	}

	for _, f := range licenseFiles {
		if _, err := os.Stat(f); err == nil {
			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			return f, string(data)
		}
	}
	return "", ""
}

func licenseDetect(content string) string {
	upper := strings.ToUpper(content)
	switch {
	case strings.Contains(upper, "MIT LICENSE"):
		return "MIT"
	case strings.Contains(upper, "APACHE LICENSE") && strings.Contains(upper, "2.0"):
		return "Apache-2.0"
	case strings.Contains(upper, "APACHE LICENSE"):
		return "Apache"
	case strings.Contains(upper, "GNU GENERAL PUBLIC LICENSE") && strings.Contains(upper, "VERSION 3"):
		return "GPL-3.0"
	case strings.Contains(upper, "GNU GENERAL PUBLIC LICENSE") && strings.Contains(upper, "VERSION 2"):
		return "GPL-2.0"
	case strings.Contains(upper, "GNU GENERAL PUBLIC LICENSE"):
		return "GPL"
	case strings.Contains(upper, "GNU LESSER GENERAL PUBLIC"):
		return "LGPL"
	case strings.Contains(upper, "BSD 3-CLAUSE") || strings.Contains(upper, "BSD 3 CLAUSE"):
		return "BSD-3-Clause"
	case strings.Contains(upper, "BSD 2-CLAUSE") || strings.Contains(upper, "BSD 2 CLAUSE"):
		return "BSD-2-Clause"
	case strings.Contains(upper, "REDISTRIBUTION AND USE IN SOURCE AND BINARY"):
		return "BSD"
	case strings.Contains(upper, "ISC LICENSE"):
		return "ISC"
	case strings.Contains(upper, "MOZILLA PUBLIC LICENSE"):
		return "MPL-2.0"
	case strings.Contains(upper, "THE UNLICENSE"):
		return "Unlicense"
	case strings.Contains(upper, "CREATIVE COMMONS"):
		return "CC"
	case strings.Contains(upper, "BOOST SOFTWARE LICENSE"):
		return "BSL-1.0"
	default:
		return "Unknown"
	}
}

func printLicenseUsage() {
	fmt.Println("Usage: gg license [--help|-h]")
	fmt.Println("Detects and summarizes the project license in one line.")
}
