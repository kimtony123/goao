package schema


import (
    "crypto/rsa"  // ✅ ADD THIS
)

// ============================================================================
// AO Protocol Constants
// ============================================================================
const (
	DataProtocol = "ao"
	Variant      = "ao.TN.1"
	TypeMessage  = "Message"
	TypeProcess  = "Process"
	SDK          = "goao"
)

// ============================================================================
// Default Module & Scheduler IDs
// ============================================================================
const (
	DefaultModule       = "xT0ogTeagEGuySbKuUoo_NaWeeBv1fZ4MqgDdKVKY0U"
	DefaultSqliteModule = "sFNHeYzhHfP9vV9CPpqZMU-4Zzq_qKGKwlwMZozWi2Y"
	DefaultScheduler    = "_GQ33BkPtZrqxA84vM8Zk-N2aO0toNNu_C-l-rawrBA"
)

// ============================================================================
// AO Unit Endpoints (Default Testnet)
// ============================================================================
const (
	DefaultMUURL          = "https://mu.ao-testnet.xyz"
	DefaultCUURL          = "https://cu.ao-testnet.xyz"
	DefaultSUURL          = "https://su.ao-testnet.xyz"
	DefaultGatewayURL     = "https://arweave.net"
	DefaultGraphQLURL     = "https://arweave.net/graphql"
	DefaultComputeGateway = "https://push.forward.computer" // HTTP Compute Endpoint
)

// ============================================================================
// Encryption Tag Constants (AES-GCM + RSA-OAEP)
// ============================================================================
const (
	TagEncryption      = "Encryption"
	TagEncryptedKey    = "Encrypted-Key"
	TagNonce           = "Nonce"
	TagContentType     = "Content-Type"
	TagEncryptResponse = "Encrypt-Response"
	EncryptionAESGCM   = "AES-GCM+RSA-OAEP"
	ContentTypeJSON    = "application/json"
	ContentTypeBinary  = "application/octet-stream"
)

// ============================================================================
// AO Action Tags
// ============================================================================
const (
	ActionSpawn   = "Spawn"
	ActionMessage = "Message"
	ActionEval    = "Eval"
	ActionBalance = "Balance"
)

// ============================================================================
// Cron Tag Prefixes
// ============================================================================
const (
	TagCronInterval  = "Cron-Interval"
	TagCronTagPrefix = "Cron-Tag-"
)

// ============================================================================
// HTTP Methods
// ============================================================================
const (
	HTTPMethodGet    = "GET"
	HTTPMethodPost   = "POST"
	HTTPMethodPut    = "PUT"
	HTTPMethodDelete = "DELETE"
)

// ============================================================================
// Sort Order for Results Listing
// ============================================================================
type SortOrder string

const (
	SortASC  SortOrder = "ASC"
	SortDESC SortOrder = "DESC"
)

// ============================================================================
// Tag - Name/Value pair for AO messages
// ============================================================================
type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ============================================================================
// DataItem - ANS-104 Data Item Structure
// ============================================================================
type DataItem struct {
	ID        string `json:"id"`
	Owner     []byte `json:"owner"`
	Target    string `json:"target"`
	Anchor    []byte `json:"anchor"`
	Tags      []Tag  `json:"tags"`
	Data      []byte `json:"data"`
	Signature []byte `json:"signature"`
}

// ============================================================================
// ResponseMu - Messenger Unit Response
// ============================================================================
type ResponseMu struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

// ============================================================================
// ResponseCu - Compute Unit Response (Message Evaluation Result)
// ============================================================================
type ResponseCu struct {
	Messages    []interface{} `json:"Messages"`
	Assignments []interface{} `json:"Assignments"`
	Spawns      []interface{} `json:"Spawns"`
	Output      interface{}   `json:"Output"`
	Error       interface{}   `json:"Error"`
	GasUsed     int64         `json:"GasUsed"`
}

// ============================================================================
// ComputeResponse - HTTP Compute Endpoint Response (NEW)
// Replaces deprecated DryRun for simple state reads
// ============================================================================
type ComputeResponse struct {
	ProcessID string      `json:"processId"`
	MessageID string      `json:"messageId"`
	Output    interface{} `json:"output"`
	Messages  []Message   `json:"messages"`
	Spawns    []Spawn     `json:"spawns"`
	Error     string      `json:"error,omitempty"`
}

// ============================================================================
// Message - AO Message Structure
// ============================================================================
type Message struct {
	ID        string `json:"Id"`
	Owner     string `json:"Owner"`
	Target    string `json:"Target"`
	Anchor    string `json:"Anchor"`
	Tags      []Tag  `json:"Tags"`
	Data      interface{} `json:"Data"`
	Timestamp int64  `json:"Timestamp"`
	Hash      string `json:"Hash"`
}

// ============================================================================
// Spawn - AO Process Spawn Structure
// ============================================================================
type Spawn struct {
	ID    string `json:"Id"`
	Owner string `json:"Owner"`
	Tags  []Tag  `json:"Tags"`
}

