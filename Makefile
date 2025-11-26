.PHONY: build run clean install-deps generate-key help

help:
	@echo "Discord GitHub Bot - Makefile commands:"
	@echo "  make install-deps  - Install Go dependencies"
	@echo "  make generate-key  - Generate a random encryption key"
	@echo "  make build        - Build the bot binary"
	@echo "  make run          - Run the bot (requires .env file)"
	@echo "  make clean        - Remove built binary and database"

install-deps:
	go mod download
	go mod tidy

generate-key:
	@go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); key := base64.StdEncoding.EncodeToString(b)[:32]; fmt.Printf("Generated encryption key: %s\n", key); fmt.Println("Add this to your .env file as ENCRYPTION_KEY") }' 2>/dev/null || openssl rand -base64 32 | head -c 32 && echo

build:
	go build -o discord-github-bot main.go

run: build
	./discord-github-bot

clean:
	rm -f discord-github-bot
	rm -f bot.db bot.db-shm bot.db-wal
