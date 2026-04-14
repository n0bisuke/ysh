package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ysh")
}

func configPath() string {
	return filepath.Join(configDir(), ".env")
}

func tokenPath() string {
	return filepath.Join(configDir(), "token.json")
}

func readLine(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func readLineDefault(prompt, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Print(prompt)
	}
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultVal
	}
	return text
}

// runSetup checks for required env vars. Search order:
//  1. Existing environment variables (e.g. exported in shell)
//  2. ~/.config/ysh/config (global config)
//  3. ./.env (local project config, for development)
//
// If required values are still missing, launches interactive setup
// and saves to ~/.config/ysh/config.
func runSetup() (apiKey, channelID, clientID, clientSecret string) {
	// Load global config first
	_ = godotenv.Load(configPath())
	// Then local config (overrides global)
	_ = godotenv.Load()

	apiKey = os.Getenv("YOUTUBE_API_KEY")
	channelID = os.Getenv("YOUTUBE_CHANNEL_ID")
	clientID = os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret = os.Getenv("YOUTUBE_CLIENT_SECRET")

	if apiKey != "" && channelID != "" {
		return
	}

	fmt.Println("=== ysh first-time setup ===")
	fmt.Println("Required values are missing. Let's configure them.")
	fmt.Println()

	if apiKey == "" {
		fmt.Println("[1/2] YouTube Data API v3 key (required)")
		fmt.Println("  Get one at: https://console.cloud.google.com/apis/credentials")
		apiKey = readLine("  YOUTUBE_API_KEY: ")
		if apiKey == "" {
			fmt.Println("API key is required. Exiting.")
			os.Exit(1)
		}
	} else {
		fmt.Println("[1/2] YOUTUBE_API_KEY: already set")
	}

	if channelID == "" {
		fmt.Println("[2/2] Your YouTube Channel ID (required)")
		fmt.Println("  Format: UCxxxxxxxxxxxxxxxx")
		fmt.Println("  Find at: https://www.youtube.com/account_advanced")
		channelID = readLine("  YOUTUBE_CHANNEL_ID: ")
		if channelID == "" {
			fmt.Println("Channel ID is required. Exiting.")
			os.Exit(1)
		}
	} else {
		fmt.Println("[2/2] YOUTUBE_CHANNEL_ID: already set")
	}

	fmt.Println()
	fmt.Println("--- Optional: OAuth credentials (for 'add' command) ---")
	fmt.Println("  Leave blank to skip (read-only mode).")
	fmt.Println("  Create at: https://console.cloud.google.com/apis/credentials")
	fmt.Println("  Redirect URI: http://localhost:8089/callback")
	fmt.Println()
	clientID = readLineDefault("  YOUTUBE_CLIENT_ID", clientID)
	clientSecret = readLineDefault("  YOUTUBE_CLIENT_SECRET", clientSecret)

	saveConfig(apiKey, channelID, clientID, clientSecret)
	fmt.Println()
	fmt.Printf("Configuration saved to %s\n", configPath())
	return
}

func saveConfig(apiKey, channelID, clientID, clientSecret string) {
	dir := configDir()
	os.MkdirAll(dir, 0700)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("YOUTUBE_API_KEY=%s\n", apiKey))
	b.WriteString(fmt.Sprintf("YOUTUBE_CHANNEL_ID=%s\n", channelID))
	if clientID != "" {
		b.WriteString(fmt.Sprintf("YOUTUBE_CLIENT_ID=%s\n", clientID))
	}
	if clientSecret != "" {
		b.WriteString(fmt.Sprintf("YOUTUBE_CLIENT_SECRET=%s\n", clientSecret))
	}
	os.WriteFile(configPath(), []byte(b.String()), 0600)

	os.Setenv("YOUTUBE_API_KEY", apiKey)
	os.Setenv("YOUTUBE_CHANNEL_ID", channelID)
	if clientID != "" {
		os.Setenv("YOUTUBE_CLIENT_ID", clientID)
	}
	if clientSecret != "" {
		os.Setenv("YOUTUBE_CLIENT_SECRET", clientSecret)
	}
}