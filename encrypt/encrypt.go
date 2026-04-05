package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// AESKeySize is the size of the AES-256 key in bytes
	AESKeySize = 32

	// GCMNonceSize is the size of the GCM nonce in bytes (12 bytes recommended)
	GCMNonceSize = 12

	// EncryptionAlgorithm is the identifier for this encryption scheme
	EncryptionAlgorithm = "AES-GCM+RSA-OAEP"
)

// ============================================================================
// Encryption (Sender → AO Process)
// ============================================================================

// EncryptPayload encrypts data using AES-GCM, then encrypts the AES key with RSA-OAEP
// This is a hybrid encryption scheme:
//   - AES-GCM for fast symmetric encryption of the actual data
//   - RSA-OAEP for secure asymmetric encryption of the AES key
//
// Parameters:
//   - plaintext: The data to encrypt
//   - rsaPub: The recipient's RSA public key (to encrypt the AES key)
//
// Returns:
//   - encryptedData: AES-GCM encrypted data (includes auth tag)
//   - encryptedKey: RSA-OAEP encrypted AES key
//   - nonce: The GCM nonce (public, needed for decryption)
//   - err: Any error during encryption
func EncryptPayload(plaintext []byte, rsaPub *rsa.PublicKey) (
	encryptedData, encryptedKey, nonce []byte, err error) {

	// 1. Generate random AES-256 key (32 bytes)
	// This key is used ONLY for this message, then discarded
	aesKey := make([]byte, AESKeySize)
	if _, err = io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// 2. Generate random nonce (12 bytes for GCM)
	// Nonce must be unique per encryption with the same key
	nonce = make([]byte, GCMNonceSize)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 3. Encrypt data with AES-GCM
	// GCM provides both confidentiality AND authentication
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Seal encrypts and authenticates the data
	// Output includes: ciphertext + auth tag (16 bytes)
	encryptedData = gcm.Seal(nil, nonce, plaintext, nil)

	// 4. Encrypt AES key with RSA-OAEP
	// Only the recipient with the RSA private key can decrypt this
	encryptedKey, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, aesKey, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encrypt AES key with RSA: %w", err)
	}

	// 5. Clear the AES key from memory (security best practice)
	clear(aesKey)

	return encryptedData, encryptedKey, nonce, nil
}

// ============================================================================
// Decryption (AO Process → Receives Message OR Client Receives Response)
// ============================================================================

// DecryptPayload reverses the encryption process
// This is used when:
//   1. An AO process receives an encrypted message from a client
//   2. A client receives an encrypted response from an AO process
//
// Parameters:
//   - encryptedData: AES-GCM encrypted data (from DataItem.Data)
//   - encryptedKey: RSA-OAEP encrypted AES key (from "Encrypted-Key" tag)
//   - nonce: GCM nonce (from "Nonce" tag)
//   - rsaPriv: The recipient's RSA private key (to decrypt the AES key)
//
// Returns:
//   - plaintext: The decrypted original data
//   - err: Any error during decryption
func DecryptPayload(encryptedData, encryptedKey, nonce []byte, rsaPriv *rsa.PrivateKey) ([]byte, error) {
	// Validate inputs
	if len(encryptedKey) == 0 {
		return nil, fmt.Errorf("encrypted key is empty")
	}
	if len(nonce) != GCMNonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", GCMNonceSize, len(nonce))
	}
	if len(encryptedData) == 0 {
		return nil, fmt.Errorf("encrypted data is empty")
	}

	// 1. Decrypt AES key with RSA-OAEP
	// This recovers the original AES key that was used to encrypt the data
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPriv, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// 2. Decrypt data with AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		clear(aesKey)
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		clear(aesKey)
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Open decrypts and verifies the authentication tag
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		clear(aesKey)
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	// 3. Clear the AES key from memory (security best practice)
	clear(aesKey)

	return plaintext, nil
}

// ============================================================================
// Base64URL Helpers (for storing in Tags)
// ============================================================================

// Base64URLEncode encodes bytes to Base64URL without padding
// This is safe for use in HTTP headers and AO tags
func Base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Base64URLDecode decodes Base64URL string to bytes
// Handles missing padding automatically
func Base64URLDecode(s string) ([]byte, error) {
	// Add padding if needed (Base64URL may omit padding)
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

// ============================================================================
// Key Generation Helpers
// ============================================================================

// GenerateRSAKey generates a new RSA key pair for testing
// In production, load keys from secure storage (JWK file, HSM, etc.)
func GenerateRSAKey(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	return privKey, &privKey.PublicKey, nil
}
