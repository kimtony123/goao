package goao

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ar-aostore/goao/schema"
)

// GetResult fetches the result of a message evaluation from the Compute Unit (CU)
// This is used to retrieve the output, messages, spawns, and errors from a processed message
//
// Parameters:
//   - messageID: The ID of the message (DataItem ID) to get results for
//   - processID: The AO process ID that handled the message
//
// Returns:
//   - *schema.ResponseCu: The result containing Messages, Spawns, Output, Error, GasUsed
//   - error: Any error during the request
//
// Example:
//   result, err := client.GetResult("message-id-123", "process-id-456")
//   if err != nil { panic(err) }
//   fmt.Println("Output:", result.Output)
func (c *Client) GetResult(messageID, processID string) (*schema.ResponseCu, error) {
	// Validate inputs
	if messageID == "" {
		return nil, fmt.Errorf("messageID is required")
	}
	if processID == "" {
		return nil, fmt.Errorf("processID is required")
	}

	// Build CU result endpoint URL
	// Format: {CU_URL}/result/{messageID}?process={processID}
	resultURL := fmt.Sprintf("%s/result/%s?process=%s",
		c.CUURL,
		messageID,
		url.QueryEscape(processID),
	)

	// Create HTTP GET request
	req, err := http.NewRequest("GET", resultURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "goao/"+schema.Version)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CU result request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
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

// ListResults fetches a batch of results from a process with pagination support
// This is useful for polling for new results or browsing message history
//
// Parameters:
//   - opts: ListResultsOptions containing process ID and pagination parameters
//     - ProcessID: The AO process ID (required)
//     - From: Cursor starting point (optional)
//     - To: Cursor ending point (optional)
//     - Sort: Sort order - "ASC" or "DESC" (optional, default: "ASC")
//     - Limit: Number of results to return (optional, default: 25)
//
// Returns:
//   - []*schema.Result: Array of result objects
//   - error: Any error during the request
//
// Example:
//   opts := &schema.ListResultsOptions{
//       ProcessID: "process-id-456",
//       From: "cursor-start",
//       Sort: schema.SortASC,
//       Limit: 25,
//   }
//   results, err := client.ListResults(opts)
func (c *Client) ListResults(opts *schema.ListResultsOptions) ([]*schema.Result, error) {
	// Validate options
	if opts == nil {
		return nil, fmt.Errorf("options are required")
	}
	if opts.ProcessID == "" {
		return nil, fmt.Errorf("processID is required")
	}

	// Build query parameters
	params := url.Values{}
	params.Set("process", opts.ProcessID)

	if opts.From != "" {
		params.Set("from", opts.From)
	}
	if opts.To != "" {
		params.Set("to", opts.To)
	}
	if opts.Sort != "" {
		params.Set("sort", string(opts.Sort))
	} else {
		params.Set("sort", string(schema.SortASC))
	}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	} else {
		params.Set("limit", "25") // Default limit
	}

	// Build CU results endpoint URL
	// Format: {CU_URL}/results?process={processID}&from={from}&to={to}&sort={sort}&limit={limit}
	resultsURL := fmt.Sprintf("%s/results?%s", c.CUURL, params.Encode())

	// Create HTTP GET request
	req, err := http.NewRequest("GET", resultsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "goao/"+schema.Version)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CU results request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CU returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var results []*schema.Result
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetResultWithRetry fetches a result with automatic retries (for handling CU eventual consistency)
// AO Compute Units may take time to process messages, so this method retries on 404
//
// Parameters:
//   - messageID: The ID of the message to get results for
//   - processID: The AO process ID
//   - maxRetries: Maximum number of retry attempts
//   - retryDelay: Delay between retries (e.g., 500 * time.Millisecond)
//
// Returns:
//   - *schema.ResponseCu: The result
//   - error: Any error after all retries exhausted
//
// Example:
//   result, err := client.GetResultWithRetry("msg-id", "proc-id", 5, 500*time.Millisecond)
func (c *Client) GetResultWithRetry(messageID, processID string, maxRetries int, retryDelay time.Duration) (*schema.ResponseCu, error) {
	var result *schema.ResponseCu
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, lastErr = c.GetResult(messageID, processID)

		// Success - return immediately
		if lastErr == nil {
			return result, nil
		}

		// If not a 404 (not found), return error immediately
		if !strings.Contains(lastErr.Error(), "404") {
			return nil, lastErr
		}

		// Wait before retry (except on last attempt)
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("result not available after %d retries: %w", maxRetries, lastErr)
}

// GetResultOutput is a convenience method to extract just the Output field from a result
// Useful when you only care about the process output, not messages/spawns
//
// Parameters:
//   - messageID: The ID of the message
//   - processID: The AO process ID
//
// Returns:
//   - interface{}: The output value (can be string, map, number, etc.)
//   - error: Any error during the request
//
// Example:
//   output, err := client.GetResultOutput("msg-id", "proc-id")
//   if err != nil { panic(err) }
//   fmt.Printf("Output: %v\n", output)
func (c *Client) GetResultOutput(messageID, processID string) (interface{}, error) {
	result, err := c.GetResult(messageID, processID)
	if err != nil {
		return nil, err
	}
	return result.Output, nil
}

// GetResultOutputString is a convenience method to get the Output as a string
// Useful for simple text outputs from AO processes
//
// Parameters:
//   - messageID: The ID of the message
//   - processID: The AO process ID
//
// Returns:
//   - string: The output as a string
//   - error: Any error during the request
//
// Example:
//   output, err := client.GetResultOutputString("msg-id", "proc-id")
//   if err != nil { panic(err) }
//   fmt.Println("Output:", output)
func (c *Client) GetResultOutputString(messageID, processID string) (string, error) {
	result, err := c.GetResult(messageID, processID)
	if err != nil {
		return "", err
	}

	// Convert output to string
	if result.Output == nil {
		return "", nil
	}

	switch v := result.Output.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		// Try to marshal to JSON string
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal output: %w", err)
		}
		return string(jsonBytes), nil
	}
}

