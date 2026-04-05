package signer

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// ECDSASigner implements the Signer interface for Ethereum ECDSA keys (secp256k1)
type ECDSASigner struct {
	privateKey string
	address    string
	pubKey     []byte
	privKey    *ecdsa.PrivateKey
}

// NewECDSASigner creates a new ECDSASigner from a hex-encoded private key
func NewECDSASigner(privateKey string) (*ECDSASigner, error) {
	privKey, err := ethcrypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Get compressed public key (33 bytes)
	pubKey := ethcrypto.CompressPubkey(&privKey.PublicKey)

	// Derive Arweave-style address: SHA-256(publicKey) base64url encoded
	addr := sha256.Sum256(pubKey)

	return &ECDSASigner{
		privateKey: privateKey,
		address:    base64.RawURLEncoding.EncodeToString(addr[:]),
		pubKey:     pubKey,
		privKey:    privKey,
	}, nil
}

// NewECDSASignerRandom creates a new ECDSASigner with a randomly generated private key
func NewECDSASignerRandom() (*ECDSASigner, error) {
	privKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	privateKeyHex := fmt.Sprintf("%x", ethcrypto.FromECDSA(privKey))
	return NewECDSASigner(privateKeyHex)
}

// Sign signs the given data using ECDSA secp256k1
// Returns 64-byte signature (r + s)
func (s *ECDSASigner) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := ethcrypto.Sign(hash[:], s.privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	// ethcrypto.Sign returns 65 bytes (r, s, v), but we only need 64 bytes (r, s)
	return signature[:64], nil
}

// PublicKey returns the compressed ECDSA public key (33 bytes)
func (s *ECDSASigner) PublicKey() []byte {
	return s.pubKey
}

// Address returns the Arweave-style address (base64url encoded)
func (s *ECDSASigner) Address() string {
	return s.address
}

// Algorithm returns the signing algorithm identifier
func (s *ECDSASigner) Algorithm() string {
	return "secp256k1"
}

// Type returns the key type (2 = Ethereum/ECDSA)
func (s *ECDSASigner) Type() int {
	return 2
}

// GetPrivateKey returns the underlying ECDSA private key (for advanced use)
func (s *ECDSASigner) GetPrivateKey() *ecdsa.PrivateKey {
	return s.privKey
}

// GetPrivateKeyHex returns the hex-encoded private key
func (s *ECDSASigner) GetPrivateKeyHex() string {
	return s.privateKey
}

// VerifySignature verifies an ECDSA signature
func (s *ECDSASigner) VerifySignature(data []byte, signature []byte) (bool, error) {
	if len(signature) != 64 {
		return false, fmt.Errorf("invalid signature length: expected 64, got %d", len(signature))
	}

	hash := sha256.Sum256(data)

	// Try both recovery bytes (0 and 1) since we don't know which one was used
	for v := byte(0); v <= 1; v++ {
		// Add recovery byte for SigToPub (go-ethereum expects 65 bytes)
		sigWithV := make([]byte, 65)
		copy(sigWithV[:64], signature)
		sigWithV[64] = v

		// Recover public key
		pubKey, err := ethcrypto.SigToPub(hash[:], sigWithV)
		if err != nil {
			continue
		}

		// Check if recovered public key matches our public key
		recoveredPubKey := ethcrypto.CompressPubkey(pubKey)
		if string(recoveredPubKey) == string(s.pubKey) {
			// Verify signature (needs 64 bytes without recovery byte)
			valid := ethcrypto.VerifySignature(
				recoveredPubKey,
				hash[:],
				signature,
			)
			if valid {
				return true, nil
			}
		}
	}

	return false, nil
}

// RecoverPublicKey recovers the public key from a signature
func RecoverPublicKey(data []byte, signature []byte) ([]byte, error) {
	if len(signature) != 64 {
		return nil, fmt.Errorf("invalid signature length: expected 64, got %d", len(signature))
	}

	hash := sha256.Sum256(data)

	// Add recovery byte for SigToPub
	sigWithV := make([]byte, 65)
	copy(sigWithV[:64], signature)
	sigWithV[64] = 0

	pubKey, err := ethcrypto.SigToPub(hash[:], sigWithV)
	if err != nil {
		return nil, fmt.Errorf("failed to recover public key: %w", err)
	}

	return ethcrypto.CompressPubkey(pubKey), nil
}

// PublicKeyToAddress derives an Arweave-style address from a public key
func PublicKeyToAddress(publicKey []byte) string {
	addr := sha256.Sum256(publicKey)
	return base64.RawURLEncoding.EncodeToString(addr[:])
}