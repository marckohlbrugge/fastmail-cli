package attachment

import (
	"bytes"
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
			"apiUrl":      "https://api.test.com/jmap/api",
			"downloadUrl": "https://api.test.com/jmap/download/{accountId}/{blobId}/{name}?type={type}",
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

// registerEmailWithAttachments registers an Email/get responder for an email
// with the given attachments.
func registerEmailWithAttachments(attachments []map[string]interface{}) {
	httpmock.RegisterResponder("POST", "https://api.test.com/jmap/api",
		httpmock.NewJsonResponderOrPanic(200, map[string]interface{}{
			"methodResponses": [][]interface{}{
				{"Email/get", map[string]interface{}{
					"list": []map[string]interface{}{
						{
							"id":          "email-1",
							"subject":     "With attachments",
							"attachments": attachments,
						},
					},
				}, "email"},
			},
		}))
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// List command tests

func TestListCommand(t *testing.T) {
	t.Run("lists attachments", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "application/pdf", "size": 2048, "name": "report.pdf"},
			{"partId": "4", "blobId": "blob-2", "type": "image/png", "size": 100, "name": "logo.png"},
		})

		cmd := NewCmdList(f)
		cmd.SetArgs([]string{"email-1"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "report.pdf")
		assert.Contains(t, stdout.String(), "application/pdf")
		assert.Contains(t, stdout.String(), "2.0 KB")
		assert.Contains(t, stdout.String(), "blob-1")
		assert.Contains(t, stdout.String(), "logo.png")
	})

	t.Run("reports no attachments", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		registerEmailWithAttachments(nil)

		cmd := NewCmdList(f)
		cmd.SetArgs([]string{"email-1"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "No attachments.")
	})

	t.Run("outputs JSON", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "application/pdf", "size": 2048, "name": "report.pdf"},
		})

		cmd := NewCmdList(f)
		cmd.SetArgs([]string{"email-1", "--json"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), `"blobId": "blob-1"`)
		assert.Contains(t, stdout.String(), `"name": "report.pdf"`)
	})

	t.Run("requires email ID argument", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdList(f)
		cmd.SetArgs([]string{})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "email ID required")
	})
}

// Download command tests

func TestDownloadCommand(t *testing.T) {
	registerDownload := func(content string) {
		httpmock.RegisterResponder("GET", `=~^https://api\.test\.com/jmap/download/`,
			httpmock.NewStringResponder(200, content))
	}

	t.Run("downloads attachment by name", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "application/pdf", "size": 11, "name": "report.pdf"},
		})
		registerDownload("pdf content")

		outputPath := filepath.Join(t.TempDir(), "report.pdf")

		cmd := NewCmdDownload(f)
		cmd.SetArgs([]string{"email-1", "report.pdf", "--output", outputPath})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Downloaded")

		data, err := os.ReadFile(outputPath)
		require.NoError(t, err)
		assert.Equal(t, "pdf content", string(data))
	})

	t.Run("downloads attachment by blob ID", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "application/pdf", "size": 11, "name": "report.pdf"},
		})
		registerDownload("pdf content")

		outputPath := filepath.Join(t.TempDir(), "out.pdf")

		cmd := NewCmdDownload(f)
		cmd.SetArgs([]string{"email-1", "blob-1", "--output", outputPath})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		assert.FileExists(t, outputPath)
	})

	t.Run("defaults to sanitized attachment name", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "text/plain", "size": 4, "name": "../../evil.txt"},
		})
		registerDownload("data")

		chdir(t, t.TempDir())

		cmd := NewCmdDownload(f)
		cmd.SetArgs([]string{"email-1", "blob-1"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		// File must be written inside the current directory, not two levels up.
		assert.FileExists(t, "evil.txt")
		assert.NoFileExists(t, filepath.Join("..", "..", "evil.txt"))
	})

	t.Run("errors when attachment not found", func(t *testing.T) {
		f, _, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "application/pdf", "size": 11, "name": "report.pdf"},
		})

		cmd := NewCmdDownload(f)
		cmd.SetArgs([]string{"email-1", "missing.pdf"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no attachment named 'missing.pdf'")
	})

	t.Run("errors on ambiguous name", func(t *testing.T) {
		f, _, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "application/pdf", "size": 11, "name": "report.pdf"},
			{"partId": "4", "blobId": "blob-2", "type": "application/pdf", "size": 22, "name": "report.pdf"},
		})

		cmd := NewCmdDownload(f)
		cmd.SetArgs([]string{"email-1", "report.pdf"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "multiple attachments named 'report.pdf'")
	})

	t.Run("refuses to overwrite existing file without --force", func(t *testing.T) {
		f, _, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "application/pdf", "size": 11, "name": "report.pdf"},
		})
		registerDownload("new content")

		outputPath := filepath.Join(t.TempDir(), "report.pdf")
		require.NoError(t, os.WriteFile(outputPath, []byte("existing"), 0644))

		cmd := NewCmdDownload(f)
		cmd.SetArgs([]string{"email-1", "report.pdf", "--output", outputPath})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		data, _ := os.ReadFile(outputPath)
		assert.Equal(t, "existing", string(data), "existing file must not be modified")
	})

	t.Run("overwrites existing file with --force", func(t *testing.T) {
		f, stdout, _ := setupTest(t)

		registerEmailWithAttachments([]map[string]interface{}{
			{"partId": "3", "blobId": "blob-1", "type": "application/pdf", "size": 11, "name": "report.pdf"},
		})
		registerDownload("new content")

		outputPath := filepath.Join(t.TempDir(), "report.pdf")
		require.NoError(t, os.WriteFile(outputPath, []byte("existing"), 0644))

		cmd := NewCmdDownload(f)
		cmd.SetArgs([]string{"email-1", "report.pdf", "--output", outputPath, "--force"})
		cmd.SetOut(stdout)
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.NoError(t, err)
		data, _ := os.ReadFile(outputPath)
		assert.Equal(t, "new content", string(data))
	})

	t.Run("requires email ID and attachment arguments", func(t *testing.T) {
		f := &cmdutil.Factory{}
		cmd := NewCmdDownload(f)
		cmd.SetArgs([]string{"email-1"})
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "email ID and attachment name or blob ID required")
	})
}
