package inbox

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

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

	// Register session endpoint
	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"apiUrl": "https://api.test.com/jmap/api",
			"accounts": map[string]interface{}{
				"account-1": map[string]interface{}{},
			},
		}))

	// Create test client
	client := jmap.NewClient("test-token")
	client.SetBaseURL("https://api.test.com")

	// Create factory with test streams
	ios, _, stdout, stderr := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams: ios,
	}
	f.SetJMAPClient(client)

	return f, stdout, stderr
}

func TestInboxCommand(t *testing.T) {
	t.Run("lists emails in human format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		// Register API responses
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
									{"id": "inbox-1", "name": "Inbox", "role": "inbox"},
								},
							}, "mailboxes"},
						},
					})
				case "Email/query":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/query", map[string]interface{}{"ids": []string{"email-1", "email-2"}}, "query"},
							{"Email/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{
										"id":         "email-1",
										"threadId":   "thread-1",
										"subject":    "Hello World",
										"from":       []map[string]string{{"name": "Alice", "email": "alice@example.com"}},
										"receivedAt": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
										"preview":    "This is a test email",
										"keywords":   map[string]bool{},
									},
									{
										"id":         "email-2",
										"threadId":   "thread-2",
										"subject":    "Meeting Tomorrow",
										"from":       []map[string]string{{"name": "Bob", "email": "bob@example.com"}},
										"receivedAt": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
										"preview":    "Don't forget our meeting",
										"keywords":   map[string]bool{"$seen": true},
									},
								},
							}, "emails"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdInbox(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "email-1")
		assert.Contains(t, output, "Hello World")
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "email-2")
		assert.Contains(t, output, "Meeting Tomorrow")
		assert.Contains(t, output, "2 emails")
	})

	t.Run("outputs JSON format", func(t *testing.T) {
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
									{"id": "inbox-1", "role": "inbox"},
								},
							}, "mailboxes"},
						},
					})
				case "Email/query":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/query", map[string]interface{}{"ids": []string{"email-1"}}, "query"},
							{"Email/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{
										"id":         "email-1",
										"threadId":   "thread-1",
										"subject":    "Test Email",
										"from":       []map[string]string{{"email": "test@example.com"}},
										"receivedAt": "2024-01-15T10:30:00Z",
										"keywords":   map[string]bool{},
									},
								},
							}, "emails"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdInbox(f)
		cmd.SetArgs([]string{"--json"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()

		// Verify it's valid JSON
		var result []map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "email-1", result[0]["id"])
		assert.Equal(t, "Test Email", result[0]["subject"])
		assert.Equal(t, true, result[0]["isUnread"])
	})

	t.Run("shows empty message when no emails", func(t *testing.T) {
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
									{"id": "inbox-1", "role": "inbox"},
								},
							}, "mailboxes"},
						},
					})
				case "Email/query":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/query", map[string]interface{}{"ids": []string{}}, "query"},
							{"Email/get", map[string]interface{}{"list": []interface{}{}}, "emails"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdInbox(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "No emails found")
	})

	t.Run("respects limit flag", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		var capturedLimit int

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
									{"id": "inbox-1", "role": "inbox"},
								},
							}, "mailboxes"},
						},
					})
				case "Email/query":
					// Capture the limit from the request
					args := jmapReq.MethodCalls[0][1].(map[string]interface{})
					if limit, ok := args["limit"].(float64); ok {
						capturedLimit = int(limit)
					}
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/query", map[string]interface{}{"ids": []string{}}, "query"},
							{"Email/get", map[string]interface{}{"list": []interface{}{}}, "emails"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdInbox(f)
		cmd.SetArgs([]string{"--limit", "5"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Equal(t, 5, capturedLimit)
	})

	t.Run("validates custom fields", func(t *testing.T) {
		f, _, stderr := setupTest(t)

		// Only need mailbox response, validation happens before email fetch
		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"methodResponses": [][]interface{}{
					{"Mailbox/get", map[string]interface{}{
						"list": []map[string]interface{}{
							{"id": "inbox-1", "role": "inbox"},
						},
					}, "mailboxes"},
				},
			}))

		cmd := NewCmdInbox(f)
		cmd.SetArgs([]string{"--fields", "invalid_field"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(stderr)

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown field")
	})
}

func TestInboxCommand_FlagParsing(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantLimit  int
		wantJSON   bool
		wantFields string
	}{
		{
			name:      "defaults",
			args:      []string{},
			wantLimit: 20,
			wantJSON:  false,
		},
		{
			name:      "custom limit",
			args:      []string{"--limit", "10"},
			wantLimit: 10,
		},
		{
			name:     "json output",
			args:     []string{"--json"},
			wantJSON: true,
		},
		{
			name:       "custom fields",
			args:       []string{"--fields", "id,subject"},
			wantFields: "id,subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}
			cmd := NewCmdInbox(f)
			cmd.SetArgs(tt.args)

			// Parse flags without executing
			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			if tt.wantLimit != 0 {
				limit, _ := cmd.Flags().GetInt("limit")
				assert.Equal(t, tt.wantLimit, limit)
			}

			jsonFlag, _ := cmd.Flags().GetBool("json")
			assert.Equal(t, tt.wantJSON, jsonFlag)

			if tt.wantFields != "" {
				fields, _ := cmd.Flags().GetString("fields")
				assert.Equal(t, tt.wantFields, fields)
			}
		})
	}
}
