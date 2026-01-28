package jmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	DefaultBaseURL = "https://api.fastmail.com"
	SessionPath    = "/jmap/session"
)

// Client is a JMAP client for Fastmail.
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
	session    *Session
}

// Session contains JMAP session information.
type Session struct {
	APIURL      string                 `json:"apiUrl"`
	DownloadURL string                 `json:"downloadUrl"`
	UploadURL   string                 `json:"uploadUrl"`
	AccountID   string                 // First account ID
	Accounts    map[string]interface{} `json:"accounts"`
}

// Request is a JMAP request.
type Request struct {
	Using       []string        `json:"using"`
	MethodCalls [][]interface{} `json:"methodCalls"`
}

// Response is a JMAP response.
type Response struct {
	MethodResponses [][]json.RawMessage `json:"methodResponses"`
	SessionState    string              `json:"sessionState"`
}

// NewClient creates a new JMAP client.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{},
	}
}

// SetBaseURL sets a custom base URL (for testing).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = strings.TrimSuffix(url, "/")
}

// SetHTTPClient sets a custom HTTP client (for testing).
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// GetSession returns the JMAP session, fetching it if necessary.
func (c *Client) GetSession() (*Session, error) {
	if c.session != nil {
		return c.session, nil
	}

	req, err := http.NewRequest("GET", c.baseURL+SessionPath, nil)
	if err != nil {
		return nil, err
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Fastmail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get session: %s - %s", resp.Status, string(body))
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode session: %w", err)
	}

	// Extract first account ID
	for id := range session.Accounts {
		session.AccountID = id
		break
	}

	c.session = &session
	return c.session, nil
}

// MakeRequest sends a JMAP request and returns the response.
func (c *Client) MakeRequest(request *Request) (*Response, error) {
	session, err := c.GetSession()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", session.APIURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	c.setAuthHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JMAP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("JMAP request failed: %s - %s", resp.Status, string(respBody))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// setAuthHeaders sets the authorization headers on a request.
func (c *Client) setAuthHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
}

// AccountID returns the current account ID.
func (c *Client) AccountID() (string, error) {
	session, err := c.GetSession()
	if err != nil {
		return "", err
	}
	return session.AccountID, nil
}

// Standard JMAP capabilities
var (
	CoreCapability       = "urn:ietf:params:jmap:core"
	MailCapability       = "urn:ietf:params:jmap:mail"
	SubmissionCapability = "urn:ietf:params:jmap:submission"
)
