package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/mdp/qrterminal"
)

func (a *App) doOpen(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: open <video_id|playlist_id>")
		return
	}
	id := args[0]

	var url string
	if a.isPlaylistID(id) {
		url = "https://www.youtube.com/playlist?list=" + id
	} else {
		url = "https://www.youtube.com/watch?v=" + id
	}

	fmt.Println(url)
	qrterminal.GenerateHalfBlock(url, qrterminal.L, os.Stdout)

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
		fmt.Println("(Could not open browser automatically)")
		return
	}
	fmt.Println("Opened in browser.")
}

func (a *App) isPlaylistID(id string) bool {
	for _, s := range a.playlistEntries {
		if s.Text == id {
			return true
		}
	}
	return false
}