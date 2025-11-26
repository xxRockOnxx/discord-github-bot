package bot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/v57/github"
)

func (b *Bot) handleAuth(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	// Check if the user is already authenticated
	user, err := b.db.GetUser(userID)
	if err != nil {
		log.Printf("Failed to get user from DB: %v", err)
		b.respondError(s, i, "An error occurred while checking authentication status.")
		return
	}

	if user != nil {
		// User is already authenticated
		b.respondEphemeral(s, i, fmt.Sprintf(
			"You are already authenticated as **%s**. If you wish to authenticate with a different GitHub account, please use the `/gh-unauth` command first, then try `/gh-auth` again. You may also need to clear your browser's GitHub cookies or use an incognito/private browsing window.",
			user.GitHubUsername,
		))
		return
	}

	// User is not authenticated, proceed with original flow
	authURL := fmt.Sprintf("%s/auth?discord_id=%s", b.config.PublicURL, userID)

	b.respondEphemeral(s, i, fmt.Sprintf(
		"Click the link below to authenticate with **one** GitHub account:\n%s\n\nThis link will expire in 10 minutes. If you wish to switch accounts later, use `/gh-unauth` first.",
		authURL,
	))
}

func (b *Bot) handleUnauth(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID

	if err := b.db.DeleteUser(userID); err != nil {
		log.Printf("Failed to delete user: %v", err)
		b.respondError(s, i, "Failed to remove authentication")
		return
	}

	b.respondEphemeral(s, i, "Your GitHub authentication has been removed.")
}

func (b *Bot) handleSetRepo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	repo := b.getStringOption(i.ApplicationCommandData().Options, "repo")
	channelID := i.ChannelID

	if !strings.Contains(repo, "/") {
		b.respondError(s, i, "Repository must be in format: owner/repo")
		return
	}

	settings, err := b.db.GetChannelSettings(channelID)
	if err != nil {
		log.Printf("Failed to get channel settings: %v", err)
		b.respondError(s, i, "Failed to get channel settings")
		return
	}

	settings.DefaultRepo = repo

	if err := b.db.SaveChannelSettings(settings); err != nil {
		log.Printf("Failed to save channel settings: %v", err)
		b.respondError(s, i, "Failed to save channel settings")
		return
	}

	b.respondSuccess(s, i, fmt.Sprintf("✅ Default repository set to: %s", repo))
}

func (b *Bot) handleSetProject(s *discordgo.Session, i *discordgo.InteractionCreate) {
	project := b.getStringOption(i.ApplicationCommandData().Options, "project")
	channelID := i.ChannelID

	settings, err := b.db.GetChannelSettings(channelID)
	if err != nil {
		log.Printf("Failed to get channel settings: %v", err)
		b.respondError(s, i, "Failed to get channel settings")
		return
	}

	settings.DefaultProject = project

	if err := b.db.SaveChannelSettings(settings); err != nil {
		log.Printf("Failed to save channel settings: %v", err)
		b.respondError(s, i, "Failed to save channel settings")
		return
	}

	b.respondSuccess(s, i, fmt.Sprintf("✅ Default project set to: %s", project))
}

func (b *Bot) handleIssueCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	title := b.getStringOption(i.ApplicationCommandData().Options, "title")
	body := b.getStringOption(i.ApplicationCommandData().Options, "body")
	repo := b.getStringOption(i.ApplicationCommandData().Options, "repo")

	if repo == "" {
		settings, err := b.db.GetChannelSettings(i.ChannelID)
		if err != nil || settings.DefaultRepo == "" {
			b.respondError(s, i, "No repository specified and no default repository set for this channel")
			return
		}
		repo = settings.DefaultRepo
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		b.respondError(s, i, "Invalid repository format. Use: owner/repo")
		return
	}
	owner, repoName := parts[0], parts[1]

	client, err := b.oauth.GetGitHubClient(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	ctx := context.Background()
	issue := &github.IssueRequest{
		Title: &title,
		Body:  &body,
	}

	createdIssue, _, err := client.Issues.Create(ctx, owner, repoName, issue)
	if err != nil {
		log.Printf("Failed to create issue: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to create issue: %v", err))
		return
	}

	b.respondSuccess(s, i, fmt.Sprintf(
		"✅ Issue created successfully!\n**#%d** %s\n%s",
		createdIssue.GetNumber(),
		createdIssue.GetTitle(),
		createdIssue.GetHTMLURL(),
	))
}

