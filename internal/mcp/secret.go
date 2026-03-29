package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/skys-mission/key-agent/internal/storage"
)

func (s *Server) registerSecretTools() {
	// secret_get
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "secret_get",
		Description: "Get a secret from the secret store",
	}, s.secretGet)

	// secret_set
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "secret_set",
		Description: "Set a secret in the secret store",
	}, s.secretSet)

	// secret_delete
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "secret_delete",
		Description: "Delete a secret from the secret store",
	}, s.secretDelete)

	// secret_list
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "secret_list",
		Description: "List all secret keys in the store",
	}, s.secretList)
}

// SecretGetArgs represents arguments for secret_get tool.
type SecretGetArgs struct {
	Key string `json:"key" jsonschema:"required,The secret key to retrieve"`
}

func (s *Server) secretGet(ctx context.Context, req *mcp.CallToolRequest, args SecretGetArgs) (*mcp.CallToolResult, any, error) {
	entry, err := s.store.GetSecret(ctx, args.Key)
	if err != nil {
		if err == storage.ErrNotFound {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Secret not found: %s", args.Key)},
				},
			}, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to get secret: %w", err)
	}

	data, _ := json.MarshalIndent(entry, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

// SecretSetArgs represents arguments for secret_set tool.
type SecretSetArgs struct {
	Key   string `json:"key" jsonschema:"required,The secret key to set"`
	Value string `json:"value" jsonschema:"required,The secret value to store"`
	Type  string `json:"type" jsonschema:"required,The secret type (password, api_key, certificate, private_key, token, other)"`
}

func (s *Server) secretSet(ctx context.Context, req *mcp.CallToolRequest, args SecretSetArgs) (*mcp.CallToolResult, any, error) {
	secretType := storage.SecretType(args.Type)
	validTypes := map[storage.SecretType]bool{
		storage.SecretTypePassword:    true,
		storage.SecretTypeAPIKey:      true,
		storage.SecretTypeCertificate: true,
		storage.SecretTypePrivateKey:  true,
		storage.SecretTypeToken:       true,
		storage.SecretTypeOther:       true,
	}
	if !validTypes[secretType] {
		return nil, nil, fmt.Errorf("invalid secret type: %s", args.Type)
	}

	// Check if entry exists
	existing, _ := s.store.GetSecret(ctx, args.Key)

	entry := &storage.SecretEntry{
		Entry: storage.Entry{
			Key:   args.Key,
			Value: args.Value,
		},
		Type: secretType,
	}
	if existing != nil {
		entry.CreatedAt = existing.CreatedAt
		entry.Version = existing.Version
	}

	if err := s.store.SetSecret(ctx, entry); err != nil {
		return nil, nil, fmt.Errorf("failed to set secret: %w", err)
	}

	// Fetch the updated entry
	entry, _ = s.store.GetSecret(ctx, args.Key)
	data, _ := json.MarshalIndent(entry, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

// SecretDeleteArgs represents arguments for secret_delete tool.
type SecretDeleteArgs struct {
	Key string `json:"key" jsonschema:"required,The secret key to delete"`
}

func (s *Server) secretDelete(ctx context.Context, req *mcp.CallToolRequest, args SecretDeleteArgs) (*mcp.CallToolResult, any, error) {
	err := s.store.DeleteSecret(ctx, args.Key)
	if err != nil {
		if err == storage.ErrNotFound {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Secret not found: %s", args.Key)},
				},
			}, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to delete secret: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Deleted secret: %s", args.Key)},
		},
	}, nil, nil
}

// SecretListArgs represents arguments for secret_list tool.
type SecretListArgs struct {
	Prefix string `json:"prefix" jsonschema:"Optional prefix to filter secret keys"`
}

func (s *Server) secretList(ctx context.Context, req *mcp.CallToolRequest, args SecretListArgs) (*mcp.CallToolResult, any, error) {
	keys, err := s.store.ListSecret(ctx, args.Prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list secrets: %w", err)
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
