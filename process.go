package goao

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ar-aostore/goao/dataitem"
	"github.com/ar-aostore/goao/encrypt"
	"github.com/ar-aostore/goao/schema"
	"github.com/ar-aostore/goao/signer"
)

// SendMessage sends a message to an AO process
// Supports optional encryption via SendMessageOptions
func (c *Client) SendMessage(process string, data []byte, tags []schema.Tag, opts *schema.SendMessageOptions) (string, error) {
	// Check if encryption is requested
	if opts != nil && opts.EncryptWithRSA != nil {
		// Encrypt the payload
		encryptedData, encryptedKey, nonce, err := encrypt.EncryptPayload(data, opts.EncryptWithRSA)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt payload: %w", err)
		}

		// Replace data with encrypted payload
		data = encryptedData

		// Add encryption metadata tags
		tags = append(tags,
			schema.Tag{Name: schema.TagEncryption, Value: schema.EncryptionAESGCM},
			schema.Tag{Name: schema.TagEncryptedKey, Value: encrypt.Base64URLEncode(encryptedKey)},
			schema.Tag{Name: schema.TagNonce, Value: encrypt.Base64URLEncode(nonce)},
		)
	}

	// Build DataItem
	di := dataitem.NewDataItem(process, data, tags, nil)

	// Set anchor if provided
	if opts != nil && len(opts.Anchor) == 32 {
		di.Anchor = make([]byte, 32)
		copy(di.Anchor, opts.Anchor)
	}

	// Sign the DataItem
	if c.Signer == nil {
		return "", fmt.Errorf("signer is not configured")
	}
	if err := di.Sign(c.Signer); err != nil {
		return "", fmt.Errorf("failed to sign DataItem: %w", err)
	}

	// Serialize to binary (ANS-104)
	binary, err := di.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize DataItem: %w", err)
	}

	// POST to Messenger Unit
	resp, err := c.doBinaryRequest("POST", c.MUURL+"/submit", binary, "application/octet-stream")
	if err != nil {
		return "", fmt.Errorf("MU request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("MU returned %d: %s", resp.StatusCode, string(body))
	}

	// Return the DataItem ID (message ID)
	return di.ID, nil
}

// SpawnProcess spawns a new AO process
// module: Transaction ID of the WASM module
// data: Initial state (optional)
// tags: Metadata tags (Module, Scheduler, etc.)
func (c *Client) SpawnProcess(module string, data []byte, tags []schema.Tag) (string, error) {
	// Add Module tag if not present
	hasModule := false
	hasScheduler := false
	for _, tag := range tags {
		if tag.Name == "Module" {
			hasModule = true
		}
		if tag.Name == "Scheduler" {
			hasScheduler = true
		}
	}

	if !hasModule {
		tags = append(tags, schema.Tag{Name: "Module", Value: module})
	}

	if !hasScheduler {
		tags = append(tags, schema.Tag{Name: "Scheduler", Value: schema.DefaultScheduler})
	}

	// Add AO protocol tags
	tags = append(tags,
		schema.Tag{Name: "Data-Protocol", Value: schema.DataProtocol},
		schema.Tag{Name: "Variant", Value: schema.Variant},
		schema.Tag{Name: "Type", Value: schema.TypeProcess},
	)

	// Build DataItem (Target is empty for spawn)
	di := dataitem.NewDataItem("", data, tags, nil)

	// Sign the DataItem
	if c.Signer == nil {
		return "", fmt.Errorf("signer is not configured")
	}
	if err := di.Sign(c.Signer); err != nil {
		return "", fmt.Errorf("failed to sign DataItem: %w", err)
	}

	// Serialize to binary
	binary, err := di.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize DataItem: %w", err)
	}

	// POST to Scheduler Unit
	resp, err := c.doBinaryRequest("POST", c.SUURL+"/submit", binary, "application/octet-stream")
	if err != nil {
		return "", fmt.Errorf("SU request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("SU returned %d: %s", resp.StatusCode, string(body))
	}

	// Return the process ID (DataItem ID)
	return di.ID, nil
}

