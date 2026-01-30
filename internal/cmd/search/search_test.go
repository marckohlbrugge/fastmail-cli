package search

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

func mockSearchResponse(emails []map[string]interface{}) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		ids := make([]string, len(emails))
		for i, e := range emails {
			ids[i] = e["id"].(string)
		}

		return httpmock.NewJsonResponse(200, map[string]interface{}{
			"methodResponses": [][]interface{}{
				{"Email/query", map[string]interface{}{"ids": ids}, "query"},
				{"Email/get", map[string]interface{}{"list": emails}, "emails"},
			},
		})
	}
}

func mockMailboxAndSearchResponse(mailboxes []map[string]interface{}, emails []map[string]interface{}) httpmock.Responder {
	return func(req *http.Request) (*http.Response, error) {
		var jmapReq jmap.Request
		json.NewDecoder(req.Body).Decode(&jmapReq)

		method := jmapReq.MethodCalls[0][0].(string)

		switch method {
		case "Mailbox/get":
			return httpmock.NewJsonResponse(200, map[string]interface{}{
				"methodResponses": [][]interface{}{
					{"Mailbox/get", map[string]interface{}{"list": mailboxes}, "mailboxes"},
				},
			})
		case "Email/query":
			ids := make([]string, len(emails))
			for i, e := range emails {
				ids[i] = e["id"].(string)
			}
			return httpmock.NewJsonResponse(200, map[string]interface{}{
				"methodResponses": [][]interface{}{
					{"Email/query", map[string]interface{}{"ids": ids}, "query"},
					{"Email/get", map[string]interface{}{"list": emails}, "emails"},
				},
			})
		default:
			return httpmock.NewStringResponse(400, "unexpected method: "+method), nil
		}
	}
}

