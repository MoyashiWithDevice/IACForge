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

	// Required fields at top level
	obj["id"] = e.ID
	obj["kind"] = string(e.Kind)
	obj["name"] = e.Name

	// Build attributes sub-key
	attrs := make(map[string]interface{})
	if e.Owner != "" {
		attrs["owner"] = e.Owner
	}
	if e.Description != "" {
		attrs["description"] = e.Description
	}
	if e.Status != "" {
		attrs["status"] = string(e.Status)
	}
	if len(e.Tags) > 0 {
		attrs["tags"] = e.Tags
	}
	if len(e.Labels) > 0 {
		attrs["labels"] = sortMap(e.Labels)
	}
	if len(e.Extensions) > 0 {
		attrs["extensions"] = sortInterfaceMap(e.Extensions)
	}
	if len(attrs) > 0 {
		obj["attributes"] = attrs
	}

	// Build spec sub-key for kind-specific properties
	if len(e.Properties) > 0 {
		spec := make(map[string]interface{})
		sortedKeys := make([]string, 0, len(e.Properties))
		for k := range e.Properties {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		for _, k := range sortedKeys {
			spec[k] = e.Properties[k]
		}
		obj["spec"] = spec
	}

	return obj
}

// buildRelation constructs the YAML representation of a relation.
func (s *Serializer) buildRelation(r *core.Relation) map[string]interface{} {
	obj := make(map[string]interface{})

	// Required fields at top level
	obj["id"] = r.ID
	obj["type"] = string(r.Type)

	// Participants at top level
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

	// Build attributes sub-key
	attrs := make(map[string]interface{})
	if r.Description != "" {
		attrs["description"] = r.Description
	}
	if r.Status != "" {
		attrs["status"] = string(r.Status)
	}
	if len(r.Tags) > 0 {
		attrs["tags"] = r.Tags
	}
	if len(r.Labels) > 0 {
		attrs["labels"] = sortMap(r.Labels)
	}
	if len(r.Extensions) > 0 {
		attrs["extensions"] = sortInterfaceMap(r.Extensions)
	}
	if len(attrs) > 0 {
		obj["attributes"] = attrs
	}

	// Build spec sub-key for relation-type-specific properties
	if len(r.Properties) > 0 {
		spec := make(map[string]interface{})
		sortedKeys := make([]string, 0, len(r.Properties))
		for k := range r.Properties {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		for _, k := range sortedKeys {
			spec[k] = r.Properties[k]
		}
		obj["spec"] = spec
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
