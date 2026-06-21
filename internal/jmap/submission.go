package jmap

import (
	"encoding/json"
	"fmt"
	"time"
)

// SendOptions configures email sending behavior.
type SendOptions struct {
	HoldFor int // Seconds to hold before sending (0 = immediate)
}

// EmailSubmission represents a pending or completed email submission.
type EmailSubmission struct {
	ID         string     `json:"id"`
	EmailID    string     `json:"emailId"`
	IdentityID string     `json:"identityId"`
	SendAt     *time.Time `json:"sendAt"`
	UndoStatus string     `json:"undoStatus"` // pending, final, canceled
}

// SendEmail sends a draft email with optional hold for undo.
// Returns the submission ID (useful for cancellation if holdFor > 0).
func (c *Client) SendEmail(draftID string, opts *SendOptions) (string, error) {
	session, err := c.GetSession()
	if err != nil {
		return "", err
	}

	// Get identity for sending
	identity, err := c.GetDefaultIdentity()
	if err != nil {
		return "", err
	}

	// Get sent mailbox
	sentMailbox, err := c.GetMailboxByRole("sent")
	if err != nil {
		return "", fmt.Errorf("could not find Sent mailbox: %w", err)
	}

	// Build submission object
	submission := map[string]interface{}{
		"emailId":    draftID,
		"identityId": identity.ID,
	}

	// Add holdFor for delayed/undo send
	if opts != nil && opts.HoldFor > 0 {
		submission["envelope"] = map[string]interface{}{
			"mailFrom": map[string]interface{}{
				"email": identity.Email,
				"parameters": map[string]interface{}{
					"holdFor": opts.HoldFor,
				},
			},
			// ponytail: rcptTo populated by server from email headers
		}
	}

	// Create EmailSubmission and update the email's mailbox in one request
	request := &Request{
		Using: []string{CoreCapability, MailCapability, SubmissionCapability},
		MethodCalls: [][]interface{}{
			{
				"EmailSubmission/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"create": map[string]interface{}{
						"submission": submission,
					},
					"onSuccessUpdateEmail": map[string]interface{}{
						"#submission": map[string]interface{}{
							"mailboxIds/" + sentMailbox.ID: true,
							"keywords/$draft":              nil,
						},
					},
				},
				"sendEmail",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return "", err
	}

	var result struct {
		Created map[string]struct {
			ID string `json:"id"`
		} `json:"created"`
		NotCreated map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"notCreated"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return "", err
	}

	if e, ok := result.NotCreated["submission"]; ok {
		return "", fmt.Errorf("failed to send email: %s - %s", e.Type, e.Description)
	}

	created, ok := result.Created["submission"]
	if !ok {
		return "", fmt.Errorf("failed to send email: no submission created")
	}

	return created.ID, nil
}

// GetPendingSubmissions returns submissions that haven't been sent yet.
func (c *Client) GetPendingSubmissions() ([]EmailSubmission, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability, SubmissionCapability},
		MethodCalls: [][]interface{}{
			{
				"EmailSubmission/query",
				map[string]interface{}{
					"accountId": session.AccountID,
					"filter": map[string]interface{}{
						"undoStatus": "pending",
					},
				},
				"query",
			},
			{
				"EmailSubmission/get",
				map[string]interface{}{
					"accountId":  session.AccountID,
					"#ids":       map[string]interface{}{"resultOf": "query", "name": "EmailSubmission/query", "path": "/ids"},
					"properties": []string{"id", "emailId", "identityId", "sendAt", "undoStatus"},
				},
				"get",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	if len(resp.MethodResponses) < 2 {
		return nil, fmt.Errorf("invalid response")
	}

	var result struct {
		List []EmailSubmission `json:"list"`
	}
	if err := json.Unmarshal(resp.MethodResponses[1][1], &result); err != nil {
		return nil, err
	}

	return result.List, nil
}

// CancelSubmission cancels a pending email submission (undo send).
func (c *Client) CancelSubmission(submissionID string) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability, SubmissionCapability},
		MethodCalls: [][]interface{}{
			{
				"EmailSubmission/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"update": map[string]interface{}{
						submissionID: map[string]interface{}{
							"undoStatus": "canceled",
						},
					},
				},
				"cancel",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return err
	}

	var result struct {
		Updated    map[string]interface{} `json:"updated"`
		NotUpdated map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"notUpdated"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return err
	}

	if e, ok := result.NotUpdated[submissionID]; ok {
		return fmt.Errorf("failed to cancel: %s - %s", e.Type, e.Description)
	}

	return nil
}

// GetEmailForSending fetches an email with the info needed for send confirmation.
func (c *Client) GetEmailForSending(emailID string) (*Email, error) {
	return c.GetEmailByID(emailID)
}
