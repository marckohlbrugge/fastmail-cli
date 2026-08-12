package jmap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// BlobUpload is the server's response after uploading a blob (RFC 8620 §6).
type BlobUpload struct {
	AccountID string `json:"accountId"`
	BlobID    string `json:"blobId"`
	Type      string `json:"type"`
	Size      int64  `json:"size"`
}

// UploadBlob uploads data to the JMAP upload endpoint. The returned blobId
// can be referenced when creating emails (e.g. as an attachment).
func (c *Client) UploadBlob(contentType string, data io.Reader) (*BlobUpload, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	if session.UploadURL == "" {
		return nil, fmt.Errorf("session has no upload URL")
	}

	uploadURL := strings.ReplaceAll(session.UploadURL, "{accountId}", session.AccountID)

	req, err := http.NewRequest("POST", uploadURL, data)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed: %s - %s", resp.Status, string(body))
	}

	var upload BlobUpload
	if err := json.NewDecoder(resp.Body).Decode(&upload); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &upload, nil
}