// DryRun evaluates a message without persisting state (LEGACY - kept for backward compatibility)
// DEPRECATED: Use GetCompute() for simple state reads
func (c *Client) DryRun(processID string, data []byte, tags []schema.Tag, anchor []byte) (*schema.ResponseCu, error) {
	// Build DataItem
	di := dataitem.NewDataItem(processID, data, tags, anchor)

	// Sign the DataItem
	if c.Signer == nil {
		return nil, fmt.Errorf("signer is not configured")
	}
	if err := di.Sign(c.Signer); err != nil {
		return nil, fmt.Errorf("failed to sign DataItem: %w", err)
	}

	// Serialize to binary
	binary, err := di.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize DataItem: %w", err)
	}

	// POST to CU dry-run endpoint
	dryRunURL := fmt.Sprintf("%s/dry-run?process=%s", c.CUURL, processID)
	resp, err := c.doBinaryRequest("POST", dryRunURL, binary, "application/octet-stream")
	if err != nil {
		return nil, fmt.Errorf("CU dry-run request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CU returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var result schema.ResponseCu
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetCompute fetches a field from an AO process via HTTP Compute Gateway (NEW - Recommended)
// URL pattern: <gateway>/<process-id>~process@1.0/compute/<field-path>
// Example: https://push.forward.computer/abc123~process@1.0/compute/counter
//
// TWO USAGE PATTERNS:
// 1. Component-based: processID="abc123", fieldPath="counter"
// 2. Full URL: processID="https://gateway.com/abc123~process@1.0/compute/counter", fieldPath=""
func (c *Client) GetCompute(processID, fieldPath string) ([]byte, error) {
	var computeURL string

	// Check if user provided a full URL
	if strings.HasPrefix(processID, "http://") || strings.HasPrefix(processID, "https://") {
		computeURL = processID
	} else {
		// Construct URL from components
		computeURL = fmt.Sprintf("%s/%s~process@1.0/compute/%s",
			c.ComputeGateway,
			processID,
			url.PathEscape(fieldPath),
		)
	}

	// Create HTTP GET request
	req, err := http.NewRequest("GET", computeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "goao/"+schema.Version)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("compute request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("compute gateway returned %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// GetComputeString fetches a field and returns it as a string
func (c *Client) GetComputeString(processID, fieldPath string) (string, error) {
	data, err := c.GetCompute(processID, fieldPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetComputeJSON fetches a field and unmarshals the JSON response into a struct
func (c *Client) GetComputeJSON(processID, fieldPath string, out interface{}) error {
	data, err := c.GetCompute(processID, fieldPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// Assign associates an L1 transaction with an AO process
func (c *Client) Assign(processID, messageID string, exclude []string, baseLayer bool) (string, error) {
	requestBody := map[string]interface{}{
		"process":   processID,
		"message":   messageID,
		"exclude":   exclude,
		"baseLayer": baseLayer,
	}

	var response map[string]interface{}
	err := c.doJSONRequest("POST", c.SUURL+"/assign", requestBody, &response)
	if err != nil {
		return "", err
	}

	if id, ok := response["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("no process ID in response")
}

// Monitor initiates cron message monitoring for a process
func (c *Client) Monitor(processID string) (string, error) {
	if c.Signer == nil {
		return "", fmt.Errorf("signer is not configured")
	}

	// Build DataItem for monitor
	di := dataitem.NewDataItem(processID, nil, []schema.Tag{
		{Name: "Action", Value: "Monitor"},
	}, nil)

	if err := di.Sign(c.Signer); err != nil {
		return "", fmt.Errorf("failed to sign DataItem: %w", err)
	}

	binary, err := di.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize DataItem: %w", err)
	}

	resp, err := c.doBinaryRequest("POST", c.SUURL+"/monitor", binary, "application/octet-stream")
	if err != nil {
		return "", fmt.Errorf("SU monitor request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("SU returned %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if id, ok := response["subscriptionId"].(string); ok {
		return id, nil
	}

	return di.ID, nil
}

// DecryptResponse decrypts an encrypted response from an AO process
func (c *Client) DecryptResponse(encryptedData, encryptedKey, nonce string, rsaPrivKey interface{}) ([]byte, error) {
	// Decode base64url strings to bytes
	encData, err := encrypt.Base64URLDecode(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	encKey, err := encrypt.Base64URLDecode(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	nce, err := encrypt.Base64URLDecode(nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	// Get the signer's private key (must be RSA)
	rsaSigner, ok := rsaPrivKey.(*signer.RSASigner)
	if !ok {
		return nil, fmt.Errorf("signer must be RSA for decryption")
	}

	// Decrypt the payload
	return encrypt.DecryptPayload(encData, encKey, nce, rsaSigner.GetPrivateKey())
}