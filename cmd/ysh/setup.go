package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/api/youtube/v3"
)

var _ = youtube.YoutubeForceSslScope // ensure youtube package is linked

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

// fetchOwnChannelID uses OAuth to get the authenticated user's channel ID.
func fetchOwnChannelID(clientID, clientSecret string) string {
	ctx := context.Background()
	svc, err := getOAuthService(ctx, clientID, clientSecret)
	if err != nil {
		return ""
	}
	channels, err := svc.Channels.List([]string{"id"}).Mine(true).Do()
	if err != nil {
		return ""
	}
	if len(channels.Items) == 0 {
		return ""
	}
	return channels.Items[0].Id
}

// runSetup checks for credentials. Either API key or OAuth is sufficient.
// Search order:
//  1. Existing environment variables (e.g. exported in shell)
//  2. ~/.ysh/.env (global config)
//  3. ./.env (local project config, for development)
func runSetup() (apiKey, channelID, clientID, clientSecret string) {
	_ = godotenv.Load(configPath())
	_ = godotenv.Load()

	apiKey = os.Getenv("YOUTUBE_API_KEY")
	channelID = os.Getenv("YOUTUBE_CHANNEL_ID")
	clientID = os.Getenv("YOUTUBE_CLIENT_ID")
	clientSecret = os.Getenv("YOUTUBE_CLIENT_SECRET")

	hasAPIKey := apiKey != ""
	hasOAuth := clientID != "" && clientSecret != ""

	// All configured — nothing to do
	if (hasAPIKey || hasOAuth) && channelID != "" {
		return
	}

	fmt.Println("=== ysh first-time setup ===")
	fmt.Println()
	fmt.Println("You need either an API key (read-only) or OAuth credentials (full access).")
	fmt.Println()

	if !hasAPIKey && !hasOAuth {
		fmt.Println("Choose authentication method:")
		fmt.Println("  1) API key (read-only access, simpler setup)")
		fmt.Println("  2) OAuth credentials (full access, auto-detect channel)")
		fmt.Println()
		var choice string
		for choice != "1" && choice != "2" {
			choice = readLine("  Select [1/2]: ")
			if choice != "1" && choice != "2" {
				fmt.Println("  Please enter 1 or 2.")
			}
		}

		switch choice {
		case "1":
			fmt.Println()
			fmt.Println("  Get an API key at: https://console.cloud.google.com/apis/credentials")
			for apiKey == "" {
				apiKey = readLine("  YOUTUBE_API_KEY: ")
				if apiKey == "" {
					fmt.Println("  API key is required. Please enter your key.")
				}
			}
		case "2":
			fmt.Println()
			fmt.Println("  Create OAuth credentials at: https://console.cloud.google.com/apis/credentials")
			fmt.Println("  Application type: Desktop app")
			fmt.Println("  Redirect URI: http://localhost:8089/callback")
			fmt.Println()
			for clientID == "" {
				clientID = readLine("  YOUTUBE_CLIENT_ID: ")
				if clientID == "" {
					fmt.Println("  Client ID is required. Please enter your client ID.")
				}
			}
			for clientSecret == "" {
				clientSecret = readLine("  YOUTUBE_CLIENT_SECRET: ")
				if clientSecret == "" {
					fmt.Println("  Client secret is required. Please enter your client secret.")
				}
			}
			// Auto-detect channel ID via OAuth
			fmt.Println()
			fmt.Println("  Authenticating with OAuth to detect your channel...")
			autoID := fetchOwnChannelID(clientID, clientSecret)
			if autoID != "" {
				fmt.Printf("  Detected channel: %s\n", autoID)
				channelID = autoID
			} else {
				fmt.Println("  Could not auto-detect your channel.")
			}
		}
	} else if hasAPIKey && !hasOAuth {
		fmt.Println("YOUTUBE_API_KEY: already set")
		fmt.Println()
		fmt.Println("--- Optional: OAuth credentials (for write operations) ---")
		fmt.Println("  Leave blank to skip (read-only mode).")
		fmt.Println("  Create at: https://console.cloud.google.com/apis/credentials")
		fmt.Println("  Redirect URI: http://localhost:8089/callback")
		fmt.Println()
		clientID = readLineDefault("  YOUTUBE_CLIENT_ID", clientID)
		clientSecret = readLineDefault("  YOUTUBE_CLIENT_SECRET", clientSecret)
	} else if !hasAPIKey && hasOAuth {
		fmt.Println("YOUTUBE_CLIENT_ID / SECRET: already set (OAuth mode)")
	}

	// If channel ID still missing, ask manually (needed for API key mode or OAuth auto-detect failed)
	if channelID == "" {
		fmt.Println()
		fmt.Println("Your YouTube Channel ID is required.")
		fmt.Println("  Format: UCxxxxxxxxxxxxxxxx")
		fmt.Println("  Find at: https://www.youtube.com/account_advanced")
		for channelID == "" {
			channelID = readLine("  YOUTUBE_CHANNEL_ID: ")
			if channelID == "" {
				fmt.Println("  Channel ID is required. Please enter your channel ID.")
			}
		}
	}

	saveConfig(apiKey, channelID, clientID, clientSecret)
	fmt.Println()
	fmt.Printf("Configuration saved to %s\n", configPath())
	return
}

func saveConfig(apiKey, channelID, clientID, clientSecret string) {
	dir := configDir()
	os.MkdirAll(dir, 0700)

	var b strings.Builder
	if apiKey != "" {
		b.WriteString(fmt.Sprintf("YOUTUBE_API_KEY=%s\n", apiKey))
	}
	b.WriteString(fmt.Sprintf("YOUTUBE_CHANNEL_ID=%s\n", channelID))
	if clientID != "" {
		b.WriteString(fmt.Sprintf("YOUTUBE_CLIENT_ID=%s\n", clientID))
	}
	if clientSecret != "" {
		b.WriteString(fmt.Sprintf("YOUTUBE_CLIENT_SECRET=%s\n", clientSecret))
	}
	os.WriteFile(configPath(), []byte(b.String()), 0600)

	if apiKey != "" {
		os.Setenv("YOUTUBE_API_KEY", apiKey)
	}
	os.Setenv("YOUTUBE_CHANNEL_ID", channelID)
	if clientID != "" {
		os.Setenv("YOUTUBE_CLIENT_ID", clientID)
	}
	if clientSecret != "" {
		os.Setenv("YOUTUBE_CLIENT_SECRET", clientSecret)
	}
}
