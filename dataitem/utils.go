package dataitem

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"sort"

	"github.com/ar-aostore/goao/schema"
)

// WriteVarInt writes a variable-length integer to the writer
// Uses the same encoding as Arweave/ANS-104
func WriteVarInt(w io.Writer, val uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, val)
	_, err := w.Write(buf[:n])
	return err
}

// ReadVarInt reads a variable-length integer from the reader
// FIXED: Changed from io.Reader to io.ByteReader
func ReadVarInt(r io.ByteReader) (uint64, error) {
	return binary.ReadUvarint(r)
}

// Base64URLEncode encodes bytes to Base64URL without padding
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

// TagsEncode encodes tags to binary format
func TagsEncode(tags []schema.Tag) ([]byte, error) {
	var buf []byte

	// Sort tags by name for consistent encoding
	sortedTags := make([]schema.Tag, len(tags))
	copy(sortedTags, tags)
	sort.Slice(sortedTags, func(i, j int) bool {
		return sortedTags[i].Name < sortedTags[j].Name
	})

	for _, tag := range sortedTags {
		nameBytes := []byte(tag.Name)
		valueBytes := []byte(tag.Value)

		// Encode name length + name
		nameLenBuf := make([]byte, binary.MaxVarintLen64)
		nameLen := binary.PutUvarint(nameLenBuf, uint64(len(nameBytes)))
		buf = append(buf, nameLenBuf[:nameLen]...)
		buf = append(buf, nameBytes...)

		// Encode value length + value
		valueLenBuf := make([]byte, binary.MaxVarintLen64)
		valueLen := binary.PutUvarint(valueLenBuf, uint64(len(valueBytes)))
		buf = append(buf, valueLenBuf[:valueLen]...)
		buf = append(buf, valueBytes...)
	}

	return buf, nil
}

// TagsDecode decodes binary tags to Tag slice
func TagsDecode(data []byte) ([]schema.Tag, error) {
	var tags []schema.Tag
	reader := bytes.NewReader(data)

	for reader.Len() > 0 {
		// Read name length
		nameLen, err := binary.ReadUvarint(reader)
		if err != nil {
			break
		}

		// Read name
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(reader, nameBytes); err != nil {
			return nil, err
		}

		// Read value length
		valueLen, err := binary.ReadUvarint(reader)
		if err != nil {
			return nil, err
		}

		// Read value
		valueBytes := make([]byte, valueLen)
		if _, err := io.ReadFull(reader, valueBytes); err != nil {
			return nil, err
		}

		tags = append(tags, schema.Tag{
			Name:  string(nameBytes),
			Value: string(valueBytes),
		})
	}

	return tags, nil
}

// GenerateAnchor generates a random 32-byte anchor
func GenerateAnchor() ([]byte, error) {
	anchor := make([]byte, 32)
	_, err := io.ReadFull(bytes.NewReader(make([]byte, 32)), anchor)
	if err != nil {
		return nil, err
	}
	return anchor, nil
}

// ValidateDataItem validates a DataItem structure
func ValidateDataItem(di *DataItem) error {
	if di == nil {
		return ErrDataItemNil
	}

	if len(di.Owner) == 0 {
		return ErrOwnerRequired
	}

	if len(di.Signature) == 0 {
		return ErrSignatureRequired
	}

	// Validate anchor length if present
	if len(di.Anchor) > 0 && len(di.Anchor) != 32 {
		return ErrInvalidAnchorLength
	}

	return nil
}

// GetDataItemSize estimates the size of a serialized DataItem
func GetDataItemSize(di *DataItem) int {
	size := 0

	// Owner (length + bytes)
	size += lenVarInt(uint64(len(di.Owner))) + len(di.Owner)

	// Target (length + bytes)
	size += lenVarInt(uint64(len(di.Target))) + len(di.Target)

	// Anchor (length + bytes)
	size += lenVarInt(uint64(len(di.Anchor))) + len(di.Anchor)

	// Tags
	size += lenVarInt(uint64(len(di.Tags)))
	for _, tag := range di.Tags {
		size += lenVarInt(uint64(len(tag.Name))) + len(tag.Name)
		size += lenVarInt(uint64(len(tag.Value))) + len(tag.Value)
	}

	// Data (length + bytes)
	size += lenVarInt(uint64(len(di.Data))) + len(di.Data)

	// Signature (length + bytes)
	size += lenVarInt(uint64(len(di.Signature))) + len(di.Signature)

	return size
}

// lenVarInt returns the number of bytes needed to encode a varint
func lenVarInt(val uint64) int {
	buf := make([]byte, binary.MaxVarintLen64)
	return binary.PutUvarint(buf, val)
}

// Error types for DataItem validation
var (
	ErrDataItemNil       = &DataItemError{"dataitem is nil"}
	ErrOwnerRequired     = &DataItemError{"owner is required"}
	ErrSignatureRequired = &DataItemError{"signature is required"}
	ErrInvalidAnchorLength = &DataItemError{"anchor must be 32 bytes or empty"}
)

// DataItemError represents a DataItem validation error
type DataItemError struct {
	Message string
}

func (e *DataItemError) Error() string {
	return e.Message
}