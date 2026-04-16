package main

import (
	"fmt"
)

func (a *App) doCat(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: cat <video_id>")
		return
	}
	videoID := args[0]

	call := a.service.Videos.List([]string{"snippet,statistics,contentDetails"}).Id(videoID)
	videos, err := call.Do()
	if err != nil {
		fmt.Printf("Error fetching video: %v\n", err)
		return
	}
	if len(videos.Items) == 0 {
		fmt.Printf("Video not found: %s\n", videoID)
		return
	}

	v := videos.Items[0]
	fmt.Printf("%sTitle:%s       %s\n", colorCyan, colorReset, v.Snippet.Title)
	fmt.Printf("%sChannel:%s     %s (%s)\n", colorCyan, colorReset, v.Snippet.ChannelTitle, v.Snippet.ChannelId)
	fmt.Printf("%sPublished:%s   %s\n", colorCyan, colorReset, v.Snippet.PublishedAt[:10])
	fmt.Printf("%sDuration:%s    %s\n", colorCyan, colorReset, formatDuration(v.ContentDetails.Duration))
	fmt.Printf("%sViews:%s       %s\n", colorCyan, colorReset, formatNumber(v.Statistics.ViewCount))
	fmt.Printf("%sLikes:%s       %s\n", colorCyan, colorReset, formatNumber(v.Statistics.LikeCount))
	if v.Statistics.CommentCount != 0 {
		fmt.Printf("%sComments:%s    %s\n", colorCyan, colorReset, formatNumber(v.Statistics.CommentCount))
	}
	if v.Snippet.Description != "" {
		fmt.Println()
		fmt.Println(v.Snippet.Description)
	}
}