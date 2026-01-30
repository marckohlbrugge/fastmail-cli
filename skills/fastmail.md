---
name: fastmail
description: Check and manage Fastmail emails, drafts, and folders using the fm CLI
---

Help the user manage their Fastmail account using the `fm` CLI tool.

User request: $ARGUMENTS

## Safety Rules (MUST FOLLOW)

**NEVER use `--unsafe` or `FM_UNSAFE=1` without explicit user consent for each specific action.** If a command fails due to safe mode, you MUST ask the user for permission before retrying with unsafe mode. Do not automatically bypass safety features.

**Destructive actions require explicit consent:**
- `fm email delete` - Ask before each delete
- `fm draft delete` - Ask before each delete
- `fm draft send` - Ask before sending (see below)

**Always prefer creating drafts over sending emails.** Unless the user explicitly says "send this email", create a draft instead. This lets the user review before sending. If you're unsure whether to send or draft, ask.

**When triaging multiple emails**, use AskUserQuestion to let the user decide on each email interactively. Suggest options like:
- Archive (recommended for newsletters, notifications)
- Reply (show draft preview)
- Delete
- Skip / Keep in inbox

Example:
```
Email from: newsletter@example.com
Subject: Weekly digest

What would you like to do?
[ ] Archive (Recommended)
[ ] Delete
[ ] Skip
```

## Command Reference

### Inbox & Search

```bash
# List recent inbox emails
fm inbox
fm inbox --limit 10

# JSON output with specific fields
fm inbox --json id,subject,from

# Search emails (returns up to 50 by default)
fm search "query"
fm search "query" --limit 100
fm search "query" --folder inbox

# Search with JSON output
fm search "query" --json id,subject,from,date
```

**Available JSON fields:** `id`, `threadId`, `subject`, `from`, `to`, `cc`, `date`, `preview`, `unread`, `attachment`

**Search operators:**
- `from:alice` - Emails from alice
- `to:bob` - Emails to bob
- `subject:hello` - Subject contains hello
- `has:attachment` - Has attachments
- `is:unread` - Unread emails only
- `is:flagged` - Flagged/starred emails
- `before:YYYY-MM-DD` - Emails before date
- `after:YYYY-MM-DD` - Emails after date

**Boolean operators:** `AND`, `OR`, `NOT`, `()` for grouping

### Reading Emails

```bash
# Read a specific email
fm email read M1234567890

# View entire conversation thread
fm email thread M1234567890
```

### Managing Emails

```bash
# Archive an email
fm email archive M1234567890

# Move to a folder
fm email move M1234567890 "Work"

# Delete (move to trash)
fm email delete M1234567890
```

### Drafts

```bash
# Create a new draft
fm draft new --to bob@example.com --subject "Hello"
fm draft new --to bob@example.com --subject "Hello" --body "Message here"

# Reply to an email
fm draft reply M1234567890
fm draft reply M1234567890 --body "Thanks for the update!"

# Forward an email
fm draft forward M1234567890 --to alice@example.com

# Edit a draft
fm draft edit M1234567890

# Send a draft
fm draft send M1234567890

# Delete a draft
fm draft delete M1234567890
```

### Folders

```bash
# List all folders
fm folders
fm folder list

# Create a folder
fm folder create "Projects"

# Rename a folder
fm folder rename abc123 "New Name"
```

### Authentication

```bash
# Login (stores token in system keychain)
fm auth login

# Check auth status
fm auth status

# Logout
fm auth logout
```

### Help

```bash
# General help
fm --help

# Command-specific help
fm email --help
fm draft --help
fm search --help
```

## Common Workflows

**Check for new emails from someone:**
```bash
fm search "from:alice is:unread"
```

**Triage inbox (interactive):**
1. List emails: `fm inbox`
2. For each email, use AskUserQuestion to let user choose: Archive / Reply / Delete / Skip
3. Only perform destructive actions after user confirms each one

**Reply to an email (prefer drafts):**
```bash
fm email read M123  # Read the email first
fm draft reply M123 --body "Sounds good, thanks!"
# Tell user: "Draft created. Would you like me to send it?"
# Only send after explicit confirmation
```

**Find emails with attachments:**
```bash
fm search "has:attachment from:team"
```

## Environment Variables

- `FASTMAIL_TOKEN` - API token (overrides stored credentials)
- `NO_COLOR` - Disable color output

Note: `FM_UNSAFE=1` exists but should NEVER be used without explicit user consent for each action.
