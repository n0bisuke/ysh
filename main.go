package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/c-bata/go-prompt"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	colorBlue   = "\033[1;34m"
	colorGreen  = "\033[1;32m"
	colorCyan   = "\033[1;36m"
	colorReset  = "\033[0m"
)

type App struct {
	service       *youtube.Service // API-key based (read-only)
	oauthService  *youtube.Service // OAuth-based (read-write, lazy init)
	cwd           string           // current path: "//" | "/UCxxxx" | "/UCxxxx/PLxxxx"
	homeChannel   string           // own channel ID from config
	entries       []prompt.Suggest // all ls results for cd completion
	videoEntries  []prompt.Suggest // video IDs for cp first arg
	playlistEntries []prompt.Suggest // playlist IDs for cp second arg / cd
}

func main() {
	apiKey, channelID, clientID, clientSecret := runSetup()

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create YouTube client: %v", err)
	}

	app := &App{
		service:     service,
		cwd:         "/" + channelID,
		homeChannel: channelID,
	}

	hasOAuth := clientID != "" && clientSecret != ""

	fmt.Println("ysh - YouTube Shell")
	if hasOAuth {
		fmt.Println("Mode: full (read + write)")
		fmt.Println("Commands: ls, cd, cp, open, pwd, exit")
	} else {
		fmt.Println("Mode: read-only (cp command unavailable)")
		fmt.Println("Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET in ~/.ysh/.env to enable write operations.")
		fmt.Println("Commands: ls, cd, open, pwd, exit")
	}
	p := prompt.New(
		app.executor,
		app.completer,
		prompt.OptionTitle("ysh: YouTube Shell"),
		prompt.OptionPrefix(app.promptStr()),
		prompt.OptionLivePrefix(app.livePrefix),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(*prompt.Buffer) {
				fmt.Println("\nBye!")
				os.Exit(0)
			},
		}),
	)
	p.Run()
}

// parsePath breaks cwd into components.
// "//"        → ("", "")       — subscriptions list
// "/UCxxxx"   → ("UCxxxx", "") — channel root
// "/UCxxxx/PLxxxx" → ("UCxxxx", "PLxxxx") — playlist
func parsePath(cwd string) (channelID, playlistID string) {
	cwd = strings.TrimPrefix(cwd, "/")
	parts := strings.SplitN(cwd, "/", 2)
	channelID = parts[0]
	if len(parts) == 2 {
		playlistID = parts[1]
	}
	return
}

func (a *App) promptStr() string {
	return fmt.Sprintf("yt:%s $ ", a.cwd)
}

func (a *App) livePrefix() (string, bool) {
	return a.promptStr(), true
}

func (a *App) executor(in string) {
	parts := strings.Fields(in)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "ls":
		a.doLs()
	case "cd":
		a.doCd(args)
	case "pwd":
		fmt.Println(a.cwd)
	case "open":
		a.doOpen(args)
	case "cp":
		a.doCp(args)
	case "exit", "quit":
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		fmt.Printf("unknown command: %s\n", cmd)
	}
}

func (a *App) completer(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	if text == "" {
		return nil
	}
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil
	}
	cmd := parts[0]

	switch cmd {
	case "cd":
		suggestions := []prompt.Suggest{
			{Text: "..", Description: "Go up to parent directory"},
			{Text: "~", Description: "Go to your channel"},
		}
		suggestions = append(suggestions, a.playlistEntries...)
		return suggestions
	case "cp":
		// First arg: video ID, second arg: playlist ID
		if len(parts) >= 2 {
			// Already typed video ID, suggest playlists
			return a.playlistEntries
		}
		return a.videoEntries
	case "open":
		return a.videoEntries
	case "exit", "quit":
		return []prompt.Suggest{}
	default:
		return []prompt.Suggest{
			{Text: "ls", Description: "List items in current directory"},
			{Text: "cd", Description: "Change directory"},
			{Text: "pwd", Description: "Print working directory"},
			{Text: "open", Description: "Open video in browser"},
			{Text: "cp", Description: "Copy video into playlist"},
			{Text: "exit", Description: "Exit the shell"},
		}
	}
}

// ── ls ──

func (a *App) doLs() {
	chID, plID := parsePath(a.cwd)
	switch {
	case chID == "" && plID == "":
		// "//" — subscriptions list
		a.listSubscriptions()
	case plID == "":
		// "/UCxxxx" — channel root
		a.listChannel(chID)
	default:
		// "/UCxxxx/PLxxxx" — playlist
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
	// Fetch playlists
	plCall := a.service.Playlists.List([]string{"snippet,contentDetails"}).
		ChannelId(chID).
		MaxResults(50)
	playlists, err := plCall.Do()
	if err != nil {
		fmt.Printf("Error fetching playlists: %v\n", err)
		return
	}

	// Fetch uploaded videos
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
		s := prompt.Suggest{Text: pl.Id, Description: pl.Snippet.Title}
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

// ── cd ──

func (a *App) doCd(args []string) {
	if len(args) == 0 {
		a.cwd = "/" + a.homeChannel
		return
	}

	target := args[0]
	switch target {
	case "..":
		chID, plID := parsePath(a.cwd)
		switch {
		case plID != "":
			// /UCxxxx/PLxxxx → /UCxxxx
			a.cwd = "/" + chID
		case chID != "":
			// /UCxxxx → //
			a.cwd = "//"
		}
	case "~":
		a.cwd = "/" + a.homeChannel
	case "/", "//":
		a.cwd = "//"
	default:
		// Navigate into entry
		chID, plID := parsePath(a.cwd)
		if chID == "" {
			// At subscriptions → target is a channel ID
			a.cwd = "/" + target
		} else if plID == "" {
			// At channel root → target is a playlist ID
			a.cwd = "/" + chID + "/" + target
		}
		// Inside playlist: cd does nothing (no deeper level)
	}
}

// ── open ──

func (a *App) doOpen(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: open <video_id|playlist_id>")
		return
	}
	id := args[0]

	// Determine if the ID is a playlist or a video
	var url string
	if a.isPlaylistID(id) {
		url = "https://www.youtube.com/playlist?list=" + id
	} else {
		url = "https://www.youtube.com/watch?v=" + id
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Could not open browser: %v\n", url)
		fmt.Println("Open this URL manually:")
		fmt.Println(url)
		return
	}
	fmt.Printf("Opened %s\n", url)
}

func (a *App) isPlaylistID(id string) bool {
	for _, s := range a.playlistEntries {
		if s.Text == id {
			return true
		}
	}
	return false
}

// ── cp ──

func (a *App) doCp(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: cp <video_id> <playlist_id>")
		return
	}
	videoID := args[0]
	playlistID := args[1]

	if a.oauthService == nil {
		clientID := os.Getenv("YOUTUBE_CLIENT_ID")
		clientSecret := os.Getenv("YOUTUBE_CLIENT_SECRET")
		if clientID == "" || clientSecret == "" {
			fmt.Println("OAuth credentials not set.")
			fmt.Println("Add YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET to your .env file.")
			fmt.Println("See: https://console.cloud.google.com/apis/credentials")
			return
		}
		ctx := context.Background()
		svc, err := getOAuthService(ctx, clientID, clientSecret)
		if err != nil {
			fmt.Printf("OAuth authentication failed: %v\n", err)
			return
		}
		a.oauthService = svc
	}

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

// ── helpers ──

func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-1]) + "…"
}