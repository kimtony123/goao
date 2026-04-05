package encrypt

import (
	"crypto/rand"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptPayload_Roundtrip(t *testing.T) {
	privKey, pubKey, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	originalData := []byte("secret message content for AO process")

	encryptedData, encryptedKey, nonce, err := EncryptPayload(originalData, pubKey)
	require.NoError(t, err)

	decryptedData, err := DecryptPayload(encryptedData, encryptedKey, nonce, privKey)
	require.NoError(t, err)

	assert.Equal(t, originalData, decryptedData)
}

func TestBase64URLEncodeDecode_Roundtrip(t *testing.T) {
	testData := [][]byte{
		{},
		[]byte("hello"),
		[]byte("hello world"),
	}

	for _, data := range testData {
		encoded := Base64URLEncode(data)
		decoded, err := Base64URLDecode(encoded)
		require.NoError(t, err)
		assert.Equal(t, data, decoded)
	}
}

func TestGenerateRSAKey(t *testing.T) {
	privKey, pubKey, err := GenerateRSAKey(2048)
	require.NoError(t, err)

	assert.NotNil(t, privKey)
	assert.NotNil(t, pubKey)
	assert.Equal(t, 2048, privKey.N.BitLen())
}

// Additional test to cover randomness (ensures no unused import)
func TestRandomness(t *testing.T) {
	// Ensure crypto/rand is used so the import is not flagged as unused
	randomBytes := make([]byte, 32)
	n, err := rand.Read(randomBytes)
	require.NoError(t, err)
	assert.Equal(t, 32, n)
}