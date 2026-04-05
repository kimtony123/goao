package signer

import (
	"crypto"      // ✅ ADDED
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/everFinance/gojwk"
)

// RSASigner implements the Signer interface for Arweave RSA keys (JWK format)
type RSASigner struct {
	address string
	pubKey  *rsa.PublicKey
	prvKey  *rsa.PrivateKey
}

// NewRSASignerFromPath creates a new RSASigner from a JWK file path
func NewRSASignerFromPath(path string) (*RSASigner, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}
	return NewRSASigner(b)
}

// NewRSASigner creates a new RSASigner from JWK bytes
func NewRSASigner(b []byte) (*RSASigner, error) {
	key, err := gojwk.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JWK: %w", err)
	}

	pubKey, err := key.DecodePublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	pub, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA type")
	}

	prvKey, err := key.DecodePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	prv, ok := prvKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA type")
	}

	// Derive Arweave address: SHA-256(publicKey modulus) base64url encoded
	addr := sha256.Sum256(pub.N.Bytes())

	return &RSASigner{
		address: base64.RawURLEncoding.EncodeToString(addr[:]),
		pubKey:  pub,
		prvKey:  prv,
	}, nil
}

// NewRSASignerByPrivateKey creates a new RSASigner from an RSA private key
func NewRSASignerByPrivateKey(privateKey *rsa.PrivateKey) *RSASigner {
	pub := &privateKey.PublicKey
	addr := sha256.Sum256(pub.N.Bytes())
	return &RSASigner{
		address: base64.RawURLEncoding.EncodeToString(addr[:]),
		pubKey:  pub,
		prvKey:  privateKey,
	}
}

// Sign signs the given data using RSA-PSS with SHA-256
func (s *RSASigner) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := rsa.SignPSS(rand.Reader, s.prvKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}
	return signature, nil
}

// PublicKey returns the RSA public key modulus as bytes
func (s *RSASigner) PublicKey() []byte {
	return s.pubKey.N.Bytes()
}

// Address returns the Arweave address (base64url encoded)
func (s *RSASigner) Address() string {
	return s.address
}

// Algorithm returns the signing algorithm identifier
func (s *RSASigner) Algorithm() string {
	return "rsa-pss-sha512"
}

// Type returns the key type (1 = Arweave/RSA)
func (s *RSASigner) Type() int {
	return 1
}

// GetPrivateKey returns the underlying RSA private key (for advanced use)
func (s *RSASigner) GetPrivateKey() *rsa.PrivateKey {
	return s.prvKey
}

// GetPublicKey returns the underlying RSA public key (for advanced use)
func (s *RSASigner) GetPublicKey() *rsa.PublicKey {
	return s.pubKey
}

// GenerateTestRSAKey generates a test RSA key pair (for tests)
func GenerateTestRSAKey(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privKey, &privKey.PublicKey, nil
}