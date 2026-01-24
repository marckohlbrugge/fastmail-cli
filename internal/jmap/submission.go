package jmap

import (
	"encoding/json"
	"fmt"
)

// SendEmail sends a draft email.
func (c *Client) SendEmail(draftID string) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	// Get identity for sending
	identity, err := c.GetDefaultIdentity()
	if err != nil {
		return err
	}

	// Get sent mailbox
	sentMailbox, err := c.GetMailboxByRole("sent")
	if err != nil {
		return fmt.Errorf("could not find Sent mailbox: %w", err)
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
						"submission": map[string]interface{}{
							"emailId":    draftID,
							"identityId": identity.ID,
						},
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
		return err
	}

	var result struct {
		Created map[string]interface{} `json:"created"`
		NotCreated map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"notCreated"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return err
	}

	if e, ok := result.NotCreated["submission"]; ok {
		return fmt.Errorf("failed to send email: %s - %s", e.Type, e.Description)
	}

	if _, ok := result.Created["submission"]; !ok {
		return fmt.Errorf("failed to send email: no submission created")
	}

	return nil
}

// GetEmailForSending fetches an email with the info needed for send confirmation.
func (c *Client) GetEmailForSending(emailID string) (*Email, error) {
	return c.GetEmailByID(emailID)
}
