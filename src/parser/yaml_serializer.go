package parser

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"

	"gopkg.in/yaml.v3"

	"IACForge/src/core"
	"IACForge/src/schema"
)

// Serializer serializes a Graph to YAML syntax.
type Serializer struct {
	indent int
	schema *schema.Schema
}

// NewSerializer creates a new YAML serializer with the core schema.
func NewSerializer() *Serializer {
	return &Serializer{
		indent: 2,
		schema: schema.CoreSchema(),
	}
}

// NewSerializerWithSchema creates a new YAML serializer with a custom schema.
func NewSerializerWithSchema(s *schema.Schema) *Serializer {
	return &Serializer{
		indent: 2,
		schema: s,
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

	// Build a map of children grouped by parent ID and nest key
	childrenByParent := make(map[string][]childEntry)

	for _, e := range g.Entities() {
		if e.IsRoot() {
			continue
		}
		parent, ok := g.GetEntity(e.Owner)
		if !ok {
			continue
		}
		nd, ok := s.schema.FindNestingByChildKind(parent.Kind, e.Kind)
		if !ok {
			continue
		}
		childrenByParent[e.Owner] = append(childrenByParent[e.Owner], childEntry{
			entity:  e,
			nestKey: nd.NestKey,
		})
	}

	// Sort children by nest key, then by ID for deterministic output
	for parentID := range childrenByParent {
		entries := childrenByParent[parentID]
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].nestKey != entries[j].nestKey {
				return entries[i].nestKey < entries[j].nestKey
			}
			return entries[i].entity.ID < entries[j].entity.ID
		})
		childrenByParent[parentID] = entries
	}

	// Collect set of entity IDs that are nested children (to exclude from top-level)
	nestedIDs := make(map[string]bool)
	for _, entries := range childrenByParent {
		for _, entry := range entries {
			nestedIDs[entry.entity.ID] = true
		}
	}

	// Add entities that are not nested children (roots and non-nestable owned entities)
	entities := g.Entities()
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].ID < entities[j].ID
	})

	for _, e := range entities {
		if nestedIDs[e.ID] {
			continue
		}
		objects = append(objects, s.buildEntityWithChildren(e, childrenByParent, false))
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

// childEntry holds a nested child entity and its nest key.
type childEntry struct {
	entity  *core.Entity
	nestKey string
}

// buildEntityWithChildren constructs the YAML representation of an entity,
// including nested children in the spec section.
// When isNested is true, the output omits owner and kind (inferred from context).
func (s *Serializer) buildEntityWithChildren(e *core.Entity, childrenByParent map[string][]childEntry, isNested bool) map[string]interface{} {
	obj := make(map[string]interface{})

	// Always include id
	obj["id"] = e.ID

	// Only output kind for top-level entities (nested kind is inferred from nest key)
	if !isNested {
		obj["kind"] = string(e.Kind)
	}

	// Always include name for top-level entities; for nested, only if different from id
	if !isNested || e.Name != e.ID {
		obj["name"] = e.Name
	}

	// Build attributes sub-key (omit owner for nested entities)
	attrs := make(map[string]interface{})
	if !isNested && e.Owner != "" {
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
	spec := make(map[string]interface{})
	sortedKeys := make([]string, 0, len(e.Properties))
	for k := range e.Properties {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	for _, k := range sortedKeys {
		spec[k] = serializePropertyValue(e.Properties[k])
	}

	// Group children by nest key and recurse
	directChildren := childrenByParent[e.ID]
	if len(directChildren) > 0 {
		nestKeyGroups := make(map[string][]childEntry)
		for _, c := range directChildren {
			nestKeyGroups[c.nestKey] = append(nestKeyGroups[c.nestKey], c)
		}

		nestKeys := make([]string, 0, len(nestKeyGroups))
		for k := range nestKeyGroups {
			nestKeys = append(nestKeys, k)
		}
		sort.Strings(nestKeys)

		for _, nestKey := range nestKeys {
			group := nestKeyGroups[nestKey]
			nestedList := make([]interface{}, 0, len(group))
			for _, ce := range group {
				nestedList = append(nestedList, s.buildEntityWithChildren(ce.entity, childrenByParent, true))
			}
			spec[nestKey] = nestedList
		}
	}

	if len(spec) > 0 {
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
			spec[k] = serializePropertyValue(r.Properties[k])
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

// serializePropertyValue converts a property value for YAML output.
// ReferenceValue is serialized with the @ prefix.
func serializePropertyValue(v interface{}) interface{} {
	if ref, ok := v.(core.ReferenceValue); ok {
		return ref.String()
	}
	return v
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
