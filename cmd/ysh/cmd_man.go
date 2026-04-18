package main

import (
	"fmt"
	"slices"
	"strings"
)

var manPages = map[string]string{
	"ls": `ls [-l]

List items in the current directory.

  -l    Show privacy status (public/unlisted/private)

At channel root (/UCxxxx), shows playlists and uploaded videos.
Inside a playlist (/UCxxxx/PLxxxx), shows videos in that playlist.
At root (//), shows your subscribed channels (requires OAuth).`,

	"cd": `cd [ID | .. | ~ | /]

Change the current directory.

  ID    Channel or playlist ID to navigate into
  ..    Go up to parent directory
  ~     Go to your own channel
  /     Go to root (subscription list)

Examples:
  cd PLxxxx        Enter a playlist
  cd ..            Go back to channel root
  cd ~             Jump to your channel
  cd /             Go to subscription list`,

	"pwd": `pwd

Print the current working directory path.`,

	"cat": `cat VIDEO_ID

Show detailed information about a video.

Displays title, description, channel, publish date,
view count, like count, and duration.`,

	"open": `open [ID | .]

Open a video or playlist in the default browser.
Also displays a QR code for mobile access.

  .    Open the current path (channel or playlist URL)

If the ID starts with "PL", it is treated as a playlist.
Otherwise it is treated as a video.`,

	"cp": `cp VIDEO_ID PLAYLIST_ID [POSITION]

Add a video to a playlist.

  VIDEO_ID      ID of the video to add
  PLAYLIST_ID   ID of the target playlist
  POSITION      (optional) Insert position (0=top, omitted=append)

Examples:
  cp dQw4w9WgXcQ PLxxxx        Add to end of playlist
  cp dQw4w9WgXcQ PLxxxx 0      Add to beginning`,

	"rm": `rm VIDEO_ID PLAYLIST_ID

Remove a video from a playlist.
This does not delete the video itself.

  VIDEO_ID      ID of the video to remove
  PLAYLIST_ID   ID of the playlist`,

	"mv": `mv VIDEO_ID SRC_PLAYLIST_ID DST_PLAYLIST_ID [POSITION]

Move a video from one playlist to another.

  VIDEO_ID          ID of the video to move
  SRC_PLAYLIST_ID   Source playlist
  DST_PLAYLIST_ID   Destination playlist
  POSITION          (optional) Insert position in destination`,

	"mkdir": `mkdir [-m MODE] TITLE

Create a new playlist.

  -m MODE   Set privacy mode: public, unlisted (default), private

Examples:
  mkdir "My Playlist"              Create unlisted playlist
  mkdir -m public "My Playlist"    Create public playlist`,

	"chmod": `chmod MODE PLAYLIST_ID

Change the privacy setting of a playlist.

  MODE   One of: public, unlisted, private

Examples:
  chmod public PLxxxx
  chmod private PLxxxx`,

	"rmdir": `rmdir PLAYLIST_ID

Delete a playlist. This cannot be undone.`,

	"whoami": `whoami

Display your authenticated channel information.
Shows channel ID, title, and creation date.`,

	"find": `find [KEYWORD] [OPTIONS]

Search for videos on your channel.

  -k, --keyword KEYWORD   Search keyword (required)
  -s, --since DURATION    Time filter: 1h, 24h (default), 7d, 30d
  -c, --channel CHANNEL   Limit to channel ID (default: your channel)
  -e, --event-type TYPE   Filter by broadcast type: live, upcoming, completed
  -l, --live              Shorthand for --event-type live
  -a, --add-to PL_ID     Auto-add found videos to this playlist
  -q, --quiet             Output only video IDs (for scripting)

Examples:
  find -k "golang" -s 1h              Search last hour
  find -k "live stream" --live         Find live streams
  find -k "tutorial" -a PLxxxx -q     Add results to playlist (quiet)`,

	"grep": `grep KEYWORD

Filter the current ls results by keyword.
Matches against both IDs and titles (case-insensitive).

You must run ls first to load entries before using grep.

Examples:
  grep golang       Filter entries containing "golang"
  grep PL            Filter entries with IDs containing "PL"`,

	"tree": `tree

Display the current path as a tree hierarchy.

  At root (//)           Shows subscribed channels
  At channel (/UCxxxx)   Shows playlists with item counts
  At playlist (/PLxxxx)  Shows videos in the playlist

Uses Unicode box-drawing characters (├── └──).`,

	"sort": `sort --by title|date PLAYLIST_ID [-r]

Sort videos in a playlist by title or date.

  -b, --by title|date   Sort criterion (required)
  -r, --reverse         Reverse sort order

  title   Sort alphabetically by video title
  date    Sort chronologically by publish date (newest first)

Note: This updates each item's position via the YouTube API.
For a 50-item playlist, up to 50 API calls may be required.

Examples:
  sort --by title PLxxxx          Alphabetical order
  sort --by date PLxxxx -r        Oldest first`,

	"logout": `logout

Reset the OAuth token. You will need to re-authenticate
on the next command that requires write access.`,

	"exit": `exit

Exit the shell. You can also press Ctrl+C.`,
}

func (a *App) doMan(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: man <command>")
		fmt.Println()
		fmt.Println("Available commands:")
		var cmds []string
		for k := range manPages {
			cmds = append(cmds, k)
		}
		slices.Sort(cmds)
		fmt.Println("  " + strings.Join(cmds, ", "))
		return
	}

	cmd := args[0]
	page, ok := manPages[cmd]
	if !ok || page == "" {
		fmt.Printf("No manual entry for: %s\n", cmd)
		return
	}
	fmt.Printf("%s\n", page)
}
