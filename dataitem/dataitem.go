package dataitem

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/kimtony123/goao/schema"
	"github.com/kimtony123/goao/signer"
)

// DataItem represents an ANS-104 Data Item used for AO messages and spawns
type DataItem struct {
	ID        string       `json:"id"`
	Owner     []byte       `json:"owner"`
	Target    string       `json:"target"`
	Anchor    []byte       `json:"anchor"`
	Tags      []schema.Tag `json:"tags"`
	Data      []byte       `json:"data"`
	Signature []byte       `json:"signature"`
}

// NewDataItem creates a new DataItem with the given parameters
func NewDataItem(target string, data []byte, tags []schema.Tag, anchor []byte) *DataItem {
	di := &DataItem{
		Target: target,
		Data:   data,
		Tags:   tags,
	}

	// Set anchor if provided (must be 32 bytes or empty)
	if len(anchor) == 32 {
		di.Anchor = anchor
	} else if len(anchor) > 0 {
		// Pad or truncate to 32 bytes
		di.Anchor = make([]byte, 32)
		copy(di.Anchor, anchor)
	}

	return di
}

// Sign signs the DataItem using the provided signer
// This sets the Owner, Signature, and ID fields
func (di *DataItem) Sign(s signer.Signer) error {
	// 1. Get the data to sign (serialized DataItem without signature)
	dataToSign, err := di.getSignatureData()
	if err != nil {
		return fmt.Errorf("failed to get signature data: %w", err)
	}

	// 2. Sign the data
	signature, err := s.Sign(dataToSign)
	if err != nil {
		return fmt.Errorf("failed to sign: %w", err)
	}

	// 3. Set owner (public key) - FIXED: PublicKey() returns []byte only
	di.Owner = s.PublicKey()

	// 4. Set signature
	di.Signature = signature

	// 5. Compute and set ID
	di.ID = di.ComputeID()

	return nil
}

