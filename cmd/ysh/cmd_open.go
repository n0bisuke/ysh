package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/mdp/qrterminal"
)

func (a *App) doOpen(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: open <video_id|playlist_id|. >")
		return
	}
	id := args[0]

	var url string

	// Handle "." — open current path
	if id == "." || id == "./" {
		chID, plID := parsePath(a.cwd)
		switch {
		case chID == "" && plID == "":
			fmt.Println("No URL for root directory.")
			return
		case plID != "":
			url = "https://www.youtube.com/playlist?list=" + plID
		default:
			url = "https://www.youtube.com/channel/" + chID
		}
	} else if a.isPlaylistID(id) {
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

	if err := cmd.Run(); err != nil {
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
	return strings.HasPrefix(id, "PL")
}
