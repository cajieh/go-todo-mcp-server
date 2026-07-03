package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cajieh/go-todo-mcp/internal/store"
	"github.com/cajieh/go-todo-mcp/pkg/protocol"
)

func main() {
	// Initialize our safe, optimized memory engine
	todoStore := store.NewTodoStore()

	// Use an input buffer scanner to read stdin line-by-line
	scanner := bufio.NewScanner(os.Stdin)

	// Keep listening to incoming input streams endlessly
	for scanner.Scan() {
		rawLine := scanner.Bytes()

		// Read the base request structure first
		var req protocol.JSONRPCRequest
		if err := json.Unmarshal(rawLine, &req); err != nil {
			sendError(0, -32700, "Parse error: invalid JSON structure")
			continue
		}

		// Route the request based on the MCP method
		switch req.Method {
		case "initialize":
			// Handshake required by Claude Code to establish connection
			sendResponse(protocol.JSONRPCResponse{
				JSONRPC: "2.0",
				Result: map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"capabilities":    map[string]interface{}{},
					"serverInfo":      map[string]string{"name": "go-todo-server", "version": "1.0.0"},
				},
				ID: req.ID,
			})

		case "tools/list":
			handleListTools(req.ID)

		case "tools/call":
			// Safely execute the tool using our database engine
			handleToolCall(req, todoStore)

		default:
			sendError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
		}
	}
}

// handleListTools returns the collection of tools this MCP server provides.
func handleListTools(requestID int64) {
	tools := []map[string]interface{}{
		{
			"name":        "add_todo",
			"description": "Create a new task in your todo list",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]string{"type": "string"},
				},
				"required": []string{"title"},
			},
		},
		{
			"name":        "list_todos",
			"description": "Retrieve all current tasks in your todo list",
		},
	}

	response := protocol.JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  map[string]interface{}{"tools": tools},
		ID:      requestID,
	}
	sendResponse(response)
}

// handleToolCall parses parameters and reads/writes to our storage engine.
func handleToolCall(req protocol.JSONRPCRequest, todoStore *store.TodoStore) {
	var params protocol.ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(req.ID, -32602, "Invalid params mapping")
		return
	}

	switch params.Name {
	case "add_todo":
		// Decode arguments dynamically from the RawMessage block
		var args map[string]string
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			sendError(req.ID, -32602, "Malformed tool arguments")
			return
		}
		
		title := args["title"]
		if title == "" {
			sendError(req.ID, -32602, "Missing required argument: title")
			return
		}

		// Write safely to our storage core using pointer operations
		newTodo := todoStore.Add(title)
		
		sendResponse(protocol.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  newTodo,
			ID:      req.ID,
		})

	case "list_todos":
		// Read safely from our thread-safe storage core
		allTodos := todoStore.List()
		
		sendResponse(protocol.JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  allTodos,
			ID:      req.ID,
		})

	default:
		sendError(req.ID, -32601, fmt.Sprintf("Unknown tool: %s", params.Name))
	}
}

func sendResponse(resp protocol.JSONRPCResponse) {
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func sendError(id int64, code int, message string) {
	errResp := protocol.JSONRPCError{
		JSONRPC: "2.0",
		Error: protocol.ErrorField{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	data, _ := json.Marshal(errResp)
	fmt.Println(string(data))
}
