package parser

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"

	"gopkg.in/yaml.v3"

	"IACForge/src/core"
)

// Serializer serializes a Graph to YAML syntax.
type Serializer struct {
	indent int
}

// NewSerializer creates a new YAML serializer.
func NewSerializer() *Serializer {
	return &Serializer{
		indent: 2,
	}
}

// SetIndent sets the indentation level.
func (s *Serializer) SetIndent(indent int) {
	s.indent = indent
}

// SerializeFile writes a Graph to a YAML file.
func (s *Serializer) SerializeFile(g *core.Graph, path string) error {
	data, err := s.Serialize(g)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Serialize writes a Graph to YAML bytes.
func (s *Serializer) Serialize(g *core.Graph) ([]byte, error) {
	doc := s.buildDocument(g)

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(s.indent)

	if err := encoder.Encode(doc); err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close YAML encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// SerializeTo writes a Graph to a writer.
func (s *Serializer) SerializeTo(g *core.Graph, w io.Writer) error {
	doc := s.buildDocument(g)

	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(s.indent)

	if err := encoder.Encode(doc); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	return encoder.Close()
}

// buildDocument constructs the YAML document structure.
func (s *Serializer) buildDocument(g *core.Graph) map[string]interface{} {
	objects := make([]interface{}, 0)

	// Add entities (sorted by ID for deterministic output)
	entities := g.Entities()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	for _, e := range entities {
		objects = append(objects, s.buildEntity(e))
	}

	// Add relations (sorted by ID for deterministic output)
	relations := g.Relations()
	sort.Slice(relations, func(i, j int) bool {
		return relations[i].ID < relations[j].ID
	})

	for _, r := range relations {
		objects = append(objects, s.buildRelation(r))
	}

	return map[string]interface{}{
		"objects": objects,
	}
}

// buildEntity constructs the YAML representation of an entity.
func (s *Serializer) buildEntity(e *core.Entity) map[string]interface{} {
	obj := make(map[string]interface{})

	// Required fields
	obj["id"] = e.ID
	obj["kind"] = string(e.Kind)
	obj["name"] = e.Name

	// Optional fields (in order)
	if e.Owner != "" {
		obj["owner"] = e.Owner
	}
	if e.Description != "" {
		obj["description"] = e.Description
	}
	if e.Status != "" {
		obj["status"] = string(e.Status)
	}
	if len(e.Tags) > 0 {
		obj["tags"] = e.Tags
	}
	if len(e.Labels) > 0 {
		obj["labels"] = sortMap(e.Labels)
	}
	if len(e.Metadata) > 0 {
		obj["metadata"] = sortInterfaceMap(e.Metadata)
	}

	// Kind-specific properties (sorted by key)
	if len(e.Properties) > 0 {
		sortedKeys := make([]string, 0, len(e.Properties))
		for k := range e.Properties {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		for _, k := range sortedKeys {
			obj[k] = e.Properties[k]
		}
	}

	return obj
}

// buildRelation constructs the YAML representation of a relation.
func (s *Serializer) buildRelation(r *core.Relation) map[string]interface{} {
	obj := make(map[string]interface{})

	// Required fields
	obj["id"] = r.ID
	obj["type"] = string(r.Type)

	// Participants
	if r.Direction == core.DirectionSymmetric {
		obj["participants"] = r.Participants.List
	} else {
		participants := make(map[string]interface{})
		if r.Participants.Source != "" {
			participants["source"] = r.Participants.Source
		}
		if r.Participants.Target != "" {
			participants["target"] = r.Participants.Target
		}
		obj["participants"] = participants
	}

	// Optional fields (in order)
	if r.Description != "" {
		obj["description"] = r.Description
	}
	if r.Status != "" {
		obj["status"] = string(r.Status)
	}
	if len(r.Tags) > 0 {
		obj["tags"] = r.Tags
	}
	if len(r.Labels) > 0 {
		obj["labels"] = sortMap(r.Labels)
	}
	if len(r.Metadata) > 0 {
		obj["metadata"] = sortInterfaceMap(r.Metadata)
	}

	// Relation-specific properties (sorted by key)
	if len(r.Properties) > 0 {
		sortedKeys := make([]string, 0, len(r.Properties))
		for k := range r.Properties {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		for _, k := range sortedKeys {
			obj[k] = r.Properties[k]
		}
	}

	return obj
}

// sortMap returns a sorted copy of a string map for deterministic YAML output.
func sortMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// sortInterfaceMap returns a sorted copy of a map[string]interface{} for deterministic YAML output.
func sortInterfaceMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
