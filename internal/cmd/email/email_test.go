package email

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

func mockEmailGetResponse(email map[string]interface{}) httpmock.Responder {
	return httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
		"methodResponses": [][]interface{}{
			{"Email/get", map[string]interface{}{
				"list": []map[string]interface{}{email},
			}, "email"},
		},
	})
}

func mockMailboxResponse(mailboxes []map[string]interface{}) httpmock.Responder {
	return httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
		"methodResponses": [][]interface{}{
			{"Mailbox/get", map[string]interface{}{
				"list": mailboxes,
			}, "mailboxes"},
		},
	})
}

func mockEmailSetResponse(updated map[string]interface{}) httpmock.Responder {
	return httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
		"methodResponses": [][]interface{}{
			{"Email/set", map[string]interface{}{
				"updated": updated,
			}, "result"},
		},
	})
}

// Read command tests

func TestReadCommand(t *testing.T) {
	t.Run("displays email content", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockEmailGetResponse(map[string]interface{}{
				"id":         "email-1",
				"threadId":   "thread-1",
				"subject":    "Hello World",
				"from":       []map[string]string{{"name": "Alice", "email": "alice@example.com"}},
				"to":         []map[string]string{{"name": "Bob", "email": "bob@example.com"}},
				"receivedAt": time.Now().Format(time.RFC3339),
				"textBody":   []map[string]string{{"partId": "1"}},
				"bodyValues": map[string]map[string]string{
					"1": {"value": "This is the email body content that is long enough to be considered substantial."},
				},
			}))

		cmd := NewCmdRead(f)
		cmd.SetArgs([]string{"email-1"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "email-1")
		assert.Contains(t, output, "Hello World")
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "alice@example.com")
		assert.Contains(t, output, "Bob")
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockEmailGetResponse(map[string]interface{}{
				"id":         "email-1",
				"threadId":   "thread-1",
				"subject":    "Test Email",
				"from":       []map[string]string{{"email": "test@example.com"}},
				"receivedAt": "2024-01-15T10:30:00Z",
			}))

		cmd := NewCmdRead(f)
		cmd.SetArgs([]string{"email-1", "--json"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "email-1", result["id"])
		assert.Equal(t, "Test Email", result["subject"])
	})

	t.Run("shows attachments", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockEmailGetResponse(map[string]interface{}{
				"id":         "email-1",
				"threadId":   "thread-1",
				"subject":    "Email with attachment",
				"from":       []map[string]string{{"email": "test@example.com"}},
				"to":         []map[string]string{{"email": "me@example.com"}},
				"receivedAt": time.Now().Format(time.RFC3339),
				"attachments": []map[string]interface{}{
					{"name": "document.pdf", "type": "application/pdf", "size": 12345},
				},
			}))

		cmd := NewCmdRead(f)
		cmd.SetArgs([]string{"email-1"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "Attachments:")
		assert.Contains(t, output, "document.pdf")
		assert.Contains(t, output, "application/pdf")
	})

	t.Run("requires email ID argument", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdRead(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "email ID required")
	})
}

// Thread command tests

func TestThreadCommand(t *testing.T) {
	t.Run("displays thread emails", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Email/get":
					// First call gets the email to find thread ID
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "email-1", "threadId": "thread-1"},
								},
							}, "email"},
						},
					})
				case "Thread/get":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Thread/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "thread-1", "emailIds": []string{"email-1", "email-2"}},
								},
							}, "getThread"},
							{"Email/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{
										"id":         "email-1",
										"threadId":   "thread-1",
										"subject":    "Original message",
										"from":       []map[string]string{{"name": "Alice", "email": "alice@example.com"}},
										"receivedAt": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
										"preview":    "First email content",
									},
									{
										"id":         "email-2",
										"threadId":   "thread-1",
										"subject":    "Re: Original message",
										"from":       []map[string]string{{"name": "Bob", "email": "bob@example.com"}},
										"receivedAt": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
										"preview":    "Reply content",
									},
								},
							}, "emails"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdThread(f)
		cmd.SetArgs([]string{"email-1"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "Thread with 2 emails")
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "Bob")
		assert.Contains(t, output, "Original message")
		assert.Contains(t, output, "Re: Original message")
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				if method == "Email/get" {
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/get", map[string]interface{}{
								"list": []map[string]interface{}{
									{"id": "email-1", "threadId": "thread-1"},
								},
							}, "email"},
						},
					})
				}

				return httpmock.NewJsonResponse(200, map[string]interface{}{
					"methodResponses": [][]interface{}{
						{"Thread/get", map[string]interface{}{
							"list": []map[string]interface{}{
								{"id": "thread-1", "emailIds": []string{"email-1"}},
							},
						}, "getThread"},
						{"Email/get", map[string]interface{}{
							"list": []map[string]interface{}{
								{
									"id":         "email-1",
									"subject":    "Test",
									"receivedAt": "2024-01-15T10:30:00Z",
								},
							},
						}, "emails"},
					},
				})
			})

		cmd := NewCmdThread(f)
		cmd.SetArgs([]string{"email-1", "--json"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)

		var result []map[string]interface{}
		err = json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})
}

// Archive command tests