// ============================================================================
// Result - Simplified Result Structure for GetResult/ListResults
// ============================================================================
type Result struct {
	MessageID string      `json:"messageId"`
	ProcessID string      `json:"processId"`
	Owner     string      `json:"owner"`
	Target    string      `json:"target"`
	Tags      []Tag       `json:"tags"`
	Data      interface{} `json:"data"`
	Output    interface{} `json:"output"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
	GasUsed   int64       `json:"gasUsed"`
}

// ============================================================================
// ListResultsOptions - Pagination options for ListResults
// ============================================================================
type ListResultsOptions struct {
	ProcessID string    `json:"process"`
	From      string    `json:"from,omitempty"`
	To        string    `json:"to,omitempty"`
	Sort      SortOrder `json:"sort,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// ============================================================================
// SpawnOptions - Options for spawning a process
// ============================================================================
type SpawnOptions struct {
	Module    string `json:"module"`
	Scheduler string `json:"scheduler,omitempty"`
	Data      []byte `json:"data,omitempty"`
	Tags      []Tag  `json:"tags,omitempty"`
}

// ============================================================================
// SendMessageOptions - Options for sending a message
// ============================================================================
// Update SendMessageOptions in schema/schema.go

type SendMessageOptions struct {
    Process        string          `json:"process"`
    Data           []byte          `json:"data"`
    Tags           []Tag           `json:"tags,omitempty"`
    Anchor         []byte          `json:"anchor,omitempty"`
    EncryptWithRSA *rsa.PublicKey  `json:"encryptWithRSA,omitempty"` // ✅ Already correct
    GatewayURL     string          `json:"gatewayURL,omitempty"`     // 🆕 ADD: Allow per-message gateway override
}



// ============================================================================
// DryRunOptions - Legacy: Use HTTP Compute Endpoint Instead
// ============================================================================
// DEPRECATED: DryRun is deprecated in favor of GetCompute() for simple state reads.
// However, DryRun is still supported for complex message evaluations.
type DryRunOptions struct {
	Process string `json:"process"`
	Data    []byte `json:"data"`
	Tags    []Tag  `json:"tags,omitempty"`
	Anchor  []byte `json:"anchor,omitempty"`
}

// ============================================================================
// AssignOptions - Options for assigning a message to a process
// ============================================================================
type AssignOptions struct {
	Process   string   `json:"process"`
	Message   string   `json:"message"`
	Exclude   []string `json:"exclude,omitempty"`
	BaseLayer bool     `json:"baseLayer,omitempty"`
}

// ============================================================================
// MonitorOptions - Options for monitoring cron messages
// ============================================================================
type MonitorOptions struct {
	Process string `json:"process"`
}

// ============================================================================
// CronConfig - Cron job configuration for serializeCron helper
// ============================================================================
type CronConfig struct {
	Interval string `json:"interval"`
	Tags     []Tag  `json:"tags,omitempty"`
}

// ============================================================================
// SignerResult - Result from signer interface
// ============================================================================
type SignerResult struct {
	Signature []byte `json:"signature"`
	Address   string `json:"address"`
	PublicKey []byte `json:"publicKey"`
	Alg       string `json:"alg,omitempty"`
	Type      int    `json:"type,omitempty"`
}

// ============================================================================
// KeyPublicMeta - Public key metadata for signing
// ============================================================================
type KeyPublicMeta struct {
	PublicKey []byte `json:"publicKey"`
	Alg       string `json:"alg,omitempty"`
	Type      int    `json:"type,omitempty"`
}

// ============================================================================
// EncryptedMessage - Structure for encrypted message payloads
// ============================================================================
type EncryptedMessage struct {
	EncryptedData string `json:"encryptedData"`
	EncryptedKey  string `json:"encryptedKey"`
	Nonce         string `json:"nonce"`
}

// ============================================================================
// Error Types
// ============================================================================
type ErrorType string

const (
	ErrorMU       ErrorType = "MU_ERROR"
	ErrorCU       ErrorType = "CU_ERROR"
	ErrorSU       ErrorType = "SU_ERROR"
	ErrorSign     ErrorType = "SIGN_ERROR"
	ErrorEncrypt  ErrorType = "ENCRYPT_ERROR"
	ErrorDecrypt  ErrorType = "DECRYPT_ERROR"
	ErrorSerialize ErrorType = "SERIALIZE_ERROR"
)

// ============================================================================
// SDK Version
// ============================================================================
const (
	Version   = "0.2.0"
	UserAgent = "goao/" + Version
)


const (
    // ComputeEndpointPattern is the URL pattern for HTTP Compute Gateway
    // Format: {gateway}/{process-id}~process@1.0/compute/{field-path}
    // Example: https://push.forward.computer/abc123~process@1.0/compute/counter
    ComputeEndpointPattern = "%s/%s~process@1.0/compute/%s"

    // ProcessSuffix is the standard suffix for AO process identifiers
    ProcessSuffix = "~process@1.0"
)


