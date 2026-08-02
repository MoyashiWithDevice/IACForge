package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"IACForge/src/core"
	"IACForge/src/query"
)

func registerQueryMCPTools(s *mcpserver.MCPServer, sm *SessionManager) {
	s.AddTool(
		mcp.NewTool("query_entities",
			mcp.WithDescription("Query entities by kind and optional conditions. Conditions are a JSON list of {\"field\",\"operator\",\"value\"}; operators include eq, ne, in, nin, gt, ge, lt, le, contains, starts_with, ends_with, matches, defined, undefined."),
			mcp.WithString("kind", mcp.Description("Entity kind to filter by (e.g. server)")),
			mcp.WithString("where_json", mcp.Description("JSON list of conditions, e.g. [{\"field\":\"status\",\"operator\":\"eq\",\"value\":\"active\"}]")),
			mcp.WithNumber("limit", mcp.Description("Maximum number of results")),
			mcp.WithNumber("offset", mcp.Description("Number of results to skip")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			q := query.NewQuery()
			q.Select = query.NewSelectClause()
			q.Select.AddEntity(core.EntityKind(req.GetString("kind", "")))

			if whereJSON := req.GetString("where_json", ""); whereJSON != "" {
				where, err := parseConditions(whereJSON)
				if err != nil {
					return toolError(fmt.Sprintf("invalid where_json: %v", err)), nil
				}
				q.Where = where
			}

			if limit := req.GetInt("limit", 0); limit > 0 {
				q.Limit = limit
			}
			if offset := req.GetInt("offset", 0); offset > 0 {
				q.Offset = offset
			}

			result, err := query.NewEngine(sd.Graph).Execute(q)
			if err != nil {
				return toolError(fmt.Sprintf("query failed: %v", err)), nil
			}

			return toolResult(string(queryResultsJSON(result))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("query_related",
			mcp.WithDescription("Traverse the graph from an entity. Operations: children, parent, ancestors, descendants, related, sources, targets, outgoing, incoming, reverse_ownership."),
			mcp.WithString("from", mcp.Required(), mcp.Description("Starting entity ID")),
			mcp.WithString("operation", mcp.Required(), mcp.Description("Traversal operation (children, parent, ancestors, descendants, related, sources, targets, outgoing, incoming, reverse_ownership)")),
			mcp.WithString("relation_type", mcp.Description("Optional relation type filter for relation traversals")),
			mcp.WithNumber("depth", mcp.Description("Traversal depth limit")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)

			from, err := req.RequireString("from")
			if err != nil {
				return toolError(err.Error()), nil
			}
			op, err := req.RequireString("operation")
			if err != nil {
				return toolError(err.Error()), nil
			}

			traverse := query.NewTraverseClause(from, query.TraverseOp(op))
			if relType := req.GetString("relation_type", ""); relType != "" {
				traverse.RelationType = core.RelationType(relType)
			}
			if depth := req.GetInt("depth", 0); depth > 0 {
				traverse.SetDepth(depth)
			}

			q := query.NewQuery()
			q.Select = query.NewSelectClause()
			q.Traverse = traverse

			result, err := query.NewEngine(sd.Graph).Execute(q)
			if err != nil {
				return toolError(fmt.Sprintf("traverse failed: %v", err)), nil
			}

			return toolResult(string(queryResultsJSON(result))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("resolve_path",
			mcp.WithDescription("Resolve a path reference (e.g. site/rack/srv/net/eth0) to an entity. Supports direct IDs, ownership paths, and interface references."),
			mcp.WithString("ref", mcp.Required(), mcp.Description("Reference string (ID or path)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			ref, err := req.RequireString("ref")
			if err != nil {
				return toolError(err.Error()), nil
			}

			e, ok := sd.Graph.ResolvePathEntity(ref)
			if !ok {
				return toolError(fmt.Sprintf("reference %q could not be resolved", ref)), nil
			}
			data, _ := json.MarshalIndent(entityToJSONMap(e), "", "  ")
			return toolResult(string(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("who_references",
			mcp.WithDescription("Find all objects that reference an entity: child entities (ownership), relations, and @-prefixed property references."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID to look up")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sd := getOrCreateSession(ctx, sm)
			id, err := req.RequireString("id")
			if err != nil {
				return toolError(err.Error()), nil
			}

			type refSummary struct {
				ID   string `json:"id"`
				Kind string `json:"kind,omitempty"`
			}

			var children []refSummary
			var relations []refSummary
			var propertyRefs []map[string]string

			for _, e := range sd.Graph.Entities() {
				if e.Owner == id {
					children = append(children, refSummary{ID: e.ID, Kind: string(e.Kind)})
				}
				for key, value := range e.Properties {
					if valueReferences(value, id) {
						propertyRefs = append(propertyRefs, map[string]string{"object": e.ID, "property": key})
					}
				}
			}

			for _, r := range sd.Graph.Relations() {
				for _, pid := range r.ParticipantIDs() {
					if pid == id {
						relations = append(relations, refSummary{ID: r.ID, Kind: string(r.Type)})
						break
					}
				}
				for key, value := range r.Properties {
					if valueReferences(value, id) {
						propertyRefs = append(propertyRefs, map[string]string{"object": r.ID, "property": key})
					}
				}
			}

			result := map[string]interface{}{
				"id":                   id,
				"children":             children,
				"relations":            relations,
				"property_references":  propertyRefs,
			}

			data, _ := json.MarshalIndent(result, "", "  ")
			return toolResult(string(data)), nil
		},
	)
}

// parseConditions parses a JSON list of conditions into a query WhereClause.
func parseConditions(jsonStr string) (*query.WhereClause, error) {
	var raw []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, fmt.Errorf("expected a JSON list of conditions: %w", err)
	}

	where := query.NewWhereClause()
	for _, c := range raw {
		field, _ := c["field"].(string)
		if field == "" {
			return nil, fmt.Errorf("condition missing 'field'")
		}
		op, _ := c["operator"].(string)
		if op == "" {
			return nil, fmt.Errorf("condition missing 'operator'")
		}
		where.AddCondition(field, query.Operator(op), c["value"])
	}
	return where, nil
}

// queryResultsJSON serializes query results as JSON, mapping objects to
// snake_case summaries consistent with other MCP tool responses.
func queryResultsJSON(result *query.Result) []byte {
	type resultItem struct {
		ID     string      `json:"id"`
		Type   string      `json:"type"`
		Path   string      `json:"path,omitempty"`
		Object interface{} `json:"object"`
	}

	items := make([]resultItem, 0, len(result.Results))
	for _, item := range result.Results {
		var obj interface{}
		switch o := item.Object.(type) {
		case *core.Entity:
			obj = entityToJSONMap(o)
		case *core.Relation:
			obj = relationToJSONMap(o)
		default:
			obj = o
		}
		items = append(items, resultItem{
			ID:     item.ID,
			Type:   item.Type,
			Path:   item.Path,
			Object: obj,
		})
	}

	resp := map[string]interface{}{
		"count":     result.Count,
		"truncated": result.Truncated,
		"results":   items,
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	return data
}

// relationToJSONMap converts a Relation to a map with snake_case keys.
func relationToJSONMap(r *core.Relation) map[string]interface{} {
	m := map[string]interface{}{
		"id":        r.ID,
		"type":      string(r.Type),
		"direction": string(r.Direction),
	}
	switch r.Direction {
	case core.DirectionSymmetric:
		m["participants"] = r.Participants.List
	default:
		m["source"] = r.Source()
		m["target"] = r.Target()
	}
	if r.Description != "" {
		m["description"] = r.Description
	}
	if r.Status != "" {
		m["status"] = string(r.Status)
	}
	if len(r.Tags) > 0 {
		m["tags"] = r.Tags
	}
	if len(r.Labels) > 0 {
		m["labels"] = r.Labels
	}
	if len(r.Properties) > 0 {
		m["spec"] = r.Properties
	}
	return m
}

// valueReferences reports whether a property value (recursively) references the given ID.
func valueReferences(value interface{}, id string) bool {
	switch v := value.(type) {
	case core.ReferenceValue:
		return v.RefTargetID() == id
	case string:
		return v == "@"+id
	case []interface{}:
		for _, item := range v {
			if valueReferences(item, id) {
				return true
			}
		}
	case map[string]interface{}:
		for _, item := range v {
			if valueReferences(item, id) {
				return true
			}
		}
	}
	return false
}
