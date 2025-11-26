package config

import (
	"errors"
	"os"
)

type Config struct {
	DiscordBotToken     string
	DiscordApplicationID string
	GitHubClientID      string
	GitHubClientSecret  string
	GitHubRedirectURL   string
	EncryptionKey       []byte
	OAuthServerPort     string
	OAuthServerHost     string
	PublicURL           string
	DatabasePath        string
}

func Load() (*Config, error) {
	discordToken := os.Getenv("DISCORD_BOT_TOKEN")
	if discordToken == "" {
		return nil, errors.New("DISCORD_BOT_TOKEN is required")
	}

	appID := os.Getenv("DISCORD_APPLICATION_ID")
	if appID == "" {
		return nil, errors.New("DISCORD_APPLICATION_ID is required")
	}

	ghClientID := os.Getenv("GITHUB_CLIENT_ID")
	if ghClientID == "" {
		return nil, errors.New("GITHUB_CLIENT_ID is required")
	}

	ghClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if ghClientSecret == "" {
		return nil, errors.New("GITHUB_CLIENT_SECRET is required")
	}

	ghRedirectURL := os.Getenv("GITHUB_REDIRECT_URL")
	if ghRedirectURL == "" {
		ghRedirectURL = "http://localhost:8080/callback"
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		return nil, errors.New("ENCRYPTION_KEY is required (32 bytes for AES-256)")
	}
	if len(encryptionKey) != 32 {
		return nil, errors.New("ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
	}

	port := os.Getenv("OAUTH_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("OAUTH_SERVER_HOST")
	if host == "" {
		host = "localhost"
	}

	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		publicURL = "http://" + host + ":" + port
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./bot.db"
	}

	return &Config{
		DiscordBotToken:      discordToken,
		DiscordApplicationID: appID,
		GitHubClientID:       ghClientID,
		GitHubClientSecret:   ghClientSecret,
		GitHubRedirectURL:    ghRedirectURL,
		EncryptionKey:        []byte(encryptionKey),
		OAuthServerPort:      port,
		OAuthServerHost:      host,
		PublicURL:            publicURL,
		DatabasePath:         dbPath,
	}, nil
}
