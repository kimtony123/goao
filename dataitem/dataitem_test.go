package dataitem

import (
	"crypto/rand"
	"testing"

	"github.com/kimtony123/goao/schema"
	"github.com/kimtony123/goao/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDataItem tests DataItem creation
func TestNewDataItem(t *testing.T) {
	tags := []schema.Tag{
		{Name: "Action", Value: "Message"},
		{Name: "Data-Protocol", Value: "ao"},
	}

	anchor := make([]byte, 32)
	_, err := rand.Read(anchor)
	require.NoError(t, err)

	di := NewDataItem("process-123", []byte("hello ao"), tags, anchor)

	assert.Equal(t, "process-123", di.Target)
	assert.Equal(t, "hello ao", string(di.Data))
	assert.Len(t, di.Tags, 2)
	assert.Len(t, di.Anchor, 32)
	assert.Empty(t, di.ID)
	assert.Empty(t, di.Signature)
}

// TestDataItem_Sign tests DataItem signing
func TestDataItem_Sign(t *testing.T) {
	s, err := signer.NewRSASignerFromPath("../testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	di := NewDataItem(
		"process-123",
		[]byte("test message"),
		[]schema.Tag{{Name: "Action", Value: "Test"}},
		make([]byte, 32),
	)

	err = di.Sign(s)
	require.NoError(t, err)

	assert.NotEmpty(t, di.ID)
	assert.NotEmpty(t, di.Signature)
	assert.NotEmpty(t, di.Owner)
}

// TestDataItem_Sign_ECDSA tests DataItem signing with ECDSA
func TestDataItem_Sign_ECDSA(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	s, err := signer.NewECDSASigner(privateKey)
	require.NoError(t, err)

	di := NewDataItem(
		"process-123",
		[]byte("test message"),
		[]schema.Tag{{Name: "Action", Value: "Test"}},
		make([]byte, 32),
	)

	err = di.Sign(s)
	require.NoError(t, err)

	assert.NotEmpty(t, di.ID)
	assert.NotEmpty(t, di.Signature)
	assert.Len(t, di.Signature, 64)
	assert.NotEmpty(t, di.Owner)
}

// TestDataItem_SerializeRoundtrip tests serialize/deserialize roundtrip
func TestDataItem_SerializeRoundtrip(t *testing.T) {
	s, err := signer.NewRSASignerFromPath("../testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	original := NewDataItem(
		"process-456",
		[]byte("roundtrip test data"),
		[]schema.Tag{
			{Name: "Action", Value: "Message"},
			{Name: "Type", Value: "Roundtrip"},
		},
		make([]byte, 32),
	)

	err = original.Sign(s)
	require.NoError(t, err)

	binary, err := original.Serialize()
	require.NoError(t, err)

	deserialized := &DataItem{}
	err = deserialized.Deserialize(binary)
	require.NoError(t, err)

	assert.Equal(t, original.ID, deserialized.ID)
	assert.Equal(t, original.Target, deserialized.Target)
	assert.Equal(t, original.Data, deserialized.Data)
}