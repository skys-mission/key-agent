// Package keysdk provides the official Go SDK for key-agent.
package keysdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is the SDK client for key-agent.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Config holds client configuration.
type Config struct {
	// BaseURL is the key-agent server URL (default: http://127.0.0.1:8080)
	BaseURL string
	// Token is the authentication token
	Token string
	// HTTPClient is an optional custom HTTP client
	HTTPClient *http.Client
}

// NewClient creates a new SDK client with the given configuration.
func NewClient(cfg *Config) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		token:      cfg.Token,
		httpClient: httpClient,
	}
}

// doRequest performs an HTTP request.
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// doRequestWithResponse performs an HTTP request and parses the response.
func (c *Client) doRequestWithResponse(method, path string, body interface{}, result interface{}) error {
	resp, err := c.doRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		if json.Unmarshal(data, &errResp) == nil {
			return &Error{Code: errResp.Error.Code, Message: errResp.Error.Message}
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	if result != nil {
		return json.Unmarshal(data, result)
	}

	return nil
}

// Error represents an API error.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// === KV Operations ===

// KVEntry represents a KV entry.
type KVEntry struct {
	Key       string                 `json:"key"`
	Value     string                 `json:"value"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt string                 `json:"created_at,omitempty"`
	UpdatedAt string                 `json:"updated_at,omitempty"`
	Version   int64                  `json:"version,omitempty"`
}

// GetKV retrieves a KV entry by key.
func (c *Client) GetKV(key string) (*KVEntry, error) {
	var result KVEntry
	err := c.doRequestWithResponse(http.MethodGet, "/api/v1/kv/"+url.PathEscape(key), nil, &result)
	return &result, err
}

// SetKVOptions provides options for setting a KV entry.
type SetKVOptions struct {
	Value    string                 `json:"value"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SetKV sets a KV entry.
func (c *Client) SetKV(key string, opts *SetKVOptions) (*KVEntry, error) {
	var result KVEntry
	err := c.doRequestWithResponse(http.MethodPut, "/api/v1/kv/"+url.PathEscape(key), opts, &result)
	return &result, err
}

// DeleteKV deletes a KV entry.
func (c *Client) DeleteKV(key string) error {
	return c.doRequestWithResponse(http.MethodDelete, "/api/v1/kv/"+url.PathEscape(key), nil, nil)
}

// ListKV lists KV keys with optional prefix filter.
func (c *Client) ListKV(prefix string) ([]string, error) {
	path := "/api/v1/kv"
	if prefix != "" {
		path += "?prefix=" + url.QueryEscape(prefix)
	}

	var result struct {
		Keys []string `json:"keys"`
	}
	err := c.doRequestWithResponse(http.MethodGet, path, nil, &result)
	return result.Keys, err
}

// === Secret Operations ===

// SecretType is the type of a secret.
type SecretType string

const (
	SecretTypePassword    SecretType = "password"
	SecretTypeAPIKey      SecretType = "api_key"
	SecretTypeCertificate SecretType = "certificate"
	SecretTypePrivateKey  SecretType = "private_key"
	SecretTypeToken       SecretType = "token"
	SecretTypeOther       SecretType = "other"
)

// SecretEntry represents a secret entry.
type SecretEntry struct {
	KVEntry
	Type SecretType `json:"type"`
}

// GetSecret retrieves a secret entry by key.
func (c *Client) GetSecret(key string) (*SecretEntry, error) {
	var result SecretEntry
	err := c.doRequestWithResponse(http.MethodGet, "/api/v1/secrets/"+url.PathEscape(key), nil, &result)
	return &result, err
}

// SetSecretOptions provides options for setting a secret.
type SetSecretOptions struct {
	Value    string                 `json:"value"`
	Type     SecretType             `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SetSecret sets a secret entry.
func (c *Client) SetSecret(key string, opts *SetSecretOptions) (*SecretEntry, error) {
	var result SecretEntry
	err := c.doRequestWithResponse(http.MethodPut, "/api/v1/secrets/"+url.PathEscape(key), opts, &result)
	return &result, err
}

// DeleteSecret deletes a secret entry.
func (c *Client) DeleteSecret(key string) error {
	return c.doRequestWithResponse(http.MethodDelete, "/api/v1/secrets/"+url.PathEscape(key), nil, nil)
}

// ListSecret lists secret keys with optional prefix filter.
func (c *Client) ListSecret(prefix string) ([]string, error) {
	path := "/api/v1/secrets"
	if prefix != "" {
		path += "?prefix=" + url.QueryEscape(prefix)
	}

	var result struct {
		Keys []string `json:"keys"`
	}
	err := c.doRequestWithResponse(http.MethodGet, path, nil, &result)
	return result.Keys, err
}

// === Token Operations ===

// CreateTokenOptions provides options for creating a token.
type CreateTokenOptions struct {
	Name      string `json:"name"`
	Type      string `json:"type,omitempty"`
	ExpiresIn string `json:"expires_in,omitempty"`
}

// TokenInfo represents token information.
type TokenInfo struct {
	Token     string `json:"token"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at,omitempty"`
}

// CreateToken creates a new token.
func (c *Client) CreateToken(opts *CreateTokenOptions) (*TokenInfo, error) {
	var result TokenInfo
	err := c.doRequestWithResponse(http.MethodPost, "/api/v1/token", opts, &result)
	return &result, err
}

// === Health ===

// HealthStatus represents the health check response.
type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Health checks the server health.
func (c *Client) Health() (*HealthStatus, error) {
	var result HealthStatus
	err := c.doRequestWithResponse(http.MethodGet, "/health", nil, &result)
	return &result, err
}
