package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"google.golang.org/api/youtube/v3"
)

func (a *App) doLs() {
	chID, plID := parsePath(a.cwd)
	switch {
	case chID == "" && plID == "":
		a.listSubscriptions()
	case plID == "":
		a.listChannel(chID)
	default:
		a.listPlaylistItems(plID)
	}
}

func (a *App) listSubscriptions() {
	hasOAuth := os.Getenv("YOUTUBE_CLIENT_ID") != "" && os.Getenv("YOUTUBE_CLIENT_SECRET") != ""

	if !hasOAuth {
		fmt.Println("Subscriptions require OAuth credentials.")
		fmt.Println("Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET in ~/.ysh/.env to list your subscriptions.")
		fmt.Println()
		fmt.Println("You can still navigate directly: cd UCxxxxxxxxxxxxxxxx")
		return
	}

	call := a.service.Subscriptions.List([]string{"snippet"}).
		Mine(true).
		MaxResults(50)
	subs, err := call.Do()
	if err != nil {
		fmt.Printf("Error fetching subscriptions: %v\n", err)
		return
	}

	fmt.Printf("%-3s  %-38s %-30s %s\n", "T", "ID", "TITLE", "INFO")
	fmt.Println(strings.Repeat("-", 80))
	a.entries = make([]prompt.Suggest, 0, len(subs.Items))
	a.playlistEntries = nil
	a.videoEntries = nil
	for _, sub := range subs.Items {
		chID := sub.Snippet.ResourceId.ChannelId
		title := sub.Snippet.Title
		fmt.Printf("%sd%s  %-38s %-30s channel\n",
			colorBlue, colorReset,
			chID, truncate(title, 29))
		a.entries = append(a.entries, prompt.Suggest{
			Text:        chID,
			Description: title,
		})
	}
}

func (a *App) listChannel(chID string) {
	plCall := a.service.Playlists.List([]string{"snippet,contentDetails"}).
		ChannelId(chID).
		MaxResults(50)
	playlists, err := plCall.Do()
	if err != nil {
		fmt.Printf("Error fetching playlists: %v\n", err)
		return
	}

	chCall := a.service.Channels.List([]string{"contentDetails"}).Id(chID)
	channels, err := chCall.Do()
	var videoItems []*youtube.PlaylistItem
	if err == nil && len(channels.Items) > 0 {
		uploadsID := channels.Items[0].ContentDetails.RelatedPlaylists.Uploads
		if uploadsID != "" {
			items, err := a.service.PlaylistItems.List([]string{"snippet,contentDetails"}).
				PlaylistId(uploadsID).
				MaxResults(50).
				Do()
			if err == nil {
				videoItems = items.Items
			}
		}
	}

	fmt.Printf("%-3s  %-38s %-30s %s\n", "T", "ID", "TITLE", "INFO")
	fmt.Println(strings.Repeat("-", 80))
	a.playlistEntries = make([]prompt.Suggest, 0, len(playlists.Items))
	a.videoEntries = make([]prompt.Suggest, 0, len(videoItems))
	a.entries = make([]prompt.Suggest, 0, len(playlists.Items)+len(videoItems))
	for _, pl := range playlists.Items {
		fmt.Printf("%sd%s  %-38s %-30s %d items\n",
			colorCyan, colorReset,
			pl.Id, truncate(pl.Snippet.Title, 29), pl.ContentDetails.ItemCount)
		s := prompt.Suggest{Text: pl.Id, Description: "[PL] " + pl.Snippet.Title}
		a.playlistEntries = append(a.playlistEntries, s)
		a.entries = append(a.entries, s)
	}
	for _, item := range videoItems {
		videoID := item.ContentDetails.VideoId
		title := item.Snippet.Title
		fmt.Printf("%s-%s  %-38s %-30s video\n",
			colorGreen, colorReset,
			videoID, truncate(title, 29))
		s := prompt.Suggest{Text: videoID, Description: title}
		a.videoEntries = append(a.videoEntries, s)
		a.entries = append(a.entries, s)
	}
}

func (a *App) listPlaylistItems(playlistID string) {
	call := a.service.PlaylistItems.List([]string{"snippet,contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(50)
	items, err := call.Do()
	if err != nil {
		fmt.Printf("Error fetching playlist items: %v\n", err)
		return
	}

	fmt.Printf("%-15s %-45s %s\n", "VIDEO_ID", "TITLE", "POSITION")
	fmt.Println(strings.Repeat("-", 80))
	a.videoEntries = make([]prompt.Suggest, 0, len(items.Items))
	a.entries = make([]prompt.Suggest, 0, len(items.Items))
	for _, item := range items.Items {
		videoID := item.ContentDetails.VideoId
		title := item.Snippet.Title
		pos := item.Snippet.Position
		fmt.Printf("%-15s %-45s %d\n", videoID, truncate(title, 44), pos)
		s := prompt.Suggest{Text: videoID, Description: title}
		a.videoEntries = append(a.videoEntries, s)
		a.entries = append(a.entries, s)
	}
}