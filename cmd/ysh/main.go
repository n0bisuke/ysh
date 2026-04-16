package main

import (
	"context"
	"fmt"
	"log"
	"os"
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
	service         *youtube.Service // API-key based (read-only)
	oauthService    *youtube.Service // OAuth-based (read-write, lazy init)
	cwd             string           // current path: "//" | "/UCxxxx" | "/UCxxxx/PLxxxx"
	homeChannel     string           // own channel ID from config
	entries         []prompt.Suggest // all ls results for cd completion
	videoEntries    []prompt.Suggest // video IDs for cp first arg
	playlistEntries []prompt.Suggest // playlist IDs for cp second arg / cd
	history         []string         // command history for up/down arrow
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
	fmt.Println("Loading...")
	app.doLs()
	fmt.Println()
	if hasOAuth {
		fmt.Println("Mode: full (read + write)")
		fmt.Println("Commands: ls, cd, cat, cp, mkdir, chmod, rm, mv, rmdir, whoami, open, pwd, exit")
	} else {
		fmt.Println("Mode: read-only (write commands unavailable)")
		fmt.Println("Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET in ~/.ysh/.env to enable write operations.")
		fmt.Println("Commands: ls, cd, cat, whoami, open, pwd, exit")
	}
	p := prompt.New(
		app.executor,
		app.completer,
		prompt.OptionTitle("ysh: YouTube Shell"),
		prompt.OptionPrefix(app.promptStr()),
		prompt.OptionLivePrefix(app.livePrefix),
		prompt.OptionHistory(app.history),
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

	a.history = append(a.history, in)

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
	case "cat":
		a.doCat(args)
	case "cp":
		a.doCp(args)
	case "mkdir":
		a.doMkdir(args)
	case "chmod":
		a.doChmod(args)
	case "rm":
		a.doRm(args)
	case "whoami":
		a.doWhoami()
	case "mv":
		a.doMv(args)
	case "rmdir":
		a.doRmdir(args)
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

	prefix := ""
	if len(parts) > 1 {
		prefix = parts[len(parts)-1]
	}
	if len(text) > 0 && text[len(text)-1] == ' ' {
		prefix = ""
	}

	switch cmd {
	case "cd":
		suggestions := []prompt.Suggest{
			{Text: "..", Description: "Go up to parent directory"},
			{Text: "~", Description: "Go to your channel"},
		}
		suggestions = append(suggestions, a.playlistEntries...)
		return filterByPrefix(suggestions, prefix)
	case "cp":
		if len(parts) >= 2 && prefix == "" {
			return a.playlistEntries
		}
		if len(parts) >= 2 && prefix != "" {
			return filterByPrefix(a.playlistEntries, prefix)
		}
		return filterByPrefix(a.videoEntries, prefix)
	case "rm":
		if len(parts) >= 2 && prefix == "" {
			return a.playlistEntries
		}
		if len(parts) >= 2 && prefix != "" {
			return filterByPrefix(a.playlistEntries, prefix)
		}
		return filterByPrefix(a.videoEntries, prefix)
	case "mv":
		switch len(parts) {
		case 2:
			if prefix == "" {
				return a.playlistEntries
			}
			return filterByPrefix(a.playlistEntries, prefix)
		case 3:
			if prefix == "" {
				return a.playlistEntries
			}
			return filterByPrefix(a.playlistEntries, prefix)
		default:
			return filterByPrefix(a.videoEntries, prefix)
		}
	case "rmdir":
		return filterByPrefix(a.playlistEntries, prefix)
	case "open":
		all := make([]prompt.Suggest, 0, len(a.videoEntries)+len(a.playlistEntries))
		all = append(all, a.videoEntries...)
		all = append(all, a.playlistEntries...)
		return filterByPrefix(all, prefix)
	case "cat":
		return filterByPrefix(a.videoEntries, prefix)
	case "mkdir":
		return []prompt.Suggest{{Text: "-m", Description: "Set mode (public/unlisted/private)"}}
	case "chmod":
		return a.playlistEntries
	case "whoami":
		return []prompt.Suggest{}
	case "exit", "quit":
		return []prompt.Suggest{}
	default:
		return []prompt.Suggest{
			{Text: "ls", Description: "List items in current directory"},
			{Text: "cd", Description: "Change directory"},
			{Text: "pwd", Description: "Print working directory"},
			{Text: "open", Description: "Open video/playlist in browser"},
			{Text: "cat", Description: "Show video details"},
			{Text: "cp", Description: "Copy video into playlist"},
			{Text: "mkdir", Description: "Create playlist"},
			{Text: "chmod", Description: "Change playlist privacy"},
			{Text: "rm", Description: "Remove video from playlist"},
			{Text: "whoami", Description: "Show your channel info"},
			{Text: "mv", Description: "Move video between playlists"},
			{Text: "rmdir", Description: "Delete playlist"},
			{Text: "exit", Description: "Exit the shell"},
		}
	}
}