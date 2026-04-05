# 🚀 goao

**a  Go SDK for interacting with AO Protocol.**

`goao` provides a comprehensive, type-safe, and efficient way to spawn processes, send messages, read state, and manage encrypted communications within the AO compute network. Built on the same principles as [`goar`](https://github.com/permadao/goar) and compatible with the [`ao-connect`](https://github.com/permaweb/ao/tree/main/connect) TypeScript SDK.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Test Status](https://github.com/permadao/goao/actions/workflows/main.yml/badge.svg)](https://github.com/permadao/goao/actions)

---

## 📦 Installation

```bash
go get github.com/permadao/goao
```

---

## 🏁 Quick Start

```go
package main

import (
    "fmt"
    "github.com/permadao/goao"
    "github.com/permadao/goao/signer"
)

func main() {
    // 1. Load your wallet (RSA JWK or ECDSA Hex)
    s, err := signer.NewRSASignerFromPath("wallet.json")
    if err != nil {
        panic(err)
    }

    // 2. Create Client (defaults to AO Testnet)
    client := goao.NewClient(
        goao.WithSigner(s),
    )

    // 3. Spawn a Process
    processID, err := client.SpawnProcess(
        "fcoN_xGpMhGJ3Kp...", // Module ID
        []byte(`{"initial": "state"}`),
        []goao.Tag{{Name: "App-Name", Value: "MyApp"}},
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("Process Spawned:", processID)

    // 4. Send a Message
    msgID, err := client.SendMessage(
        processID,
        []byte("Hello AO!"),
        []goao.Tag{{Name: "Action", Value: "Ping"}},
        nil,
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("Message Sent:", msgID)

    // 5. Read State (HTTP Compute Endpoint)
    counter, err := client.GetComputeString(processID, "counter")
    if err != nil {
        panic(err)
    }
    fmt.Println("Counter Value:", counter)
}
```

---

## ⚙️ Configuration

`goao` uses a flexible client configuration similar to `ao-connect`. You can override default URLs for MU, CU, SU, and the Compute Gateway.

### Default Endpoints
| Component | Default URL |
|-----------|-------------|
| **Messenger Unit (MU)** | `https://mu.ao-testnet.xyz` |
| **Compute Unit (CU)** | `https://cu.ao-testnet.xyz` |
| **Scheduler Unit (SU)** | `https://su.ao-testnet.xyz` |
| **Compute Gateway** | `https://push.forward.computer` |
| **Arweave Gateway** | `https://arweave.net` |

### Custom Configuration
```go
client := goao.NewClient(
    goao.WithMU("https://custom-mu.example.com"),
    goao.WithCU("https://custom-cu.example.com"),
    goao.WithSU("https://custom-su.example.com"),
    goao.WithComputeGateway("https://custom-compute.example.com"),
    goao.WithSigner(mySigner),
)
```

---

## 🔑 Signers

`goao` supports both **RSA (Arweave)** and **ECDSA (Ethereum)** signers, unified under a single interface.

### RSA Signer (Arweave JWK)
```go
import "github.com/permadao/goao/signer"

// Load from JWK file
s, err := signer.NewRSASignerFromPath("wallet.json")

// Or from raw bytes
s, err := signer.NewRSASigner(jwkBytes)
```

### ECDSA Signer (Ethereum Hex)
```go
// Load from hex private key
s, err := signer.NewECDSASigner("ccb43edab9f7fd5c24388c43579b9d50ad3c8d94cf5fce9bb0bcb8e6ddb57bce")

// Or generate random
s, err := signer.NewECDSASignerRandom()
```

---

## 📬 Core Features

### 1. Spawn Process
Create a new AO process with initial state and tags.
```go
processID, err := client.SpawnProcess(
    "module-tx-id",
    []byte(`{"owner": "address123"}`),
    []goao.Tag{
        {Name: "Module", Value: "module-tx-id"},
        {Name: "Scheduler", Value: "_GQ33BkPtZrqxA84vM8Zk-N2aO0toNNu_C-l-rawrBA"},
    },
)
```

### 2. Send Message
Send a message to a process. Supports optional **encryption**.
```go
// Plain Text
msgID, err := client.SendMessage(processID, []byte("Hello"), tags, nil)

// Encrypted (AES-GCM + RSA-OAEP)
opts := &goao.SendMessageOptions{
    EncryptWithRSA: recipientPublicKey, // *rsa.PublicKey
}
msgID, err := client.SendMessage(processID, secretData, tags, opts)
```

### 3. Read State (HTTP Compute)
**Recommended:** Fetch process state via HTTP Compute Gateway.
```go
// Get raw bytes
data, err := client.GetCompute(processID, "counter")

// Get string
val, err := client.GetComputeString(processID, "status")

// Get JSON
var state map[string]interface{}
err := client.GetComputeJSON(processID, "state", &state)

// Full URL Support (Flexible Gateway)
data, err := client.GetCompute("https://push.forward.computer/"+processID+"~process@1.0/compute/counter", "")
```

### 4. Get Result
Fetch the result of a specific message evaluation from the Compute Unit.
```go
result, err := client.GetResult(messageID, processID)
fmt.Println("Output:", result.Output)
fmt.Println("Gas Used:", result.GasUsed)
```

### 5. List Results
Paginate through message results.
```go
opts := &goao.ListResultsOptions{
    ProcessID: processID,
    From:      "cursor-start",
    Limit:     25,
    Sort:      goao.SortASC,
}
results, err := client.ListResults(opts)
```

---

## 🔐 Encryption

`goao` supports end-to-end encryption for messages using **AES-GCM** (data) and **RSA-OAEP** (key exchange).

### Sending Encrypted Data
```go
import "github.com/permadao/goao/encrypt"

// Load recipient's public key
pubKey := loadRSAPublicKey("recipient_pub.pem")

opts := &goao.SendMessageOptions{
    EncryptWithRSA: pubKey,
}

// Data is encrypted automatically before signing
msgID, err := client.SendMessage(processID, secretData, tags, opts)
```

### Decrypting Responses
```go
// If the process returns an encrypted response
decrypted, err := client.DecryptResponse(
    encryptedData,
    encryptedKey, // From Tags["Encrypted-Key"]
    nonce,        // From Tags["Nonce"]
)
```

### Encryption Flow
```
┌─────────────┐                      ┌─────────────┐
│   SENDER    │                      │  RECIPIENT  │
│   (Client)  │                      │(AO Process) │
└──────┬──────┘                      └──────┬──────┘
       │                                    │
       │ 1. Generate random AES-256 key     │
       │    (ephemeral, per-message)        │
       │                                    │
       │ 2. Encrypt data with AES-GCM       │
       │    (produces ciphertext + tag)     │
       │                                    │
       │ 3. Encrypt AES key with RSA-OAEP   │
       │    (using recipient's PUBLIC key)  │
       │                                    │
       │ 4. Send DataItem:                  │
       │    - Data = encryptedData          │
       │    - Tags["Encrypted-Key"]         │
       │    - Tags["Nonce"]                 │
       │───────────────────────────────────>│
       │                                    │
       │                                    │ 5. Decrypt AES key with
       │                                    │    RSA PRIVATE key
       │                                    │
       │                                    │ 6. Decrypt data with AES-GCM
       │                                    │
       │                                    │ 7. Process plaintext
```

---

## 🕰️ Legacy Support (DryRun)

**Note:** `DryRun` is deprecated in favor of `GetCompute` for simple state reads. However, it is retained for backward compatibility with existing dApps.

```go
// Legacy DryRun (evaluates message without persisting)
result, err := client.DryRun(processID, []byte("return counter"), tags, nil)

// Recommended: Use GetCompute instead
counter, err := client.GetComputeString(processID, "counter")
```

---

## 🧪 Testing

### Run All Tests
```bash
go test -v ./...
```

### Run with Coverage
```bash
go test -v ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run E2E Tests (Requires Testnet)
```bash
go test -v ./... -tags=e2e -timeout=5m
```

### Run Specific Package
```bash
go test -v ./signer/...
go test -v ./encrypt/...
go test -v ./compute/...
go test -v ./dataitem/...
```

---

## 📂 Project Structure

```
📁 goao/
├── 📁 .github/
│   └── 📁 workflows/
│       └── 📄 main.yml              # CI/CD (tests + Protocol Land sync)
├── 📁 schema/
│   ├── 📄 schema.go                 # Types, Constants, Tags
│   └── 📄 schema_test.go            # Schema tests
├── 📁 signer/
│   ├── 📄 signer.go                 # Signer interface
│   ├── 📄 signer_test.go            # Interface tests
│   ├── 📄 rsa.go                    # RSA signer (Arweave JWK)
│   ├── 📄 rsa_test.go               # RSA tests
│   ├── 📄 ecdsa.go                  # ECDSA signer (Ethereum)
│   └── 📄 ecdsa_test.go             # ECDSA tests
├── 📁 dataitem/
│   ├── 📄 dataitem.go               # ANS-104 DataItem struct
│   ├── 📄 dataitem_test.go          # DataItem tests
│   ├── 📄 utils.go                  # Varint, Base64, DeepHash
│   └── 📄 utils_test.go             # Utils tests
├── 📁 encrypt/
│   ├── 📄 encrypt.go                # AES-GCM + RSA-OAEP
│   └── 📄 encrypt_test.go           # Encryption tests
├── 📁 compute/
│   ├── 📄 compute.go                # HTTP Compute Gateway
│   └── 📄 compute_test.go           # Compute tests
├── 📄 client.go                     # HTTP Client (MU, CU, SU)
├── 📄 client_test.go                # Client tests
├── 📄 process.go                    # Spawn, SendMessage, DryRun
├── 📄 process_test.go               # Process tests
├── 📄 result.go                     # GetResult, ListResults
├── 📄 result_test.go                # Result tests
├── 📄 .gitignore                    # Git ignore rules
├── 📄 README.md                     # This file
├── 📄 go.mod                        # Dependencies
├── 📄 go.sum                        # Auto-generated
└── 📄 testKey.json                  # Test RSA key
```

---

## 📊 API Reference

| Method | Description | Returns |
|--------|-------------|---------|
| `SendMessage(process, data, tags, opts)` | Send message to process | `messageID string` |
| `SpawnProcess(module, data, tags)` | Spawn new AO process | `processID string` |
| `GetCompute(processID, fieldPath)` | Fetch state via HTTP | `[]byte` |
| `GetComputeString(processID, fieldPath)` | Fetch state as string | `string` |
| `GetComputeJSON(processID, fieldPath, out)` | Fetch state as JSON | `error` |
| `GetResult(messageID, processID)` | Get message result from CU | `*ResponseCu` |
| `ListResults(opts)` | List results with pagination | `[]*Result` |
| `DryRun(processID, data, tags, anchor)` | Evaluate without persisting (legacy) | `*ResponseCu` |
| `Assign(processID, messageID, exclude, baseLayer)` | Assign L1 tx to process | `string` |
| `Monitor(processID)` | Start cron monitoring | `subscriptionID string` |
| `DecryptResponse(encryptedData, encryptedKey, nonce)` | Decrypt encrypted response | `[]byte` |

---

## 🔧 Environment Variables

Configure AO components via environment variables:

```bash
export AO_MU_URL="https://mu.ao-testnet.xyz"
export AO_CU_URL="https://cu.ao-testnet.xyz"
export AO_SU_URL="https://su.ao-testnet.xyz"
export AO_COMPUTE_GATEWAY="https://push.forward.computer"
export AO_GATEWAY_URL="https://arweave.net"
```

---

## 🤝 Contributing

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/amazing-feature`).
3. Commit your changes (`git commit -m 'Add amazing feature'`).
4. Push to the branch (`git push origin feature/amazing-feature`).
5. Open a Pull Request.

**Testing Requirements:**
- All new features must include unit tests.
- Code coverage must not decrease.
- Run `make test` before submitting PR.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **[ao-connect](https://github.com/permaweb/ao/tree/main/connect)** - The TypeScript SDK that inspired the API design.
- **[goar](https://github.com/permadao/goar)** - The Arweave Go SDK that provided the foundation for signing and HTTP clients.
- **[Permaweb](https://permaweb.io/)** - For building the AO compute network.

---

## 📞 Support

- **GitHub Issues:** [Report bugs or request features](https://github.com/permadao/goao/issues)
- **Discord:** [Permaweb Discord](https://discord.gg/permaweb)
- **Documentation:** [AO Docs](https://docs.ao.arweave.dev/)

---

**Built with ❤️ for the Permaweb ecosystem.**
