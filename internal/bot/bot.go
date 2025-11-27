package bot

import (
	"fmt"
	"log"

	"discord-github-bot/internal/config"
	"discord-github-bot/internal/database"
	"discord-github-bot/internal/oauth"
	"discord-github-bot/internal/github/rest"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	config      *config.Config
	db          *database.Database
	oauth       *oauth.Server
	session     *discordgo.Session
	commands    []*discordgo.ApplicationCommand
	githubREST *rest.GitHubRESTClient
}

func New(cfg *config.Config, db *database.Database, oauthServer *oauth.Server) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		config:  cfg,
		db:      db,
		oauth:   oauthServer,
		session: session,
		githubREST: rest.NewGitHubRESTClient(),
	}

	bot.registerCommands()
	session.AddHandler(bot.handleInteraction)

	return bot, nil
}

func (b *Bot) registerCommands() {
	b.commands = []*discordgo.ApplicationCommand{
		{
			Name:        "gh-auth",
			Description: "Authenticate with GitHub to link your account",
		},
		{
			Name:        "gh-unauth",
			Description: "Remove your GitHub authentication",
		},
		{
			Name:        "gh-set-repo",
			Description: "Set the default repository for this channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "repo",
					Description: "Repository in format: owner/repo",
					Required:    true,
				},
			},
		},
		{
			Name:        "gh-set-project",
			Description: "Set the default project for this channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "project",
					Description: "Project number",
					Required:    true,
				},
			},
		},
		{
			Name:        "gh-issue-create",
			Description: "Create a new GitHub issue",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "Issue title",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "body",
					Description: "Issue description",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "repo",
					Description: "Repository (overrides channel default)",
					Required:    false,
				},
			},
		},
		{
			Name:        "gh-issue-list",
			Description: "List GitHub issues",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "repo",
					Description: "Repository (overrides channel default)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "state",
					Description: "Issue state",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "open", Value: "open"},
						{Name: "closed", Value: "closed"},
						{Name: "all", Value: "all"},
					},
				},
			},
		},
		{
			Name:        "gh-issue-view",
			Description: "View a specific GitHub issue",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "number",
					Description: "Issue number",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "repo",
					Description: "Repository (overrides channel default)",
					Required:    false,
				},
			},
		},
		{
			Name:        "gh-issue-close",
			Description: "Close a GitHub issue",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "number",
					Description: "Issue number",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "repo",
					Description: "Repository (overrides channel default)",
					Required:    false,
				},
			},
		},
		{
			Name:        "gh-issue-comment",
			Description: "Comment on a GitHub issue",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "number",
					Description: "Issue number",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comment",
					Description: "Your comment",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "repo",
					Description: "Repository (overrides channel default)",
					Required:    false,
				},
			},
		},
		{
			Name:        "gh-project-items-list",
			Description: "List items in a GitHub project",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "project-number",
					Description: "The number of the project",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "org",
					Description: "Organization name (overrides channel default)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Filter query (default: -is:closed -is:done)",
					Required:    false,
				},
			},
		},
		{
			Name:        "gh-project-list",
			Description: "List all GitHub projects in an organization",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "org",
					Description: "Organization name (overrides channel default)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Filter query (default: is:open)",
					Required:    false,
				},
			},
		},
		{
			Name:        "gh-project-add-issue",
			Description: "Add an existing issue to a GitHub project",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "issue-number",
					Description: "Issue number to add",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "project-number",
					Description: "Project number (overrides channel default)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "repo",
					Description: "Repository in format: owner/repo (overrides channel default)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "org",
					Description: "Organization name (overrides channel default)",
					Required:    false,
				},
			},
		},
	}
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return err
	}

	log.Printf("Registering commands for application %s", b.config.DiscordApplicationID)
	for _, cmd := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.config.DiscordApplicationID, "", cmd)
		if err != nil {
			log.Printf("Failed to register command %s: %v", cmd.Name, err)
			return err
		}
		log.Printf("Registered command: %s", cmd.Name)
	}

	return nil
}

func (b *Bot) Stop() {
	log.Println("Cleaning up commands...")
	commands, err := b.session.ApplicationCommands(b.config.DiscordApplicationID, "")
	if err != nil {
		log.Printf("Failed to fetch commands: %v", err)
	} else {
		for _, cmd := range commands {
			err := b.session.ApplicationCommandDelete(b.config.DiscordApplicationID, "", cmd.ID)
			if err != nil {
				log.Printf("Failed to delete command %s: %v", cmd.Name, err)
			}
		}
	}

	b.session.Close()
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	switch data.Name {
	case "gh-auth":
		b.handleAuth(s, i)
	case "gh-unauth":
		b.handleUnauth(s, i)
	case "gh-set-repo":
		b.handleSetRepo(s, i)
	case "gh-set-project":
		b.handleSetProject(s, i)
	case "gh-issue-create":
		b.handleIssueCreate(s, i)
	case "gh-issue-list":
		b.handleIssueList(s, i)
	case "gh-issue-view":
		b.handleIssueView(s, i)
	case "gh-issue-close":
		b.handleIssueClose(s, i)
	case "gh-issue-comment":
		b.handleIssueComment(s, i)
	case "gh-project-items-list":
		b.handleProjectItemsList(s, i)
	case "gh-project-list":
		b.handleProjectList(s, i)
	case "gh-project-add-issue":
		b.handleProjectAddIssue(s, i)
	default:
		b.respondError(s, i, "Unknown command")
	}
}

func (b *Bot) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("❌ Error: %s", message),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) respondSuccess(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func (b *Bot) respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) getStringOption(options []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, opt := range options {
		if opt.Name == name {
			return opt.StringValue()
		}
	}
	return ""
}

func (b *Bot) getIntOption(options []*discordgo.ApplicationCommandInteractionDataOption, name string) int {
	for _, opt := range options {
		if opt.Name == name {
			return int(opt.IntValue())
		}
	}
	return 0
}
