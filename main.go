package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"discord-github-bot/internal/bot"
	"discord-github-bot/internal/config"
	"discord-github-bot/internal/database"
	"discord-github-bot/internal/oauth"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.New(cfg.DatabasePath, cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	oauthServer := oauth.NewServer(cfg, db)
	go func() {
		log.Printf("Starting OAuth server...")
		log.Printf("  Local address:  http://%s:%s", cfg.OAuthServerHost, cfg.OAuthServerPort)
		log.Printf("  Public URL:     %s", cfg.PublicURL)
		if err := oauthServer.Start(); err != nil {
			log.Fatalf("OAuth server error: %v", err)
		}
	}()

	discordBot, err := bot.New(cfg, db, oauthServer)
	if err != nil {
		log.Fatalf("Failed to create Discord bot: %v", err)
	}

	if err := discordBot.Start(); err != nil {
		log.Fatalf("Failed to start Discord bot: %v", err)
	}
	defer discordBot.Stop()

	log.Println("Discord GitHub Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
