package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"strconv"

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

	// Parse GitHub project URL if provided
	// Format: https://github.com/orgs/{org}/projects/{number}
	var projectValue string
	if after, found := strings.CutPrefix(project, "https://github.com/orgs/"); found {
		parts := strings.Split(after, "/")
		if len(parts) >= 3 && parts[1] == "projects" {
			org := parts[0]
			projectNumber := parts[2]
			projectValue = fmt.Sprintf("%s/%s", org, projectNumber)
		} else {
			b.respondError(s, i, "Invalid GitHub project URL format. Expected: https://github.com/orgs/{org}/projects/{number}")
			return
		}
	} else {
		projectValue = project
	}

	settings, err := b.db.GetChannelSettings(channelID)
	if err != nil {
		log.Printf("Failed to get channel settings: %v", err)
		b.respondError(s, i, "Failed to get channel settings")
		return
	}

	settings.DefaultProject = projectValue

	if err := b.db.SaveChannelSettings(settings); err != nil {
		log.Printf("Failed to save channel settings: %v", err)
		b.respondError(s, i, "Failed to save channel settings")
		return
	}

	b.respondSuccess(s, i, fmt.Sprintf("✅ Default project set to: %s", projectValue))
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

func (b *Bot) handleProjectItemsList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	projectNumber := b.getIntOption(i.ApplicationCommandData().Options, "project-number")
	org := b.getStringOption(i.ApplicationCommandData().Options, "org")
	query := b.getStringOption(i.ApplicationCommandData().Options, "query")

	// Default query if not provided
	if query == "" {
		query = "-is:closed -is:done"
	}

	if projectNumber == 0 || org == "" {
		settings, err := b.db.GetChannelSettings(i.ChannelID)
		if err != nil || settings.DefaultProject == "" {
			b.respondError(s, i, "No project specified and no default project set for this channel")
			return
		}

		// Parse default project value in format "org/number"
		defaultOrg, defaultProjectNumber := b.parseProjectValue(settings.DefaultProject)
		if projectNumber == 0 {
			projectNumber = defaultProjectNumber
		}
		if org == "" {
			org = defaultOrg
		}
	}

	if org == "" || projectNumber == 0 {
		b.respondError(s, i, "No organization or project number specified and could not derive from default project.")
		return
	}

	accessToken, err := b.oauth.GetGitHubToken(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	projectItemsResponse, err := b.githubREST.ListProjectItems(org, projectNumber, accessToken, 10, query)
	if err != nil {
		log.Printf("Failed to list project items using REST: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to list project items: %v", err))
		return
	}

	if len(*projectItemsResponse) == 0 {
		b.respondSuccess(s, i, fmt.Sprintf("No items found in project #%d for organization %s with query: %s", projectNumber, org, query))
		return
	}

	var response strings.Builder
	response.WriteString(fmt.Sprintf("**Items in Project #%d for %s:**\n\n", projectNumber, org))

	for _, item := range *projectItemsResponse {
		// Skip if no content
		if item.Content == nil {
			continue
		}

		// Determine item type
		var itemType string
		switch item.ContentType {
		case "Issue":
			itemType = "Issue"
		case "PullRequest":
			itemType = "Pull Request"
		case "DraftIssue":
			itemType = "Draft"
		default:
			itemType = item.ContentType
		}

		title := item.Content.Title
		if title == "" {
			title = "Untitled Item"
		}

		// Get status from fields
		status := "Open"
		for _, field := range item.Fields {
			if name, ok := field["name"].(string); ok && (name == "Status" || name == "status") {
				if value, ok := field["value"].(string); ok {
					status = value
				}
			}
		}

		// Format: **[Issue #123]** This is a table (Status: Open)
		var line string
		if item.Content.Number != 0 {
			if item.Content.HTMLURL != "" {
				line = fmt.Sprintf("**[%s #%d](%s)** %s (Status: %s)\n",
					itemType, item.Content.Number, item.Content.HTMLURL, title, status)
			} else {
				line = fmt.Sprintf("**[%s #%d]** %s (Status: %s)\n",
					itemType, item.Content.Number, title, status)
			}
		} else {
			// Draft issues don't have numbers
			line = fmt.Sprintf("**[%s]** %s (Status: %s)\n",
				itemType, title, status)
		}
		response.WriteString(line)
	}

	b.respondSuccess(s, i, response.String())
}

