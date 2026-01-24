package jmap

import (
	"encoding/json"
	"fmt"
)

// DraftEmail contains data for creating a draft.
type DraftEmail struct {
	To         []string
	CC         []string
	BCC        []string
	Subject    string
	TextBody   string
	HTMLBody   string
	From       string
	InReplyTo  string
	References []string
}

// ForwardOptions contains options for forwarding an email.
type ForwardOptions struct {
	EmailID string
	To      []string
	CC      []string
	From    string
	Body    string
}

// SaveDraft creates a new draft email.
func (c *Client) SaveDraft(draft DraftEmail) (string, error) {
	session, err := c.GetSession()
	if err != nil {
		return "", err
	}

	draftsMailbox, err := c.GetMailboxByRole("drafts")
	if err != nil {
		return "", fmt.Errorf("could not find Drafts mailbox: %w", err)
	}

	identity, err := c.GetDefaultIdentity()
	if err != nil {
		return "", err
	}

	fromEmail := draft.From
	if fromEmail == "" {
		fromEmail = identity.Email
	}

	emailObject := map[string]interface{}{
		"mailboxIds": map[string]bool{draftsMailbox.ID: true},
		"keywords":   map[string]bool{"$draft": true},
		"from":       []map[string]string{{"email": fromEmail}},
		"to":         addressesToMap(draft.To),
		"subject":    draft.Subject,
	}

	if len(draft.CC) > 0 {
		emailObject["cc"] = addressesToMap(draft.CC)
	}
	if len(draft.BCC) > 0 {
		emailObject["bcc"] = addressesToMap(draft.BCC)
	}

	if draft.InReplyTo != "" {
		emailObject["inReplyTo"] = []string{draft.InReplyTo}
	}
	if len(draft.References) > 0 {
		emailObject["references"] = draft.References
	}

	if draft.HTMLBody != "" {
		emailObject["htmlBody"] = []map[string]string{{"partId": "html", "type": "text/html"}}
		emailObject["bodyValues"] = map[string]interface{}{"html": map[string]string{"value": draft.HTMLBody}}
	} else {
		emailObject["textBody"] = []map[string]string{{"partId": "text", "type": "text/plain"}}
		emailObject["bodyValues"] = map[string]interface{}{"text": map[string]string{"value": draft.TextBody}}
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Email/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"create": map[string]interface{}{
						"draft": emailObject,
					},
				},
				"createDraft",
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

	if e, ok := result.NotCreated["draft"]; ok {
		return "", fmt.Errorf("failed to create draft: %s", e.Description)
	}

	if created, ok := result.Created["draft"]; ok {
		return created.ID, nil
	}

	return "", fmt.Errorf("failed to create draft: no ID returned")
}

// CreateReplyDraft creates a draft reply to an email.
func (c *Client) CreateReplyDraft(emailID, body string, replyAll bool) (string, error) {
	original, err := c.GetEmailByID(emailID)
	if err != nil {
		return "", err
	}

	identity, err := c.GetDefaultIdentity()
	if err != nil {
		return "", err
	}
	myEmail := identity.Email

	// Determine recipients
	replyToAddrs := original.ReplyTo
	if len(replyToAddrs) == 0 {
		replyToAddrs = original.From
	}

	var to []string
	for _, addr := range replyToAddrs {
		if addr.Email != myEmail {
			to = append(to, addr.Email)
		}
	}

	// For reply-all, include original To and CC
	var cc []string
	if replyAll {
		allRecipients := append(original.To, original.CC...)
		for _, addr := range allRecipients {
			if addr.Email != myEmail && !contains(to, addr.Email) {
				cc = append(cc, addr.Email)
			}
		}
	}

	// If replying to own email, reply to original recipients
	if len(to) == 0 && len(original.To) > 0 {
		for _, addr := range original.To {
			to = append(to, addr.Email)
		}
	}

	// Build subject
	subject := original.Subject
	if len(subject) < 4 || subject[:4] != "Re: " {
		subject = "Re: " + subject
	}

	// Build references chain
	var references []string
	references = append(references, original.References...)
	references = append(references, original.MessageID...)

	var inReplyTo string
	if len(original.MessageID) > 0 {
		inReplyTo = original.MessageID[0]
	}

	return c.SaveDraft(DraftEmail{
		To:         to,
		CC:         cc,
		Subject:    subject,
		TextBody:   body,
		InReplyTo:  inReplyTo,
		References: references,
	})
}

// CreateForwardDraft creates a forward draft with the original message.
func (c *Client) CreateForwardDraft(opts ForwardOptions) (string, error) {
	original, err := c.GetEmailByID(opts.EmailID)
	if err != nil {
		return "", err
	}

	// Build forward subject
	subject := original.Subject
	if len(subject) < 5 || subject[:5] != "Fwd: " {
		subject = "Fwd: " + subject
	}

	// Format original message info
	fromStr := FormatAddresses(original.From)
	toStr := FormatAddresses(original.To)
	dateStr := original.ReceivedAt.Format("Mon, Jan 2, 2006 at 3:04 PM")
	origSubject := original.Subject
	if origSubject == "" {
		origSubject = "(no subject)"
	}

	// Get original body
	var originalBody string
	if original.BodyValues != nil {
		for _, part := range original.TextBody {
			if bv, ok := original.BodyValues[part.PartID]; ok {
				originalBody = bv.Value
				break
			}
		}
	}

	// Build forward body
	forwardBody := opts.Body
	if forwardBody != "" {
		forwardBody += "\n\n"
	}
	forwardBody += fmt.Sprintf(`----- Original message -----
From: %s
To: %s
Subject: %s
Date: %s

%s`, fromStr, toStr, origSubject, dateStr, originalBody)

	return c.SaveDraft(DraftEmail{
		To:       opts.To,
		CC:       opts.CC,
		From:     opts.From,
		Subject:  subject,
		TextBody: forwardBody,
	})
}

// DeleteDraft deletes a draft email.
func (c *Client) DeleteDraft(draftID string) error {
	return c.DeleteEmail(draftID)
}

// addressesToMap converts email strings to JMAP address format.
func addressesToMap(addrs []string) []map[string]string {
	result := make([]map[string]string, len(addrs))
	for i, addr := range addrs {
		result[i] = map[string]string{"email": addr}
	}
	return result
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
