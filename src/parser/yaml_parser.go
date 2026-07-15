package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"IACForge/src/core"
)

// ParseError represents a parsing error with location information.
type ParseError struct {
	Line    int
	Column  int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// Parser parses YAML syntax into a Graph.
type Parser struct {
	graph *core.Graph
}

// NewParser creates a new YAML parser.
func NewParser() *Parser {
	return &Parser{
		graph: core.NewGraph(),
	}
}

// ParseFile reads a YAML file and parses it into a Graph.
func (p *Parser) ParseFile(path string) (*core.Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return p.Parse(data)
}

// Parse parses YAML bytes into a Graph.
func (p *Parser) Parse(data []byte) (*core.Graph, error) {
	p.graph = core.NewGraph()

	var doc struct {
		Objects []map[string]interface{} `yaml:"objects"`
	}

	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := p.parseObjects(doc.Objects); err != nil {
		return nil, err
	}

	if err := p.graph.BuildOwnershipPaths(); err != nil {
		return nil, fmt.Errorf("failed to build ownership paths: %w", err)
	}

	return p.graph, nil
}

// ParseReader parses YAML from a reader into a Graph.
func (p *Parser) ParseReader(r io.Reader) (*core.Graph, error) {
	var doc struct {
		Objects []map[string]interface{} `yaml:"objects"`
	}

	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	p.graph = core.NewGraph()

	if err := p.parseObjects(doc.Objects); err != nil {
		return nil, err
	}

	if err := p.graph.BuildOwnershipPaths(); err != nil {
		return nil, fmt.Errorf("failed to build ownership paths: %w", err)
	}

	return p.graph, nil
}

// ParseDir reads all YAML files from a directory recursively and merges them into a single Graph.
// File names are sorted to ensure deterministic loading order.
func (p *Parser) ParseDir(dir string) (*core.Graph, error) {
	p.graph = core.NewGraph()

	var allObjects []map[string]interface{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var doc struct {
			Objects []map[string]interface{} `yaml:"objects"`
		}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		allObjects = append(allObjects, doc.Objects...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	if err := p.parseObjects(allObjects); err != nil {
		return nil, err
	}

	if err := p.graph.BuildOwnershipPaths(); err != nil {
		return nil, fmt.Errorf("failed to build ownership paths: %w", err)
	}

	return p.graph, nil
}

// Load is a convenience function that loads a path (file or directory) into a Graph.
func (p *Parser) Load(path string) (*core.Graph, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access %s: %w", path, err)
	}
	if info.IsDir() {
		return p.ParseDir(path)
	}
	return p.ParseFile(path)
}

// parseObjects processes a list of object definitions.
func (p *Parser) parseObjects(objects []map[string]interface{}) error {
	// First pass: parse all entities
	entities := make(map[string]*core.Entity)
	var relations []*rawRelation

	for i, obj := range objects {
		if obj == nil {
			continue
		}

		if _, hasKind := obj["kind"]; hasKind {
			entity, err := p.parseEntity(obj)
			if err != nil {
				return fmt.Errorf("object at index %d: %w", i, err)
			}
			entities[entity.ID] = entity
		} else if _, hasType := obj["type"]; hasType {
			relations = append(relations, &rawRelation{index: i, data: obj})
		}
	}

	// Add entities to graph
	for _, entity := range entities {
		if err := p.graph.AddEntity(entity); err != nil {
			return fmt.Errorf("failed to add entity %s: %w", entity.ID, err)
		}
	}

	// Second pass: parse relations (need entities to be added first for reference validation)
	for _, raw := range relations {
		relation, err := p.parseRelation(raw.data, entities)
		if err != nil {
			return fmt.Errorf("object at index %d: %w", raw.index, err)
		}
		if err := p.graph.AddRelation(relation); err != nil {
			return fmt.Errorf("failed to add relation %s: %w", relation.ID, err)
		}
	}

	return nil
}

// rawRelation holds a relation definition before parsing.
type rawRelation struct {
	index int
	data  map[string]interface{}
}

// parseEntity parses an entity from its YAML representation.
func (p *Parser) parseEntity(obj map[string]interface{}) (*core.Entity, error) {
	// Required fields at top level
	id, err := getString(obj, "id")
	if err != nil {
		return nil, fmt.Errorf("entity missing required field 'id': %w", err)
	}

	kindStr, err := getString(obj, "kind")
	if err != nil {
		return nil, fmt.Errorf("entity missing required field 'kind': %w", err)
	}

	name, err := getString(obj, "name")
	if err != nil {
		return nil, fmt.Errorf("entity missing required field 'name': %w", err)
	}

	entity := core.NewEntity(id, core.EntityKind(kindStr), name)

	// Parse attributes sub-key
	if attrs, ok := getMapOptional(obj, "attributes"); ok {
		if owner, ok := getStringOptional(attrs, "owner"); ok {
			entity.SetOwner(owner)
		}
		if description, ok := getStringOptional(attrs, "description"); ok {
			entity.Description = description
		}
		if statusStr, ok := getStringOptional(attrs, "status"); ok {
			entity.SetStatus(core.Status(statusStr))
		}
		if tags, ok := getSliceOptional(attrs, "tags"); ok {
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					entity.AddTag(tagStr)
				}
			}
		}
		if labels, ok := getMapOptional(attrs, "labels"); ok {
			for k, v := range labels {
				if vStr, ok := v.(string); ok {
					entity.SetLabel(k, vStr)
				}
			}
		}
		if extensions, ok := getMapOptional(attrs, "extensions"); ok {
			entity.Extensions = extensions
		}
	}

	// Parse spec sub-key for kind-specific properties
	if spec, ok := getMapOptional(obj, "spec"); ok {
		for k, v := range spec {
			entity.SetProperty(k, v)
		}
	}

	return entity, nil
}

