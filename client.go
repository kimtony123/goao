package goao

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kimtony123/goao/schema"
	"github.com/kimtony123/goao/signer"
)

// Default URLs for AO Testnet
const (
	DefaultMUURL          = "https://mu.ao-testnet.xyz"
	DefaultCUURL          = "https://cu.ao-testnet.xyz"
	DefaultSUURL          = "https://su.ao-testnet.xyz"
	DefaultGatewayURL     = "https://arweave.net"
	DefaultGraphQLURL     = "https://arweave.net/graphql"
	DefaultComputeGateway = "https://push.forward.computer"
)

// Client is the main HTTP client for interacting with AO units
type Client struct {
	MUURL               string
	CUURL               string
	SUURL               string
	GatewayURL          string
	GraphQLURL          string
	ComputeGateway      string
	GraphQLMaxRetries   int
	GraphQLRetryBackoff time.Duration
	Signer              signer.Signer
	httpClient          *http.Client
}

// Option is a functional option for configuring the Client
type Option func(*Client)

// NewClient creates a new AO Client with optional configuration
// Empty strings will use default URLs
func NewClient(opts ...Option) *Client {
	c := &Client{
		MUURL:               DefaultMUURL,
		CUURL:               DefaultCUURL,
		SUURL:               DefaultSUURL,
		GatewayURL:          DefaultGatewayURL,
		GraphQLURL:          DefaultGraphQLURL,
		ComputeGateway:      DefaultComputeGateway,
		GraphQLMaxRetries:   0,
		GraphQLRetryBackoff: 300 * time.Millisecond,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithMU sets the Messenger Unit URL
func WithMU(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.MUURL = url
		}
	}
}

// WithCU sets the Compute Unit URL
func WithCU(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.CUURL = url
		}
	}
}

// WithSU sets the Scheduler Unit URL
func WithSU(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.SUURL = url
		}
	}
}

// WithGatewayURL sets the Arweave Gateway URL
func WithGatewayURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.GatewayURL = url
		}
	}
}

// WithComputeGateway sets the HTTP Compute Gateway URL
func WithComputeGateway(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.ComputeGateway = url
		}
	}
}

// WithSigner sets the signer for signing messages
func WithSigner(signer signer.Signer) Option {
	return func(c *Client) {
		c.Signer = signer
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// GetMUURL returns the configured MU URL
func (c *Client) GetMUURL() string {
	return c.MUURL
}

// GetCUURL returns the configured CU URL
func (c *Client) GetCUURL() string {
	return c.CUURL
}

// GetSUURL returns the configured SU URL
func (c *Client) GetSUURL() string {
	return c.SUURL
}

// GetComputeGateway returns the configured Compute Gateway URL
func (c *Client) GetComputeGateway() string {
	return c.ComputeGateway
}

// doRequest performs an HTTP request with error handling
func (c *Client) doRequest(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", "goao/"+schema.Version)
	req.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// doJSONRequest performs an HTTP request and parses JSON response
func (c *Client) doJSONRequest(method, url string, requestBody, responseBody interface{}) error {
	var bodyReader io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	resp, err := c.doRequest(method, url, bodyReader, map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if responseBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// doBinaryRequest performs an HTTP request with binary body
func (c *Client) doBinaryRequest(method, url string, body []byte, contentType string) (*http.Response, error) {
	headers := map[string]string{
		"Content-Type": contentType,
	}
	return c.doRequest(method, url, bytes.NewReader(body), headers)
}