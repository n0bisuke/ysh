package main

import (
	"fmt"
)

func (a *App) doTree(_ []string) {
	chID, plID := parsePath(a.cwd)

	switch {
	case chID == "" && plID == "":
		a.treeRoot()
	case plID == "":
		a.treeChannel(chID)
	default:
		a.treePlaylist(chID, plID)
	}
}

func (a *App) treeRoot() {
	if len(a.entries) == 0 {
		fmt.Println("No channels loaded. Run ls first.")
		return
	}
	fmt.Println("My Channels")
	for i, e := range a.entries {
		pfx := treePrefix(i == len(a.entries)-1)
		fmt.Printf("%s%s%s%s - %s\n", pfx, colorBlue, e.Text, colorReset, e.Description)
	}
}

func (a *App) treeChannel(chID string) {
	svc := a.readService(chID)

	chCall := svc.Channels.List([]string{"snippet"}).Id(chID)
	channels, err := chCall.Do()
	if err != nil {
		fmt.Printf("Error fetching channel: %v\n", err)
		return
	}
	if len(channels.Items) == 0 {
		fmt.Printf("Channel not found: %s\n", chID)
		return
	}
	ch := channels.Items[0]
	fmt.Printf("%s%s%s\n", colorBlue, ch.Snippet.Title, colorReset)

	plCall := svc.Playlists.List([]string{"snippet,contentDetails"}).
		ChannelId(chID).MaxResults(50)
	playlists, err := plCall.Do()
	if err != nil {
		fmt.Printf("Error fetching playlists: %v\n", err)
		return
	}

	for i, pl := range playlists.Items {
		pfx := treePrefix(i == len(playlists.Items)-1)
		count := pl.ContentDetails.ItemCount
		fmt.Printf("%s%s%s%s - %s [%d item%s]\n",
			pfx, colorCyan, pl.Id, colorReset,
			truncate(pl.Snippet.Title, 40), count, pluralS(count))
	}
}

func (a *App) treePlaylist(chID, plID string) {
	svc := a.readService(chID)

	plCall := svc.Playlists.List([]string{"snippet"}).Id(plID)
	playlists, err := plCall.Do()
	if err != nil || len(playlists.Items) == 0 {
		fmt.Printf("Error fetching playlist: %v\n", err)
		return
	}
	pl := playlists.Items[0]
	fmt.Printf("%s%s%s\n", colorCyan, pl.Snippet.Title, colorReset)

	itemsCall := svc.PlaylistItems.List([]string{"snippet,contentDetails"}).
		PlaylistId(plID).MaxResults(50)
	items, err := itemsCall.Do()
	if err != nil {
		fmt.Printf("Error fetching items: %v\n", err)
		return
	}

	for i, item := range items.Items {
		pfx := treePrefix(i == len(items.Items)-1)
		videoID := item.ContentDetails.VideoId
		title := item.Snippet.Title
		fmt.Printf("%s%s%s%s - %s\n",
			pfx, colorGreen, videoID, colorReset, truncate(title, 50))
	}
}

func treePrefix(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}

func pluralS(n int64) string {
	if n != 1 {
		return "s"
	}
	return ""
}