func TestArchiveCommand(t *testing.T) {
	t.Run("archives single email", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return mockMailboxResponse([]map[string]interface{}{
						{"id": "archive-1", "name": "Archive", "role": "archive"},
					})(req)
				case "Email/set":
					return mockEmailSetResponse(map[string]interface{}{
						"email-1": nil,
					})(req)
				default:
					return httpmock.NewStringResponse(400, "unexpected: "+method), nil
				}
			})

		cmd := NewCmdArchive(f)
		cmd.SetArgs([]string{"email-1"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Moved to Archive")
	})

	t.Run("archives multiple emails", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return mockMailboxResponse([]map[string]interface{}{
						{"id": "archive-1", "name": "Archive", "role": "archive"},
					})(req)
				case "Email/set":
					return httpmock.NewJsonResponse(200, map[string]interface{}{
						"methodResponses": [][]interface{}{
							{"Email/set", map[string]interface{}{
								"updated": map[string]interface{}{
									"email-1": nil,
									"email-2": nil,
									"email-3": nil,
								},
							}, "bulkArchive"},
						},
					})
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdArchive(f)
		cmd.SetArgs([]string{"email-1", "email-2", "email-3"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Archived 3 emails")
	})

	t.Run("requires at least one email ID", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdArchive(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one email ID required")
	})
}

// Move command tests

func TestMoveCommand(t *testing.T) {
	t.Run("moves email to folder by name", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return mockMailboxResponse([]map[string]interface{}{
						{"id": "inbox-1", "name": "Inbox", "role": "inbox"},
						{"id": "work-1", "name": "Work", "role": ""},
					})(req)
				case "Email/set":
					return mockEmailSetResponse(map[string]interface{}{
						"email-1": nil,
					})(req)
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdMove(f)
		cmd.SetArgs([]string{"email-1", "Work"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Moved to Work")
	})

	t.Run("moves email to folder by role", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return mockMailboxResponse([]map[string]interface{}{
						{"id": "inbox-1", "name": "Inbox", "role": "inbox"},
					})(req)
				case "Email/set":
					return mockEmailSetResponse(map[string]interface{}{
						"email-1": nil,
					})(req)
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdMove(f)
		cmd.SetArgs([]string{"email-1", "inbox"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Moved to Inbox")
	})

	t.Run("requires email ID and folder arguments", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdMove(f)
		cmd.SetArgs([]string{"email-1"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "email ID and folder required")
	})
}

// Delete command tests

func TestDeleteCommand(t *testing.T) {
	t.Run("deletes email with --yes and --unsafe flags", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return mockMailboxResponse([]map[string]interface{}{
						{"id": "trash-1", "name": "Trash", "role": "trash"},
					})(req)
				case "Email/set":
					return mockEmailSetResponse(map[string]interface{}{
						"email-1": nil,
					})(req)
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdDelete(f)
		// In test environment (non-TTY), both --yes and --unsafe are needed
		cmd.SetArgs([]string{"email-1", "--yes", "--unsafe"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Moved to Trash")
	})

	t.Run("blocks in safe mode without --unsafe", func(t *testing.T) {
		// Test streams have stdinIsTTY=false, which triggers safe mode
		f, _, _ := setupTest(t)

		cmd := NewCmdDelete(f)
		cmd.SetArgs([]string{"email-1", "--yes"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		var safeModeErr *cmdutil.SafeModeError
		assert.ErrorAs(t, err, &safeModeErr)
	})

	t.Run("allows delete in safe mode with --unsafe", func(t *testing.T) {
		// Test streams have stdinIsTTY=false (safe mode), but --unsafe overrides
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			func(req *http.Request) (*http.Response, error) {
				var jmapReq jmap.Request
				json.NewDecoder(req.Body).Decode(&jmapReq)

				method := jmapReq.MethodCalls[0][0].(string)

				switch method {
				case "Mailbox/get":
					return mockMailboxResponse([]map[string]interface{}{
						{"id": "trash-1", "name": "Trash", "role": "trash"},
					})(req)
				case "Email/set":
					return mockEmailSetResponse(map[string]interface{}{
						"email-1": nil,
					})(req)
				default:
					return httpmock.NewStringResponse(400, "unexpected"), nil
				}
			})

		cmd := NewCmdDelete(f)
		cmd.SetArgs([]string{"email-1", "--yes", "--unsafe"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Moved to Trash")
	})

	t.Run("requires email ID argument", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdDelete(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "email ID required")
	})
}

// HTML to text conversion tests

func TestHtmlToText(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple text",
			html:     "<p>Hello world</p>",
			expected: "Hello world",
		},
		{
			name:     "line breaks",
			html:     "Line 1<br>Line 2<br/>Line 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "paragraphs",
			html:     "<p>First paragraph</p><p>Second paragraph</p>",
			expected: "First paragraph\n\nSecond paragraph",
		},
		{
			name:     "HTML entities",
			html:     "Tom &amp; Jerry &lt;3 &quot;movies&quot;",
			expected: "Tom & Jerry <3 \"movies\"",
		},
		{
			name:     "strips style tags",
			html:     "<style>body { color: red; }</style><p>Content</p>",
			expected: "Content",
		},
		{
			name:     "strips script tags",
			html:     "<script>alert('hi');</script><p>Content</p>",
			expected: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := htmlToText(tt.html)
			assert.Equal(t, tt.expected, result)
		})
	}
}
