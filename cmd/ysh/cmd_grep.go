package main

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

func (a *App) doGrep(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: grep <keyword>")
		return
	}
	keyword := args[0]
	lower := strings.ToLower(keyword)

	if len(a.entries) == 0 {
		fmt.Println("No entries to filter. Run ls first.")
		return
	}

	playlistSet := make(map[string]bool, len(a.playlistEntries))
	for _, s := range a.playlistEntries {
		playlistSet[s.Text] = true
	}
	videoSet := make(map[string]bool, len(a.videoEntries))
	for _, s := range a.videoEntries {
		videoSet[s.Text] = true
	}

	var matched []prompt.Suggest
	for _, entry := range a.entries {
		if strings.Contains(strings.ToLower(entry.Text), lower) ||
			strings.Contains(strings.ToLower(entry.Description), lower) {
			matched = append(matched, entry)
		}
	}

	if len(matched) == 0 {
		fmt.Printf("No matches for %q\n", keyword)
		return
	}

	chID, plID := parsePath(a.cwd)

	switch {
	case chID == "" && plID == "":
		fmt.Printf("%-3s  %-38s %-15s %s\n", "T", "ID", "INFO", "TITLE")
		fmt.Println(strings.Repeat("-", 90))
		for _, e := range matched {
			fmt.Printf("%sd%s  %-38s %-15s %s\n",
				colorBlue, colorReset, e.Text, "my channel", e.Description)
		}
	case plID == "":
		fmt.Printf("%-3s  %-38s %-15s %s\n", "T", "ID", "INFO", "TITLE")
		fmt.Println(strings.Repeat("-", 90))
		for _, e := range matched {
			if playlistSet[e.Text] {
				desc := strings.TrimPrefix(e.Description, "[PL] ")
				fmt.Printf("%sd%s  %-38s %-15s %s\n",
					colorCyan, colorReset, e.Text, "playlist", desc)
			} else if videoSet[e.Text] {
				fmt.Printf("%s-%s  %-38s %-15s %s\n",
					colorGreen, colorReset, e.Text, "video", e.Description)
			}
		}
	default:
		fmt.Printf("%-15s %-10s %s\n", "VIDEO_ID", "POSITION", "TITLE")
		fmt.Println(strings.Repeat("-", 90))
		for _, e := range matched {
			fmt.Printf("%-15s %-10s %s\n",
				e.Text, "-", e.Description)
		}
	}
}
