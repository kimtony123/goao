package goao

// mockSigner is a test implementation of the Signer interface
type mockSigner struct {
    address string
    // Add any other fields needed for tests
}

func (m *mockSigner) Sign(data []byte) ([]byte, error) {
    return []byte("signature"), nil
}

func (m *mockSigner) PublicKey() []byte {
    return []byte("public-key")
}

func (m *mockSigner) Address() string {
    return m.address
}

func (m *mockSigner) Algorithm() string {
    return "rsa-pss-sha512"
}

func (m *mockSigner) Type() int {
    return 1
}