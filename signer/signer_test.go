package signer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRSASigner_ImplementsSigner verifies RSA signer implements the Signer interface
func TestRSASigner_ImplementsSigner(t *testing.T) {
	var _ Signer = (*RSASigner)(nil)
}

// TestECDSASigner_ImplementsSigner verifies ECDSA signer implements the Signer interface
func TestECDSASigner_ImplementsSigner(t *testing.T) {
	var _ Signer = (*ECDSASigner)(nil)
}

// TestSigner_InterfaceMethods verifies all interface methods work correctly
func TestSigner_InterfaceMethods(t *testing.T) {
	// Test RSA signer
	t.Run("RSA Signer", func(t *testing.T) {
		signer, err := NewRSASignerFromPath("testKey.json")
		if err != nil {
			t.Skip("Skipping RSA test: testKey.json not found")
			return
		}

		data := []byte("test data for signing")
		sig, err := signer.Sign(data)
		assert.NoError(t, err)
		assert.NotEmpty(t, sig)

		pubKey := signer.PublicKey()
		assert.NotEmpty(t, pubKey)

		addr := signer.Address()
		assert.NotEmpty(t, addr)

		alg := signer.Algorithm()
		assert.Equal(t, "rsa-pss-sha512", alg)

		keyType := signer.Type()
		assert.Equal(t, 1, keyType)
	})

	// Test ECDSA signer
	t.Run("ECDSA Signer", func(t *testing.T) {
		// Use a known test private key
		privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
		signer, err := NewECDSASigner(privateKey)
		assert.NoError(t, err)

		data := []byte("test data for signing")
		sig, err := signer.Sign(data)
		assert.NoError(t, err)
		assert.Len(t, sig, 64) // ECDSA signature is 64 bytes

		pubKey := signer.PublicKey()
		assert.NotEmpty(t, pubKey)

		addr := signer.Address()
		assert.NotEmpty(t, addr)

		alg := signer.Algorithm()
		assert.Equal(t, "secp256k1", alg)

		keyType := signer.Type()
		assert.Equal(t, 2, keyType)
	})
}