package jmap

import "time"

// Mailbox represents a JMAP mailbox (folder).
type Mailbox struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ParentID     string `json:"parentId,omitempty"`
	Role         string `json:"role,omitempty"`
	SortOrder    int    `json:"sortOrder"`
	TotalEmails  int    `json:"totalEmails"`
	UnreadEmails int    `json:"unreadEmails"`
	TotalThreads int    `json:"totalThreads"`
	UnreadThreads int   `json:"unreadThreads"`
}

// EmailAddress represents an email address with optional name.
type EmailAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

// Email represents a JMAP email.
type Email struct {
	ID            string                  `json:"id"`
	ThreadID      string                  `json:"threadId"`
	MailboxIDs    map[string]bool         `json:"mailboxIds,omitempty"`
	Keywords      map[string]bool         `json:"keywords,omitempty"`
	Subject       string                  `json:"subject"`
	From          []EmailAddress          `json:"from,omitempty"`
	To            []EmailAddress          `json:"to,omitempty"`
	CC            []EmailAddress          `json:"cc,omitempty"`
	BCC           []EmailAddress          `json:"bcc,omitempty"`
	ReplyTo       []EmailAddress          `json:"replyTo,omitempty"`
	ReceivedAt    time.Time               `json:"receivedAt"`
	Preview       string                  `json:"preview,omitempty"`
	HasAttachment bool                    `json:"hasAttachment"`
	TextBody      []BodyPart              `json:"textBody,omitempty"`
	HTMLBody      []BodyPart              `json:"htmlBody,omitempty"`
	BodyValues    map[string]BodyValue    `json:"bodyValues,omitempty"`
	Attachments   []Attachment            `json:"attachments,omitempty"`
	MessageID     []string                `json:"messageId,omitempty"`
	InReplyTo     []string                `json:"inReplyTo,omitempty"`
	References    []string                `json:"references,omitempty"`
}

// BodyPart represents a part of the email body.
type BodyPart struct {
	PartID string `json:"partId"`
	Type   string `json:"type"`
}

// BodyValue contains the actual body content.
type BodyValue struct {
	Value string `json:"value"`
}

// Attachment represents an email attachment.
type Attachment struct {
	PartID string `json:"partId"`
	BlobID string `json:"blobId"`
	Type   string `json:"type"`
	Size   int64  `json:"size"`
	Name   string `json:"name,omitempty"`
}

// Identity represents a sender identity.
type Identity struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name,omitempty"`
	MayDelete bool   `json:"mayDelete"`
}

// Thread represents a JMAP thread.
type Thread struct {
	ID       string   `json:"id"`
	EmailIDs []string `json:"emailIds"`
}

// IsUnread returns true if the email hasn't been read.
func (e *Email) IsUnread() bool {
	return !e.Keywords["$seen"]
}

// IsDraft returns true if the email is a draft.
func (e *Email) IsDraft() bool {
	return e.Keywords["$draft"]
}

// String returns a formatted string for an EmailAddress.
func (a EmailAddress) String() string {
	if a.Name != "" {
		return a.Name + " <" + a.Email + ">"
	}
	return a.Email
}

// FormatAddresses formats a slice of email addresses as a string.
func FormatAddresses(addrs []EmailAddress) string {
	if len(addrs) == 0 {
		return ""
	}
	result := ""
	for i, addr := range addrs {
		if i > 0 {
			result += ", "
		}
		result += addr.String()
	}
	return result
}
