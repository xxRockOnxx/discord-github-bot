<div align="center">

# 🤖 Discord GitHub Bot

### Manage GitHub Issues Without Leaving Discord

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![Discord](https://img.shields.io/badge/Discord-Bot-5865F2?style=for-the-badge&logo=discord&logoColor=white)](https://discord.com)
[![GitHub](https://img.shields.io/badge/GitHub-API-181717?style=for-the-badge&logo=github)](https://github.com)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)

[Features](#-features) • [Quick Start](#-quick-start) • [Usage](#-usage) • [Docker](#-docker-deployment) • [Security](#-security)

---

</div>

## 🎯 Why This Bot?

Stop context-switching between Discord and GitHub. Manage your entire issue workflow from the comfort of your Discord server. Whether you're tracking bugs, managing features, or collaborating with your team, this bot brings GitHub's power directly to Discord.

**Built with security in mind** — Personal OAuth authentication ensures all actions are attributed correctly, and AES-256-GCM encryption keeps your tokens safe.

## ✨ Features

<table>
<tr>
<td width="50%">

### 🔐 Secure Authentication
- **Personal OAuth** - Each user authenticates with their own GitHub account
- **AES-256-GCM Encryption** - Tokens encrypted at rest
- **Ephemeral Auth Links** - Private authentication messages

### ⚡ Issue Management
- ✅ Create issues with rich descriptions
- 📋 List issues (open, closed, or all)
- 🔍 View detailed issue information
- ❌ Close issues directly from Discord
- 💬 Comment and collaborate seamlessly

</td>
<td width="50%">

### 🎛️ Smart Configuration
- **Channel-Specific Defaults** - Set repository per channel
- **Project Integration** - Link to GitHub Projects with pagination and filtering
- **Modern Slash Commands** - Intuitive autocomplete

### 🐳 Deployment Ready
- **Docker Support** - One-command deployment
- **Docker Compose** - Production-ready setup
- **Persistent Storage** - Volume-mounted database

</td>
</tr>
</table>

---

## 🚀 Getting Started

Choose the best option for your needs:

<table>
<tr>
<td width="50%" align="center">

### ☁️ Use Hosted Version
**Recommended for most users**

✅ Zero configuration required
✅ Always up-to-date
✅ Instant setup (1 click)
✅ No server costs

**[➡️ Invite Bot to Discord](https://discord.com/oauth2/authorize?client_id=1443137251159572532&permissions=2147485696&integration_type=0&scope=applications.commands+bot)**

Then jump to [Usage](#-usage) to get started!

</td>
<td width="50%" align="center">

### 🏠 Self-Host
**For advanced users**

✅ Full control over your data
✅ Custom modifications possible
✅ Run on private networks
✅ No external dependencies

**[➡️ Self-Hosting Guide](#-self-hosting)**

Requires Discord Bot App & GitHub OAuth App setup.

</td>
</tr>
</table>

### 🤔 Which Option Should I Choose?

| Feature | Hosted Version | Self-Hosted |
|---------|---------------|-------------|
| Setup Time | < 1 minute | ~15-30 minutes |
| Technical Knowledge | None required | Basic Docker/Go knowledge |
| Cost | Free | Server costs (if applicable) |
| Maintenance | Automatic | Manual updates |
| Data Control | Hosted by bot provider | Full control |
| Custom Features | Not available | Modify as needed |
| Private Network | ❌ | ✅ |

> 💡 **Not sure?** Start with the hosted version - you can always self-host later!

---

## 🏠 Self-Hosting

> ⚠️ **Important:** Self-hosting requires creating your own Discord Bot application in the [Discord Developer Portal](https://discord.com/developers/applications). This is different from the hosted version which uses a pre-configured bot application. You'll get your own invite link after setup.

<details>
<summary>⚡ <b>Quick Deploy (Docker - Recommended)</b></summary>

### Prerequisites

- Docker & Docker Compose installed
- Discord Bot Application ([Create one](https://discord.com/developers/applications)) - **Required for self-hosting**
- GitHub OAuth App ([Create one](https://github.com/settings/developers))

### Deploy in 3 Steps

```bash
# 1. Clone and configure
git clone <your-repo>
cd discord-github-bot
cp .env.example .env

# 2. Edit .env with your credentials (see configuration guide below)
nano .env

# 3. Launch!
docker-compose up -d
```

**Your bot is now running!** View logs with `docker-compose logs -f`

</details>

<details>
<summary>📖 <b>Detailed Self-Hosting Setup</b></summary>

### Self-Hosting Configuration

> **Note:** For self-hosting, you need to create your own Discord Bot application since you'll be running your own instance.

### 1. Create a Discord Bot Application

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **"New Application"** and give it a name
3. Go to the **"Bot"** section and click **"Add Bot"**
4. Under **"Privileged Gateway Intents"**, enable:
   - ✅ SERVER MEMBERS INTENT
   - ✅ MESSAGE CONTENT INTENT
5. Copy the **bot token** (you'll need this for `DISCORD_BOT_TOKEN`)
6. Go to **"OAuth2" > "General"** and copy the **Application ID** (needed for `DISCORD_APPLICATION_ID`)
7. Go to **"OAuth2" > "URL Generator"**:
   - Select scopes: `bot`, `applications.commands`
   - Select bot permissions: `Send Messages`, `Use Slash Commands`
   - Copy the generated URL - this is your **personal invite link** for your self-hosted bot

### 2. Create a GitHub OAuth App

1. Go to [GitHub Settings > Developer settings > OAuth Apps](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in the details:
   - **Application name**: Discord GitHub Bot
   - **Homepage URL**: `http://localhost:8080` (or your domain)
   - **Authorization callback URL**: `http://localhost:8080/callback` (or your domain)
4. Click "Register application"
5. Copy the Client ID
6. Generate a new client secret and copy it

### 3. Configure the Bot

1. Copy the example environment file:

   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and fill in your credentials:

   ```bash
   # Discord Configuration
   DISCORD_BOT_TOKEN=your_discord_bot_token
   DISCORD_APPLICATION_ID=your_application_id

   # GitHub OAuth Configuration
   GITHUB_CLIENT_ID=your_github_client_id
   GITHUB_CLIENT_SECRET=your_github_client_secret
   GITHUB_REDIRECT_URL=http://localhost:8080/callback
   # For Docker: GITHUB_REDIRECT_URL=http://your-domain.com/callback

   # Generate a random 32-byte encryption key
   ENCRYPTION_KEY=your_32_character_encryption_key

   # Server Configuration (optional)
   OAUTH_SERVER_PORT=8080
   OAUTH_SERVER_HOST=localhost
   # For Docker: OAUTH_SERVER_HOST=0.0.0.0

   # Public URL - The publicly accessible URL for OAuth callbacks
   PUBLIC_URL=http://localhost:8080
   # For Docker/Production: PUBLIC_URL=https://your-domain.com

   # Database (optional)
   DATABASE_PATH=./bot.db
   # For Docker: DATABASE_PATH=/home/botuser/data/bot.db
   ```

3. Generate a secure encryption key:

   ```bash
   # On Linux/Mac:
   openssl rand -base64 32 | head -c 32

   # Or use Go:
   go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)[:32]) }'
   ```

### 4. Install Dependencies

```bash
go mod download
```

### 5. Run the Bot

#### Option A: Run with Go (Local Development)

The bot automatically loads environment variables from the `.env` file using godotenv:

```bash
go run main.go
```

Or build and run:

```bash
go build -o discord-github-bot
./discord-github-bot
```

#### Option B: Run with Docker

Using Docker Compose (recommended):

```bash
# Build and start the container
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the container
docker-compose down
```

Using Docker directly:

```bash
# Build the image
docker build -t discord-github-bot .

# Run the container
docker run -d \
  --name discord-github-bot \
  -p 8080:8080 \
  --env-file .env \
  -v bot-data:/home/botuser/data \
  discord-github-bot

# View logs
docker logs -f discord-github-bot

# Stop the container
docker stop discord-github-bot
docker rm discord-github-bot
```

</details>

---

## 📚 Usage

> **For Hosted Version Users:** If you invited the [hosted bot](https://discord.com/oauth2/authorize?client_id=1443137251159572532&permissions=2147485696&integration_type=0&scope=applications.commands+bot), start here!
>
> **For Self-Hosters:** These commands work the same way on your self-hosted instance.

### 🎯 Getting Started

The first time you use the bot, authenticate with your GitHub account:

```
/gh-auth          # Get your personal OAuth link (only you can see it)
```

Click the link, authorize on GitHub, and you're ready! To revoke access later:

```
/gh-unauth        # Remove authentication
```

> 🔒 **Privacy First:** Authentication links are ephemeral (only visible to you)

#### 2️⃣ Configure Channel Defaults

Set up defaults so you don't repeat yourself:

```bash
/gh-set-repo repo:owner/repository    # Set default repo for this channel
/gh-set-project project:123           # Link to GitHub Project
```

### 🎮 Command Reference

<table>
<tr><th>Command</th><th>Description</th><th>Example</th></tr>

<tr>
<td><code>/gh-issue-create</code></td>
<td>Create a new issue</td>
<td><code>/gh-issue-create title:"Login bug" body:"Users can't sign in"</code></td>
</tr>

<tr>
<td><code>/gh-issue-list</code></td>
<td>List issues (open/closed/all), with pagination and filtering</td>
<td><code>/gh-issue-list state:open</code></td>
</tr>

<tr>
<td><code>/gh-issue-view</code></td>
<td>View detailed issue info</td>
<td><code>/gh-issue-view number:42</code></td>
</tr>

<tr>
<td><code>/gh-issue-close</code></td>
<td>Close an issue with a reason (completed, not_planned, or duplicate)</td>
<td><code>/gh-issue-close number:42 state_reason:completed</code></td>
</tr>

<tr>
<td><code>/gh-issue-comment</code></td>
<td>Add a comment to an issue</td>
<td><code>/gh-issue-comment number:42 comment:"Fixed!"</code></td>
</tr>

<tr>
<td><code>/gh-project-item-list</code></td>
<td>List project items (open/closed/all), with pagination and filtering</td>
<td><code>/gh-project-item-list project:123 state:open</code></td>
</tr>

<tr>
<td><code>/gh-project-item-create</code></td>
<td>Create a new project item</td>
<td><code>/gh-project-item-create project:123 title:"New Feature" body:"Implement X"</code></td>
</tr>

<tr>
<td><code>/gh-project-item-view</code></td>
<td>View detailed project item information</td>
<td><code>/gh-project-item-view project:123 item-id:456</code></td>
</tr>

<tr>
<td><code>/gh-project-item-archive</code></td>
<td>Archive a project item</td>
<td><code>/gh-project-item-archive project:123 item-id:456</code></td>
</tr>

</table>

> 💡 **Pro Tip:** All commands support optional `repo:owner/repository` parameter to override channel defaults

---

## 🏗️ Architecture

```
discord-github-bot/
├── 🎯 main.go                      # Application entry point
├── 📦 internal/
│   ├── config/                     # Configuration management
│   ├── database/                   # SQLite + encrypted token storage
│   ├── oauth/                      # GitHub OAuth flow handler
│   └── bot/                        # Discord bot & command handlers
├── 🐳 Dockerfile                   # Container configuration
├── 🐳 docker-compose.yml           # Orchestration setup
└── 📋 .env.example                 # Configuration template
```

<details>
<summary><b>Click to view detailed architecture</b></summary>

### Component Overview

- **`config/`** - Loads and validates environment variables
- **`database/`** - SQLite with AES-256-GCM encrypted token storage
- **`oauth/`** - Handles GitHub OAuth 2.0 flow with state validation
- **`bot/`** - Discord bot initialization and slash command routing

### Data Flow

```
Discord User → /command → Bot Handler → GitHub API
                                ↓
                         Database (encrypted tokens)
```

</details>

---

## 🔒 Security

| Feature | Implementation |
|---------|---------------|
| 🔐 **Token Encryption** | AES-256-GCM encryption at rest |
| 👤 **Personal Auth** | Each user uses their own GitHub account |
| 🙈 **Ephemeral Messages** | Auth links visible only to requesting user |
| ✅ **OAuth 2.0** | Standard flow with state validation |
| 🛡️ **No Shared Secrets** | Zero token sharing between users |

---

## ⚠️ Known Limitations

- 📊 GitHub Projects API support is limited (partial implementation)
- 🌐 OAuth server requires public accessibility for production
- 📄 Issue listings capped at 10 most recent per query
- 🧪 Consider using [ngrok](https://ngrok.com) for local testing

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

```bash
# 1. Fork & clone
git clone https://github.com/xxRockOnxx/discord-github-bot.git

# 2. Create a feature branch
git checkout -b feature/amazing-feature

# 3. Make your changes & commit
git commit -m "Add amazing feature"

# 4. Push & create PR
git push origin feature/amazing-feature
```

### Development Guidelines

- ✅ Write tests for new features
- 📝 Update documentation as needed
- 🎨 Follow existing code style
- 🔍 Test thoroughly before submitting

---

## 🔧 Advanced Self-Hosting

<details>
<summary><b>Docker Management Commands</b></summary>

### Common Docker Operations

```bash
# Start the bot
docker-compose up -d

# View logs
docker-compose logs -f

# Restart the bot
docker-compose restart

# Stop the bot
docker-compose down

# Update to latest version
git pull
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Important Docker Environment Variables

```bash
OAUTH_SERVER_HOST=0.0.0.0              # Accept external connections
DATABASE_PATH=/home/botuser/data/bot.db # Use mounted volume
PUBLIC_URL=https://your-domain.com      # Your public URL for OAuth
```

</details>

<details>
<summary><b>Running Without Docker (Go)</b></summary>

### Local Development

```bash
# Install dependencies
go mod download

# Run directly
go run main.go

# Or build and run
go build -o discord-github-bot
./discord-github-bot
```

The bot automatically loads `.env` file for configuration.

</details>

<details>
<summary><b>Production Deployment Tips</b></summary>

### For Production Self-Hosting:

1. **Use a reverse proxy** (nginx/Caddy) for HTTPS
2. **Set up a domain** pointing to your server
3. **Configure OAuth redirect URL** to use HTTPS
4. **Enable automatic restarts** with systemd or Docker restart policies
5. **Set up log rotation** to prevent disk space issues
6. **Back up your database** regularly (it contains encrypted tokens)

### Example nginx config:

```nginx
server {
    listen 443 ssl;
    server_name bot.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

</details>

---

## 🛠️ Troubleshooting

<details>
<summary>❌ <b>Bot doesn't respond to commands</b></summary>

- ✅ Check bot permissions in Discord server settings
- ✅ Verify slash commands registered (check startup logs)
- ✅ Confirm `DISCORD_BOT_TOKEN` and `DISCORD_APPLICATION_ID` are correct
- ✅ Ensure bot has "Use Slash Commands" permission

</details>

<details>
<summary>🔐 <b>Authentication fails</b></summary>

- ✅ Verify GitHub OAuth App credentials (`GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`)
- ✅ Confirm redirect URL matches exactly (http vs https matters!)
- ✅ Check OAuth server is accessible at configured host/port
- ✅ For Docker: Ensure port 8080 is exposed and accessible externally
- ✅ Try using [ngrok](https://ngrok.com) for local testing

</details>

<details>
<summary>💾 <b>Database errors</b></summary>

- ✅ Ensure `DATABASE_PATH` directory exists and is writable
- ✅ Verify `ENCRYPTION_KEY` is exactly 32 characters
- ✅ Check file permissions: `chmod 644 bot.db`
- ✅ For Docker: Verify volume is mounted correctly (`docker volume ls`)

</details>

<details>
<summary>🐳 <b>Docker-specific issues</b></summary>

**Container exits immediately:**
```bash
docker-compose logs      # Check for error messages
docker logs discord-github-bot
```

**Port already in use:**
- Change `OAUTH_SERVER_PORT` in `.env`
- Or modify port mapping in `docker-compose.yml`: `"8081:8080"`

**Database not persisting:**
```bash
docker volume ls         # Verify volume exists
docker-compose down -v   # Remove and recreate
docker-compose up -d
```

</details>

---

## 📄 License

MIT License - Free to use, modify, and distribute. See [LICENSE](LICENSE) for details.

---

## 💬 Support & Community

### Using the Hosted Version?
- ☁️ **Questions or issues?** The hosted bot is provided as-is
- 💡 **Feature requests?** [Open a discussion](https://github.com/xxRockOnxx/discord-github-bot/discussions)

### Self-Hosting?
- 🐛 **Found a bug?** [Open an issue](https://github.com/xxRockOnxx/discord-github-bot/issues)
- 💬 **Need help?** Check the [troubleshooting guide](#-troubleshooting)
- 🤝 **Want to contribute?** See [contributing guidelines](#-contributing)

### Show Your Support
- ⭐ **Star this repo** if you find it useful!
- 🔄 **Share** with others who might benefit
- 📢 **Spread the word** in your communities

---

<div align="center">

**Made with ❤️ for the Discord & GitHub communities**

[⬆ Back to Top](#-discord-github-bot)

</div>
