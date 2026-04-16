package main

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
)

// parsePath breaks cwd into components.
// "//"        → ("", "")       — subscriptions list
// "/UCxxxx"   → ("UCxxxx", "") — channel root
// "/UCxxxx/PLxxxx" → ("UCxxxx", "PLxxxx") — playlist
func parsePath(cwd string) (channelID, playlistID string) {
	cwd = strings.TrimPrefix(cwd, "/")
	parts := strings.SplitN(cwd, "/", 2)
	channelID = parts[0]
	if len(parts) == 2 {
		playlistID = parts[1]
	}
	return
}

func filterByPrefix(suggestions []prompt.Suggest, prefix string) []prompt.Suggest {
	if prefix == "" {
		return suggestions
	}
	lower := strings.ToLower(prefix)
	var filtered []prompt.Suggest
	for _, s := range suggestions {
		if strings.HasPrefix(strings.ToLower(s.Text), lower) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-1]) + "…"
}

func formatDuration(d string) string {
	d = strings.TrimPrefix(d, "PT")
	h, m, s := "", "", ""
	if idx := strings.Index(d, "H"); idx != -1 {
		h = d[:idx]
		d = d[idx+1:]
	}
	if idx := strings.Index(d, "M"); idx != -1 {
		m = d[:idx]
		d = d[idx+1:]
	}
	if idx := strings.Index(d, "S"); idx != -1 {
		s = d[:idx]
	}
	parts := []string{}
	if h != "" {
		parts = append(parts, h+"h")
	}
	if m != "" {
		parts = append(parts, m+"m")
	}
	if s != "" {
		parts = append(parts, s+"s")
	}
	if len(parts) == 0 {
		return "0s"
	}
	return strings.Join(parts, " ")
}

func formatNumber(n uint64) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}