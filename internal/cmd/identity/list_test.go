package identity

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

func mockIdentitiesResponse(identities []map[string]interface{}) httpmock.Responder {
	return httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
		"methodResponses": [][]interface{}{
			{"Identity/get", map[string]interface{}{
				"list": identities,
			}, "identities"},
		},
	})
}

func TestListCommand(t *testing.T) {
	t.Run("lists identities in human format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockIdentitiesResponse([]map[string]interface{}{
				{"id": "id-1", "email": "primary@example.com", "name": "John Doe", "mayDelete": false},
				{"id": "id-2", "email": "work@example.com", "name": "John at Work", "mayDelete": true},
			}))

		cmd := NewCmdList(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		output := stdout.String()
		assert.Contains(t, output, "primary@example.com")
		assert.Contains(t, output, "John Doe")
		assert.Contains(t, output, "*")
		assert.Contains(t, output, "work@example.com")
	})

	t.Run("outputs JSON format", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
			mockIdentitiesResponse([]map[string]interface{}{
				{"id": "id-1", "email": "test@example.com", "name": "Test User", "mayDelete": false},
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
		assert.Equal(t, "test@example.com", result[0]["email"])
	})
}
