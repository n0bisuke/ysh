package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type App struct {
	service   *youtube.Service
	cwd       string // current path, e.g. "/" or "/PLxxxx"
	channelID string
	entries   []prompt.Suggest // cached ls results for tab completion
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY is not set in .env")
	}

	channelID := os.Getenv("YOUTUBE_CHANNEL_ID")
	if channelID == "" {
		log.Fatal("YOUTUBE_CHANNEL_ID is not set in .env")
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create YouTube client: %v", err)
	}

	app := &App{
		service:   service,
		cwd:       "/",
		channelID: channelID,
	}

	fmt.Println("ysh - YouTube Shell")
	fmt.Println("Type ls, cd, or exit.")
	p := prompt.New(
		app.executor,
		app.completer,
		prompt.OptionTitle("ysh: YouTube Shell"),
		prompt.OptionPrefix(app.promptStr()),
		prompt.OptionLivePrefix(app.livePrefix),
	)
	p.Run()
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
	case "ls":
		return []prompt.Suggest{}
	case "cd":
		suggestions := []prompt.Suggest{
			{Text: "..", Description: "Go up to parent directory"},
		}
		suggestions = append(suggestions, a.entries...)
		return suggestions
	case "exit", "quit":
		return []prompt.Suggest{}
	default:
		return []prompt.Suggest{
			{Text: "ls", Description: "List items in current directory"},
			{Text: "cd", Description: "Change directory"},
			{Text: "exit", Description: "Exit the shell"},
		}
	}
}

func (a *App) doLs() {
	if a.cwd == "/" {
		a.listPlaylists()
	} else {
		playlistID := strings.TrimPrefix(a.cwd, "/")
		a.listPlaylistItems(playlistID)
	}
}

func (a *App) listPlaylists() {
	call := a.service.Playlists.List([]string{"snippet,contentDetails"}).
		ChannelId(a.channelID).
		MaxResults(50)
	playlists, err := call.Do()
	if err != nil {
		fmt.Printf("Error fetching playlists: %v\n", err)
		return
	}

	fmt.Printf("%-40s %-30s %s\n", "ID", "TITLE", "ITEMS")
	fmt.Println(strings.Repeat("-", 80))
	a.entries = make([]prompt.Suggest, 0, len(playlists.Items))
	for _, pl := range playlists.Items {
		fmt.Printf("%-40s %-30s %d\n", pl.Id, truncate(pl.Snippet.Title, 29), pl.ContentDetails.ItemCount)
		a.entries = append(a.entries, prompt.Suggest{
			Text:        pl.Id,
			Description: pl.Snippet.Title,
		})
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
	a.entries = make([]prompt.Suggest, 0, len(items.Items))
	for _, item := range items.Items {
		videoID := item.ContentDetails.VideoId
		title := item.Snippet.Title
		pos := item.Snippet.Position
		fmt.Printf("%-15s %-45s %d\n", videoID, truncate(title, 44), pos)
		a.entries = append(a.entries, prompt.Suggest{
			Text:        videoID,
			Description: title,
		})
	}
}

func (a *App) doCd(args []string) {
	if len(args) == 0 {
		a.cwd = "/"
		return
	}

	target := args[0]
	switch target {
	case "..":
		if a.cwd != "/" {
			a.cwd = "/"
		}
	case "/":
		a.cwd = "/"
	default:
		// Navigate into a playlist
		a.cwd = "/" + target
	}
}

func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-1]) + "…"
}