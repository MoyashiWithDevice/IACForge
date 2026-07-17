package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"IACForge/src/parser"
)

func registerYAMLMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("load_yaml",
			mcp.WithDescription("Load a YAML infrastructure model file and build the in-memory graph."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Path to the YAML file")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			path, err := req.RequireString("path")
			if err != nil {
				return toolError(err.Error()), nil
			}

			p := parser.NewParserWithSchema(sd.Schema)
			g, err := p.ParseFile(path)
			if err != nil {
				return toolError(fmt.Sprintf("failed to parse %s: %v", path, err)), nil
			}
			sd.Graph = g
			return toolResult(fmt.Sprintf("Loaded %d entities and %d relations from %s", len(g.Entities()), len(g.Relations()), path)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("save_yaml",
			mcp.WithDescription("Save the current graph to a YAML file."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Path to write the YAML file")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			path, err := req.RequireString("path")
			if err != nil {
				return toolError(err.Error()), nil
			}

			ser := parser.NewSerializerWithSchema(sd.Schema)
			if err := ser.SerializeFile(sd.Graph, path); err != nil {
				return toolError(fmt.Sprintf("failed to serialize: %v", err)), nil
			}
			return toolResult(fmt.Sprintf("Saved graph to %s", path)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("parse_yaml_string",
			mcp.WithDescription("Parse a YAML string and build the in-memory graph."),
			mcp.WithString("yaml_content", mcp.Required(), mcp.Description("YAML content string")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			content, err := req.RequireString("yaml_content")
			if err != nil {
				return toolError(err.Error()), nil
			}

			p := parser.NewParserWithSchema(sd.Schema)
			g, err := p.Parse([]byte(content))
			if err != nil {
				return toolError(fmt.Sprintf("failed to parse YAML: %v", err)), nil
			}
			sd.Graph = g
			return toolResult(fmt.Sprintf("Parsed %d entities and %d relations", len(g.Entities()), len(g.Relations()))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("serialize_to_string",
			mcp.WithDescription("Serialize the current graph to a YAML string."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			ser := parser.NewSerializerWithSchema(sd.Schema)
			data, err := ser.Serialize(sd.Graph)
			if err != nil {
				return toolError(fmt.Sprintf("failed to serialize: %v", err)), nil
			}
			return toolResult(string(data)), nil
		},
	)
}

func getOrCreateSession(ctx context.Context, sm *SessionManager) *SessionData {
	session := mcpserver.ClientSessionFromContext(ctx)
	if session != nil {
		return sm.GetOrCreate(session.SessionID())
	}
	return sm.GetOrCreate("default")
}
