//go:build integration
// +build integration

package goao

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kimtony123/goao/schema"
	"github.com/kimtony123/goao/signer"
	"github.com/stretchr/testify/require"
)

func TestRealProcess(t *testing.T) {
	// Load RSA signer from testdata
	rsaSigner, err := signer.NewRSASignerFromPath("testdata/testKey.json")
	require.NoError(t, err, "failed to load RSA key")

	// Create client with default gateways
	client := NewClient(
		WithSigner(rsaSigner),
		WithMU("https://mu.ao.arweave.net"),
		WithCU("https://cu.ao.arweave.net"),
		WithComputeGateway("https://push.forward.computer"),
	)

	processID := "6wqH8ue2-bnJG7j--FV0KGYzSs53ObFDofDITb7qtxI"

	// 1. Verify compute endpoint returns 42
	t.Log("Checking compute counter...")
	counter, err := client.GetComputeString(processID, "counter")
	require.NoError(t, err)
	require.Equal(t, "42", counter)
	t.Logf("Compute counter: %s", counter)

	// 2. Send AddProduct message
	addProductData := `{
		"product_type": "Website DApp",
		"category": "Infrastructure",
		"name": "Aostore",
		"description": "Aostore serves as the Playstore for Arweave and Aocomputer, enabling seamless collaboration between users and project owners. It empowers individuals to build and showcase their reputation on the Permaweb while rewarding users for their meaningful contributions to projects and the ecosystem.",
		"website_url": "https://aostore-orpin.vercel.app/",
		"logo_url": "https://pbs.twimg.com/profile_images/1878960977489571841/9tdFyNf9_400x400.jpg",
		"blockchain": "Arweave",
		"referral_fee": 0.02
	}`

	tags := []schema.Tag{
		{Name: "Action", Value: "AddProduct"},
	}

	t.Log("Sending AddProduct message...")
	msgID, err := client.SendMessage(processID, []byte(addProductData), tags, nil)
	require.NoError(t, err)
	t.Logf("Message sent, ID: %s", msgID)

	// Wait for result (WaitForResult expects 4 args)
	result, err := client.WaitForResult(msgID, processID, 30*time.Second, 2*time.Second)
	require.NoError(t, err)
	t.Logf("Result output: %v", result.Output)

	// Convert result.Output (interface{}) to byte slice for unmarshal
	var outputBytes []byte
	switch v := result.Output.(type) {
	case string:
		outputBytes = []byte(v)
	case []byte:
		outputBytes = v
	default:
		t.Fatalf("unexpected Output type: %T", v)
	}

	var addResp struct {
		Success   bool   `json:"success"`
		Message   string `json:"message"`
		ProductID string `json:"product_id"`
	}
	err = json.Unmarshal(outputBytes, &addResp)
	require.NoError(t, err, "failed to parse AddProduct response")
	require.True(t, addResp.Success, "AddProduct was not successful")
	require.NotEmpty(t, addResp.ProductID, "ProductID missing")
	t.Logf("Product created: %s", addResp.ProductID)

	// 3. Send FetchAllApps message
	fetchTags := []schema.Tag{
		{Name: "Action", Value: "FetchAllApps"},
	}
	t.Log("Sending FetchAllApps message...")
	fetchMsgID, err := client.SendMessage(processID, []byte(""), fetchTags, nil)
	require.NoError(t, err)
	t.Logf("FetchAllApps message ID: %s", fetchMsgID)

	fetchResult, err := client.WaitForResult(fetchMsgID, processID, 30*time.Second, 2*time.Second)
	require.NoError(t, err)
	t.Logf("FetchAllApps output: %v", fetchResult.Output)

	// Convert output similarly
	var fetchBytes []byte
	switch v := fetchResult.Output.(type) {
	case string:
		fetchBytes = []byte(v)
	case []byte:
		fetchBytes = v
	default:
		t.Fatalf("unexpected Output type: %T", v)
	}

	var apps []interface{}
	err = json.Unmarshal(fetchBytes, &apps)
	if err == nil {
		t.Logf("Number of apps: %d", len(apps))
	} else {
		t.Logf("Raw output (non-JSON): %s", fetchBytes)
	}
}