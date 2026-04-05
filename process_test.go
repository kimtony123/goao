package goao

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kimtony123/goao/encrypt"
	"github.com/kimtony123/goao/schema"
	"github.com/kimtony123/goao/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_SendMessage tests sending a message to an AO process
func TestClient_SendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/submit", r.URL.Path)
		assert.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "message-id-123"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithMU(server.URL),
		WithSigner(&mockSigner{address: "test-address"}),
	)

	messageID, err := client.SendMessage(
		"process-123",
		[]byte("hello ao"),
		[]schema.Tag{{Name: "Action", Value: "Message"}},
		nil,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, messageID)
}

// TestClient_SendMessage_Encrypted tests sending an encrypted message
func TestClient_SendMessage_Encrypted(t *testing.T) {
	privKey, pubKey, err := signer.GenerateTestRSAKey(2048)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/submit", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "encrypted-message-id"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithMU(server.URL),
		WithSigner(signer.NewRSASignerByPrivateKey(privKey)),
	)

	opts := &schema.SendMessageOptions{
		EncryptWithRSA: pubKey,
	}

	messageID, err := client.SendMessage(
		"process-123",
		[]byte("secret message"),
		[]schema.Tag{{Name: "Action", Value: "Transfer"}},
		opts,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, messageID)
}

// TestClient_SpawnProcess tests spawning a new AO process
func TestClient_SpawnProcess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/submit", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "process-id-456"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithSU(server.URL),
		WithSigner(&mockSigner{address: "test-address"}),
	)

	processID, err := client.SpawnProcess(
		"module-tx-id",
		[]byte(`{"initial": "state"}`),
		[]schema.Tag{{Name: "App-Name", Value: "TestApp"}},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, processID)
}

// TestClient_DryRun tests legacy DryRun functionality
func TestClient_DryRun(t *testing.T) {
	expectedResponse := &schema.ResponseCu{
		Output:  "counter: 42",
		GasUsed: 1000,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/dry-run")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := NewClient(
		WithCU(server.URL),
		WithSigner(&mockSigner{address: "test-address"}),
	)

	result, err := client.DryRun(
		"process-123",
		[]byte("return counter"),
		[]schema.Tag{{Name: "Action", Value: "Eval"}},
		nil,
	)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "counter: 42", result.Output)
}

// TestClient_GetCompute tests new HTTP Compute endpoint
func TestClient_GetCompute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		expectedPath := "/process-123~process@1.0/compute/counter"
		assert.Equal(t, expectedPath, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("42"))
	}))
	defer server.Close()

	client := NewClient(WithComputeGateway(server.URL))

	data, err := client.GetCompute("process-123", "counter")
	require.NoError(t, err)
	assert.Equal(t, "42", string(data))
}

// TestClient_GetComputeString tests string response helper
func TestClient_GetComputeString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("active"))
	}))
	defer server.Close()

	client := NewClient(WithComputeGateway(server.URL))

	result, err := client.GetComputeString("process-123", "status")
	require.NoError(t, err)
	assert.Equal(t, "active", result)
}

// TestClient_GetComputeJSON tests JSON response helper
func TestClient_GetComputeJSON(t *testing.T) {
	type State struct {
		Count  int    `json:"count"`
		Status string `json:"status"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(State{Count: 100, Status: "active"})
	}))
	defer server.Close()

	client := NewClient(WithComputeGateway(server.URL))

	var state State
	err := client.GetComputeJSON("process-123", "state", &state)
	require.NoError(t, err)
	assert.Equal(t, 100, state.Count)
	assert.Equal(t, "active", state.Status)
}

// TestClient_DecryptResponse tests decrypting encrypted responses
func TestClient_DecryptResponse(t *testing.T) {
	privKey, pubKey, err := signer.GenerateTestRSAKey(2048)
	require.NoError(t, err)

	originalData := []byte("secret response")
	encryptedData, encryptedKey, nonce, err := encrypt.EncryptPayload(originalData, pubKey)
	require.NoError(t, err)

	rsaSigner := signer.NewRSASignerByPrivateKey(privKey)

	client := NewClient(
		WithSigner(rsaSigner),
	)

	decrypted, err := client.DecryptResponse(
		encrypt.Base64URLEncode(encryptedData),
		encrypt.Base64URLEncode(encryptedKey),
		encrypt.Base64URLEncode(nonce),
		rsaSigner,
	)
	require.NoError(t, err)
	assert.Equal(t, originalData, decrypted)
}