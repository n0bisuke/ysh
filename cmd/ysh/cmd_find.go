package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/api/youtube/v3"
)

func (a *App) doFind(args []string) {
	var keyword, since, channelID, addToPlaylist, eventType string
	quiet := false

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--keyword", "-k":
			i++
			if i < len(args) {
				keyword = args[i]
			}
		case "--since", "-s":
			i++
			if i < len(args) {
				since = args[i]
			}
		case "--channel", "-c":
			i++
			if i < len(args) {
				channelID = args[i]
			}
		case "--add-to", "-a":
			i++
			if i < len(args) {
				addToPlaylist = args[i]
			}
		case "--event-type", "-e":
			i++
			if i < len(args) {
				eventType = args[i]
			}
		case "--live", "-l":
			eventType = "live"
		case "--quiet", "-q":
			quiet = true
		default:
			if keyword == "" {
				keyword = args[i]
			}
		}
		i++
	}

	if keyword == "" {
		fmt.Println("Usage: find [--keyword|-k] <keyword> [--since|-s] <duration> [--channel|-c] <channel_id> [--add-to|-a] <playlist_id> [--event-type|-e] <type> [--live|-l] [--quiet|-q]")
		fmt.Println()
		fmt.Println("  --keyword, -k    Search keyword (required)")
		fmt.Println("  --since, -s      Time filter: 1h, 24h, 7d, 30d (default: 24h)")
		fmt.Println("  --channel, -c    Limit to channel ID (default: your channel)")
		fmt.Println("  --add-to, -a     Auto-add found videos to this playlist")
		fmt.Println("  --event-type, -e Filter by broadcast type: live, upcoming, completed")
		fmt.Println("  --live, -l        Shorthand for --event-type live")
		fmt.Println("  --quiet, -q      Output only video IDs (for scripting)")
		return
	}

	if since == "" {
		since = "24h"
	}
	publishedAfter, err := parseDuration(since)
	if err != nil {
		fmt.Printf("Invalid --since value: %v (use: 1h, 24h, 7d, 30d)\n", err)
		return
	}

	if channelID == "" {
		channelID = a.homeChannel
	}

	svc := a.readService(channelID)

	call := svc.Search.List([]string{"snippet"}).
		Q(keyword).
		ChannelId(channelID).
		Type("video").
		PublishedAfter(publishedAfter).
		MaxResults(50)
	if eventType != "" {
		call = call.EventType(eventType)
	}
	results, err := call.Do()
	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return
	}

	if len(results.Items) == 0 {
		if !quiet {
			fmt.Println("No videos found.")
		}
		return
	}

	for _, item := range results.Items {
		videoID := item.Id.VideoId
		title := item.Snippet.Title
		publishedAt := item.Snippet.PublishedAt

		if quiet {
			fmt.Println(videoID)
			continue
		}

		fmt.Printf("%s%-15s%s %s %-30s %s\n",
			colorGreen, videoID, colorReset,
			publishedAt[:10],
			truncate(title, 29), "")
	}

	if addToPlaylist != "" {
		if !a.ensureOAuth() {
			return
		}
		fmt.Println()
		for _, item := range results.Items {
			videoID := item.Id.VideoId
			plItem := &youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					PlaylistId: addToPlaylist,
					ResourceId: &youtube.ResourceId{
						Kind:    "youtube#video",
						VideoId: videoID,
					},
				},
			}
			_, err := a.oauthService.PlaylistItems.Insert([]string{"snippet"}, plItem).Do()
			if err != nil {
				fmt.Printf("Error adding %s: %v\n", videoID, err)
			} else {
				fmt.Printf("Added %s to %s\n", videoID, addToPlaylist)
			}
		}
	}
}

func parseDuration(d string) (string, error) {
	d = strings.ToLower(strings.TrimSpace(d))
	var hours int
	if strings.HasSuffix(d, "h") {
		n, err := parseNum(d[:len(d)-1])
		if err != nil {
			return "", err
		}
		hours = n
	} else if strings.HasSuffix(d, "d") {
		n, err := parseNum(d[:len(d)-1])
		if err != nil {
			return "", err
		}
		hours = n * 24
	} else {
		return "", fmt.Errorf("unsupported format: %s", d)
	}

	t := time.Now().Add(-time.Duration(hours) * time.Hour)
	return t.Format("2006-01-02T15:04:05Z"), nil
}

func parseNum(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func init() {
	// Ensure non-interactive mode works without go-prompt
	_ = os.Stdout
}
