# fm

A command-line interface for [Fastmail](https://www.fastmail.com/) by [Marc Köhlbrugge](https://x.com/marckohlbrugge).

> **Disclaimer:** This is an unofficial tool, not affiliated with Fastmail. Use at your own risk.

## Why fm?

Fastmail has a great web interface, but sometimes you want to:

- **Quickly triage your inbox** without leaving the terminal
- **Let AI agents manage your email** with structured JSON output
- **Script email workflows** for automation
- **Search across your mailbox** with powerful JMAP queries

`fm` gives you all of this through a clean, intuitive CLI inspired by [GitHub CLI](https://cli.github.com/).

## Quick Start

```bash
# Authenticate (stores token in system keychain)
fm auth login

# Check your inbox
fm inbox

# Read an email
fm email read M1234567890

# Search for emails
fm search "from:alice subject:meeting"

# Create and send a draft
fm draft new --to bob@example.com --subject "Hello" --body "Hi Bob!"
fm draft send M9876543210
```

## Installation

### Homebrew

```bash
brew install marckohlbrugge/tap/fm --HEAD
```

### From Source

```bash
git clone https://github.com/marckohlbrugge/fastmail-cli.git
cd fastmail-cli
go build -o fm ./cmd/fm
```

Then move `fm` to somewhere in your PATH, or add the directory to your PATH.

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `fm inbox` | List recent emails in your inbox |
| `fm search <query>` | Search emails with JMAP query syntax |
| `fm folders` | List all mailboxes |

### Email Commands

| Command | Description |
|---------|-------------|
| `fm email read <id>` | Display full email content |
| `fm email thread <id>` | View entire conversation thread |
| `fm email archive <id>` | Archive email(s) |
| `fm email move <id> <folder>` | Move email to a folder |
| `fm email delete <id>` | Move email to trash |

### Draft Commands

| Command | Description |
|---------|-------------|
| `fm draft new` | Create a new draft |
| `fm draft reply <id>` | Reply to an email |
| `fm draft forward <id>` | Forward an email |
| `fm draft edit <id>` | Edit an existing draft |
| `fm draft send <id>` | Send a draft |
| `fm draft delete <id>` | Delete a draft |

### Folder Commands

| Command | Description |
|---------|-------------|
| `fm folder list` | List all folders |
| `fm folder create <name>` | Create a new folder |
| `fm folder rename <id> <name>` | Rename a folder |

## AI-Friendly Output

Every command supports `--json` for machine-readable output, making `fm` perfect for AI agents and automation:

```bash
# Get inbox as JSON
fm inbox --json

# AI agent can parse and act on emails
fm inbox --json | jq '.[0].id' | xargs fm email read --json
```

Example JSON output:

```json
[
  {
    "id": "M1234567890",
    "threadId": "T9876543210",
    "subject": "Meeting tomorrow",
    "from": [{"name": "Alice", "email": "alice@example.com"}],
    "receivedAt": "2024-01-15T10:30:00Z",
    "isUnread": true,
    "preview": "Hi, just wanted to confirm..."
  }
]
```

## Claude Code Integration

If you use [Claude Code](https://docs.anthropic.com/en/docs/claude-code), you can add the included skill to let Claude manage your email:

```bash
mkdir -p ~/.claude/skills
cp claude-code/SKILL.md ~/.claude/skills/fastmail.md
```

Then ask Claude things like:
- "Check my inbox for unread emails"
- "Search for emails from Alice about the project"
- "Draft a reply to the last email from Bob"

See [claude-code/SKILL.md](claude-code/SKILL.md) for the full command reference.

## Authentication

`fm` stores your API token securely in your system's credential store (macOS Keychain, Windows Credential Manager, or Linux Secret Service).

### Setup

1. Go to [Fastmail Settings](https://app.fastmail.com/settings/security/integrations) > Privacy & Security > Integrations
2. Click "New API Token"
3. Give it a name (e.g., "fm-cli") and select permissions
4. Run `fm auth login` and paste your token

```bash
# Interactive login
fm auth login

# Check authentication status
fm auth status

# Log out (removes token from keychain)
fm auth logout
```

### Environment Variable

Alternatively, set the `FASTMAIL_TOKEN` environment variable:

```bash
export FASTMAIL_TOKEN="fmu1-..."
fm inbox
```

## Safety Features

`fm` includes safety measures to prevent accidental data loss:

### Safe Mode

When running non-interactively (piped input, AI agents, scripts), destructive commands are blocked by default:

```bash
# This will fail in safe mode
echo "" | fm draft send M123
# Error: 'fm draft send' is disabled in safe mode.

# Override with --unsafe flag
echo "" | fm draft send M123 --unsafe --yes

# Or via environment variable
FM_UNSAFE=1 fm draft send M123 --yes
```

### Confirmation Prompts

Destructive actions require confirmation:

```bash
fm email delete M123
# Delete email M123? [y/N]

# Skip with --yes flag
fm email delete M123 --yes
```

## Shell Completion

Generate completions for your shell:

```bash
# Zsh
fm completion zsh > "${fpath[1]}/_fm"

# Bash
fm completion bash > /etc/bash_completion.d/fm

# Fish
fm completion fish > ~/.config/fish/completions/fm.fish
```

## Contributing

I'm not accepting pull requests at this time—reviewing external code for security in a tool that handles email requires more time than I can commit to. Feel free to [open an issue](https://github.com/marckohlbrugge/fastmail-cli/issues) for bug reports or feature requests.

## License

MIT
