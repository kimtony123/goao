package compute

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient tests client initialization with default gateway
func TestNewClient(t *testing.T) {
	client := NewClient("")

	assert.NotNil(t, client)
	assert.Equal(t, DefaultComputeGateway, client.gatewayURL)
	assert.NotNil(t, client.httpClient)
}

// TestNewClient_CustomGateway tests client initialization with custom gateway
func TestNewClient_CustomGateway(t *testing.T) {
	customGateway := "https://custom-node.example.com"
	client := NewClient(customGateway)

	assert.Equal(t, customGateway, client.gatewayURL)
}

// TestNewClient_TrailingSlash tests that trailing slashes are removed
func TestNewClient_TrailingSlash(t *testing.T) {
	client := NewClient("https://push.forward.computer/")

	assert.Equal(t, "https://push.forward.computer", client.gatewayURL)
}

// TestGetCompute_ComponentBased tests component-based URL construction
func TestGetCompute_ComponentBased(t *testing.T) {
	expectedProcessID := "Ma32vsWVPtedS3Rav_vUR9JvUfffQzPlhBpwWBSf2YU"
	expectedField := "counter"
	expectedResponse := "42"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/" + expectedProcessID + "~process@1.0/compute/" + expectedField
		assert.Equal(t, expectedPath, r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	data, err := client.GetCompute(expectedProcessID, expectedField)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, string(data))
}

// TestGetCompute_FullURL tests full URL pattern (user provides complete URL)
func TestGetCompute_FullURL(t *testing.T) {
	expectedResponse := "100"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the full path is used
		assert.Contains(t, r.URL.Path, "custom-process~process@1.0/compute/status")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	client := NewClient("https://unused-default.com") // This should be ignored

	// User provides full URL
	fullURL := server.URL + "/custom-process~process@1.0/compute/status"
	data, err := client.GetCompute(fullURL, "") // fieldPath ignored
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, string(data))
}

// TestGetComputeWithURL tests the convenience method
func TestGetComputeWithURL(t *testing.T) {
	expectedResponse := "active"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	client := NewClient("")

	fullURL := server.URL + "/process-123~process@1.0/compute/status"
	data, err := client.GetComputeWithURL(fullURL)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, string(data))
}

// TestGetComputeJSON tests JSON unmarshaling with component-based URL
func TestGetComputeJSON(t *testing.T) {
	type State struct {
		Count  int    `json:"count"`
		Status string `json:"status"`
	}

	expectedState := State{
		Count:  100,
		Status: "active",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedState)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	var result State
	err := client.GetComputeJSON("process-123", "state", &result)
	require.NoError(t, err)

	assert.Equal(t, expectedState.Count, result.Count)
	assert.Equal(t, expectedState.Status, result.Status)
}

// TestGetComputeString tests string response
func TestGetComputeString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello ao"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetComputeString("process-123", "greeting")
	require.NoError(t, err)
	assert.Equal(t, "hello ao", result)
}

// TestGetCompute_DifferentGateways tests switching between different gateways
func TestGetCompute_DifferentGateways(t *testing.T) {
	// Create two mock servers (simulating different nodes)
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("node1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("node2"))
	}))
	defer server2.Close()

	// Start with server1
	client := NewClient(server1.URL)
	data, err := client.GetCompute("process-123", "field")
	require.NoError(t, err)
	assert.Equal(t, "node1", string(data))

	// Switch to server2
	client.SetGatewayURL(server2.URL)
	data, err = client.GetCompute("process-123", "field")
	require.NoError(t, err)
	assert.Equal(t, "node2", string(data))
}

// TestGetCompute_Error tests error handling
func TestGetCompute_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.GetCompute("process-123", "field")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// TestGetCompute_FullURL_IgnoresGateway tests that full URL ignores configured gateway
func TestGetCompute_FullURL_IgnoresGateway(t *testing.T) {
	expectedResponse := "from-full-url"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer server.Close()

	// Client configured with different gateway (should be ignored)
	client := NewClient("https://unused-gateway.com")

	fullURL := server.URL + "/process~process@1.0/compute/data"
	data, err := client.GetCompute(fullURL, "")
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, string(data))
}

// TestSetGatewayURL tests runtime gateway switching
func TestSetGatewayURL(t *testing.T) {
	client := NewClient("")
	assert.Equal(t, DefaultComputeGateway, client.GetGatewayURL())

	client.SetGatewayURL("https://new-node.example.com")
	assert.Equal(t, "https://new-node.example.com", client.GetGatewayURL())

	// Empty string should reset to default
	client.SetGatewayURL("")
	assert.Equal(t, DefaultComputeGateway, client.GetGatewayURL())
}

// BenchmarkGetCompute_ComponentBased benchmarks component-based URL construction
func BenchmarkGetCompute_ComponentBased(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("42"))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetCompute("process-123", "counter")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetCompute_FullURL benchmarks full URL pattern
func BenchmarkGetCompute_FullURL(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("42"))
	}))
	defer server.Close()

	client := NewClient("")
	fullURL := server.URL + "/process-123~process@1.0/compute/counter"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetCompute(fullURL, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}