package main

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/api/youtube/v3"
)

func (a *App) doSort(args []string) {
	var sortBy, playlistID string
	reverse := false

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--by", "-b":
			i++
			if i < len(args) {
				sortBy = args[i]
			}
		case "--reverse", "-r":
			reverse = true
		default:
			if playlistID == "" {
				playlistID = args[i]
			}
		}
		i++
	}

	if sortBy == "" || playlistID == "" {
		fmt.Println("Usage: sort --by title|date <playlist_id> [--reverse|-r]")
		fmt.Println()
		fmt.Println("  --by title   Sort alphabetically by video title")
		fmt.Println("  --by date    Sort chronologically by published date")
		fmt.Println("  --reverse, -r  Reverse sort order")
		return
	}

	switch sortBy {
	case "title", "date":
	default:
		fmt.Printf("Invalid --by value: %s (use: title, date)\n", sortBy)
		return
	}

	if !a.ensureOAuth() {
		return
	}

	items, err := a.oauthService.PlaylistItems.List([]string{"snippet,contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(50).
		Do()
	if err != nil {
		fmt.Printf("Error fetching playlist items: %v\n", err)
		return
	}
	if len(items.Items) == 0 {
		fmt.Println("Playlist is empty.")
		return
	}

	sorted := make([]*youtube.PlaylistItem, len(items.Items))
	copy(sorted, items.Items)

	switch sortBy {
	case "title":
		sort.SliceStable(sorted, func(i, j int) bool {
			cmp := strings.Compare(
				strings.ToLower(sorted[i].Snippet.Title),
				strings.ToLower(sorted[j].Snippet.Title),
			)
			if reverse {
				return cmp > 0
			}
			return cmp < 0
		})
	case "date":
		sort.SliceStable(sorted, func(i, j int) bool {
			cmp := sorted[i].Snippet.PublishedAt < sorted[j].Snippet.PublishedAt
			if reverse {
				return !cmp
			}
			return cmp
		})
	}

	fmt.Printf("Sorting %d items by %s", len(sorted), sortBy)
	if reverse {
		fmt.Print(" (reverse)")
	}
	fmt.Println()

	for newpos, item := range sorted {
		if item.Snippet.Position == int64(newpos) {
			continue
		}
		item.Snippet.Position = int64(newpos)
		_, err := a.oauthService.PlaylistItems.Update([]string{"snippet"}, item).Do()
		if err != nil {
			fmt.Printf("Error repositioning %s: %v\n", item.ContentDetails.VideoId, err)
			continue
		}
		fmt.Printf("  %s -> position %d\n", item.ContentDetails.VideoId, newpos)
	}
	fmt.Println("Done.")
}
