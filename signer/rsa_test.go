package signer

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRSASignerFromPath(t *testing.T) {
	signer, err := NewRSASignerFromPath("testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	assert.NotNil(t, signer)
	assert.NotEmpty(t, signer.Address())
	assert.NotNil(t, signer.GetPublicKey())
	assert.NotNil(t, signer.GetPrivateKey())
}

func TestRSASigner_Sign(t *testing.T) {
	signer, err := NewRSASignerFromPath("testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	data := []byte("test message for signing")
	signature, err := signer.Sign(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)
}

func TestRSASigner_PublicKey(t *testing.T) {
	signer, err := NewRSASignerFromPath("testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	pubKey := signer.PublicKey()
	assert.NotEmpty(t, pubKey)
}

func TestRSASigner_Address(t *testing.T) {
	signer, err := NewRSASignerFromPath("testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	addr := signer.Address()
	assert.NotEmpty(t, addr)
}

func TestRSASigner_Algorithm(t *testing.T) {
	signer, err := NewRSASignerFromPath("testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	alg := signer.Algorithm()
	assert.Equal(t, "rsa-pss-sha512", alg)
}

func TestRSASigner_Type(t *testing.T) {
	signer, err := NewRSASignerFromPath("testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	keyType := signer.Type()
	assert.Equal(t, 1, keyType)
}

func TestRSASigner_SignVerify_Roundtrip(t *testing.T) {
	signer, err := NewRSASignerFromPath("testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	data := []byte("test data for roundtrip verification")

	// Sign
	signature, err := signer.Sign(data)
	assert.NoError(t, err)

	// Verify using RSA public key
	hash := sha256.Sum256(data)
	err = rsa.VerifyPSS(signer.GetPublicKey(), crypto.SHA256, hash[:], signature, nil)
	assert.NoError(t, err, "Signature verification failed")
}

func TestNewRSASignerByPrivateKey(t *testing.T) {
	signer1, err := NewRSASignerFromPath("testKey.json")
	if err != nil {
		t.Skip("Skipping test: testKey.json not found")
		return
	}

	signer2 := NewRSASignerByPrivateKey(signer1.GetPrivateKey())

	assert.Equal(t, signer1.Address(), signer2.Address())
	assert.Equal(t, signer1.PublicKey(), signer2.PublicKey())
}

func TestGenerateTestRSAKey(t *testing.T) {
	privKey, pubKey, err := GenerateTestRSAKey(2048)
	require.NoError(t, err)
	assert.NotNil(t, privKey)
	assert.NotNil(t, pubKey)
	assert.Equal(t, 2048, privKey.N.BitLen())
}