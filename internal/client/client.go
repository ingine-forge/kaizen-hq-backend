package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ClientOption allows configuring the torn client with functional options
type ClientOption func(*client)

// WithBaseURL configures the base URL for the API
func WithBaseURL(baseURL string) ClientOption {
	return func(c *client) {
		c.baseURL = baseURL
	}
}

// WithVersion configures the API version
func WithVersion(version string) ClientOption {
	return func(c *client) {
		c.version = version
	}
}

// WithTimeout configures the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *client) {
		c.client.Timeout = timeout
	}
}

// WithHTTPClient allows using a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *client) {
		c.client = httpClient
	}
}

// Client defines the interface for interacting with Torn API
type Client interface {
	FetchGymEnergy(ctx context.Context, apiKey, stat string) (StatMap, error)
	FetchTornUser(ctx context.Context, apiKey, tornID string) (*User, error)
	FetchDiscordID(ctx context.Context, apiKey string, tornID int) (string, error)
	FetchKeyDetails(ctx context.Context, apiKey string) (int, error)

	// SwitchVersion changes the API version at runtime
	SwitchVersion(version string)
	// SwitchBaseURL changes the base URL at runtime
	SwitchBaseURL(baseURL string)
}

// TornAPIProvider represents different Torn API providers
type TornAPIProvider string

const (
	// Predefined API providers
	TornMainAPI  TornAPIProvider = "https://api.torn.com"
	TornStatsAPI TornAPIProvider = "https://tornstats.com/api/v2"
	YataAPI      TornAPIProvider = "https://yata.yt/api"

	// Default API version
	DefaultVersion = ""
)

type client struct {
	baseURL string
	version string
	client  *http.Client
}

/*
NewClient creates a new client based on options provided.
It defaults to use the torn api v1 but can be configured as per need.
*/
func NewClient(opts ...ClientOption) Client {
	client := &client{
		baseURL: string(TornMainAPI),
		version: DefaultVersion,
		client:  &http.Client{Timeout: 10 * time.Second},
	}

	// Apply all options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// NewClientWithProvider creates a new client with a predefined provider
func NewClientWithProvider(provider TornAPIProvider, opts ...ClientOption) Client {
	// Start with the provider option
	providerOpt := WithBaseURL(string(provider))

	// Prepend it to the other options
	allOpts := append([]ClientOption{providerOpt}, opts...)

	return NewClient(allOpts...)
}

// SwitchVersion changes the API version at runtime
func (t *client) SwitchVersion(version string) {
	t.version = version
}

// SwitchBaseURL changes the base URL at runtime
func (t *client) SwitchBaseURL(baseURL string) {
	t.baseURL = baseURL
}

// buildURL constructs the complete API URL (API key is passed dynamically)
func (t *client) buildURL(apiKey, endpoint, selections string, params map[string]string) (string, error) {
	base, err := url.Parse(t.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Add version if not empty
	path := "/"
	if t.version != "" {
		path += t.version + "/"
	}
	path += endpoint

	base.Path += path

	// Add query parameters
	q := base.Query()
	q.Set("key", apiKey) // API key passed dynamically
	if selections != "" {
		q.Set("selections", selections)
	}

	// Add additional parameters
	for k, v := range params {
		q.Set(k, v)
	}

	base.RawQuery = q.Encode()
	return base.String(), nil
}

// makeRequest handles the HTTP request and response
func (t *client) makeRequest(ctx context.Context, url string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Check for API error response
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("API error: status=%d, body=%s", res.StatusCode, string(body))
	}

	return json.NewDecoder(res.Body).Decode(result)
}

func (t *client) FetchGymEnergy(ctx context.Context, apiKey, stat string) (StatMap, error) {
	params := map[string]string{"stat": stat}
	url, err := t.buildURL(apiKey, "faction", "contributors", params)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Contributors StatMap   `json:"contributors"`
		Error        *APIError `json:"error"`
	}

	if err := t.makeRequest(ctx, url, &parsed); err != nil {
		return nil, err
	}

	if parsed.Error != nil {
		return nil, parsed.Error
	}

	return parsed.Contributors, nil
}

func (t *client) FetchTornUser(ctx context.Context, apiKey, tornID string) (*User, error) {
	url, err := t.buildURL(apiKey, fmt.Sprintf("user/%s", tornID), "profile", nil)
	if err != nil {
		return nil, err
	}

	var user struct {
		User  *User     `json:"-"`
		Error *APIError `json:"error"`
	}

	// First read the body
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// First check for error
	if err := json.Unmarshal(body, &user); err == nil && user.Error != nil {
		return nil, user.Error
	}

	// Then parse the full user
	var fullUser User
	if err := json.Unmarshal(body, &fullUser); err != nil {
		return nil, err
	}

	return &fullUser, nil
}

func (t *client) FetchDiscordID(ctx context.Context, apiKey string, tornID int) (string, error) {
	url, err := t.buildURL(apiKey, fmt.Sprintf("user/%d", tornID), "discord", nil)
	if err != nil {
		return "", err
	}

	var parsed struct {
		Discord Discord   `json:"discord"`
		Error   *APIError `json:"error"`
	}

	if err := t.makeRequest(ctx, url, &parsed); err != nil {
		return "", err
	}

	if parsed.Error != nil {
		return "", parsed.Error
	}

	return parsed.Discord.DiscordID, nil
}

func (t *client) FetchKeyDetails(ctx context.Context, apiKey string) (int, error) {
	url, err := t.buildURL(apiKey, "key", "info", nil)
	if err != nil {
		return 0, err
	}

	var key Key

	if err := t.makeRequest(ctx, url, &key); err != nil {
		return 0, err
	}

	return key.AccessLevel, nil
}

// APIError represents an error from the Torn API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("Torn API error %d: %s", e.Code, e.Message)
}
