package schema

import (
	"encoding/json"
	"testing"
)

// ============================================================================
// Test Constants
// ============================================================================

func TestConstants_AOProtocol(t *testing.T) {
	if DataProtocol != "ao" {
		t.Errorf("Expected DataProtocol 'ao', got '%s'", DataProtocol)
	}
	if Variant != "ao.TN.1" {
		t.Errorf("Expected Variant 'ao.TN.1', got '%s'", Variant)
	}
	if TypeMessage != "Message" {
		t.Errorf("Expected TypeMessage 'Message', got '%s'", TypeMessage)
	}
	if TypeProcess != "Process" {
		t.Errorf("Expected TypeProcess 'Process', got '%s'", TypeProcess)
	}
}

func TestConstants_DefaultIDs(t *testing.T) {
	if DefaultModule == "" {
		t.Error("DefaultModule should not be empty")
	}
	if DefaultSqliteModule == "" {
		t.Error("DefaultSqliteModule should not be empty")
	}
	if DefaultScheduler == "" {
		t.Error("DefaultScheduler should not be empty")
	}
}

func TestConstants_Endpoints(t *testing.T) {
	if DefaultMUURL == "" {
		t.Error("DefaultMUURL should not be empty")
	}
	if DefaultCUURL == "" {
		t.Error("DefaultCUURL should not be empty")
	}
	if DefaultSUURL == "" {
		t.Error("DefaultSUURL should not be empty")
	}
	if DefaultComputeGateway == "" {
		t.Error("DefaultComputeGateway should not be empty")
	}
}

func TestConstants_EncryptionTags(t *testing.T) {
	if TagEncryption != "Encryption" {
		t.Errorf("Expected TagEncryption 'Encryption', got '%s'", TagEncryption)
	}
	if TagEncryptedKey != "Encrypted-Key" {
		t.Errorf("Expected TagEncryptedKey 'Encrypted-Key', got '%s'", TagEncryptedKey)
	}
	if TagNonce != "Nonce" {
		t.Errorf("Expected TagNonce 'Nonce', got '%s'", TagNonce)
	}
	if EncryptionAESGCM != "AES-GCM+RSA-OAEP" {
		t.Errorf("Expected EncryptionAESGCM 'AES-GCM+RSA-OAEP', got '%s'", EncryptionAESGCM)
	}
}

func TestConstants_Version(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
}

// ============================================================================
// Test Struct Marshaling
// ============================================================================

func TestTag_MarshalJSON(t *testing.T) {
	tag := Tag{
		Name:  "Action",
		Value: "Transfer",
	}

	data, err := json.Marshal(tag)
	if err != nil {
		t.Fatalf("Failed to marshal Tag: %v", err)
	}

	var unmarshaled Tag
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Tag: %v", err)
	}

	if unmarshaled.Name != tag.Name {
		t.Errorf("Expected Name '%s', got '%s'", tag.Name, unmarshaled.Name)
	}
	if unmarshaled.Value != tag.Value {
		t.Errorf("Expected Value '%s', got '%s'", tag.Value, unmarshaled.Value)
	}
}

func TestDataItem_MarshalJSON(t *testing.T) {
	dataItem := DataItem{
		ID:        "test-id-123",
		Owner:     []byte("owner-bytes"),
		Target:    "process-id-456",
		Anchor:    []byte("anchor-bytes"),
		Tags:      []Tag{{Name: "Action", Value: "Message"}},
		Data:      []byte("hello ao"),
		Signature: []byte("signature-bytes"),
	}

	data, err := json.Marshal(dataItem)
	if err != nil {
		t.Fatalf("Failed to marshal DataItem: %v", err)
	}

	var unmarshaled DataItem
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal DataItem: %v", err)
	}

	if unmarshaled.ID != dataItem.ID {
		t.Errorf("Expected ID '%s', got '%s'", dataItem.ID, unmarshaled.ID)
	}
	if unmarshaled.Target != dataItem.Target {
		t.Errorf("Expected Target '%s', got '%s'", dataItem.Target, unmarshaled.Target)
	}
	if len(unmarshaled.Tags) != len(dataItem.Tags) {
		t.Errorf("Expected %d tags, got %d", len(dataItem.Tags), len(unmarshaled.Tags))
	}
}

func TestResponseMu_MarshalJSON(t *testing.T) {
	resp := ResponseMu{
		Id:      "msg-id-123",
		Message: "success",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ResponseMu: %v", err)
	}

	var unmarshaled ResponseMu
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ResponseMu: %v", err)
	}

	if unmarshaled.Id != resp.Id {
		t.Errorf("Expected Id '%s', got '%s'", resp.Id, unmarshaled.Id)
	}
	if unmarshaled.Message != resp.Message {
		t.Errorf("Expected Message '%s', got '%s'", resp.Message, unmarshaled.Message)
	}
}

func TestResponseCu_MarshalJSON(t *testing.T) {
	resp := ResponseCu{
		Messages:    []interface{}{"msg1", "msg2"},
		Assignments: []interface{}{},
		Spawns:      []interface{}{"spawn1"},
		Output:      "result output",
		GasUsed:     1000,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ResponseCu: %v", err)
	}

	var unmarshaled ResponseCu
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ResponseCu: %v", err)
	}

	if len(unmarshaled.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(unmarshaled.Messages))
	}
	if unmarshaled.GasUsed != 1000 {
		t.Errorf("Expected GasUsed 1000, got %d", unmarshaled.GasUsed)
	}
}

