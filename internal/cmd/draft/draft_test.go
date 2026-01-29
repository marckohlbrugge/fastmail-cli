package draft

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

// New command tests

func TestNewCommand(t *testing.T) {
	t.Run("creates draft with required fields", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Mailbox/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "drafts-1", "name": "Drafts", "role": "drafts"},
								},
							}, "mailboxes"},
						},
					})
				case "Identity/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Identity/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "id-1", "email": "me@example.com"},
								},
							}, "identities"},
						},
					})
				case "Email/set":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/set", map[string]interface{}{
								"created": map[string]interface{}{
									"draft": map[string]interface{}{"id": "new-draft-1"},
								},
							}, "createDraft"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected: "+method), nil
				}
			})

		cmd := NewCmdNew(f)
		cmd.SetArgs([]string{"--to", "bob@example.com", "--subject", "Hello", "--body", "Hi Bob!"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Draft created: new-draft-1")
	})

	t.Run("creates draft with body from file", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		// Create temp file
		tmpDir := t.TempDir()
		bodyFile := filepath.Join(tmpDir, "body.txt")
		err := os.WriteFile(bodyFile, []byte("Content from file"), 0644)
		require.NoError(t, err)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Mailbox/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "drafts-1", "role": "drafts"},
								},
							}, "mailboxes"},
						},
					})
				case "Identity/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Identity/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "id-1", "email": "me@example.com"},
								},
							}, "identities"},
						},
					})
				case "Email/set":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/set", map[string]interface{}{
								"created": map[string]interface{}{
									"draft": map[string]interface{}{"id": "draft-2"},
								},
							}, "createDraft"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdNew(f)
		cmd.SetArgs([]string{"--to", "bob@example.com", "--subject", "Hello", "--body-file", bodyFile})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err = cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Draft created")
	})

	t.Run("requires --to flag", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdNew(f)
		cmd.SetArgs([]string{"--subject", "Hello"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})

	t.Run("requires --subject flag", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdNew(f)
		cmd.SetArgs([]string{"--to", "bob@example.com"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "required flag")
	})
}

// Reply command tests

func TestReplyCommand(t *testing.T) {
	t.Run("creates reply draft", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Email/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{
										"id":        "original-1",
										"subject":   "Original Message",
										"from":      []map[string]string{{"email": "alice@example.com"}},
										"to":        []map[string]string{{"email": "me@example.com"}},
										"messageId": []string{"<msg-1@example.com>"},
									},
								},
							}, "email"},
						},
					})
				case "Mailbox/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Mailbox/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "drafts-1", "role": "drafts"},
								},
							}, "mailboxes"},
						},
					})
				case "Identity/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Identity/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "id-1", "email": "me@example.com"},
								},
							}, "identities"},
						},
					})
				case "Email/set":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/set", map[string]interface{}{
								"created": map[string]interface{}{
									"draft": map[string]interface{}{"id": "reply-draft-1"},
								},
							}, "createDraft"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected: "+method), nil
				}
			})

		cmd := NewCmdReply(f)
		cmd.SetArgs([]string{"original-1", "--body", "Thanks for your email!"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Reply draft created")
	})

	t.Run("requires --body or --body-file", func(t *testing.T) {
		f, _, _ := setupTest(t)

		cmd := NewCmdReply(f)
		cmd.SetArgs([]string{"original-1"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--body or --body-file required")
	})

	t.Run("requires email ID argument", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdReply(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "email ID required")
	})
}

// Send command tests

func TestSendCommand(t *testing.T) {
	t.Run("blocks in safe mode without --unsafe", func(t *testing.T) {
		f, _, _ := setupTest(t)

		cmd := NewCmdSend(f)
		cmd.SetArgs([]string{"draft-1", "--yes"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		var safeModeErr *cmdutil.SafeModeError
		assert.ErrorAs(t, err, &safeModeErr)
	})

	t.Run("sends draft with --unsafe and --yes", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Email/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{
										"id":       "draft-1",
										"subject":  "Test Draft",
										"to":       []map[string]string{{"email": "bob@example.com"}},
										"keywords": map[string]bool{"$draft": true},
									},
								},
							}, "email"},
						},
					})
				case "Identity/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Identity/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "id-1", "email": "me@example.com"},
								},
							}, "identities"},
						},
					})
				case "Mailbox/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Mailbox/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "sent-1", "name": "Sent", "role": "sent"},
								},
							}, "mailboxes"},
						},
					})
				case "EmailSubmission/set":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"EmailSubmission/set", map[string]interface{}{
								"created": map[string]interface{}{
									"submission": map[string]interface{}{"id": "sub-1"},
								},
							}, "sendEmail"},
							{"Email/set", map[string]interface{}{
								"updated": map[string]interface{}{"draft-1": nil},
							}, "updateEmail"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected: "+method), nil
				}
			})

		cmd := NewCmdSend(f)
		cmd.SetArgs([]string{"draft-1", "--unsafe", "--yes"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Email sent successfully")
	})

	t.Run("rejects non-draft email", func(t *testing.T) {
		f, _, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"methodResponses": [][]interface{}{
					{"Email/get", map[string]interface{}{
						"list": []map[string]interface{}{
							{
								"id":       "email-1",
								"subject":  "Not a draft",
								"keywords": map[string]bool{"$seen": true}, // Not a draft
							},
						},
					}, "email"},
				},
			}))

		cmd := NewCmdSend(f)
		cmd.SetArgs([]string{"email-1", "--unsafe", "--yes"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a draft")
	})

	t.Run("requires draft ID argument", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdSend(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "draft ID required")
	})
}

// Delete command tests

func TestDeleteCommand(t *testing.T) {
	t.Run("blocks in safe mode without --unsafe", func(t *testing.T) {
		f, _, _ := setupTest(t)

		cmd := NewCmdDraftDelete(f)
		cmd.SetArgs([]string{"draft-1", "--yes"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		var safeModeErr *cmdutil.SafeModeError
		assert.ErrorAs(t, err, &safeModeErr)
	})

	t.Run("deletes draft with --unsafe and --yes", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Mailbox/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "trash-1", "name": "Trash", "role": "trash"},
								},
							}, "mailboxes"},
						},
					})
				case "Email/set":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/set", map[string]interface{}{
								"destroyed": []string{"draft-1"},
							}, "deleteDraft"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected: "+method), nil
				}
			})

		cmd := NewCmdDraftDelete(f)
		cmd.SetArgs([]string{"draft-1", "--unsafe", "--yes"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Draft deleted")
	})

	t.Run("requires draft ID argument", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdDraftDelete(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "draft ID required")
	})
}
