package main

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"google.golang.org/api/youtube/v3"
)

func (a *App) doLs(args []string) {
	longFormat := false
	for _, arg := range args {
		if arg == "-l" {
			longFormat = true
		}
	}

	chID, plID := parsePath(a.cwd)
	switch {
	case chID == "" && plID == "":
		a.listMyChannels()
	case plID == "":
		a.listChannel(chID, longFormat)
	default:
		a.listPlaylistItems(plID, longFormat)
	}
}

func (a *App) listMyChannels() {
	svc := a.service
	if a.oauthService != nil {
		svc = a.oauthService
	}

	call := svc.Channels.List([]string{"snippet"}).Mine(true).MaxResults(50)
	channels, err := call.Do()
	if err != nil {
		fmt.Printf("Error fetching your channels: %v\n", err)
		return
	}

	fmt.Printf("%-3s  %-38s %-15s %s\n", "T", "ID", "INFO", "TITLE")
	fmt.Println(strings.Repeat("-", 90))
	a.entries = make([]prompt.Suggest, 0, len(channels.Items))
	a.playlistEntries = nil
	a.videoEntries = nil
	for _, ch := range channels.Items {
		title := ch.Snippet.Title
		fmt.Printf("%sd%s  %-38s %-15s %s\n",
			colorBlue, colorReset,
			ch.Id, "my channel", title)
		a.entries = append(a.entries, prompt.Suggest{
			Text:        ch.Id,
			Description: title,
		})
	}
}

func (a *App) listChannel(chID string, longFormat bool) {
	svc := a.readService(chID)

	parts := []string{"snippet,contentDetails"}
	if longFormat {
		parts = []string{"snippet,contentDetails,status"}
	}

	plCall := svc.Playlists.List(parts).
		ChannelId(chID).
		MaxResults(50)
	playlists, err := plCall.Do()
	if err != nil {
		fmt.Printf("Error fetching playlists: %v\n", err)
		return
	}

	chCall := svc.Channels.List([]string{"contentDetails"}).Id(chID)
	channels, err := chCall.Do()
	if err != nil {
		fmt.Printf("Error fetching channel: %v\n", err)
		return
	}
	if len(channels.Items) == 0 {
		fmt.Printf("Channel not found: %s\n", chID)
		fmt.Println("Check your YOUTUBE_CHANNEL_ID in ~/.ysh/.env")
		return
	}

	var videoItems []*youtube.PlaylistItem
	uploadsID := channels.Items[0].ContentDetails.RelatedPlaylists.Uploads
	if uploadsID != "" {
		videoParts := []string{"snippet,contentDetails"}
		if longFormat {
			videoParts = []string{"snippet,contentDetails,status"}
		}
		items, err := svc.PlaylistItems.List(videoParts).
			PlaylistId(uploadsID).
			MaxResults(50).
			Do()
		if err == nil {
			videoItems = items.Items
		}
	}

	if len(playlists.Items) == 0 && len(videoItems) == 0 {
		fmt.Println("No playlists or videos found for this channel.")
		return
	}

	if longFormat {
		fmt.Printf("%-3s  %-10s %-38s %-15s %s\n", "T", "MODE", "ID", "INFO", "TITLE")
	} else {
		fmt.Printf("%-3s  %-38s %-15s %s\n", "T", "ID", "INFO", "TITLE")
	}
	fmt.Println(strings.Repeat("-", 90))
	a.playlistEntries = make([]prompt.Suggest, 0, len(playlists.Items))
	a.videoEntries = make([]prompt.Suggest, 0, len(videoItems))
	a.entries = make([]prompt.Suggest, 0, len(playlists.Items)+len(videoItems))
	for _, pl := range playlists.Items {
		info := fmt.Sprintf("%d items", pl.ContentDetails.ItemCount)
		if longFormat && pl.Status != nil {
			mode := pl.Status.PrivacyStatus
			fmt.Printf("%sd%s  %-10s %-38s %-15s %s\n",
				colorCyan, colorReset,
				mode, pl.Id, info, pl.Snippet.Title)
		} else {
			fmt.Printf("%sd%s  %-38s %-15s %s\n",
				colorCyan, colorReset,
				pl.Id, info, pl.Snippet.Title)
		}
		s := prompt.Suggest{Text: pl.Id, Description: "[PL] " + pl.Snippet.Title}
		a.playlistEntries = append(a.playlistEntries, s)
		a.entries = append(a.entries, s)
	}
	for _, item := range videoItems {
		videoID := item.ContentDetails.VideoId
		title := item.Snippet.Title
		if longFormat {
			mode := "public"
			if item.Status != nil && item.Status.PrivacyStatus != "" {
				mode = item.Status.PrivacyStatus
			}
			fmt.Printf("%s-%s  %-10s %-38s %-15s %s\n",
				colorGreen, colorReset,
				mode, videoID, "video", title)
		} else {
			fmt.Printf("%s-%s  %-38s %-15s %s\n",
				colorGreen, colorReset,
				videoID, "video", title)
		}
		s := prompt.Suggest{Text: videoID, Description: title}
		a.videoEntries = append(a.videoEntries, s)
		a.entries = append(a.entries, s)
	}
}

func (a *App) listPlaylistItems(playlistID string, longFormat bool) {
	chID, _ := parsePath(a.cwd)
	svc := a.readService(chID)

	parts := []string{"snippet,contentDetails"}
	if longFormat {
		parts = []string{"snippet,contentDetails,status"}
	}

	call := svc.PlaylistItems.List(parts).
		PlaylistId(playlistID).
		MaxResults(50)
	items, err := call.Do()
	if err != nil {
		fmt.Printf("Error fetching playlist items: %v\n", err)
		return
	}

	if len(items.Items) == 0 {
		fmt.Println("This playlist is empty.")
		return
	}

	if longFormat {
		fmt.Printf("%-15s %-10s %-10s %s\n", "VIDEO_ID", "MODE", "POSITION", "TITLE")
	} else {
		fmt.Printf("%-15s %-10s %s\n", "VIDEO_ID", "POSITION", "TITLE")
	}
	fmt.Println(strings.Repeat("-", 90))
	a.videoEntries = make([]prompt.Suggest, 0, len(items.Items))
	a.entries = make([]prompt.Suggest, 0, len(items.Items))
	for _, item := range items.Items {
		videoID := item.ContentDetails.VideoId
		title := item.Snippet.Title
		pos := item.Snippet.Position
		if longFormat {
			mode := "public"
			if item.Status != nil && item.Status.PrivacyStatus != "" {
				mode = item.Status.PrivacyStatus
			}
			fmt.Printf("%-15s %-10s %-10d %s\n", videoID, mode, pos, title)
		} else {
			fmt.Printf("%-15s %-10d %s\n", videoID, pos, title)
		}
		s := prompt.Suggest{Text: videoID, Description: title}
		a.videoEntries = append(a.videoEntries, s)
		a.entries = append(a.entries, s)
	}
}