func (b *Bot) handleProjectList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	org := b.getStringOption(i.ApplicationCommandData().Options, "org")

	if org == "" {
		settings, err := b.db.GetChannelSettings(i.ChannelID)
		if err == nil && settings.DefaultRepo != "" {
			parts := strings.Split(settings.DefaultRepo, "/")
			if len(parts) == 2 {
				org = parts[0]
			}
		}
	}

	if org == "" {
		b.respondError(s, i, "No organization specified and could not derive from default repository.")
		return
	}

	accessToken, err := b.oauth.GetGitHubToken(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	projectsResponse, err := b.githubREST.ListProjects(org, accessToken)
	if err != nil {
		log.Printf("Failed to list projects using REST: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to list projects: %v", err))
		return
	}

	if len(projectsResponse.Projects) == 0 {
		b.respondSuccess(s, i, fmt.Sprintf("No projects found for organization %s.", org))
		return
	}

	var response strings.Builder
	response.WriteString(fmt.Sprintf("**Projects for %s:**\n\n", org))

	for _, project := range projectsResponse.Projects {
		status := "Open"
		if project.Closed {
			status = "Closed"
		}
		response.WriteString(fmt.Sprintf("**#%d** %s - %s\n%s\n\n", project.Number, project.Title, status, project.HTMLURL))
	}

	b.respondSuccess(s, i, response.String())
}

func (b *Bot) handleProjectAddIssue(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	issueNumber := b.getIntOption(i.ApplicationCommandData().Options, "issue-number")
	projectNumber := b.getIntOption(i.ApplicationCommandData().Options, "project-number")
	repo := b.getStringOption(i.ApplicationCommandData().Options, "repo")
	org := b.getStringOption(i.ApplicationCommandData().Options, "org")

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

	if projectNumber == 0 || org == "" {
		settings, err := b.db.GetChannelSettings(i.ChannelID)
		if err != nil || settings.DefaultProject == "" {
			if org == "" {
				org = owner
			}
			if projectNumber == 0 {
				b.respondError(s, i, "No project number specified and no default project set for this channel")
				return
			}
		} else {
			// Parse default project value in format "org/number"
			defaultOrg, defaultProjectNumber := b.parseProjectValue(settings.DefaultProject)
			if projectNumber == 0 {
				projectNumber = defaultProjectNumber
			}
			if org == "" {
				org = defaultOrg
			}
		}
	}

	if org == "" {
		org = owner
	}

	accessToken, err := b.oauth.GetGitHubToken(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	// Get the issue to retrieve its node ID
	client, err := b.oauth.GetGitHubClient(userID)
	if err != nil {
		b.respondError(s, i, "You must authenticate first. Use /gh-auth")
		return
	}

	ctx := context.Background()
	issue, _, err := client.Issues.Get(ctx, owner, repoName, issueNumber)
	if err != nil {
		log.Printf("Failed to get issue: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to get issue: %v", err))
		return
	}

	issueNodeID := issue.GetNodeID()

	// Add issue to project using REST API
	addItemResponse, err := b.githubREST.AddIssueToProject(org, projectNumber, issueNodeID, accessToken)
	if err != nil {
		log.Printf("Failed to add issue to project using REST: %v", err)
		b.respondError(s, i, fmt.Sprintf("Failed to add issue to project: %v", err))
		return
	}

	b.respondSuccess(s, i, fmt.Sprintf(
		"✅ Issue #%d added to project #%d\nItem ID: %s",
		issueNumber,
		projectNumber,
		addItemResponse.ID,
	))
}

// stringToIntOption converts a string to an int.
// It's a helper function for getting int options when they might come from string settings.
func (b *Bot) stringToIntOption(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("Failed to convert string to int: %v", err)
		return 0
	}
	return val
}

// parseProjectValue parses a project value in format "org/number" and returns org and project number.
// If the format is invalid, it returns empty string and 0.
func (b *Bot) parseProjectValue(projectValue string) (org string, projectNumber int) {
	parts := strings.Split(projectValue, "/")
	if len(parts) != 2 {
		log.Printf("Invalid project value format: %s", projectValue)
		return "", 0
	}

	projectNumber = b.stringToIntOption(parts[1])
	return parts[0], projectNumber
}
