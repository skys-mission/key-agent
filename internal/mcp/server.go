// Package mcp provides MCP (Model Context Protocol) server implementation.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/skys-mission/key-agent/internal/storage"
)

// Server provides MCP server functionality.
type Server struct {
	server *mcp.Server
	store  storage.Store
}

// NewServer creates a new MCP server with KV and Secret tools.
func NewServer(store storage.Store, version string) *Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "key-agent",
		Version: version,
	}, nil)

	s := &Server{
		server: server,
		store:  store,
	}

	// Register KV tools
	s.registerKVTools()

	// Register Secret tools
	s.registerSecretTools()

	return s
}

// Handler returns the MCP HTTP handler.
func (s *Server) Handler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.server
	}, nil)
}

// KV Tools

func (s *Server) registerKVTools() {
	// kv_get
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "kv_get",
		Description: "Get a value from the KV store",
	}, s.kvGet)

	// kv_set
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "kv_set",
		Description: "Set a value in the KV store",
	}, s.kvSet)

	// kv_delete
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "kv_delete",
		Description: "Delete a key from the KV store",
	}, s.kvDelete)

	// kv_list
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "kv_list",
		Description: "List all keys in the KV store",
	}, s.kvList)
}

// KVGetArgs represents arguments for kv_get tool.
type KVGetArgs struct {
	Key string `json:"key" jsonschema:"required,The key to retrieve"`
}

func (s *Server) kvGet(ctx context.Context, req *mcp.CallToolRequest, args KVGetArgs) (*mcp.CallToolResult, any, error) {
	entry, err := s.store.GetKV(ctx, args.Key)
	if err != nil {
		if err == storage.ErrNotFound {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Key not found: %s", args.Key)},
				},
			}, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to get key: %w", err)
	}

	data, _ := json.MarshalIndent(entry, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

// KVSetArgs represents arguments for kv_set tool.
type KVSetArgs struct {
	Key   string `json:"key" jsonschema:"required,The key to set"`
	Value string `json:"value" jsonschema:"required,The value to store"`
}

func (s *Server) kvSet(ctx context.Context, req *mcp.CallToolRequest, args KVSetArgs) (*mcp.CallToolResult, any, error) {
	// Check if entry exists
	existing, _ := s.store.GetKV(ctx, args.Key)

	entry := &storage.Entry{
		Key:   args.Key,
		Value: args.Value,
	}
	if existing != nil {
		entry.CreatedAt = existing.CreatedAt
		entry.Version = existing.Version
	}

	if err := s.store.SetKV(ctx, entry); err != nil {
		return nil, nil, fmt.Errorf("failed to set key: %w", err)
	}

	// Fetch the updated entry
	entry, _ = s.store.GetKV(ctx, args.Key)
	data, _ := json.MarshalIndent(entry, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

// KVDeleteArgs represents arguments for kv_delete tool.
type KVDeleteArgs struct {
	Key string `json:"key" jsonschema:"required,The key to delete"`
}

func (s *Server) kvDelete(ctx context.Context, req *mcp.CallToolRequest, args KVDeleteArgs) (*mcp.CallToolResult, any, error) {
	err := s.store.DeleteKV(ctx, args.Key)
	if err != nil {
		if err == storage.ErrNotFound {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Key not found: %s", args.Key)},
				},
			}, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to delete key: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Deleted key: %s", args.Key)},
		},
	}, nil, nil
}

// KVListArgs represents arguments for kv_list tool.
type KVListArgs struct {
	Prefix string `json:"prefix" jsonschema:"Optional prefix to filter keys"`
}

func (s *Server) kvList(ctx context.Context, req *mcp.CallToolRequest, args KVListArgs) (*mcp.CallToolResult, any, error) {
	keys, err := s.store.ListKV(ctx, args.Prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list keys: %w", err)
	}

	if keys == nil {
		keys = []string{}
	}

	data, _ := json.MarshalIndent(map[string]interface{}{"keys": keys}, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}
