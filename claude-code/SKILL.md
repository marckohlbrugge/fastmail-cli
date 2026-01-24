---
name: fastmail
description: Check and manage Fastmail emails, drafts, and folders using the fm CLI
---

Help the user manage their Fastmail account using the `fm` CLI tool.

User request: $ARGUMENTS

## Command Reference

### Inbox & Search

```bash
# List recent inbox emails
fm inbox

# Search emails (returns up to 50 by default)
fm search "query"
fm search "query" --limit 100
fm search "query" --folder inbox
fm search "query" --json
```

**Search operators:**
- `from:alice` - Emails from alice
- `to:bob` - Emails to bob
- `subject:hello` - Subject contains hello
- `has:attachment` - Has attachments
- `is:unread` - Unread emails only

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

**Archive all newsletters:**
```bash
fm search "from:newsletter@example.com" --limit 100
# Then archive each by ID
```

**Quick reply:**
```bash
fm email read M123  # Read the email first
fm draft reply M123 --body "Sounds good, thanks!"
fm draft send <draft-id>
```

**Find emails with attachments:**
```bash
fm search "has:attachment from:team"
```

## Environment Variables

- `FASTMAIL_TOKEN` - API token (overrides stored credentials)
- `FM_UNSAFE=1` - Allow destructive operations in non-interactive mode
- `NO_COLOR` - Disable color output