func (b *Bot) handleIssueList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	repo := b.getStringOption(i.ApplicationCommandData().Options, "repo")
	state := b.getStringOption(i.ApplicationCommandData().Options, "state")

	if state == "" {
		state = "open"
	}

	if repo == "" {
		settings, err := b.db.GetChannelSettings(i.ChannelID)
		if err != nil || settings.DefaultRepo == "" {
			b.respondError(s, i, "No repository specified and no default repository set for this channel")
			return
		}
		repo = settings.DefaultRepo
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		b.respondError(s, i, "Invalid repository format. Use: owner/repo")
		return
	}
	owner, repoName := parts[0], parts[1]

	client, err := b.oauth.GetGitHubClient(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	ctx := context.Background()
	opts := &github.IssueListByRepoOptions{
		State:       state,
		ListOptions: github.ListOptions{PerPage: 10},
	}

	issues, _, err := client.Issues.ListByRepo(ctx, owner, repoName, opts)
	if err != nil {
		log.Printf("Failed to list issues: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to list issues: %v", err))
		return
	}

	if len(issues) == 0 {
		b.respondSuccess(s, i, fmt.Sprintf("No %s issues found in %s", state, repo))
		return
	}

	var response strings.Builder
	response.WriteString(fmt.Sprintf("**Issues in %s (%s):**\n\n", repo, state))

	for _, issue := range issues {
		response.WriteString(fmt.Sprintf(
			"**#%d** %s\n%s\n\n",
			issue.GetNumber(),
			issue.GetTitle(),
			issue.GetHTMLURL(),
		))
	}

	b.respondSuccess(s, i, response.String())
}

func (b *Bot) handleIssueView(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	number := b.getIntOption(i.ApplicationCommandData().Options, "number")
	repo := b.getStringOption(i.ApplicationCommandData().Options, "repo")

	if repo == "" {
		settings, err := b.db.GetChannelSettings(i.ChannelID)
		if err != nil || settings.DefaultRepo == "" {
			b.respondError(s, i, "No repository specified and no default repository set for this channel")
			return
		}
		repo = settings.DefaultRepo
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		b.respondError(s, i, "Invalid repository format. Use: owner/repo")
		return
	}
	owner, repoName := parts[0], parts[1]

	client, err := b.oauth.GetGitHubClient(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	ctx := context.Background()
	issue, _, err := client.Issues.Get(ctx, owner, repoName, number)
	if err != nil {
		log.Printf("Failed to get issue: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to get issue: %v", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("#%d %s", issue.GetNumber(), issue.GetTitle()),
		Description: issue.GetBody(),
		URL:         issue.GetHTMLURL(),
		Color:       0x2ea44f,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "State",
				Value:  issue.GetState(),
				Inline: true,
			},
			{
				Name:   "Author",
				Value:  issue.GetUser().GetLogin(),
				Inline: true,
			},
			{
				Name:   "Comments",
				Value:  fmt.Sprintf("%d", issue.GetComments()),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Created: %s", issue.GetCreatedAt().Format("Jan 2, 2006")),
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (b *Bot) handleIssueClose(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	number := b.getIntOption(i.ApplicationCommandData().Options, "number")
	repo := b.getStringOption(i.ApplicationCommandData().Options, "repo")

	if repo == "" {
		settings, err := b.db.GetChannelSettings(i.ChannelID)
		if err != nil || settings.DefaultRepo == "" {
			b.respondError(s, i, "No repository specified and no default repository set for this channel")
			return
		}
		repo = settings.DefaultRepo
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		b.respondError(s, i, "Invalid repository format. Use: owner/repo")
		return
	}
	owner, repoName := parts[0], parts[1]

	client, err := b.oauth.GetGitHubClient(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	ctx := context.Background()
	state := "closed"
	issueRequest := &github.IssueRequest{
		State: &state,
	}

	closedIssue, _, err := client.Issues.Edit(ctx, owner, repoName, number, issueRequest)
	if err != nil {
		log.Printf("Failed to close issue: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to close issue: %v", err))
		return
	}

	b.respondSuccess(s, i, fmt.Sprintf(
		"✅ Issue #%d closed successfully!\n%s",
		closedIssue.GetNumber(),
		closedIssue.GetHTMLURL(),
	))
}

func (b *Bot) handleIssueComment(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	number := b.getIntOption(i.ApplicationCommandData().Options, "number")
	comment := b.getStringOption(i.ApplicationCommandData().Options, "comment")
	repo := b.getStringOption(i.ApplicationCommandData().Options, "repo")

	if repo == "" {
		settings, err := b.db.GetChannelSettings(i.ChannelID)
		if err != nil || settings.DefaultRepo == "" {
			b.respondError(s, i, "No repository specified and no default repository set for this channel")
			return
		}
		repo = settings.DefaultRepo
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		b.respondError(s, i, "Invalid repository format. Use: owner/repo")
		return
	}
	owner, repoName := parts[0], parts[1]

	client, err := b.oauth.GetGitHubClient(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	ctx := context.Background()
	issueComment := &github.IssueComment{
		Body: &comment,
	}

	createdComment, _, err := client.Issues.CreateComment(ctx, owner, repoName, number, issueComment)
	if err != nil {
		log.Printf("Failed to create comment: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to create comment: %v", err))
		return
	}

	b.respondSuccess(s, i, fmt.Sprintf(
		"✅ Comment added to issue #%d\n%s",
		number,
		createdComment.GetHTMLURL(),
	))
}
