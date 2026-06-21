package jmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// BlobUploadResponse is returned after uploading a blob.
type BlobUploadResponse struct {
	BlobID string `json:"blobId"`
	Type   string `json:"type"`
	Size   int64  `json:"size"`
}

// UploadBlob uploads a file and returns the blob info.
// The blobId can be used to attach the file to an email.
func (c *Client) UploadBlob(filename string, contentType string, data []byte) (*BlobUploadResponse, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	if session.UploadURL == "" {
		return nil, fmt.Errorf("upload URL not available")
	}

	// Detect content type from filename if not provided
	if contentType == "" {
		ext := filepath.Ext(filename)
		if ext != "" {
			contentType = mime.TypeByExtension(ext)
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// Build upload URL by substituting template variables
	url := session.UploadURL
	url = strings.ReplaceAll(url, "{accountId}", session.AccountID)

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("blob upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("blob upload failed: %s - %s", resp.Status, string(body))
	}

	var result BlobUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &result, nil
}
