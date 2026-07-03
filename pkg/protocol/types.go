package protocol

import "encoding/json"

// JSONRPCRequest represents a standard incoming MCP request frame.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      int64           `json:"id"`
}

// ToolCallParams maps specifically to the params block when Method is "tools/call".
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// JSONRPCResponse represents the standard success frame sent back to the client.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"` // Flexibilty: Define the result type as Interface since it can be in any structure
	ID      int64       `json:"id"`
}

// JSONRPCError represents the error frame if something goes wrong (Chapter 5).
type JSONRPCError struct {
	JSONRPC string     `json:"jsonrpc"`
	Error   ErrorField `json:"error"`
	ID      int64      `json:"id"`
}

type ErrorField struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
