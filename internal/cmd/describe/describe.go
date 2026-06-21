package describe

import (
	"fmt"

	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdDescribe creates the describe command.
func NewCmdDescribe(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "describe",
		Short:   "Describe fm capabilities for AI assistants",
		Long:    "Output a comprehensive description of fm commands and capabilities, designed for AI assistants to understand and use.",
		GroupID: "utility",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(f.IOStreams.Out, description)
			return nil
		},
	}
	return cmd
}

const description = `# fm - Fastmail CLI

fm is a command-line interface for Fastmail. It uses the JMAP protocol to interact with Fastmail's API.

## Authentication

Before using fm, authenticate with:
  fm auth login

This stores the API token in the system keychain. Alternatively, set FASTMAIL_TOKEN environment variable.

Check auth status:
  fm auth status

## Commands

### Inbox & Search

List recent inbox emails:
  fm inbox
  fm inbox --limit 20

Search emails:
  fm search "query"
  fm search "from:alice"
  fm search "subject:meeting is:unread"
  fm search "has:attachment after:2024-01-01"

Search operators:
  from:ADDRESS      - Emails from address
  to:ADDRESS        - Emails to address
  subject:TEXT      - Subject contains text
  has:attachment    - Has attachments
  is:unread         - Unread only
  is:flagged        - Starred/flagged only
  before:YYYY-MM-DD - Before date
  after:YYYY-MM-DD  - After date
  in:FOLDER         - In specific folder

Boolean: AND, OR, NOT, parentheses for grouping

JSON output (for parsing):
  fm inbox --json id,subject,from,date
  fm search "query" --json id,subject,from,to,date,preview

JSON fields: id, threadId, subject, from, to, cc, date, preview, unread, attachment

### Reading Emails

Read a specific email by ID:
  fm email read EMAIL_ID

View entire thread:
  fm email thread EMAIL_ID

### Managing Emails

Archive (remove from inbox, keep in archive):
  fm email archive EMAIL_ID

Move to folder:
  fm email move EMAIL_ID "FolderName"

Delete (move to trash):
  fm email delete EMAIL_ID

### Drafts

Create a new draft:
  fm draft new --to "bob@example.com" --subject "Hello"
  fm draft new --to "bob@example.com" --subject "Hello" --body "Message"
  fm draft new --to "bob@example.com" --cc "alice@example.com" --subject "Hello"

Reply to an email (creates draft):
  fm draft reply EMAIL_ID
  fm draft reply EMAIL_ID --body "Thanks!"

Forward an email (creates draft):
  fm draft forward EMAIL_ID --to "someone@example.com"

List drafts:
  fm draft list

Edit a draft:
  fm draft edit DRAFT_ID

Send a draft:
  fm draft send DRAFT_ID

Delete a draft:
  fm draft delete DRAFT_ID

### Folders

List all folders:
  fm folders

Create a folder:
  fm folder create "NewFolder"

Rename a folder:
  fm folder rename FOLDER_ID "NewName"

### Identities (Sender Addresses)

List available sender identities:
  fm identities

When composing, use --from to specify sender:
  fm draft new --from "other@example.com" --to "bob@example.com" --subject "Hi"

## Output Formats

Most list commands support --json for machine-readable output:
  fm inbox --json id,subject,from
  fm search "query" --json id,subject,date
  fm drafts --json id,subject,to

## Environment Variables

FASTMAIL_TOKEN  - API token (overrides keychain)
NO_COLOR        - Disable colored output

## Common Workflows

### Check new emails from someone:
  fm search "from:alice is:unread"

### Read and archive an email:
  fm email read EMAIL_ID
  fm email archive EMAIL_ID

### Reply to an email:
  fm email read EMAIL_ID
  fm draft reply EMAIL_ID --body "Got it, thanks!"
  fm draft send DRAFT_ID

### Find emails with attachments:
  fm search "has:attachment from:team"

### Triage inbox:
  fm inbox --json id,subject,from,preview
  # For each email, decide: archive, reply, delete, or keep

## Error Handling

Exit codes:
  0 - Success
  1 - General error
  2 - Authentication error
  3 - Not found error

## Notes for AI Assistants

- Always read an email before replying to understand context
- Prefer creating drafts over sending directly; let user review
- Use --json output for parsing email lists
- Email IDs look like: Mxxxxxxxxxxxxxxx (M followed by alphanumeric)
- When replying to threads, check which identity received the email and use --from accordingly
- Search is case-insensitive
- Folder names are case-sensitive
`
