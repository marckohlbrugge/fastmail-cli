package jmap

import (
	"encoding/json"
	"fmt"
)

// GetMaskedEmails fetches all masked email addresses.
func (c *Client) GetMaskedEmails() ([]MaskedEmail, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	request := &Request{
		Using: []string{CoreCapability, MaskedEmailCapability},
		MethodCalls: [][]interface{}{
			{
				"MaskedEmail/get",
				map[string]interface{}{
					"accountId": session.AccountID,
					"ids":       nil, // Get all
				},
				"maskedEmails",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	var result struct {
		List []MaskedEmail `json:"list"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return nil, fmt.Errorf("failed to parse masked emails: %w", err)
	}

	return result.List, nil
}

// CreateMaskedEmail creates a new masked email address.
func (c *Client) CreateMaskedEmail(forDomain, description string) (*MaskedEmail, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	create := map[string]interface{}{
		"state": "enabled",
	}
	if forDomain != "" {
		create["forDomain"] = forDomain
	}
	if description != "" {
		create["description"] = description
	}

	request := &Request{
		Using: []string{CoreCapability, MaskedEmailCapability},
		MethodCalls: [][]interface{}{
			{
				"MaskedEmail/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"create": map[string]interface{}{
						"new": create,
					},
				},
				"createMasked",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	var result struct {
		Created map[string]MaskedEmail `json:"created"`
		NotCreated map[string]struct {
			Type        string `json:"type"`
			Description string `json:"description"`
		} `json:"notCreated"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return nil, err
	}

	if e, ok := result.NotCreated["new"]; ok {
		return nil, fmt.Errorf("%s: %s", e.Type, e.Description)
	}

	if created, ok := result.Created["new"]; ok {
		return &created, nil
	}

	return nil, fmt.Errorf("failed to create masked email: no result returned")
}

// SetMaskedEmailState updates a masked email's state (enabled, disabled).
func (c *Client) SetMaskedEmailState(id, state string) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	request := &Request{
		Using: []string{CoreCapability, MaskedEmailCapability},
		MethodCalls: [][]interface{}{
			{
				"MaskedEmail/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"update": map[string]interface{}{
						id: map[string]interface{}{
							"state": state,
						},
					},
				},
				"updateMasked",
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

	if e, ok := result.NotUpdated[id]; ok {
		return fmt.Errorf("%s: %s", e.Type, e.Description)
	}

	return nil
}

// DeleteMaskedEmail deletes a masked email (sets state to deleted).
func (c *Client) DeleteMaskedEmail(id string) error {
	return c.SetMaskedEmailState(id, "deleted")
}
