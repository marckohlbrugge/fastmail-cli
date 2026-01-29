package folder

import (
	"bytes"
	"encoding/json"
	"net/http"
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

// List command tests

func TestListCommand(t *testing.T) {
	t.Run("lists folders in human format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"methodResponses": [][]interface{}{
					{"Mailbox/get", map[string]interface{}{
						"list": []map[string]interface{}{
							{"id": "inbox-1", "name": "Inbox", "role": "inbox", "unreadEmails": 3},
							{"id": "work-1", "name": "Work", "role": "", "unreadEmails": 0},
						},
					}, "mailboxes"},
				},
			}))

		cmd := NewCmdList(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "inbox-1")
		assert.Contains(t, output, "Inbox")
		assert.Contains(t, output, "(inbox)")
		assert.Contains(t, output, "[3 unread]")
		assert.Contains(t, output, "work-1")
		assert.Contains(t, output, "Work")
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"methodResponses": [][]interface{}{
					{"Mailbox/get", map[string]interface{}{
						"list": []map[string]interface{}{
							{"id": "inbox-1", "name": "Inbox", "role": "inbox"},
						},
					}, "mailboxes"},
				},
			}))

		cmd := NewCmdList(f)
		cmd.SetArgs([]string{"--json"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)

		var result []map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "inbox-1", result[0]["id"])
	})
}

// Create command tests

func TestCreateCommand(t *testing.T) {
	t.Run("creates folder", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"methodResponses": [][]interface{}{
					{"Mailbox/set", map[string]interface{}{
						"created": map[string]interface{}{
							"newMailbox": map[string]interface{}{"id": "new-folder-1"},
						},
					}, "createMailbox"},
				},
			}))

		cmd := NewCmdCreate(f)
		cmd.SetArgs([]string{"Work Projects"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Folder created: new-folder-1")
	})

	t.Run("creates nested folder with parent", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		var capturedParent string

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				args := jmapReq.MethodCalls[0][1].(map[string]interface{})
				if create, ok := args["create"].(map[string]interface{}); ok {
					if newFolder, ok := create["newMailbox"].(map[string]interface{}); ok {
						if parent, ok := newFolder["parentId"].(string); ok {
							capturedParent = parent
						}
					}
				}

				return httpmock.NewJsonResponse(200, map[string]interface{}{
					"methodResponses": [][]interface{}{
						{"Mailbox/set", map[string]interface{}{
							"created": map[string]interface{}{
								"newMailbox": map[string]interface{}{"id": "nested-folder-1"},
							},
						}, "createMailbox"},
					},
				})
			})

		cmd := NewCmdCreate(f)
		cmd.SetArgs([]string{"Q1 Reports", "--parent", "parent-folder-id"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Folder created")
		assert.Equal(t, "parent-folder-id", capturedParent)
	})

	t.Run("requires folder name argument", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdCreate(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "folder name required")
	})
}

// Rename command tests

func TestRenameCommand(t *testing.T) {
	t.Run("renames folder", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"methodResponses": [][]interface{}{
					{"Mailbox/set", map[string]interface{}{
						"updated": map[string]interface{}{
							"folder-1": nil,
						},
					}, "renameMailbox"},
				},
			}))

		cmd := NewCmdRename(f)
		cmd.SetArgs([]string{"folder-1", "New Name"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Folder renamed")
	})

	t.Run("sends correct rename request", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		var capturedFolderID, capturedNewName string

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				args := jmapReq.MethodCalls[0][1].(map[string]interface{})
				if update, ok := args["update"].(map[string]interface{}); ok {
					for id, changes := range update {
						capturedFolderID = id
						if changeMap, ok := changes.(map[string]interface{}); ok {
							if name, ok := changeMap["name"].(string); ok {
								capturedNewName = name
							}
						}
					}
				}

				return httpmock.NewJsonResponse(200, map[string]interface{}{
					"methodResponses": [][]interface{}{
						{"Mailbox/set", map[string]interface{}{
							"updated": map[string]interface{}{
								capturedFolderID: nil,
							},
						}, "renameMailbox"},
					},
				})
			})

		cmd := NewCmdRename(f)
		cmd.SetArgs([]string{"abc123", "Renamed Folder"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Equal(t, "abc123", capturedFolderID)
		assert.Equal(t, "Renamed Folder", capturedNewName)
	})

	t.Run("requires folder ID and new name", func(t *testing.T) {
		f := &cmdutil.Factory{}

		tests := []struct {
			name string
			args []string
		}{
			{"no arguments", []string{}},
			{"only folder ID", []string{"folder-1"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cmd := NewCmdRename(f)
				cmd.SetArgs(tt.args)
				cmd.SetOut(&bytes.Buffer{})
				cmd.SetErr(&bytes.Buffer{})

				err := cmd.Execute()

				require.Error(t, err)
				assert.Contains(t, err.Error(), "folder ID and new name required")
			})
		}
	})
}
