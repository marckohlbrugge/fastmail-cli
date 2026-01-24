package jmap

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GetMailboxes fetches all mailboxes.
func (c *Client) GetMailboxes() ([]Mailbox, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Mailbox/get",
				map[string]interface{}{
					"accountId": session.AccountID,
				},
				"mailboxes",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	var result struct {
		List []Mailbox `json:"list"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return nil, fmt.Errorf("failed to parse mailboxes: %w", err)
	}

	return result.List, nil
}

// GetMailboxByRole finds a mailbox by its role (e.g., "inbox", "archive", "trash").
func (c *Client) GetMailboxByRole(role string) (*Mailbox, error) {
	mailboxes, err := c.GetMailboxes()
	if err != nil {
		return nil, err
	}

	role = strings.ToLower(role)

	// First try to find by exact role
	for _, mb := range mailboxes {
		if strings.ToLower(mb.Role) == role {
			return &mb, nil
		}
	}

	// Fall back to name matching
	for _, mb := range mailboxes {
		if strings.ToLower(mb.Name) == role {
			return &mb, nil
		}
	}

	return nil, fmt.Errorf("mailbox with role '%s' not found", role)
}

// GetMailboxByID finds a mailbox by ID.
func (c *Client) GetMailboxByID(id string) (*Mailbox, error) {
	mailboxes, err := c.GetMailboxes()
	if err != nil {
		return nil, err
	}

	for _, mb := range mailboxes {
		if mb.ID == id {
			return &mb, nil
		}
	}

	return nil, fmt.Errorf("mailbox with ID '%s' not found", id)
}

// GetMailboxByName finds a mailbox by name.
func (c *Client) GetMailboxByName(name string) (*Mailbox, error) {
	mailboxes, err := c.GetMailboxes()
	if err != nil {
		return nil, err
	}

	name = strings.ToLower(name)

	for _, mb := range mailboxes {
		if strings.ToLower(mb.Name) == name {
			return &mb, nil
		}
	}

	return nil, fmt.Errorf("mailbox with name '%s' not found", name)
}

// CreateMailbox creates a new mailbox.
func (c *Client) CreateMailbox(name string, parentID string) (string, error) {
	session, err := c.GetSession()
	if err != nil {
		return "", err
	}

	mailboxData := map[string]interface{}{
		"name": name,
	}
	if parentID != "" {
		mailboxData["parentId"] = parentID
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Mailbox/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"create": map[string]interface{}{
						"newMailbox": mailboxData,
					},
				},
				"createMailbox",
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

	if e, ok := result.NotCreated["newMailbox"]; ok {
		return "", fmt.Errorf("failed to create mailbox: %s", e.Description)
	}

	if created, ok := result.Created["newMailbox"]; ok {
		return created.ID, nil
	}

	return "", fmt.Errorf("failed to create mailbox: no ID returned")
}

// RenameMailbox renames a mailbox.
func (c *Client) RenameMailbox(mailboxID, newName string) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Mailbox/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"update": map[string]interface{}{
						mailboxID: map[string]interface{}{
							"name": newName,
						},
					},
				},
				"renameMailbox",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return err
	}

	var result struct {
		NotUpdated map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"notUpdated"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return err
	}

	if e, ok := result.NotUpdated[mailboxID]; ok {
		return fmt.Errorf("failed to rename mailbox: %s", e.Description)
	}

	return nil
}

// DeleteMailbox deletes a mailbox.
func (c *Client) DeleteMailbox(mailboxID string) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	request := &Request{
		Using: []string{CoreCapability, MailCapability},
		MethodCalls: [][]interface{}{
			{
				"Mailbox/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"destroy":   []string{mailboxID},
				},
				"deleteMailbox",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return err
	}

	var result struct {
		NotDestroyed map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"notDestroyed"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return err
	}

	if e, ok := result.NotDestroyed[mailboxID]; ok {
		return fmt.Errorf("failed to delete mailbox: %s", e.Description)
	}

	return nil
}
