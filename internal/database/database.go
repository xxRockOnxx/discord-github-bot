package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db    *sql.DB
	gcm   cipher.AEAD
}

type User struct {
	DiscordID      string
	GitHubUsername string
	GitHubToken    string
}

type ChannelSettings struct {
	ChannelID      string
	DefaultRepo    string
	DefaultProject string
}

func New(dbPath string, encryptionKey []byte) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	d := &Database{
		db:  db,
		gcm: gcm,
	}

	if err := d.createTables(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Database) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		discord_id TEXT PRIMARY KEY,
		github_username TEXT NOT NULL,
		github_token TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS channel_settings (
		channel_id TEXT PRIMARY KEY,
		default_repo TEXT,
		default_project TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_discord_id ON users(discord_id);
	CREATE INDEX IF NOT EXISTS idx_channel_settings_channel_id ON channel_settings(channel_id);
	`

	_, err := d.db.Exec(schema)
	return err
}

func (d *Database) encrypt(plaintext string) (string, error) {
	nonce := make([]byte, d.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := d.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (d *Database) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonceSize := d.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := d.gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (d *Database) SaveUser(user *User) error {
	encryptedToken, err := d.encrypt(user.GitHubToken)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO users (discord_id, github_username, github_token, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(discord_id) DO UPDATE SET
		github_username = excluded.github_username,
		github_token = excluded.github_token,
		updated_at = CURRENT_TIMESTAMP
	`

	_, err = d.db.Exec(query, user.DiscordID, user.GitHubUsername, encryptedToken)
	return err
}

func (d *Database) GetUser(discordID string) (*User, error) {
	query := `SELECT discord_id, github_username, github_token FROM users WHERE discord_id = ?`

	var user User
	var encryptedToken string

	err := d.db.QueryRow(query, discordID).Scan(&user.DiscordID, &user.GitHubUsername, &encryptedToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	token, err := d.decrypt(encryptedToken)
	if err != nil {
		return nil, err
	}

	user.GitHubToken = token
	return &user, nil
}

func (d *Database) DeleteUser(discordID string) error {
	_, err := d.db.Exec("DELETE FROM users WHERE discord_id = ?", discordID)
	return err
}

func (d *Database) SaveChannelSettings(settings *ChannelSettings) error {
	query := `
	INSERT INTO channel_settings (channel_id, default_repo, default_project, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(channel_id) DO UPDATE SET
		default_repo = excluded.default_repo,
		default_project = excluded.default_project,
		updated_at = CURRENT_TIMESTAMP
	`

	_, err := d.db.Exec(query, settings.ChannelID, settings.DefaultRepo, settings.DefaultProject)
	return err
}

func (d *Database) GetChannelSettings(channelID string) (*ChannelSettings, error) {
	query := `SELECT channel_id, default_repo, default_project FROM channel_settings WHERE channel_id = ?`

	var settings ChannelSettings
	err := d.db.QueryRow(query, channelID).Scan(&settings.ChannelID, &settings.DefaultRepo, &settings.DefaultProject)
	if err != nil {
		if err == sql.ErrNoRows {
			return &ChannelSettings{ChannelID: channelID}, nil
		}
		return nil, err
	}

	return &settings, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}
