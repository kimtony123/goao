package signer

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewECDSASigner(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signer)
}

func TestECDSASigner_Sign(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	data := []byte("test message for signing")
	signature, err := signer.Sign(data)
	assert.NoError(t, err)
	assert.Len(t, signature, 64)
}

func TestECDSASigner_PublicKey(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	pubKey := signer.PublicKey()
	assert.NotEmpty(t, pubKey)
	assert.Len(t, pubKey, 33)
}

func TestECDSASigner_Address(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	addr := signer.Address()
	assert.NotEmpty(t, addr)
}

func TestECDSASigner_Algorithm(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	alg := signer.Algorithm()
	assert.Equal(t, "secp256k1", alg)
}

func TestECDSASigner_Type(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	keyType := signer.Type()
	assert.Equal(t, 2, keyType)
}

func TestECDSASigner_SignVerify_Roundtrip(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	data := []byte("test data for roundtrip verification")

	signature, err := signer.Sign(data)
	assert.NoError(t, err)

	valid, err := signer.VerifySignature(data, signature)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestECDSASigner_Random(t *testing.T) {
	signer, err := NewECDSASignerRandom()
	require.NoError(t, err)
	assert.NotNil(t, signer)
}

func TestRecoverPublicKey(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	data := []byte("test data for public key recovery")
	signature, err := signer.Sign(data)
	assert.NoError(t, err)

	recoveredPubKey, err := RecoverPublicKey(data, signature)
	assert.NoError(t, err)
	assert.Equal(t, signer.PublicKey(), recoveredPubKey)
}

func TestPublicKeyToAddress(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	addr := PublicKeyToAddress(signer.PublicKey())
	assert.Equal(t, signer.Address(), addr)
}

func TestECDSASigner_GetPrivateKeyHex(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	assert.Equal(t, privateKey, signer.GetPrivateKeyHex())
}

func TestECDSASigner_InvalidSignature(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	data := []byte("test data")
	signature, err := signer.Sign(data)
	assert.NoError(t, err)

	signature[0] ^= 0xFF

	valid, err := signer.VerifySignature(data, signature)
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestECDSASigner_EthereumCompatibility(t *testing.T) {
	privateKey := "ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce"
	signer, err := NewECDSASigner(privateKey)
	assert.NoError(t, err)

	ethPrivKey, err := crypto.HexToECDSA(privateKey)
	assert.NoError(t, err)

	expectedPubKey := crypto.CompressPubkey(&ethPrivKey.PublicKey)
	assert.Equal(t, expectedPubKey, signer.PublicKey())
}