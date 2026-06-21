package jmap

import (
	"encoding/json"
	"fmt"
)

// ContactsCapability is the JMAP capability for contacts.
const ContactsCapability = "urn:ietf:params:jmap:contacts"

// ContactCard represents a JSContact ContactCard (RFC 9553).
type ContactCard struct {
	ID      string                   `json:"id,omitempty"`
	Name    *ContactName             `json:"name,omitempty"`
	Emails  map[string]ContactEmail  `json:"emails,omitempty"`
	Phones  map[string]ContactPhone  `json:"phones,omitempty"`
	Orgs    map[string]ContactOrg    `json:"organizations,omitempty"`
	Notes   map[string]ContactNote   `json:"notes,omitempty"`
}

// ContactName holds name components.
type ContactName struct {
	Full       string       `json:"full,omitempty"`
	Components []NameComponent `json:"components,omitempty"`
}

// NameComponent is a single name part (given, surname, etc).
type NameComponent struct {
	Kind  string `json:"kind"`  // "given", "surname", etc.
	Value string `json:"value"`
}

// ContactEmail represents a contact email address.
type ContactEmail struct {
	Address string `json:"address"`
	Label   string `json:"label,omitempty"`
	Pref    int    `json:"pref,omitempty"`
}

// ContactPhone represents a contact phone number.
type ContactPhone struct {
	Number string `json:"number"`
	Label  string `json:"label,omitempty"`
	Pref   int    `json:"pref,omitempty"`
}

// ContactOrg represents an organization.
type ContactOrg struct {
	Name string `json:"name"`
}

// ContactNote represents a note on a contact.
type ContactNote struct {
	Note string `json:"note"`
}

// GetContacts fetches contacts up to the given limit.
func (c *Client) GetContacts(limit int) ([]ContactCard, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 100
	}

	request := &Request{
		Using: []string{CoreCapability, ContactsCapability},
		MethodCalls: [][]interface{}{
			{
				"ContactCard/query",
				map[string]interface{}{
					"accountId": session.AccountID,
					"limit":     limit,
				},
				"query",
			},
			{
				"ContactCard/get",
				map[string]interface{}{
					"accountId": session.AccountID,
					"#ids":      map[string]interface{}{"resultOf": "query", "name": "ContactCard/query", "path": "/ids"},
				},
				"contacts",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	return c.parseContactsFromResponse(resp, 1)
}

// GetContact fetches a single contact by ID.
func (c *Client) GetContact(id string) (*ContactCard, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	request := &Request{
		Using: []string{CoreCapability, ContactsCapability},
		MethodCalls: [][]interface{}{
			{
				"ContactCard/get",
				map[string]interface{}{
					"accountId": session.AccountID,
					"ids":       []string{id},
				},
				"contact",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	contacts, err := c.parseContactsFromResponse(resp, 0)
	if err != nil {
		return nil, err
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("contact with ID '%s' not found", id)
	}

	return &contacts[0], nil
}

// SearchContacts searches contacts by text query.
func (c *Client) SearchContacts(query string, limit int) ([]ContactCard, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 30
	}

	request := &Request{
		Using: []string{CoreCapability, ContactsCapability},
		MethodCalls: [][]interface{}{
			{
				"ContactCard/query",
				map[string]interface{}{
					"accountId": session.AccountID,
					"filter": map[string]interface{}{
						"text": query,
					},
					"limit": limit,
				},
				"query",
			},
			{
				"ContactCard/get",
				map[string]interface{}{
					"accountId": session.AccountID,
					"#ids":      map[string]interface{}{"resultOf": "query", "name": "ContactCard/query", "path": "/ids"},
				},
				"contacts",
			},
		},
	}

	resp, err := c.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	return c.parseContactsFromResponse(resp, 1)
}

// CreateContact creates a new contact and returns its ID.
func (c *Client) CreateContact(card *ContactCard) (string, error) {
	session, err := c.GetSession()
	if err != nil {
		return "", err
	}

	// Build the contact data (omit ID, server assigns it)
	contactData := map[string]interface{}{}
	if card.Name != nil {
		contactData["name"] = card.Name
	}
	if card.Emails != nil {
		contactData["emails"] = card.Emails
	}
	if card.Phones != nil {
		contactData["phones"] = card.Phones
	}
	if card.Orgs != nil {
		contactData["organizations"] = card.Orgs
	}
	if card.Notes != nil {
		contactData["notes"] = card.Notes
	}

	request := &Request{
		Using: []string{CoreCapability, ContactsCapability},
		MethodCalls: [][]interface{}{
			{
				"ContactCard/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"create": map[string]interface{}{
						"newContact": contactData,
					},
				},
				"createContact",
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

	if e, ok := result.NotCreated["newContact"]; ok {
		return "", fmt.Errorf("failed to create contact: %s", e.Description)
	}

	if created, ok := result.Created["newContact"]; ok {
		return created.ID, nil
	}

	return "", fmt.Errorf("failed to create contact: no ID returned")
}

// DeleteContact deletes a contact by ID.
func (c *Client) DeleteContact(id string) error {
	session, err := c.GetSession()
	if err != nil {
		return err
	}

	request := &Request{
		Using: []string{CoreCapability, ContactsCapability},
		MethodCalls: [][]interface{}{
			{
				"ContactCard/set",
				map[string]interface{}{
					"accountId": session.AccountID,
					"destroy":   []string{id},
				},
				"deleteContact",
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

	if e, ok := result.NotDestroyed[id]; ok {
		return fmt.Errorf("failed to delete contact: %s", e.Description)
	}

	return nil
}

// parseContactsFromResponse extracts contacts from a JMAP response.
func (c *Client) parseContactsFromResponse(resp *Response, index int) ([]ContactCard, error) {
	if len(resp.MethodResponses) <= index {
		return nil, fmt.Errorf("invalid response: missing method response at index %d", index)
	}

	var result struct {
		List     []ContactCard `json:"list"`
		NotFound []string      `json:"notFound"`
	}

	if err := json.Unmarshal(resp.MethodResponses[index][1], &result); err != nil {
		return nil, fmt.Errorf("failed to parse contacts: %w", err)
	}

	return result.List, nil
}
