package main

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/youtube/v3"
)

func (a *App) ensureOAuth() bool {
	if a.oauthService != nil {
		return true
	}
	clientID := os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		fmt.Println("OAuth credentials not set.")
		fmt.Println("Add YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET to ~/.ysh/.env")
		fmt.Println("See: https://console.cloud.google.com/apis/credentials")
		return false
	}
	ctx := context.Background()
	svc, err := getOAuthService(ctx, clientID, clientSecret)
	if err != nil {
		fmt.Printf("OAuth authentication failed: %v\n", err)
		return false
	}
	a.oauthService = svc
	return true
}

func (a *App) doCp(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: cp <video_id> <playlist_id>")
		return
	}
	if !a.ensureOAuth() {
		return
	}
	videoID := args[0]
	playlistID := args[1]

	item := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}
	_, err := a.oauthService.PlaylistItems.Insert([]string{"snippet"}, item).Do()
	if err != nil {
		fmt.Printf("Error adding video: %v\n", err)
		return
	}
	fmt.Printf("Added video %s to playlist %s\n", videoID, playlistID)
}

func (a *App) doMkdir(args []string) {
	mode := "unlisted"
	title := ""

	i := 0
	for i < len(args) {
		if args[i] == "-m" && i+1 < len(args) {
			mode = args[i+1]
			i += 2
		} else {
			title = args[i]
			i++
		}
	}

	if title == "" {
		fmt.Println("Usage: mkdir [-m public|unlisted|private] <title>")
		return
	}

	switch mode {
	case "public", "unlisted", "private":
	default:
		fmt.Printf("Invalid mode: %s (use: public, unlisted, private)\n", mode)
		return
	}

	if !a.ensureOAuth() {
		return
	}

	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:     title,
			ChannelId: a.homeChannel,
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: mode,
		},
	}
	created, err := a.oauthService.Playlists.Insert([]string{"snippet,status"}, playlist).Do()
	if err != nil {
		fmt.Printf("Error creating playlist: %v\n", err)
		return
	}
	fmt.Printf("Created playlist: %sd%s %s (%s)\n",
		colorCyan, colorReset, created.Id, created.Snippet.Title)
	fmt.Printf("  Mode: %s\n", mode)
}

func (a *App) doChmod(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: chmod <public|unlisted|private> <playlist_id>")
		return
	}
	mode := args[0]
	playlistID := args[1]

	switch mode {
	case "public", "unlisted", "private":
	default:
		fmt.Printf("Invalid mode: %s (use: public, unlisted, private)\n", mode)
		return
	}

	if !a.ensureOAuth() {
		return
	}

	playlist := &youtube.Playlist{
		Id: playlistID,
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: mode,
		},
	}
	_, err := a.oauthService.Playlists.Update([]string{"status"}, playlist).Do()
	if err != nil {
		fmt.Printf("Error updating playlist: %v\n", err)
		return
	}
	fmt.Printf("Set playlist %s to %s\n", playlistID, mode)
}

func (a *App) doRm(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: rm <video_id> <playlist_id>")
		return
	}
	if !a.ensureOAuth() {
		return
	}
	videoID := args[0]
	playlistID := args[1]

	// Find the playlist item ID for this video in the playlist
	items, err := a.oauthService.PlaylistItems.List([]string{"snippet,contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(50).
		Do()
	if err != nil {
		fmt.Printf("Error fetching playlist items: %v\n", err)
		return
	}

	var itemID string
	for _, item := range items.Items {
		if item.ContentDetails.VideoId == videoID {
			itemID = item.Id
			break
		}
	}
	if itemID == "" {
		fmt.Printf("Video %s not found in playlist %s\n", videoID, playlistID)
		return
	}

	err = a.oauthService.PlaylistItems.Delete(itemID).Do()
	if err != nil {
		fmt.Printf("Error removing video: %v\n", err)
		return
	}
	fmt.Printf("Removed video %s from playlist %s\n", videoID, playlistID)
}

func (a *App) doMv(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: mv <video_id> <src_playlist_id> <dst_playlist_id>")
		return
	}
	if !a.ensureOAuth() {
		return
	}
	videoID := args[0]
	srcPlaylistID := args[1]
	dstPlaylistID := args[2]

	// Add to destination playlist
	item := &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: dstPlaylistID,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoID,
			},
		},
	}
	_, err := a.oauthService.PlaylistItems.Insert([]string{"snippet"}, item).Do()
	if err != nil {
		fmt.Printf("Error adding video to destination: %v\n", err)
		return
	}

	// Remove from source playlist
	items, err := a.oauthService.PlaylistItems.List([]string{"snippet,contentDetails"}).
		PlaylistId(srcPlaylistID).
		MaxResults(50).
		Do()
	if err != nil {
		fmt.Printf("Error fetching source playlist items: %v\n", err)
		return
	}

	var itemID string
	for _, it := range items.Items {
		if it.ContentDetails.VideoId == videoID {
			itemID = it.Id
			break
		}
	}
	if itemID == "" {
		fmt.Printf("Video %s not found in source playlist %s (added to destination only)\n", videoID, srcPlaylistID)
		return
	}

	err = a.oauthService.PlaylistItems.Delete(itemID).Do()
	if err != nil {
		fmt.Printf("Error removing video from source: %v\n", err)
		return
	}
	fmt.Printf("Moved video %s from %s to %s\n", videoID, srcPlaylistID, dstPlaylistID)
}

func (a *App) doRmdir(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: rmdir <playlist_id>")
		return
	}
	if !a.ensureOAuth() {
		return
	}
	playlistID := args[0]

	err := a.oauthService.Playlists.Delete(playlistID).Do()
	if err != nil {
		fmt.Printf("Error deleting playlist: %v\n", err)
		return
	}
	fmt.Printf("Deleted playlist %s\n", playlistID)
}