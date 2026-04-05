package dataitem

import (
	"bytes"
	"testing"

	"github.com/kimtony123/goao/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteVarInt tests variable-length integer writing
func TestWriteVarInt(t *testing.T) {
	tests := []struct {
		name     string
		value    uint64
		expected []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"small", 1, []byte{0x01}},
		{"medium", 127, []byte{0x7f}},
		{"large", 128, []byte{0x80, 0x01}},
		{"very large", 16384, []byte{0x80, 0x80, 0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteVarInt(&buf, tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.Bytes())
		})
	}
}

// TestReadVarInt tests variable-length integer reading
func TestReadVarInt(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint64
	}{
		{"zero", []byte{0x00}, 0},
		{"small", []byte{0x01}, 1},
		{"medium", []byte{0x7f}, 127},
		{"large", []byte{0x80, 0x01}, 128},
		{"very large", []byte{0x80, 0x80, 0x01}, 16384},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			value, err := ReadVarInt(reader)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, value)
		})
	}
}

// TestVarInt_Roundtrip tests varint write/read roundtrip
func TestVarInt_Roundtrip(t *testing.T) {
	values := []uint64{0, 1, 127, 128, 255, 256, 16383, 16384, 65535, 1000000}

	for _, val := range values {
		var buf bytes.Buffer
		err := WriteVarInt(&buf, val)
		require.NoError(t, err)

		reader := bytes.NewReader(buf.Bytes())
		readVal, err := ReadVarInt(reader)
		require.NoError(t, err)

		assert.Equal(t, val, readVal, "Roundtrip failed for value %d", val)
	}
}

// TestBase64URLEncode tests Base64URL encoding
func TestBase64URLEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", []byte{}, ""},
		{"simple", []byte("hello"), "aGVsbG8"},
		{"binary", []byte{0x00, 0x01, 0x02}, "AAEC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base64URLEncode(tt.input)
			assert.Equal(t, tt.expected, result)
			// Verify no padding
			assert.NotContains(t, result, "=")
		})
	}
}

// TestBase64URLDecode tests Base64URL decoding
func TestBase64URLDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{"empty", "", []byte{}},
		{"simple", "aGVsbG8", []byte("hello")},
		{"with padding", "AAEC", []byte{0x00, 0x01, 0x02}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Base64URLDecode(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBase64URL_Roundtrip tests Base64URL encode/decode roundtrip
func TestBase64URL_Roundtrip(t *testing.T) {
	testData := [][]byte{
		{},
		[]byte("hello"),
		[]byte("hello world"),
		{0x00, 0x01, 0x02, 0x03},
		make([]byte, 1000),
	}

	for _, data := range testData {
		encoded := Base64URLEncode(data)
		decoded, err := Base64URLDecode(encoded)
		require.NoError(t, err)
		assert.Equal(t, data, decoded)
	}
}

// TestTagsEncode tests tag encoding
func TestTagsEncode(t *testing.T) {
	tags := []schema.Tag{
		{Name: "Action", Value: "Message"},
		{Name: "Type", Value: "Test"},
	}

	encoded, err := TagsEncode(tags)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
}

// TestTagsDecode tests tag decoding
func TestTagsDecode(t *testing.T) {
	original := []schema.Tag{
		{Name: "Action", Value: "Message"},
		{Name: "Type", Value: "Test"},
	}

	encoded, err := TagsEncode(original)
	require.NoError(t, err)

	decoded, err := TagsDecode(encoded)
	require.NoError(t, err)

	assert.Len(t, decoded, len(original))
	for i, tag := range original {
		assert.Equal(t, tag.Name, decoded[i].Name)
		assert.Equal(t, tag.Value, decoded[i].Value)
	}
}

// TestTags_Roundtrip tests tag encode/decode roundtrip
func TestTags_Roundtrip(t *testing.T) {
	testTags := [][]schema.Tag{
		{},
		{{Name: "Action", Value: "Message"}},
		{
			{Name: "Action", Value: "Message"},
			{Name: "Type", Value: "Test"},
			{Name: "Version", Value: "1.0"},
		},
	}

	for _, tags := range testTags {
		encoded, err := TagsEncode(tags)
		require.NoError(t, err)

		decoded, err := TagsDecode(encoded)
		require.NoError(t, err)

		assert.Len(t, decoded, len(tags))
		for i, tag := range tags {
			assert.Equal(t, tag.Name, decoded[i].Name)
			assert.Equal(t, tag.Value, decoded[i].Value)
		}
	}
}

// TestValidateDataItem tests DataItem validation
func TestValidateDataItem(t *testing.T) {
	// Valid DataItem
	valid := &DataItem{
		Owner:     []byte("owner"),
		Signature: []byte("signature"),
		Anchor:    make([]byte, 32),
	}
	err := ValidateDataItem(valid)
	assert.NoError(t, err)

	// Nil DataItem
	err = ValidateDataItem(nil)
	assert.Error(t, err)

	// Missing owner
	noOwner := &DataItem{
		Signature: []byte("signature"),
	}
	err = ValidateDataItem(noOwner)
	assert.Error(t, err)

	// Missing signature
	noSig := &DataItem{
		Owner: []byte("owner"),
	}
	err = ValidateDataItem(noSig)
	assert.Error(t, err)

	// Invalid anchor length
	badAnchor := &DataItem{
		Owner:     []byte("owner"),
		Signature: []byte("signature"),
		Anchor:    []byte("short"),
	}
	err = ValidateDataItem(badAnchor)
	assert.Error(t, err)
}

// TestGetDataItemSize tests DataItem size estimation
func TestGetDataItemSize(t *testing.T) {
	di := &DataItem{
		Owner:     make([]byte, 256),
		Target:    "process-123",
		Anchor:    make([]byte, 32),
		Tags:      []schema.Tag{{Name: "Action", Value: "Message"}},
		Data:      []byte("test data"),
		Signature: make([]byte, 512),
	}

	size := GetDataItemSize(di)
	assert.Greater(t, size, 0)

	// Size should be reasonable (not too small or too large)
	assert.Greater(t, size, 800) // Owner + Signature + overhead
	assert.Less(t, size, 2000)
}

// TestLenVarInt tests varint length calculation
func TestLenVarInt(t *testing.T) {
	tests := []struct {
		value    uint64
		expected int
	}{
		{0, 1},
		{127, 1},
		{128, 2},
		{16383, 2},
		{16384, 3},
	}

	for _, tt := range tests {
		length := lenVarInt(tt.value)
		assert.Equal(t, tt.expected, length)
	}
}