package signer

// Signer is the interface that wraps the basic signing operations.
type Signer interface {
    // Sign signs the given data and returns the raw signature bytes.
    Sign(data []byte) ([]byte, error)

    // PublicKey returns the raw public key bytes.
    PublicKey() []byte

    // Address returns the Arweave address derived from the public key.
    Address() string

    // Algorithm returns the signing algorithm identifier.
    Algorithm() string

    // Type returns the key type identifier (1 = Arweave/RSA, 2 = Ethereum/ECDSA).
    Type() int
}