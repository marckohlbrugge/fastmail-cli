package jmap

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
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

	// Set up body - prefer both HTML and text if available
	if draft.HTMLBody != "" && draft.TextBody != "" {
		// Both HTML and plain text (best compatibility)
		emailObject["htmlBody"] = []map[string]string{{"partId": "html", "type": "text/html"}}
		emailObject["textBody"] = []map[string]string{{"partId": "text", "type": "text/plain"}}
		emailObject["bodyValues"] = map[string]interface{}{
			"html": map[string]string{"value": draft.HTMLBody},
			"text": map[string]string{"value": draft.TextBody},
		}
	} else if draft.HTMLBody != "" {
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

	// Get original body content
	var originalTextBody, originalHTMLBody string
	if original.BodyValues != nil {
		for _, part := range original.TextBody {
			if bv, ok := original.BodyValues[part.PartID]; ok {
				originalTextBody = bv.Value
				break
			}
		}
		for _, part := range original.HTMLBody {
			if bv, ok := original.BodyValues[part.PartID]; ok {
				originalHTMLBody = bv.Value
				break
			}
		}
	}

	// Build attribution line
	fromStr := FormatAddresses(original.From)
	dateStr := original.ReceivedAt.Format("Mon, Jan 2, 2006 at 3:04 PM")
	attribution := fmt.Sprintf("On %s, %s wrote:", dateStr, fromStr)

	// Build plain text reply with quoted original
	textBody := body + "\n\n" + attribution + "\n" + quoteText(originalTextBody)

	// Build HTML reply with quoted original
	htmlBody := formatReplyHTML(body, attribution, originalHTMLBody, originalTextBody)

	return c.SaveDraft(DraftEmail{
		To:         to,
		CC:         cc,
		Subject:    subject,
		TextBody:   textBody,
		HTMLBody:   htmlBody,
		InReplyTo:  inReplyTo,
		References: references,
	})
}

// quoteText prefixes each line with "> " for plain text quoting.
func quoteText(text string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = "> " + line
	}
	return strings.Join(lines, "\n")
}

// formatReplyHTML creates an HTML reply body with blockquoted original.
func formatReplyHTML(replyText, attribution, originalHTML, originalText string) string {
	// Convert reply text to HTML divs (Fastmail style)
	replyHTML := textToHTMLDivs(replyText)

	// Use original HTML if available, otherwise convert text to HTML divs
	// Strip outer document tags from HTML to avoid style conflicts
	var quotedContent string
	if originalHTML != "" {
		quotedContent = extractHTMLBody(originalHTML)
	} else if originalText != "" {
		quotedContent = textToHTMLDivs(originalText)
	}

	// Match Fastmail's exact format with #qt styles
	return fmt.Sprintf(`<!DOCTYPE html><html><head><title></title></head><body>%s<div><br></div><div>%s</div><blockquote type="cite" id="qt">%s</blockquote><div><br></div></body></html>`,
		replyHTML, html.EscapeString(attribution), quotedContent)
}

// extractHTMLBody extracts just the body content from an HTML document,
// stripping DOCTYPE, html, head, and body tags to avoid style conflicts.
func extractHTMLBody(htmlContent string) string {
	content := htmlContent

	// Remove DOCTYPE
	if idx := strings.Index(strings.ToLower(content), "<!doctype"); idx != -1 {
		if end := strings.Index(content[idx:], ">"); end != -1 {
			content = content[:idx] + content[idx+end+1:]
		}
	}

	// Remove <html> and </html>
	content = removeTag(content, "html")

	// Remove <head>...</head> entirely
	if start := strings.Index(strings.ToLower(content), "<head"); start != -1 {
		if end := strings.Index(strings.ToLower(content[start:]), "</head>"); end != -1 {
			content = content[:start] + content[start+end+7:]
		}
	}

	// Remove <body> and </body> but keep the content
	content = removeTag(content, "body")

	return strings.TrimSpace(content)
}

// removeTag removes opening and closing tags but keeps inner content.
func removeTag(content, tagName string) string {
	lower := strings.ToLower(content)

	// Remove opening tag (may have attributes)
	if start := strings.Index(lower, "<"+tagName); start != -1 {
		if end := strings.Index(content[start:], ">"); end != -1 {
			content = content[:start] + content[start+end+1:]
			lower = strings.ToLower(content)
		}
	}

	// Remove closing tag
	closeTag := "</" + tagName + ">"
	if idx := strings.Index(lower, closeTag); idx != -1 {
		content = content[:idx] + content[idx+len(closeTag):]
	}

	return content
}

// textToHTMLDivs converts plain text to HTML with each line in a <div>.
// Empty lines become <div><br></div> (Fastmail style).
func textToHTMLDivs(text string) string {
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if line == "" {
			result = append(result, "<div><br></div>")
		} else {
			result = append(result, fmt.Sprintf("<div>%s</div>", html.EscapeString(line)))
		}
	}
	return strings.Join(result, "")
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