// parseRelation parses a relation from its YAML representation.
func (p *Parser) parseRelation(obj map[string]interface{}, entities map[string]*core.Entity) (*core.Relation, error) {
	// Required fields at top level
	id, err := getString(obj, "id")
	if err != nil {
		return nil, fmt.Errorf("relation missing required field 'id': %w", err)
	}

	typeStr, err := getString(obj, "type")
	if err != nil {
		return nil, fmt.Errorf("relation missing required field 'type': %w", err)
	}

	relType := core.RelationType(typeStr)

	// Parse participants
	participants, direction, err := p.parseParticipants(obj, relType, entities)
	if err != nil {
		return nil, fmt.Errorf("failed to parse participants: %w", err)
	}

	relation := &core.Relation{
		ID:          id,
		Type:        relType,
		Direction:   direction,
		Participants: *participants,
		Properties:  make(map[string]interface{}),
	}

	// Parse attributes sub-key
	if attrs, ok := getMapOptional(obj, "attributes"); ok {
		if description, ok := getStringOptional(attrs, "description"); ok {
			relation.Description = description
		}
		if statusStr, ok := getStringOptional(attrs, "status"); ok {
			relation.SetStatus(core.Status(statusStr))
		}
		if tags, ok := getSliceOptional(attrs, "tags"); ok {
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					relation.AddTag(tagStr)
				}
			}
		}
		if labels, ok := getMapOptional(attrs, "labels"); ok {
			for k, v := range labels {
				if vStr, ok := v.(string); ok {
					relation.SetLabel(k, vStr)
				}
			}
		}
		if extensions, ok := getMapOptional(attrs, "extensions"); ok {
			relation.Extensions = extensions
		}
	}

	// Parse spec sub-key for relation-type-specific properties
	if spec, ok := getMapOptional(obj, "spec"); ok {
		for k, v := range spec {
			relation.SetProperty(k, v)
		}
	}

	return relation, nil
}

// parseParticipants parses the participants field from a relation.
func (p *Parser) parseParticipants(obj map[string]interface{}, relType core.RelationType, entities map[string]*core.Entity) (*core.Participants, core.Direction, error) {
	participantsRaw, ok := obj["participants"]
	if !ok {
		return nil, "", fmt.Errorf("missing required field 'participants'")
	}

	switch participants := participantsRaw.(type) {
	case []interface{}:
		// List format (symmetric relations)
		ids := make([]string, 0, len(participants))
		for _, p := range participants {
			idStr, ok := p.(string)
			if !ok {
				return nil, "", fmt.Errorf("participant must be a string, got %T", p)
			}
			ids = append(ids, idStr)
		}
		return &core.Participants{List: ids}, core.DirectionSymmetric, nil

	case map[string]interface{}:
		// Map format (directed relations)
		source, _ := participants["source"].(string)
		target, _ := participants["target"].(string)

		if source == "" || target == "" {
			return nil, "", fmt.Errorf("directed relation must have both 'source' and 'target' participants")
		}

		return &core.Participants{
			Source: source,
			Target: target,
		}, core.DirectionDirected, nil

	default:
		return nil, "", fmt.Errorf("invalid participants format: expected list or map, got %T", participantsRaw)
	}
}

// getString extracts a string value from a map.
func getString(obj map[string]interface{}, key string) (string, error) {
	val, ok := obj[key]
	if !ok {
		return "", fmt.Errorf("field %q not found", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("field %q must be a string, got %T", key, val)
	}
	return str, nil
}

// getStringOptional extracts an optional string value from a map.
func getStringOptional(obj map[string]interface{}, key string) (string, bool) {
	val, ok := obj[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	if !ok {
		return "", false
	}
	return str, true
}

// getSliceOptional extracts an optional slice value from a map.
func getSliceOptional(obj map[string]interface{}, key string) ([]interface{}, bool) {
	val, ok := obj[key]
	if !ok {
		return nil, false
	}
	slice, ok := val.([]interface{})
	if !ok {
		return nil, false
	}
	return slice, true
}

// getMapOptional extracts an optional map value from a map.
func getMapOptional(obj map[string]interface{}, key string) (map[string]interface{}, bool) {
	val, ok := obj[key]
	if !ok {
		return nil, false
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return m, true
}

// ResolveReferences checks that all references in the graph point to existing objects.
func ResolveReferences(g *core.Graph) []error {
	var errs []error

	// Check entity owners
	for _, e := range g.Entities() {
		if !e.IsRoot() {
			if _, ok := g.GetEntity(e.Owner); !ok {
				errs = append(errs, fmt.Errorf("entity %s references non-existent owner %s", e.ID, e.Owner))
			}
		}
	}

	// Check relation participants
	for _, r := range g.Relations() {
		for _, pid := range r.ParticipantIDs() {
			// Handle interface references (entity/interface format)
			entityID := pid
			if idx := strings.Index(pid, "/"); idx != -1 {
				entityID = pid[:idx]
			}
			if _, ok := g.GetEntity(entityID); !ok {
				errs = append(errs, fmt.Errorf("relation %s references non-existent entity %s", r.ID, pid))
			}
		}
	}

	return errs
}
