package goao

import (
    "encoding/json"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestNewClient tests client initialization with defaults
func TestNewClient(t *testing.T) {
    client := NewClient()

    assert.NotNil(t, client)
    assert.Equal(t, DefaultMUURL, client.MUURL)
    assert.Equal(t, DefaultCUURL, client.CUURL)
    assert.Equal(t, DefaultSUURL, client.SUURL)
    assert.Equal(t, DefaultComputeGateway, client.ComputeGateway)
    assert.NotNil(t, client.httpClient)
}

// TestNewClient_WithOptions tests client initialization with custom options
func TestNewClient_WithOptions(t *testing.T) {
    customMU := "https://custom-mu.example.com"
    customCU := "https://custom-cu.example.com"
    customCompute := "https://custom-compute.example.com"

    client := NewClient(
        WithMU(customMU),
        WithCU(customCU),
        WithComputeGateway(customCompute),
    )

    assert.Equal(t, customMU, client.MUURL)
    assert.Equal(t, customCU, client.CUURL)
    assert.Equal(t, customCompute, client.ComputeGateway)
}

// TestClient_doRequest tests HTTP request handling
func TestClient_doRequest(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "GET", r.Method)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status": "ok"}`))
    }))
    defer server.Close()

    client := NewClient()
    resp, err := client.doRequest("GET", server.URL, nil, nil)
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestClient_doJSONRequest tests JSON request/response handling
func TestClient_doJSONRequest(t *testing.T) {
    type Request struct {
        Name string `json:"name"`
    }
    type Response struct {
        Message string `json:"message"`
    }

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

        var req Request
        err := json.NewDecoder(r.Body).Decode(&req)
        require.NoError(t, err)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(Response{Message: "success"})
    }))
    defer server.Close()

    client := NewClient()
    var resp Response
    err := client.doJSONRequest("POST", server.URL, Request{Name: "test"}, &resp)
    require.NoError(t, err)
    assert.Equal(t, "success", resp.Message)
}

// TestClient_doBinaryRequest tests binary request handling
func TestClient_doBinaryRequest(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))

        body, _ := io.ReadAll(r.Body)
        assert.Equal(t, []byte("binary data"), body)

        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    client := NewClient()
    resp, err := client.doBinaryRequest("POST", server.URL, []byte("binary data"), "application/octet-stream")
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestClient_WithHTTPClient tests custom HTTP client
func TestClient_WithHTTPClient(t *testing.T) {
    customClient := &http.Client{
        Timeout: 60 * time.Second,
    }

    client := NewClient(WithHTTPClient(customClient))
    assert.Equal(t, customClient, client.httpClient)
}

// TestClient_WithSigner tests signer configuration
func TestClient_WithSigner(t *testing.T) {
    mockSigner := &mockSigner{
        address: "test-address",
    }

    client := NewClient(WithSigner(mockSigner))
    assert.Equal(t, mockSigner, client.Signer)
}