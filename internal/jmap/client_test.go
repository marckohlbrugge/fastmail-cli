package jmap

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient() *Client {
	client := NewClient("test-token")
	client.SetBaseURL("https://api.test.com")
	return client
}

func TestNewClient(t *testing.T) {
	client := NewClient("my-token")

	assert.Equal(t, DefaultBaseURL, client.baseURL)
	assert.NotNil(t, client.httpClient)
}

func TestClient_SetBaseURL(t *testing.T) {
	client := NewClient("token")

	client.SetBaseURL("https://custom.api.com/")
	assert.Equal(t, "https://custom.api.com", client.baseURL, "should strip trailing slash")

	client.SetBaseURL("https://another.api.com")
	assert.Equal(t, "https://another.api.com", client.baseURL)
}

func TestClient_GetSession(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := newTestClient()

	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"apiUrl":      "https://api.test.com/jmap/api",
			"downloadUrl": "https://api.test.com/jmap/download",
			"uploadUrl":   "https://api.test.com/jmap/upload",
			"accounts": map[string]interface{}{
				"account-123": map[string]interface{}{
					"name": "test@example.com",
				},
			},
		}))

	session, err := client.GetSession()

	require.NoError(t, err)
	assert.Equal(t, "https://api.test.com/jmap/api", session.APIURL)
	assert.Equal(t, "https://api.test.com/jmap/download", session.DownloadURL)
	assert.Equal(t, "account-123", session.AccountID)
}

func TestClient_GetSession_Cached(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := newTestClient()

	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"apiUrl":   "https://api.test.com/jmap/api",
			"accounts": map[string]interface{}{"acc-1": map[string]interface{}{}},
		}))

	// First call
	_, err := client.GetSession()
	require.NoError(t, err)

	// Second call should use cache
	_, err = client.GetSession()
	require.NoError(t, err)

	assert.Equal(t, 1, httpmock.GetTotalCallCount(), "should only make one HTTP call")
}

func TestClient_GetSession_Unauthorized(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := newTestClient()

	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewStringResponder(401, "Unauthorized"))

	_, err := client.GetSession()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestClient_GetSession_SetsAuthHeader(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := NewClient("secret-token")
	client.SetBaseURL("https://api.test.com")

	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "Bearer secret-token", req.Header.Get("Authorization"))
			return httpmock.NewJsonResponse(200, map[string]interface{}{
				"apiUrl":   "https://api.test.com/jmap/api",
				"accounts": map[string]interface{}{"acc": map[string]interface{}{}},
			})
		})

	_, err := client.GetSession()
	require.NoError(t, err)
}

func TestClient_MakeRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := newTestClient()

	// Register session endpoint
	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"apiUrl":   "https://api.test.com/jmap/api",
			"accounts": map[string]interface{}{"acc-1": map[string]interface{}{}},
		}))

	// Register API endpoint
	httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"methodResponses": [][]interface{}{
				{"Mailbox/get", map[string]interface{}{"list": []interface{}{}}, "0"},
			},
			"sessionState": "state-123",
		}))

	request := &Request{
		Using:       []string{MailCapability},
		MethodCalls: [][]interface{}{{"Mailbox/get", map[string]interface{}{}, "0"}},
	}

	resp, err := client.MakeRequest(request)

	require.NoError(t, err)
	assert.Equal(t, "state-123", resp.SessionState)
	assert.Len(t, resp.MethodResponses, 1)
}

func TestClient_MakeRequest_Error(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := newTestClient()

	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"apiUrl":   "https://api.test.com/jmap/api",
			"accounts": map[string]interface{}{"acc-1": map[string]interface{}{}},
		}))

	httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
		httpmock.NewStringResponder(500, "Internal Server Error"))

	request := &Request{
		Using:       []string{MailCapability},
		MethodCalls: [][]interface{}{{"Mailbox/get", map[string]interface{}{}, "0"}},
	}

	_, err := client.MakeRequest(request)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestClient_AccountID(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := newTestClient()

	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"apiUrl": "https://api.test.com/jmap/api",
			"accounts": map[string]interface{}{
				"my-account-id": map[string]interface{}{},
			},
		}))

	accountID, err := client.AccountID()

	require.NoError(t, err)
	assert.Equal(t, "my-account-id", accountID)
}
