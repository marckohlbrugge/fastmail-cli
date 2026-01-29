package folders

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/marckohlbrugge/fastmail-cli/internal/cmdutil"
	"github.com/marckohlbrugge/fastmail-cli/internal/iostreams"
	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	httpmock.Activate()
	t.Cleanup(httpmock.DeactivateAndReset)

	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"apiUrl": "https://api.test.com/jmap/api",
			"accounts": map[string]interface{}{
				"account-1": map[string]interface{}{},
			},
		}))

	client := jmap.NewClient("test-token")
	client.SetBaseURL("https://api.test.com")

	ios, _, stdout, stderr := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams: ios,
	}
	f.SetJMAPClient(client)

	return f, stdout, stderr
}

func mockMailboxesResponse(mailboxes []map[string]interface{}) httpmock.Responder {
	return httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
		"methodResponses": [][]interface{}{
			{"Mailbox/get", map[string]interface{}{
				"list": mailboxes,
			}, "mailboxes"},
		},
	})
}

func TestFoldersCommand(t *testing.T) {
	t.Run("lists folders in human format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockMailboxesResponse([]map[string]interface{}{
				{"id": "inbox-1", "name": "Inbox", "role": "inbox", "unreadEmails": 5},
				{"id": "sent-1", "name": "Sent", "role": "sent", "unreadEmails": 0},
				{"id": "work-1", "name": "Work", "role": "", "unreadEmails": 2},
			}))

		cmd := NewCmdFolders(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "inbox-1")
		assert.Contains(t, output, "Inbox")
		assert.Contains(t, output, "(inbox)")
		assert.Contains(t, output, "[5 unread]")
		assert.Contains(t, output, "sent-1")
		assert.Contains(t, output, "Sent")
		assert.Contains(t, output, "work-1")
		assert.Contains(t, output, "Work")
		assert.Contains(t, output, "[2 unread]")
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockMailboxesResponse([]map[string]interface{}{
				{"id": "inbox-1", "name": "Inbox", "role": "inbox", "unreadEmails": 3},
				{"id": "archive-1", "name": "Archive", "role": "archive", "unreadEmails": 0},
			}))

		cmd := NewCmdFolders(f)
		cmd.SetArgs([]string{"--json"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)

		var result []map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "inbox-1", result[0]["id"])
		assert.Equal(t, "Inbox", result[0]["name"])
		assert.Equal(t, "inbox", result[0]["role"])
	})

	t.Run("shows message when no mailboxes", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockMailboxesResponse([]map[string]interface{}{}))

		cmd := NewCmdFolders(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "No mailboxes found")
	})

	t.Run("does not show unread count when zero", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockMailboxesResponse([]map[string]interface{}{
				{"id": "sent-1", "name": "Sent", "role": "sent", "unreadEmails": 0},
			}))

		cmd := NewCmdFolders(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "Sent")
		assert.NotContains(t, output, "unread")
	})

	t.Run("does not show role when empty", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockMailboxesResponse([]map[string]interface{}{
				{"id": "custom-1", "name": "My Folder", "role": "", "unreadEmails": 0},
			}))

		cmd := NewCmdFolders(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "My Folder")
		assert.NotContains(t, output, "()")
	})
}
