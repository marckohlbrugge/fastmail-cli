package jmap

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// EmailQueryFilter defines filters for email queries.
type EmailQueryFilter struct {
	InMailbox     string `json:"inMailbox,omitempty"`
	Text          string `json:"text,omitempty"`
	From          string `json:"from,omitempty"`
	To            string `json:"to,omitempty"`
	Subject       string `json:"subject,omitempty"`
	HasAttachment *bool  `json:"hasAttachment,omitempty"`
	HasKeyword    string `json:"hasKeyword,omitempty"`
	NotKeyword    string `json:"notKeyword,omitempty"`
	Before        string `json:"before,omitempty"`
	After         string `json:"after,omitempty"`
}

// SearchFilters contains search parameters.
type SearchFilters struct {
	Query         string
	From          string
	To            string
	Subject       string
	HasAttachment *bool
	IsUnread      *bool
	MailboxID     string
	Before        string
	After         string
	Limit         int
}

// Standard email properties for list views
var emailListProperties = []string{
	"id", "threadId", "subject", "from", "to", "receivedAt",
	"preview", "hasAttachment", "keywords",
}

// Extended email properties for full view
var emailFullProperties = []string{
	"id", "threadId", "subject", "from", "to", "cc", "bcc", "replyTo",
	"receivedAt", "textBody", "htmlBody", "attachments", "bodyValues",
	"messageId", "inReplyTo", "references", "keywords",
}

// GetRecentEmails fetches recent emails from a mailbox.
func (c *Client) GetRecentEmails(mailboxID string, limit int) ([]Email, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Email/query",
				map[string]interface{}{
					"accountId": session.AccountID,
					"filter":    map[string]interface{}{"inMailbox": mailboxID},
					"sort":      []map[string]interface{}{{"property": "receivedAt", "isAscending": false}},
					"limit":     limit,
				},
				"query",
			},
			{
				"Email/get",
				map[string]interface{}{
					"accountId":  session.AccountID,
					"#ids":       map[string]interface{}{"resultOf": "query", "name": "Email/query", "path": "/ids"},
					"properties": emailListProperties,
				},
				"emails",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	return c.parseEmailsFromResponse(resp, 1)
}

// GetEmailByID fetches a single email by ID.
func (c *Client) GetEmailByID(emailID string) (*Email, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Email/get",
				map[string]interface{}{
					"accountId":            session.AccountID,
					"ids":                  []string{emailID},
					"properties":           emailFullProperties,
					"bodyProperties":       []string{"partId", "blobId", "type", "size"},
					"fetchTextBodyValues":  true,
					"fetchHTMLBodyValues":  true,
				},
				"email",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	emails, err := c.parseEmailsFromResponse(resp, 0)
	if err != nil {
		return nil, err
	}

	if len(emails) == 0 {
		return nil, fmt.Errorf("email with ID '%s' not found", emailID)
	}

	return &emails[0], nil
}

// GetThread fetches all emails in a thread.
func (c *Client) GetThread(emailOrThreadID string) ([]Email, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	// First, try to get the threadId from the email
	threadID := emailOrThreadID
	email, err := c.GetEmailByID(emailOrThreadID)
	if err == nil && email.ThreadID != "" {
		threadID = email.ThreadID
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Thread/get",
				map[string]interface{}{
					"accountId": session.AccountID,
					"ids":       []string{threadID},
				},
				"getThread",
			},
			{
				"Email/get",
				map[string]interface{}{
					"accountId":           session.AccountID,
					"#ids":                map[string]interface{}{"resultOf": "getThread", "name": "Thread/get", "path": "/list/*/emailIds"},
					"properties":          emailFullProperties,
					"fetchTextBodyValues": true,
				},
				"emails",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	return c.parseEmailsFromResponse(resp, 1)
}

// Search searches for emails matching the given filters.
func (c *Client) Search(filters SearchFilters) ([]Email, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	filter := make(map[string]interface{})

	if filters.Query != "" {
		filter["text"] = filters.Query
	}
	if filters.From != "" {
		filter["from"] = filters.From
	}
	if filters.To != "" {
		filter["to"] = filters.To
	}
	if filters.Subject != "" {
		filter["subject"] = filters.Subject
	}
	if filters.HasAttachment != nil {
		filter["hasAttachment"] = *filters.HasAttachment
	}
	if filters.IsUnread != nil {
		if *filters.IsUnread {
			filter["notKeyword"] = "$seen"
		} else {
			filter["hasKeyword"] = "$seen"
		}
	}
	if filters.MailboxID != "" {
		filter["inMailbox"] = filters.MailboxID
	}
	if filters.Before != "" {
		filter["before"] = filters.Before
	}
	if filters.After != "" {
		filter["after"] = filters.After
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Email/query",
				map[string]interface{}{
					"accountId": session.AccountID,
					"filter":    filter,
					"sort":      []map[string]interface{}{{"property": "receivedAt", "isAscending": false}},
					"limit":     limit,
				},
				"query",
			},
			{
				"Email/get",
				map[string]interface{}{
					"accountId":  session.AccountID,
					"#ids":       map[string]interface{}{"resultOf": "query", "name": "Email/query", "path": "/ids"},
					"properties": emailListProperties,
				},
				"emails",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	return c.parseEmailsFromResponse(resp, 1)
}

// MoveEmail moves an email to a different mailbox.
func (c *Client) MoveEmail(emailID, mailboxID string) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Email/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"update": map[string]interface{}{
						emailID: map[string]interface{}{
							"mailboxIds": map[string]bool{mailboxID: true},
						},
					},
				},
				"moveEmail",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return err
	}

	return c.checkSetError(resp, 0, emailID)
}

