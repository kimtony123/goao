package goao

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kimtony123/goao/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_GetResultWithRetry tests retry logic for eventual consistency
func TestClient_GetResultWithRetry(t *testing.T) {
	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// Return 404 for first 2 attempts, then success
		if attemptCount < 3 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("result not ready"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&schema.ResponseCu{
			Output: "success",
		})
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))

	result, err := client.GetResultWithRetry("message-id", "process-id", 5, 100*time.Millisecond)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "success", result.Output)
	assert.Equal(t, 3, attemptCount) // Should have retried twice
}

// TestClient_GetResultWithRetry_ExhaustedRetries tests retry exhaustion
func TestClient_GetResultWithRetry_ExhaustedRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("result not ready"))
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))

	_, err := client.GetResultWithRetry("message-id", "process-id", 2, 100*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "after 2 retries")
}

// TestClient_WaitForResult tests polling with timeout
func TestClient_WaitForResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&schema.ResponseCu{
			Output: "polled result",
		})
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))

	result, err := client.WaitForResult("message-id", "process-id", 5*time.Second, 1*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "polled result", result.Output)
}

// TestClient_GetResultOutput tests convenience method for output extraction
func TestClient_GetResultOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&schema.ResponseCu{
			Output: "hello ao",
		})
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))

	output, err := client.GetResultOutput("message-id", "process-id")
	require.NoError(t, err)
	assert.Equal(t, "hello ao", output)
}

// TestClient_GetResultOutputString tests string output extraction
func TestClient_GetResultOutputString(t *testing.T) {
	t.Run("String output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&schema.ResponseCu{
				Output: "counter: 42",
			})
		}))
		defer server.Close()

		client := NewClient(WithCU(server.URL))
		output, err := client.GetResultOutputString("message-id", "process-id")
		require.NoError(t, err)
		assert.Equal(t, "counter: 42", output)
	})

	t.Run("JSON output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&schema.ResponseCu{
				Output: map[string]interface{}{
					"balance": 1000,
					"status":  "active",
				},
			})
		}))
		defer server.Close()

		client := NewClient(WithCU(server.URL))
		output, err := client.GetResultOutputString("message-id", "process-id")
		require.NoError(t, err)
		assert.Contains(t, output, "balance")
	})

	t.Run("Nil output", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&schema.ResponseCu{
				Output: nil,
			})
		}))
		defer server.Close()

		client := NewClient(WithCU(server.URL))
		output, err := client.GetResultOutputString("message-id", "process-id")
		require.NoError(t, err)
		assert.Equal(t, "", output)
	})
}

// TestClient_GetResultMessages tests message extraction
func TestClient_GetResultMessages(t *testing.T) {
	expectedMessages := []interface{}{
		map[string]interface{}{"type": "transfer", "amount": 100},
		map[string]interface{}{"type": "mint", "amount": 50},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&schema.ResponseCu{
			Messages: expectedMessages,
		})
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))

	messages, err := client.GetResultMessages("message-id", "process-id")
	require.NoError(t, err)
	assert.Len(t, messages, 2)
}

// TestClient_GetResultSpawns tests spawn extraction
func TestClient_GetResultSpawns(t *testing.T) {
	expectedSpawns := []interface{}{
		map[string]interface{}{"process": "child-process-1"},
		map[string]interface{}{"process": "child-process-2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&schema.ResponseCu{
			Spawns: expectedSpawns,
		})
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))

	spawns, err := client.GetResultSpawns("message-id", "process-id")
	require.NoError(t, err)
	assert.Len(t, spawns, 2)
}

// TestClient_HasResultError tests error checking helper
func TestClient_HasResultError(t *testing.T) {
	t.Run("No error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&schema.ResponseCu{
				Error: nil,
			})
		}))
		defer server.Close()

		client := NewClient(WithCU(server.URL))
		hasError, errMsg, err := client.HasResultError("message-id", "process-id")
		require.NoError(t, err)
		assert.False(t, hasError)
		assert.Empty(t, errMsg)
	})

	t.Run("With error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&schema.ResponseCu{
				Error: "Insufficient balance",
			})
		}))
		defer server.Close()

		client := NewClient(WithCU(server.URL))
		hasError, errMsg, err := client.HasResultError("message-id", "process-id")
		require.NoError(t, err)
		assert.True(t, hasError)
		assert.Equal(t, "Insufficient balance", errMsg)
	})
}

// TestClient_DecodeResultOutput tests JSON unmarshaling into struct
func TestClient_DecodeResultOutput(t *testing.T) {
	type TokenBalance struct {
		Balance int    `json:"balance"`
		Ticker  string `json:"ticker"`
	}

	expectedBalance := TokenBalance{
		Balance: 1000,
		Ticker:  "MTK",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&schema.ResponseCu{
			Output: expectedBalance,
		})
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))

	var balance TokenBalance
	err := client.DecodeResultOutput("message-id", "process-id", &balance)
	require.NoError(t, err)
	assert.Equal(t, 1000, balance.Balance)
	assert.Equal(t, "MTK", balance.Ticker)
}

// BenchmarkGetResult benchmarks single result fetching
func BenchmarkGetResult(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&schema.ResponseCu{
			Output: "benchmark",
		})
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetResult("message-id", "process-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListResults benchmarks batch result fetching
func BenchmarkListResults(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*schema.Result{
			{MessageID: "msg1", Output: "out1"},
			{MessageID: "msg2", Output: "out2"},
		})
	}))
	defer server.Close()

	client := NewClient(WithCU(server.URL))
	opts := &schema.ListResultsOptions{
		ProcessID: "process-123",
		Limit:     25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ListResults(opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}