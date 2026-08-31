package main

import (
	"fmt"
	"os"
	"time"
)

func handleTime() {
	if len(os.Args) >= 3 {
		switch os.Args[2] {
		case "--iso":
			fmt.Println(time.Now().Format(time.RFC3339))
			return
		case "--unix":
			fmt.Println(time.Now().Unix())
			return
		case "--utc":
			fmt.Println(time.Now().UTC().Format("Mon 2006-01-02 15:04 MST"))
			return
		case "--week":
			now := time.Now()
			_, week := now.ISOWeek()
			fmt.Printf("%s (week %d, day %d)\n", now.Format("Mon 2006-01-02 15:04 MST"), week, now.YearDay())
			return
		case "--help", "-h":
			fmt.Println("Usage: gg time [--iso|--unix|--utc|--week]")
			fmt.Println()
			fmt.Println("Modes:")
			fmt.Println("  (default)    Sat 2026-08-09 14:23 MST")
			fmt.Println("  --iso        2026-08-09T14:23:45-07:00")
			fmt.Println("  --unix       1754765025")
			fmt.Println("  --utc        Sat 2026-08-09 21:23 UTC")
			fmt.Println("  --week       Sat 2026-08-09 14:23 MST (week 32, day 221)")
			return
		}
	}
	fmt.Println(time.Now().Format("Mon 2006-01-02 15:04 MST"))
}
