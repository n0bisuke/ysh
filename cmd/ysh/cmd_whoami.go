package main

import (
	"fmt"
)

func (a *App) doWhoami() {
	call := a.service.Channels.List([]string{"snippet,statistics"}).Id(a.homeChannel)
	channels, err := call.Do()
	if err != nil {
		fmt.Printf("Error fetching channel info: %v\n", err)
		return
	}
	if len(channels.Items) == 0 {
		fmt.Printf("Channel not found: %s\n", a.homeChannel)
		return
	}

	ch := channels.Items[0]
	fmt.Printf("%sChannel:%s     %s\n", colorCyan, colorReset, ch.Snippet.Title)
	fmt.Printf("%sID:%s          %s\n", colorCyan, colorReset, ch.Id)
	if ch.Snippet.CustomUrl != "" {
		fmt.Printf("%sCustom URL:%s  %s\n", colorCyan, colorReset, ch.Snippet.CustomUrl)
	}
	fmt.Printf("%sPublished:%s   %s\n", colorCyan, colorReset, ch.Snippet.PublishedAt[:10])
	fmt.Printf("%sSubscribers:%s %s\n", colorCyan, colorReset, formatNumber(ch.Statistics.SubscriberCount))
	fmt.Printf("%sVideos:%s      %s\n", colorCyan, colorReset, formatNumber(ch.Statistics.VideoCount))
	fmt.Printf("%sViews:%s       %s\n", colorCyan, colorReset, formatNumber(ch.Statistics.ViewCount))
}