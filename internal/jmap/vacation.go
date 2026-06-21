package jmap

import (
	"encoding/json"
	"fmt"
)

// GetVacationResponse fetches the current vacation response settings.
func (c *Client) GetVacationResponse() (*VacationResponse, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	request := &Request{
		Using: []string{CoreCapability, VacationResponseCapability},
		MethodCalls: [][]interface{}{
			{
				"VacationResponse/get",
				map[string]interface{}{
					"accountId": session.AccountID,
					"ids":       nil, // Get all (there's only one singleton)
				},
				"vacation",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	var result struct {
		List []VacationResponse `json:"list"`
	}

	if err := json.Unmarshal(resp.MethodResponses[0][1], &result); err != nil {
		return nil, fmt.Errorf("failed to parse vacation response: %w", err)
	}

	if len(result.List) == 0 {
		return nil, fmt.Errorf("no vacation response found")
	}

	return &result.List[0], nil
}

// SetVacationResponse updates the vacation response settings.
func (c *Client) SetVacationResponse(enabled bool, subject, textBody *string, fromDate, toDate *string) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	// Get current vacation response to get its ID
	current, err := c.GetVacationResponse()
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"isEnabled": enabled,
	}
	if subject != nil {
		update["subject"] = *subject
	}
	if textBody != nil {
		update["textBody"] = *textBody
	}
	if fromDate != nil {
		update["fromDate"] = *fromDate
	}
	if toDate != nil {
		update["toDate"] = *toDate
	}

	request := &Request{
		Using: []string{CoreCapability, VacationResponseCapability},
		MethodCalls: [][]interface{}{
			{
				"VacationResponse/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"update": map[string]interface{}{
						current.ID: update,
					},
				},
				"setVacation",
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

	if e, ok := result.NotUpdated[current.ID]; ok {
		return fmt.Errorf("%s: %s", e.Type, e.Description)
	}

	return nil
}
