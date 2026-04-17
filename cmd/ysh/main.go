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

var version = "dev"

const (
	colorBlue  = "\033[1;34m"
	colorGreen = "\033[1;32m"
	colorCyan  = "\033[1;36m"
	colorReset = "\033[0m"
)

type App struct {
	service         *youtube.Service // read operations (API key or OAuth)
	oauthService    *youtube.Service // write operations (OAuth, lazy init if service uses API key)
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

	hasOAuth := clientID != "" && clientSecret != ""
	hasAPIKey := apiKey != ""

	app := &App{
		cwd:         "/" + channelID,
		homeChannel: channelID,
	}

	if hasAPIKey {
		service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			log.Fatalf("Failed to create YouTube client: %v", err)
		}
		app.service = service
	}

	if hasOAuth {
		svc, err := getOAuthService(ctx, clientID, clientSecret)
		if err != nil {
			fmt.Printf("Warning: OAuth authentication failed: %v\n", err)
			if hasAPIKey {
				fmt.Println("Falling back to API key (read-only mode).")
			} else {
				log.Fatalf("OAuth authentication failed and no API key set. Exiting.")
			}
			hasOAuth = false
		} else {
			app.oauthService = svc
			if !hasAPIKey {
				app.service = svc
			}
			// Override homeChannel with the OAuth-authenticated channel
			channels, err := svc.Channels.List([]string{"id"}).Mine(true).Do()
			if err == nil && len(channels.Items) > 0 {
				oauthChannelID := channels.Items[0].Id
				if oauthChannelID != channelID {
					app.homeChannel = oauthChannelID
					app.cwd = "/" + oauthChannelID
				}
			}
		}
	}

	// Non-interactive (one-shot) mode: ysh <command> [args...]
	if len(os.Args) > 1 {
		args := os.Args[1:]
		for i := 0; i < len(args); i++ {
			if args[i] == "-C" && i+1 < len(args) {
				app.cwd = args[i+1]
				args = append(args[:i], args[i+2:]...)
				i -= 1
				continue
			}
			switch args[i] {
			case "--version", "-v":
				fmt.Printf("ysh %s\n", version)
				return
			case "--help", "-h":
				fmt.Printf("ysh %s - YouTube Shell\n\n", version)
				fmt.Println("Usage: ysh [-C <path>] [command] [args...]")
				fmt.Println()
				fmt.Println("Options:")
				fmt.Println("  -C <path>  Set working path (e.g. -C /UCxxxx/PLxxxx)")
				fmt.Println()
				fmt.Println("Commands:")
				fmt.Println("  ls [-l]              List playlists/videos")
				fmt.Println("  cat <video_id>       Show video details")
				fmt.Println("  cp <video> <pl>      Add video to playlist")
				fmt.Println("  rm <video> <pl>      Remove video from playlist")
				fmt.Println("  mv <video> <src> <dst> Move video between playlists")
				fmt.Println("  mkdir [-m MODE] <title> Create playlist")
				fmt.Println("  chmod <mode> <pl>    Change playlist privacy")
				fmt.Println("  rmdir <pl>           Delete playlist")
				fmt.Println("  whoami               Show your channel info")
				fmt.Println("  pwd                  Print current path")
				fmt.Println("  logout               Reset OAuth token")
				fmt.Println("  grep <keyword>       Filter ls results by keyword")
				fmt.Println("  tree                 Display tree view of hierarchy")
				fmt.Println()
				fmt.Println("Run without arguments for interactive shell.")
				return
			}
		}
		if len(args) > 0 {
			app.executor(strings.Join(args, " "))
		}
		return
	}

	// Interactive mode
	fmt.Printf("ysh %s - YouTube Shell\n", version)
	fmt.Println("Loading...")
	app.doLs(nil)
	fmt.Println()
	if hasOAuth {
		if hasAPIKey {
			fmt.Println("Mode: full (API key + OAuth)")
		} else {
			fmt.Println("Mode: full (OAuth)")
		}
		fmt.Println("Commands: ls, cd, cat, cp, mkdir, chmod, rm, mv, rmdir, whoami, open, pwd, logout, exit, find, grep, tree")
	} else {
		fmt.Println("Mode: read-only (API key)")
		fmt.Println("Set YOUTUBE_CLIENT_ID and YOUTUBE_CLIENT_SECRET in ~/.ysh/.env to enable write operations.")
		fmt.Println("Commands: ls, cd, cat, whoami, open, pwd, exit, find, grep, tree")
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
		a.doLs(args)
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
	case "logout":
		a.doLogout()
	case "mv":
		a.doMv(args)
	case "rmdir":
		a.doRmdir(args)
	case "find":
		a.doFind(args)
	case "grep":
		a.doGrep(args)
	case "tree":
		a.doTree(args)
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
	case "ls":
		return []prompt.Suggest{{Text: "-l", Description: "Show privacy status"}}
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
	case "find":
		return []prompt.Suggest{
			{Text: "-k", Description: "Search keyword"},
			{Text: "-s", Description: "Time filter (1h, 24h, 7d, 30d)"},
			{Text: "-c", Description: "Channel ID"},
			{Text: "-a", Description: "Auto-add to playlist"},
			{Text: "-q", Description: "Quiet mode (IDs only)"},
		}
	case "open":
		all := make([]prompt.Suggest, 0, len(a.videoEntries)+len(a.playlistEntries))
		all = append(all, a.videoEntries...)
		all = append(all, a.playlistEntries...)
		return filterByPrefix(all, prefix)
	case "grep":
		return filterByPrefix(a.entries, prefix)
	case "tree":
		return []prompt.Suggest{}
	case "cat":
		return filterByPrefix(a.videoEntries, prefix)
	case "mkdir":
		return []prompt.Suggest{{Text: "-m", Description: "Set mode (public/unlisted/private)"}}
	case "chmod":
		return a.playlistEntries
	case "whoami", "logout":
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
			{Text: "find", Description: "Search videos"},
			{Text: "grep", Description: "Filter ls results by keyword"},
			{Text: "tree", Description: "Display tree view of hierarchy"},
			{Text: "cp", Description: "Copy video into playlist"},
			{Text: "mkdir", Description: "Create playlist"},
			{Text: "chmod", Description: "Change playlist privacy"},
			{Text: "rm", Description: "Remove video from playlist"},
			{Text: "whoami", Description: "Show your channel info"},
			{Text: "mv", Description: "Move video between playlists"},
			{Text: "rmdir", Description: "Delete playlist"},
			{Text: "logout", Description: "Reset OAuth token"},
			{Text: "exit", Description: "Exit the shell"},
		}
	}
}
