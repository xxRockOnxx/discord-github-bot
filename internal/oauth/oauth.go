package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"discord-github-bot/internal/config"
	"discord-github-bot/internal/database"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
	oauth2gh "golang.org/x/oauth2/github"
)

type Server struct {
	config      *config.Config
	db          *database.Database
	oauthConfig *oauth2.Config
	states      map[string]string
	statesMu    sync.RWMutex
	tmpl        *template.Template
}

func NewServer(cfg *config.Config, db *database.Database) *Server {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  cfg.GitHubRedirectURL,
		Scopes:       []string{"repo", "user:email", "read:org"},
		Endpoint:     oauth2gh.Endpoint,
	}

	tmpl, err := template.ParseFiles("templates/success.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	return &Server{
		config:      cfg,
		db:          db,
		oauthConfig: oauthConfig,
		states:      make(map[string]string),
		tmpl:        tmpl,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/auth", s.handleAuth)
	http.HandleFunc("/callback", s.handleCallback)

	addr := fmt.Sprintf("%s:%s", s.config.OAuthServerHost, s.config.OAuthServerPort)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) GenerateAuthURL(discordID string) string {
	state := s.generateState()
	s.statesMu.Lock()
	s.states[state] = discordID
	s.statesMu.Unlock()

	go func() {
		time.Sleep(10 * time.Minute)
		s.statesMu.Lock()
		delete(s.states, state)
		s.statesMu.Unlock()
	}()

	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline, oauth2.SetAuthURLParam("prompt", "select_account"))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	filePath := "templates/index.html"

	// Get file info for Last-Modified and ETag
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	lastModified := fileInfo.ModTime().UTC()
	etag := fmt.Sprintf(`"%x-%x"`, lastModified.Unix(), fileInfo.Size())

	// Set caching headers
	w.Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
	w.Header().Set("Last-Modified", lastModified.Format(http.TimeFormat))
	w.Header().Set("ETag", etag)

	// Check If-None-Match (ETag)
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Check If-Modified-Since
	if modifiedSince := r.Header.Get("If-Modified-Since"); modifiedSince != "" {
		t, err := http.ParseTime(modifiedSince)
		if err == nil && !lastModified.After(t) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	http.ServeFile(w, r, filePath)
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	discordID := r.URL.Query().Get("discord_id")
	if discordID == "" {
		http.Error(w, "Missing discord_id parameter", http.StatusBadRequest)
		return
	}

	authURL := s.GenerateAuthURL(discordID)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		http.Error(w, "Missing state or code parameter", http.StatusBadRequest)
		return
	}

	s.statesMu.Lock()
	discordID, exists := s.states[state]
	if exists {
		delete(s.states, state)
	}
	s.statesMu.Unlock()

	if !exists {
		http.Error(w, "Invalid or expired state parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		log.Printf("OAuth exchange error: %v", err)
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	client := github.NewClient(s.oauthConfig.Client(ctx, token))
	ghUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Printf("Failed to get GitHub user: %v", err)
		http.Error(w, "Failed to get GitHub user information", http.StatusInternalServerError)
		return
	}

	user := &database.User{
		DiscordID:      discordID,
		GitHubUsername: ghUser.GetLogin(),
		GitHubToken:    token.AccessToken,
	}

	if err := s.db.SaveUser(user); err != nil {
		log.Printf("Failed to save user: %v", err)
		http.Error(w, "Failed to save user information", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	data := struct {
		GitHubUsername string
	}{
		GitHubUsername: ghUser.GetLogin(),
	}

	if err := s.tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to execute template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func (s *Server) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *Server) GetGitHubClient(discordID string) (*github.Client, error) {
	user, err := s.db.GetUser(discordID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not authenticated with GitHub")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: user.GitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), nil
}
