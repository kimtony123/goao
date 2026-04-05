package compute

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DefaultComputeGateway is the default AO Compute Gateway
const DefaultComputeGateway = "https://push.forward.computer"

// Client is the HTTP client for interacting with AO Compute Gateways
type Client struct {
	gatewayURL string
	httpClient *http.Client
}

// NewClient creates a new Compute Client
// If gatewayURL is empty, uses DefaultComputeGateway
func NewClient(gatewayURL string) *Client {
	if gatewayURL == "" {
		gatewayURL = DefaultComputeGateway
	}
	// Ensure gatewayURL doesn't have trailing slash
	gatewayURL = strings.TrimRight(gatewayURL, "/")

	return &Client{
		gatewayURL: gatewayURL,
		httpClient: &http.Client{},
	}
}

// NewClientWithHTTP creates a new Compute Client with a custom HTTP client
func NewClientWithHTTP(gatewayURL string, httpClient *http.Client) *Client {
	if gatewayURL == "" {
		gatewayURL = DefaultComputeGateway
	}
	gatewayURL = strings.TrimRight(gatewayURL, "/")

	return &Client{
		gatewayURL: gatewayURL,
		httpClient: httpClient,
	}
}

// GetCompute fetches a field from an AO process via the HTTP Compute Gateway
//
// TWO USAGE PATTERNS SUPPORTED:
//
// 1. Component-based (library constructs URL):
//    processID: "Ma32vsWVPtedS3Rav_vUR9JvUfffQzPlhBpwWBSf2YU"
//    fieldPath: "counter"
//    → URL: https://push.forward.computer/Ma32vsWVPtedS3Rav_vUR9JvUfffQzPlhBpwWBSf2YU~process@1.0/compute/counter
//
// 2. Full URL (user provides complete URL):
//    processID: "https://push.forward.computer/Ma32vsWVPtedS3Rav_vUR9JvUfffQzPlhBpwWBSf2YU~process@1.0/compute/counter"
//    fieldPath: "" (ignored)
//    → Uses the provided URL directly
//
// Parameters:
//   - processID: Either the process ID OR a full compute URL
//   - fieldPath: The field to fetch (ignored if processID is a full URL)
//
// Returns:
//   - []byte: The raw response data
//   - error: Any error during the HTTP request
func (c *Client) GetCompute(processID, fieldPath string) ([]byte, error) {
	var computeURL string

	// Check if user provided a full URL (starts with http:// or https://)
	if strings.HasPrefix(processID, "http://") || strings.HasPrefix(processID, "https://") {
		// Use the full URL directly
		computeURL = processID
	} else {
		// Construct URL from components
		// Format: {gateway}/{processID}~process@1.0/compute/{fieldPath}
		computeURL = fmt.Sprintf("%s/%s~process@1.0/compute/%s",
			c.gatewayURL,
			processID,
			url.PathEscape(fieldPath),
		)
	}

	// Create HTTP GET request
	req, err := http.NewRequest("GET", computeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "goao-compute/0.2.0")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("compute request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("compute gateway returned %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// GetComputeWithURL fetches data using a complete URL provided by the user
// This is a convenience method for when you already have the full URL
//
// Example:
//   url := "https://push.forward.computer/Ma32vsWVPtedS3Rav_vUR9JvUfffQzPlhBpwWBSf2YU~process@1.0/compute/counter"
//   data, err := client.GetComputeWithURL(url)
func (c *Client) GetComputeWithURL(fullURL string) ([]byte, error) {
	return c.GetCompute(fullURL, "")
}

// GetComputeJSON fetches a field and unmarshals the JSON response into a struct
// Supports both component-based and full URL patterns (same as GetCompute)
func (c *Client) GetComputeJSON(processID, fieldPath string, out interface{}) error {
	data, err := c.GetCompute(processID, fieldPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, out)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// GetComputeString fetches a field and returns it as a string
// Supports both component-based and full URL patterns (same as GetCompute)
func (c *Client) GetComputeString(processID, fieldPath string) (string, error) {
	data, err := c.GetCompute(processID, fieldPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetGatewayURL returns the configured gateway URL
func (c *Client) GetGatewayURL() string {
	return c.gatewayURL
}

// SetGatewayURL updates the gateway URL (for switching nodes at runtime)
func (c *Client) SetGatewayURL(gatewayURL string) {
	if gatewayURL == "" {
		gatewayURL = DefaultComputeGateway
	}
	c.gatewayURL = strings.TrimRight(gatewayURL, "/")
}
