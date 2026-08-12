package jmap

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func registerBlobSession() {
	httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"apiUrl":      "https://api.test.com/jmap/api",
			"downloadUrl": "https://api.test.com/jmap/download/{accountId}/{blobId}/{name}?type={type}",
			"uploadUrl":   "https://api.test.com/jmap/upload/{accountId}/",
			"accounts": map[string]interface{}{
				"account-123": map[string]interface{}{},
			},
		}))
}

func TestClient_UploadBlob(t *testing.T) {
	t.Run("uploads data and returns blob info", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		client := newTestClient()
		registerBlobSession()

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/upload/account-123/",
			func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
				assert.Equal(t, "application/pdf", req.Header.Get("Content-Type"))

				return httpmock.NewJsonResponse(200, map[string]interface{}{
					"accountId": "account-123",
					"blobId":    "blob-1",
					"type":      "application/pdf",
					"size":      11,
				})
			})

		upload, err := client.UploadBlob("application/pdf", strings.NewReader("pdf content"))

		require.NoError(t, err)
		assert.Equal(t, "blob-1", upload.BlobID)
		assert.Equal(t, "application/pdf", upload.Type)
		assert.Equal(t, int64(11), upload.Size)
	})

	t.Run("returns error on failed upload", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		client := newTestClient()
		registerBlobSession()

		httpmock.RegisterResponder("POST", "https://api.test.com/jmap/upload/account-123/",
			httpmock.NewStringResponder(413, "too large"))

		_, err := client.UploadBlob("application/pdf", strings.NewReader("pdf content"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "upload failed")
		assert.Contains(t, err.Error(), "too large")
	})

	t.Run("returns error when session has no upload URL", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		client := newTestClient()

		httpmock.RegisterResponder("GET", "https://api.test.com/jmap/session",
			httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
				"apiUrl": "https://api.test.com/jmap/api",
				"accounts": map[string]interface{}{
					"account-123": map[string]interface{}{},
				},
			}))

		_, err := client.UploadBlob("text/plain", strings.NewReader("hi"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "upload URL")
	})
}

func TestClient_DownloadBlob(t *testing.T) {
	t.Run("downloads blob with auth header", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		client := newTestClient()
		registerBlobSession()

		httpmock.RegisterResponder("GET", `=~^https://api\.test\.com/jmap/download/account-123/blob-1/report\.pdf`,
			func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
				return httpmock.NewStringResponse(200, "file content"), nil
			})

		data, err := client.DownloadBlob("blob-1", "report.pdf", "application/pdf")

		require.NoError(t, err)
		assert.Equal(t, "file content", string(data))
	})

	t.Run("escapes attachment name in URL", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		client := newTestClient()
		registerBlobSession()

		var requestedPath string
		httpmock.RegisterResponder("GET", `=~^https://api\.test\.com/jmap/download/`,
			func(req *http.Request) (*http.Response, error) {
				requestedPath = req.URL.EscapedPath()
				return httpmock.NewStringResponse(200, "data"), nil
			})

		_, err := client.DownloadBlob("blob-1", "weird/../name.pdf", "application/pdf")

		require.NoError(t, err)
		assert.Contains(t, requestedPath, "weird%2F..%2Fname.pdf")
	})

	t.Run("returns error on failed download", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		client := newTestClient()
		registerBlobSession()

		httpmock.RegisterResponder("GET", `=~^https://api\.test\.com/jmap/download/account-123/blob-1/report\.pdf`,
			httpmock.NewStringResponder(401, "unauthorized"))

		_, err := client.DownloadBlob("blob-1", "report.pdf", "application/pdf")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "download failed")
	})
}
