package draft

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
)

// uploadAttachments uploads local files and returns attachment metadata
// that can be included in a draft.
func uploadAttachments(client *jmap.Client, paths []string) ([]jmap.Attachment, error) {
	var attachments []jmap.Attachment
	for _, path := range paths {
		att, err := uploadAttachment(client, path)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}
	return attachments, nil
}

func uploadAttachment(client *jmap.Client, path string) (jmap.Attachment, error) {
	file, err := os.Open(path)
	if err != nil {
		return jmap.Attachment{}, fmt.Errorf("failed to read attachment: %w", err)
	}
	defer file.Close()

	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	upload, err := client.UploadBlob(contentType, file)
	if err != nil {
		return jmap.Attachment{}, fmt.Errorf("failed to upload attachment %s: %w", path, err)
	}

	return jmap.Attachment{
		BlobID: upload.BlobID,
		Type:   upload.Type,
		Name:   filepath.Base(path),
		Size:   upload.Size,
	}, nil
}