// ArchiveEmail moves an email to the archive mailbox.
func (c *Client) ArchiveEmail(emailID string) error {
	archive, err := c.GetMailboxByRole("archive")
	if err != nil {
		return fmt.Errorf("could not find Archive mailbox: %w", err)
	}

	return c.MoveEmail(emailID, archive.ID)
}

// ArchiveEmails archives multiple emails.
func (c *Client) ArchiveEmails(emailIDs []string) (archived int, failed []string, err error) {
	if len(emailIDs) == 0 {
		return 0, nil, nil
	}

	archive, err := c.GetMailboxByRole("archive")
	if err != nil {
		return 0, emailIDs, fmt.Errorf("could not find Archive mailbox: %w", err)
	}

	session, err := c.GetSession()
	if err != nil {
		return 0, emailIDs, err
	}

	update := make(map[string]interface{})
	for _, id := range emailIDs {
		update[id] = map[string]interface{}{
			"mailboxIds": map[string]bool{archive.ID: true},
		}
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Email/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"update":    update,
				},
				"bulkArchive",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return 0, emailIDs, err
	}

	// Parse results
	var result struct {
		Updated    map[string]interface{} `json:"updated"`
		NotUpdated map[string]interface{} `json:"notUpdated"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return 0, emailIDs, err
	}

	for id := range result.NotUpdated {
		failed = append(failed, id)
	}
	archived = len(emailIDs) - len(failed)

	return archived, failed, nil
}

// DeleteEmail moves an email to trash.
func (c *Client) DeleteEmail(emailID string) error {
	trash, err := c.GetMailboxByRole("trash")
	if err != nil {
		return fmt.Errorf("could not find Trash mailbox: %w", err)
	}

	return c.MoveEmail(emailID, trash.ID)
}

// MarkRead marks an email as read or unread.
func (c *Client) MarkRead(emailID string, read bool) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	keywords := map[string]bool{}
	if read {
		keywords["$seen"] = true
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Email/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"update": map[string]interface{}{
						emailID: map[string]interface{}{
							"keywords": keywords,
						},
					},
				},
				"markRead",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return err
	}

	return c.checkSetError(resp, 0, emailID)
}

// DownloadBlob downloads an attachment blob.
func (c *Client) DownloadBlob(blobID, name, contentType string) ([]byte, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	if session.DownloadURL == "" {
		return nil, fmt.Errorf("download URL not available")
	}

	// Build download URL
	url := session.DownloadURL
	url = strings.ReplaceAll(url, "{accountId}", session.AccountID)
	url = strings.ReplaceAll(url, "{blobId}", blobID)
	url = strings.ReplaceAll(url, "{name}", name)
	url = strings.ReplaceAll(url, "{type}", contentType)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// parseEmailsFromResponse extracts emails from a JMAP response.
func (c *Client) parseEmailsFromResponse(resp *Response, index int) ([]Email, error) {
	if len(resp.MethodResponses) <= index {
		return nil, fmt.Errorf("invalid response: missing method response at index %d", index)
	}

	var result struct {
		List     []Email  `json:"list"`
		NotFound []string `json:"notFound"`
	}

	if err := json.Unmarshal(resp.MethodResponses[index][1], &result); err != nil {
		return nil, fmt.Errorf("failed to parse emails: %w", err)
	}

	return result.List, nil
}

// checkSetError checks for errors in an Email/set response.
func (c *Client) checkSetError(resp *Response, index int, id string) error {
	var result struct {
		NotUpdated map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"notUpdated"`
		NotCreated map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"notCreated"`
	}

	if err := json.Unmarshal(resp.MethodResponses[index][1], &result); err != nil {
		return err
	}

	if e, ok := result.NotUpdated[id]; ok {
		return fmt.Errorf("%s: %s", e.Type, e.Description)
	}
	if e, ok := result.NotCreated[id]; ok {
		return fmt.Errorf("%s: %s", e.Type, e.Description)
	}

	return nil
}
