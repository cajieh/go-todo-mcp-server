# Go Todo MCP Server (`go-todo-mcp`)

A production-grade, highly concurrent Model Context Protocol (MCP) server written in Go. This server integrates seamlessly with **Claude Code** (Anthropic's developer CLI) to manage a todo list using custom-built tools over standard I/O (`stdin`/`stdout`).

This project was built from scratch following standard enterprise Go layout conventions (`cmd/`, `pkg/`, `internal/`) to internalize advanced Go concepts including strict JSON serialization, pointer management, implicit interface contracts, and thread-safe memory management.

## 🏗️ Production Project Architecture

The project adheres to the standard Go ecosystem layout, establishing strong boundaries between execution entry points, private internal logic, and public protocol types:

```text
go-todo-mcp/
├── go.mod                 # Module tracking configuration
├── cmd/
│   └── mcp-server/
│       └── main.go        # Application entry point & stdin/stdout parsing loop
├── internal/
│   └── store/
│       └── store.go       # Thread-safe, in-memory map engine (Private Core)
└── pkg/
    └── protocol/
        └── types.go       # Strict JSON-RPC 2.0 MCP data models (Public Types)

```

---

## 💡 Go Architecture & Core Concepts Applied

### 1. Composite Types & JSON Mapping (Chapter 4)

Unlike dynamic languages, Go requires strict definitions for data stream mutations. The communication shell uses precise **JSON Struct Tags** combined with `json.RawMessage` to defer parsing nested JSON arguments until routing conditions are met:

```go
type JSONRPCRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params,omitempty"`
    ID      int64           `json:"id"`
}

```

### 2. Constructor Functions & Long-lived Allocation (Chapter 2 & 4)

To ensure memory safety and maintain identical state context across goroutines, storage insta forces the underlying struct to escape to the **Heap** for long-term survival, while protecting map storage from fatal `nil` assignment panics using `make()`:

```go
func NewTodoStore() *TodoStore {
    return &TodoStore{
        todos: make(map[string]Todo),
    }
}

```

### 3. Shared Variables & Concurrency Controls (Chapter 6 & 9)

MCP clients fire tool executions simultaneously. To prevent fatal race conditions (concurrent map read/write violations), an encapsulated **Mutual Exclusion Lock (`sync.RWMutex`)** protects memory maps. Read-heavy actions (`List()`) leverage non-blocking shared locks (`RLock`), while mutations (`Add()`) completely serialize access using absolute write locks (`Lock`), safely unwinding via `defer` blocks:

```go
func (s *TodoStore) Add(title string) Todo {
    s.mu.Lock()
    defer s.mu.Unlock()
    // ... thread-safe insertion logic
}

```

### 4. Dynamic Return Flexibilities (Chapter 7)

The server uses Go's empty interface (`interface{}`) dynamic pattern inside response payloads. This provides structural polymorphism, allowing a single response packet to dynamically encode lists of tools, map collections, or individual task elements depending on runtime demands.

---

## 🏃‍♂️ Getting Started (Local Development)

### Prerequisites

* Go 1.20 or higher installed.
* Claude Code CLI installed and authenticated.

### 1. Clone & Build the Server

Navigate to your project directory and build the static executable:

```bash
# Build the binary into a custom root directory
go build -o ./go-todo-mcp cmd/mcp-server/main.go

```

### 2. Manual Test (Happy Path Validation)

You can test the server directly in your terminal using raw JSON-RPC input strings. Run the binary with the **Go Race Detector flag** enabled to verify absolute memory concurrency optimization:

```bash
go run -race cmd/mcp-server/main.go

```

Paste the following JSON block into the terminal and hit `Enter`:

```json
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"add_todo","arguments":{"title":"Master Go Memory Layout"}},"id":1}

```

**Expected Output:**

```json
{"jsonrpc":"2.0","result":{"id":"1","title":"Master Go Memory Layout","done":false},"id":1}

```

---

## 🔌 Connecting to Claude Code CLI

### 1. Register the Server

Tell Claude Code how to execute your local Go backend binary sub-process by using the dedicated `claude mcp` utility. Make sure to replace the path below with your **absolute project folder path**:

```bash
claude mcp add go-todo -- /Users/yourusername/Documents/projects/go-todo-mcp-server/go-todo-mcp

```

*(The `--` flag is required to distinguish your custom binary command path from regular Claude CLI parameters).*

### 2. Verify Connection

Run the following command to ensure the protocol handshake was completed successfully:

```bash
claude mcp list

```

You should see a green verification indicator:

```text
go-todo: /Users/yourusername/.../go-todo-mcp - ✓ Connected

```

### 3. Launch and Chat

Boot up your interactive development session:

```bash
claude

```

You can now interact with your custom Go server tools natively using slash commands or conversational prompts:

* **Add a Todo:** `/add_todo title="Review OpenShift PRs"` or *"Add a task to review OpenShift PRs"*
* **List All Todos:** `/todos` or *"Show me my todo list"*

### 4. Live Debug Streaming

To inspect raw JSON-RPC frames moving back and forth between Claude Code and your Go engine in real-time, run your main session in debug mode:

```bash
claude --debug mcp

```