// GetResultMessages extracts the Messages array from a result
// Useful when the process sends messages as output (e.g., token transfers)
//
// Parameters:
//   - messageID: The ID of the message
//   - processID: The AO process ID
//
// Returns:
//   - []interface{}: Array of message objects
//   - error: Any error during the request
func (c *Client) GetResultMessages(messageID, processID string) ([]interface{}, error) {
	result, err := c.GetResult(messageID, processID)
	if err != nil {
		return nil, err
	}
	return result.Messages, nil
}

// GetResultSpawns extracts the Spawns array from a result
// Useful when the process spawns child processes
//
// Parameters:
//   - messageID: The ID of the message
//   - processID: The AO process ID
//
// Returns:
//   - []interface{}: Array of spawn objects
//   - error: Any error during the request
func (c *Client) GetResultSpawns(messageID, processID string) ([]interface{}, error) {
	result, err := c.GetResult(messageID, processID)
	if err != nil {
		return nil, err
	}
	return result.Spawns, nil
}

// HasResultError checks if a result contains an error
// Useful for quick error checking without parsing the full result
//
// Parameters:
//   - messageID: The ID of the message
//   - processID: The AO process ID
//
// Returns:
//   - bool: true if result has an error
//   - string: The error message (empty if no error)
//   - error: Any error during the request
func (c *Client) HasResultError(messageID, processID string) (bool, string, error) {
	result, err := c.GetResult(messageID, processID)
	if err != nil {
		return false, "", err
	}

	if result.Error != nil {
		switch v := result.Error.(type) {
		case string:
			return true, v, nil
		case map[string]interface{}:
			if msg, ok := v["message"].(string); ok {
				return true, msg, nil
			}
			return true, fmt.Sprintf("%v", v), nil
		default:
			return true, fmt.Sprintf("%v", v), nil
		}
	}

	return false, "", nil
}

// WaitForResult polls for a result until it's available or timeout is reached
// This is useful for synchronous-style interactions with AO processes
//
// Parameters:
//   - messageID: The ID of the message
//   - processID: The AO process ID
//   - timeout: Maximum time to wait (e.g., 30 * time.Second)
//   - pollInterval: Time between polls (e.g., 1 * time.Second)
//
// Returns:
//   - *schema.ResponseCu: The result
//   - error: Any error (including timeout)
//
// Example:
//   result, err := client.WaitForResult("msg-id", "proc-id", 30*time.Second, 1*time.Second)
func (c *Client) WaitForResult(messageID, processID string, timeout, pollInterval time.Duration) (*schema.ResponseCu, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		result, err := c.GetResult(messageID, processID)
		if err == nil {
			return result, nil
		}
		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("timeout waiting for result")
}

// DecodeResultOutput unmarshals the result Output into a provided struct
// Useful for strongly-typed result handling
//
// Parameters:
//   - messageID: The ID of the message
//   - processID: The AO process ID
//   - out: Pointer to struct to unmarshal into
//
// Returns:
//   - error: Any error during request or unmarshaling
//
// Example:
//   type TokenBalance struct {
//       Balance int `json:"balance"`
//       Ticker  string `json:"ticker"`
//   }
//   var balance TokenBalance
//   err := client.DecodeResultOutput("msg-id", "proc-id", &balance)
func (c *Client) DecodeResultOutput(messageID, processID string, out interface{}) error {
	result, err := c.GetResult(messageID, processID)
	if err != nil {
		return err
	}

	if result.Output == nil {
		return fmt.Errorf("result output is nil")
	}

	// If output is already bytes, unmarshal directly
	if outputBytes, ok := result.Output.([]byte); ok {
		return json.Unmarshal(outputBytes, out)
	}

	// Otherwise, marshal and re-unmarshal
	jsonBytes, err := json.Marshal(result.Output)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %w", err)
	}

	return json.Unmarshal(jsonBytes, out)
}