func TestComputeResponse_MarshalJSON(t *testing.T) {
	resp := ComputeResponse{
		ProcessID: "process-123",
		MessageID: "msg-456",
		Output:    "counter value: 42",
		Error:     "",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ComputeResponse: %v", err)
	}

	var unmarshaled ComputeResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ComputeResponse: %v", err)
	}

	if unmarshaled.ProcessID != resp.ProcessID {
		t.Errorf("Expected ProcessID '%s', got '%s'", resp.ProcessID, unmarshaled.ProcessID)
	}
	if unmarshaled.Output != resp.Output {
		t.Errorf("Expected Output '%v', got '%v'", resp.Output, unmarshaled.Output)
	}
}

func TestEncryptedMessage_MarshalJSON(t *testing.T) {
	msg := EncryptedMessage{
		EncryptedData: "base64url-encoded-data",
		EncryptedKey:  "base64url-encoded-key",
		Nonce:         "base64url-encoded-nonce",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal EncryptedMessage: %v", err)
	}

	var unmarshaled EncryptedMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal EncryptedMessage: %v", err)
	}

	if unmarshaled.EncryptedData != msg.EncryptedData {
		t.Errorf("Expected EncryptedData '%s', got '%s'", msg.EncryptedData, unmarshaled.EncryptedData)
	}
	if unmarshaled.Nonce != msg.Nonce {
		t.Errorf("Expected Nonce '%s', got '%s'", msg.Nonce, unmarshaled.Nonce)
	}
}

func TestSendMessageOptions_MarshalJSON(t *testing.T) {
	opts := SendMessageOptions{
		Process: "process-123",
		Data:    []byte("secret data"),
		Tags:    []Tag{{Name: "Action", Value: "Transfer"}},
		Anchor:  []byte("32-byte-anchor-value-here!"),
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("Failed to marshal SendMessageOptions: %v", err)
	}

	var unmarshaled SendMessageOptions
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SendMessageOptions: %v", err)
	}

	if unmarshaled.Process != opts.Process {
		t.Errorf("Expected Process '%s', got '%s'", opts.Process, unmarshaled.Process)
	}
	if len(unmarshaled.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(unmarshaled.Tags))
	}
}

func TestListResultsOptions_MarshalJSON(t *testing.T) {
	opts := ListResultsOptions{
		ProcessID: "process-123",
		From:      "cursor-start",
		To:        "cursor-end",
		Sort:      SortASC,
		Limit:     25,
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("Failed to marshal ListResultsOptions: %v", err)
	}

	var unmarshaled ListResultsOptions
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ListResultsOptions: %v", err)
	}

	if unmarshaled.ProcessID != opts.ProcessID {
		t.Errorf("Expected ProcessID '%s', got '%s'", opts.ProcessID, unmarshaled.ProcessID)
	}
	if unmarshaled.Sort != SortASC {
		t.Errorf("Expected Sort ASC, got '%s'", unmarshaled.Sort)
	}
	if unmarshaled.Limit != 25 {
		t.Errorf("Expected Limit 25, got %d", unmarshaled.Limit)
	}
}

func TestSortOrder_Values(t *testing.T) {
	if SortASC != "ASC" {
		t.Errorf("Expected SortASC 'ASC', got '%s'", SortASC)
	}
	if SortDESC != "DESC" {
		t.Errorf("Expected SortDESC 'DESC', got '%s'", SortDESC)
	}
}

func TestErrorType_Values(t *testing.T) {
	if ErrorMU != "MU_ERROR" {
		t.Errorf("Expected ErrorMU 'MU_ERROR', got '%s'", ErrorMU)
	}
	if ErrorCU != "CU_ERROR" {
		t.Errorf("Expected ErrorCU 'CU_ERROR', got '%s'", ErrorCU)
	}
	if ErrorSign != "SIGN_ERROR" {
		t.Errorf("Expected ErrorSign 'SIGN_ERROR', got '%s'", ErrorSign)
	}
	if ErrorEncrypt != "ENCRYPT_ERROR" {
		t.Errorf("Expected ErrorEncrypt 'ENCRYPT_ERROR', got '%s'", ErrorEncrypt)
	}
	if ErrorDecrypt != "DECRYPT_ERROR" {
		t.Errorf("Expected ErrorDecrypt 'DECRYPT_ERROR', got '%s'", ErrorDecrypt)
	}
}

// ============================================================================
// Test Helper Functions
// ============================================================================

func TestCronConfig_MarshalJSON(t *testing.T) {
	config := CronConfig{
		Interval: "10-minutes",
		Tags:     []Tag{{Name: "Foo", Value: "bar"}},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal CronConfig: %v", err)
	}

	var unmarshaled CronConfig
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal CronConfig: %v", err)
	}

	if unmarshaled.Interval != config.Interval {
		t.Errorf("Expected Interval '%s', got '%s'", config.Interval, unmarshaled.Interval)
	}
}

func TestSignerResult_MarshalJSON(t *testing.T) {
	result := SignerResult{
		Signature: []byte("signature-bytes"),
		Address:   "arweave-address-123",
		PublicKey: []byte("public-key-bytes"),
		Alg:       "rsa-pss-sha512",
		Type:      1,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal SignerResult: %v", err)
	}

	var unmarshaled SignerResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SignerResult: %v", err)
	}

	if unmarshaled.Address != result.Address {
		t.Errorf("Expected Address '%s', got '%s'", result.Address, unmarshaled.Address)
	}
	if unmarshaled.Alg != result.Alg {
		t.Errorf("Expected Alg '%s', got '%s'", result.Alg, unmarshaled.Alg)
	}
}
