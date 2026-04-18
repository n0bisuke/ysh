package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func oauthConfig(clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8089/callback",
		Scopes:       []string{youtube.YoutubeForceSslScope},
	}
}

func loadToken(_ *oauth2.Config) (*oauth2.Token, error) {
	// CI environment: restore token from env var
	if envToken := os.Getenv("YOUTUBE_TOKEN_JSON"); envToken != "" {
		var tok oauth2.Token
		if err := json.Unmarshal([]byte(envToken), &tok); err != nil {
			return nil, fmt.Errorf("invalid YOUTUBE_TOKEN_JSON: %w", err)
		}
		return &tok, nil
	}
	data, err := os.ReadFile(tokenPath())
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func saveToken(tok *oauth2.Token) error {
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	os.MkdirAll(configDir(), 0700)
	return os.WriteFile(tokenPath(), data, 0600)
}

// startAuthServer launches a local HTTP server for the OAuth callback,
// opens the browser, and returns the authorization code.
func startAuthServer(config *oauth2.Config) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Authorization failed: no code received.")
			return
		}
		codeCh <- code
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Authorization successful! You can close this tab.")
	})

	server := &http.Server{Addr: ":8089", Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("Open the following URL in your browser to authorize ysh:")
	fmt.Println(authURL)

	select {
	case code := <-codeCh:
		server.Close()
		return code, nil
	case err := <-errCh:
		server.Close()
		return "", err
	}
}

// getOAuthService builds a *youtube.Service authenticated via OAuth2.
// Reuses a cached token from .token.json when available.
func getOAuthService(ctx context.Context, clientID, clientSecret string) (*youtube.Service, error) {
	config := oauthConfig(clientID, clientSecret)

	tok, err := loadToken(config)
	if err != nil {
		// No cached token — run the OAuth flow
		code, err := startAuthServer(config)
		if err != nil {
			return nil, fmt.Errorf("OAuth flow failed: %w", err)
		}
		tok, err = config.Exchange(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("token exchange failed: %w", err)
		}
		if err := saveToken(tok); err != nil {
			fmt.Printf("Warning: could not save token: %v\n", err)
		}
	}

	client := config.Client(ctx, tok)
	return youtube.NewService(ctx, option.WithHTTPClient(client))
}