func TestSearchCommand(t *testing.T) {
	t.Run("searches with query", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockSearchResponse([]map[string]interface{}{
				{
					"id":         "email-1",
					"threadId":   "thread-1",
					"subject":    "Hello World",
					"from":       []map[string]string{{"name": "Alice", "email": "alice@example.com"}},
					"receivedAt": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					"preview":    "This is a test email",
					"keywords":   map[string]bool{},
				},
			}))

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"hello"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "email-1")
		assert.Contains(t, output, "Hello World")
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "1 results")
	})

	t.Run("searches without query using folder", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockMailboxAndSearchResponse(
				[]map[string]interface{}{
					{"id": "drafts-1", "name": "Drafts", "role": "drafts"},
				},
				[]map[string]interface{}{
					{
						"id":         "draft-1",
						"threadId":   "thread-1",
						"subject":    "My Draft",
						"from":       []map[string]string{{"email": "me@example.com"}},
						"receivedAt": time.Now().Format(time.RFC3339),
						"preview":    "Draft content",
						"keywords":   map[string]bool{"$draft": true},
					},
				}))

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"--folder", "drafts"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "draft-1")
		assert.Contains(t, output, "My Draft")
	})

	t.Run("sends correct filter for OR operator", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		var capturedFilter map[string]interface{}

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				args := jmapReq.MethodCalls[0][1].(map[string]interface{})
				if filter, ok := args["filter"].(map[string]interface{}); ok {
					capturedFilter = filter
				}

				return httpmock.NewJsonResponse(200, map[string]interface{}{
					"methodResponses": [][]interface{}{
						{"Email/query", map[string]interface{}{"ids": []string{}}, "query"},
						{"Email/get", map[string]interface{}{"list": []interface{}{}}, "emails"},
					},
				})
			})

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"hiring OR discount"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Equal(t, "OR", capturedFilter["operator"])
		conditions := capturedFilter["conditions"].([]interface{})
		assert.Len(t, conditions, 2)
	})

	t.Run("sends correct filter for AND operator", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		var capturedFilter map[string]interface{}

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				args := jmapReq.MethodCalls[0][1].(map[string]interface{})
				if filter, ok := args["filter"].(map[string]interface{}); ok {
					capturedFilter = filter
				}

				return httpmock.NewJsonResponse(200, map[string]interface{}{
					"methodResponses": [][]interface{}{
						{"Email/query", map[string]interface{}{"ids": []string{}}, "query"},
						{"Email/get", map[string]interface{}{"list": []interface{}{}}, "emails"},
					},
				})
			})

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"from:alice AND subject:meeting"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Equal(t, "AND", capturedFilter["operator"])
	})

	t.Run("sends correct filter for NOT operator", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		var capturedFilter map[string]interface{}

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				args := jmapReq.MethodCalls[0][1].(map[string]interface{})
				if filter, ok := args["filter"].(map[string]interface{}); ok {
					capturedFilter = filter
				}

				return httpmock.NewJsonResponse(200, map[string]interface{}{
					"methodResponses": [][]interface{}{
						{"Email/query", map[string]interface{}{"ids": []string{}}, "query"},
						{"Email/get", map[string]interface{}{"list": []interface{}{}}, "emails"},
					},
				})
			})

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"NOT spam"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Equal(t, "NOT", capturedFilter["operator"])
	})

	t.Run("outputs JSON format with specified fields", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockSearchResponse([]map[string]interface{}{
				{
					"id":         "email-1",
					"threadId":   "thread-1",
					"subject":    "Test Email",
					"from":       []map[string]string{{"email": "test@example.com"}},
					"receivedAt": "2024-01-15T10:30:00Z",
					"keywords":   map[string]bool{},
				},
			}))

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"test", "--json", "id,subject"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)

		var result []map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "email-1", result[0]["id"])
		assert.Equal(t, "Test Email", result[0]["subject"])
		// Should not include fields that weren't requested
		_, hasThreadId := result[0]["threadId"]
		assert.False(t, hasThreadId)
	})

	t.Run("shows empty message when no results with query", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockSearchResponse([]map[string]interface{}{}))

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"nonexistent"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "No emails found matching: nonexistent")
	})

	t.Run("shows empty message when no results without query", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockMailboxAndSearchResponse(
				[]map[string]interface{}{
					{"id": "folder-1", "name": "Empty", "role": ""},
				},
				[]map[string]interface{}{}))

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"--folder", "Empty"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "No emails found")
		assert.NotContains(t, stdout.String(), "matching:")
	})

	t.Run("respects limit flag", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		var capturedLimit int

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

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
			})

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"test", "--limit", "10"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Equal(t, 10, capturedLimit)
	})

	t.Run("combines folder filter with query", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		var capturedFilter map[string]interface{}

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
					args := jmapReq.MethodCalls[0][1].(map[string]interface{})
					if filter, ok := args["filter"].(map[string]interface{}); ok {
						capturedFilter = filter
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

		cmd := NewCmdSearch(f)
		cmd.SetArgs([]string{"hello", "--folder", "inbox"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		// Should be an AND filter combining text and inMailbox
		assert.Equal(t, "AND", capturedFilter["operator"])
	})
}

func TestSearchCommand_FlagParsing(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantLimit  int
		wantFields []string
	}{
		{
			name:      "defaults",
			args:      []string{"query"},
			wantLimit: 50,
		},
		{
			name:      "custom limit",
			args:      []string{"query", "--limit", "100"},
			wantLimit: 100,
		},
		{
			name:       "json flag with fields",
			args:       []string{"query", "--json", "id,subject,from"},
			wantFields: []string{"id", "subject", "from"},
		},
		{
			name:      "no query with folder",
			args:      []string{"--folder", "drafts"},
			wantLimit: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}
			cmd := NewCmdSearch(f)
			cmd.SetArgs(tt.args)

			err := cmd.ParseFlags(tt.args)
			require.NoError(t, err)

			if tt.wantLimit != 0 {
				limit, _ := cmd.Flags().GetInt("limit")
				assert.Equal(t, tt.wantLimit, limit)
			}

			if tt.wantFields != nil {
				jsonFields, _ := cmd.Flags().GetStringSlice("json")
				assert.Equal(t, tt.wantFields, jsonFields)
			}
		})
	}
}
