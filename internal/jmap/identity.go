package jmap

import (
	"encoding/json"
	"fmt"
)

// GetIdentities fetches all sender identities.
func (c *Client) GetIdentities() ([]Identity, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	request := &Request{
		Using: []string{CoreCapability, SubmissionCapability},
		MethodCalls: [][]interface{}{
			{
				"Identity/get",
				map[string]interface{}{
					"accountId": session.AccountID,
				},
				"identities",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	var result struct {
		List []Identity `json:"list"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return nil, fmt.Errorf("failed to parse identities: %w", err)
	}

	return result.List, nil
}

// GetDefaultIdentity returns the primary sender identity.
func (c *Client) GetDefaultIdentity() (*Identity, error) {
	identities, err := c.GetIdentities()
	if err != nil {
		return nil, err
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no identities found")
	}

	// Prefer non-deletable identity (usually the primary)
	for _, id := range identities {
		if !id.MayDelete {
			return &id, nil
		}
	}

	return &identities[0], nil
}