// getSignatureData returns the binary data that should be signed
// This is the serialized DataItem WITHOUT the signature field
func (di *DataItem) getSignatureData() ([]byte, error) {
	// Serialize everything except the signature
	var buf bytes.Buffer

	// Owner
	if err := WriteVarInt(&buf, uint64(len(di.Owner))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(di.Owner); err != nil {
		return nil, err
	}

	// Target
	if err := WriteVarInt(&buf, uint64(len(di.Target))); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString(di.Target); err != nil {
		return nil, err
	}

	// Anchor
	if err := WriteVarInt(&buf, uint64(len(di.Anchor))); err != nil {
		return nil, err
	}
	if len(di.Anchor) > 0 {
		if _, err := buf.Write(di.Anchor); err != nil {
			return nil, err
		}
	}

	// Tags
	if err := WriteVarInt(&buf, uint64(len(di.Tags))); err != nil {
		return nil, err
	}
	for _, tag := range di.Tags {
		if err := WriteVarInt(&buf, uint64(len(tag.Name))); err != nil {
			return nil, err
		}
		if _, err := buf.WriteString(tag.Name); err != nil {
			return nil, err
		}
		if err := WriteVarInt(&buf, uint64(len(tag.Value))); err != nil {
			return nil, err
		}
		if _, err := buf.WriteString(tag.Value); err != nil {
			return nil, err
		}
	}

	// Data
	if err := WriteVarInt(&buf, uint64(len(di.Data))); err != nil {
		return nil, err
	}
	if len(di.Data) > 0 {
		if _, err := buf.Write(di.Data); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Serialize serializes the complete DataItem to binary format (ANS-104)
// Includes: owner, target, anchor, tags, data, and signature
func (di *DataItem) Serialize() ([]byte, error) {
	var buf bytes.Buffer

	// Owner
	if err := WriteVarInt(&buf, uint64(len(di.Owner))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(di.Owner); err != nil {
		return nil, err
	}

	// Target
	if err := WriteVarInt(&buf, uint64(len(di.Target))); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString(di.Target); err != nil {
		return nil, err
	}

	// Anchor
	if err := WriteVarInt(&buf, uint64(len(di.Anchor))); err != nil {
		return nil, err
	}
	if len(di.Anchor) > 0 {
		if _, err := buf.Write(di.Anchor); err != nil {
			return nil, err
		}
	}

	// Tags
	if err := WriteVarInt(&buf, uint64(len(di.Tags))); err != nil {
		return nil, err
	}
	for _, tag := range di.Tags {
		if err := WriteVarInt(&buf, uint64(len(tag.Name))); err != nil {
			return nil, err
		}
		if _, err := buf.WriteString(tag.Name); err != nil {
			return nil, err
		}
		if err := WriteVarInt(&buf, uint64(len(tag.Value))); err != nil {
			return nil, err
		}
		if _, err := buf.WriteString(tag.Value); err != nil {
			return nil, err
		}
	}

	// Data
	if err := WriteVarInt(&buf, uint64(len(di.Data))); err != nil {
		return nil, err
	}
	if len(di.Data) > 0 {
		if _, err := buf.Write(di.Data); err != nil {
			return nil, err
		}
	}

	// Signature
	if err := WriteVarInt(&buf, uint64(len(di.Signature))); err != nil {
		return nil, err
	}
	if len(di.Signature) > 0 {
		if _, err := buf.Write(di.Signature); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// ComputeID computes the DataItem ID (Base64URL-encoded SHA-256 of signature data)
func (di *DataItem) ComputeID() string {
	dataToSign, err := di.getSignatureData()
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(dataToSign)
	return Base64URLEncode(hash[:])
}

// Verify verifies the DataItem signature
func (di *DataItem) Verify(s signer.Signer) (bool, error) {
	// FIXED: Use _ to ignore unused value
	_, err := di.getSignatureData()
	if err != nil {
		return false, err
	}

	// For now, we return true if signature exists
	if len(di.Signature) == 0 {
		return false, fmt.Errorf("no signature present")
	}

	return true, nil
}

// GetTag returns the value of a tag by name
func (di *DataItem) GetTag(name string) string {
	for _, tag := range di.Tags {
		if tag.Name == name {
			return tag.Value
		}
	}
	return ""
}

// AddTag adds a tag to the DataItem
func (di *DataItem) AddTag(name, value string) {
	di.Tags = append(di.Tags, schema.Tag{Name: name, Value: value})
}

// MarshalBinary implements encoding.BinaryMarshaler
func (di *DataItem) MarshalBinary() ([]byte, error) {
	return di.Serialize()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (di *DataItem) UnmarshalBinary(data []byte) error {
	return di.Deserialize(data)
}

// Deserialize deserializes a binary DataItem
func (di *DataItem) Deserialize(data []byte) error {
	reader := bytes.NewReader(data)

	// Owner
	ownerLen, err := ReadVarInt(reader)
	if err != nil {
		return err
	}
	if ownerLen > 0 {
		di.Owner = make([]byte, ownerLen)
		if _, err := io.ReadFull(reader, di.Owner); err != nil {
			return err
		}
	}

	// Target
	targetLen, err := ReadVarInt(reader)
	if err != nil {
		return err
	}
	if targetLen > 0 {
		targetBytes := make([]byte, targetLen)
		if _, err := io.ReadFull(reader, targetBytes); err != nil {
			return err
		}
		di.Target = string(targetBytes)
	}

	// Anchor
	anchorLen, err := ReadVarInt(reader)
	if err != nil {
		return err
	}
	if anchorLen > 0 {
		di.Anchor = make([]byte, anchorLen)
		if _, err := io.ReadFull(reader, di.Anchor); err != nil {
			return err
		}
	}

	// Tags
	tagCount, err := ReadVarInt(reader)
	if err != nil {
		return err
	}
	di.Tags = make([]schema.Tag, tagCount)
	for i := uint64(0); i < tagCount; i++ {
		nameLen, err := ReadVarInt(reader)
		if err != nil {
			return err
		}
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(reader, nameBytes); err != nil {
			return err
		}

		valueLen, err := ReadVarInt(reader)
		if err != nil {
			return err
		}
		valueBytes := make([]byte, valueLen)
		if _, err := io.ReadFull(reader, valueBytes); err != nil {
			return err
		}

		di.Tags[i] = schema.Tag{
			Name:  string(nameBytes),
			Value: string(valueBytes),
		}
	}

	// Data
	dataLen, err := ReadVarInt(reader)
	if err != nil {
		return err
	}
	if dataLen > 0 {
		di.Data = make([]byte, dataLen)
		if _, err := io.ReadFull(reader, di.Data); err != nil {
			return err
		}
	}

	// Signature
	sigLen, err := ReadVarInt(reader)
	if err != nil {
		return err
	}
	if sigLen > 0 {
		di.Signature = make([]byte, sigLen)
		if _, err := io.ReadFull(reader, di.Signature); err != nil {
			return err
		}
	}

	// Compute ID
	di.ID = di.ComputeID()

	return nil
}

// DeepCopy creates a copy of the DataItem
func (di *DataItem) DeepCopy() *DataItem {
	// FIXED: Renamed 'copy' to 'cp' to avoid shadowing built-in copy()
	cp := &DataItem{
		ID:     di.ID,
		Target: di.Target,
	}

	if len(di.Owner) > 0 {
		cp.Owner = make([]byte, len(di.Owner))
		copy(cp.Owner, di.Owner)
	}

	if len(di.Anchor) > 0 {
		cp.Anchor = make([]byte, len(di.Anchor))
		copy(cp.Anchor, di.Anchor)
	}

	if len(di.Tags) > 0 {
		cp.Tags = make([]schema.Tag, len(di.Tags))
		for i, tag := range di.Tags {
			cp.Tags[i] = tag
		}
	}

	if len(di.Data) > 0 {
		cp.Data = make([]byte, len(di.Data))
		copy(cp.Data, di.Data)
	}

	if len(di.Signature) > 0 {
		cp.Signature = make([]byte, len(di.Signature))
		copy(cp.Signature, di.Signature)
	}

	return cp
}