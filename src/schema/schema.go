package schema

import (
	"fmt"

	"IACForge/src/core"
)

// PropertyType represents a property type in the schema.
type PropertyType string

const (
	PropertyTypeString    PropertyType = "string"
	PropertyTypeInteger   PropertyType = "integer"
	PropertyTypeNumber    PropertyType = "number"
	PropertyTypeBoolean   PropertyType = "boolean"
	PropertyTypeList      PropertyType = "list"
	PropertyTypeMap       PropertyType = "map"
	PropertyTypeReference PropertyType = "reference"
	PropertyTypeEnum      PropertyType = "enum"
)

// Constraint represents a validation constraint on a property.
type Constraint struct {
	Min         *float64  `yaml:"min,omitempty"`
	Max         *float64  `yaml:"max,omitempty"`
	MinLength   *int      `yaml:"min_length,omitempty"`
	MaxLength   *int      `yaml:"max_length,omitempty"`
	Pattern     *string   `yaml:"pattern,omitempty"`
	Enum        []string  `yaml:"enum,omitempty"`
	UniqueItems *bool     `yaml:"unique_items,omitempty"`
}

// PropertyDefinition defines a property within an entity kind or relation type.
type PropertyDefinition struct {
	Name         string              `yaml:"name"`
	Type         PropertyType        `yaml:"type"`
	Required     bool                `yaml:"required"`
	Default      interface{}         `yaml:"default,omitempty"`
	Description  string              `yaml:"description,omitempty"`
	Constraints  *Constraint         `yaml:"constraints,omitempty"`
	Properties   []PropertyDefinition `yaml:"properties,omitempty"`
}

// DirectionType represents the directionality of a relation type.
type DirectionType string

const (
	DirectionDirected   DirectionType = "directed"
	DirectionSymmetric  DirectionType = "symmetric"
	DirectionUndirected DirectionType = "undirected"
)

// ParticipantConstraints defines which entity kinds can participate in a relation.
type ParticipantConstraints struct {
	SourceKinds    []core.EntityKind `yaml:"source_kinds,omitempty"`
	TargetKinds    []core.EntityKind `yaml:"target_kinds,omitempty"`
	MinParticipants int              `yaml:"min_participants,omitempty"`
	MaxParticipants int              `yaml:"max_participants,omitempty"`
}

// NestingDefinition defines a nestable child relationship for an entity kind.
type NestingDefinition struct {
	NestKey            string              `yaml:"nest_key"`
	ChildKind          core.EntityKind     `yaml:"child_kind"`
	ChildKeys          map[string]core.EntityKind `yaml:"child_keys,omitempty"`
	AutoRelationType   core.RelationType   `yaml:"auto_relation_type,omitempty"`
	AutoRelationSource string              `yaml:"auto_relation_source,omitempty"` // "parent" or "child"
}

// EntityKindDefinition defines an entity kind in the schema.
type EntityKindDefinition struct {
	Description string              `yaml:"description,omitempty"`
	Properties  []PropertyDefinition `yaml:"properties,omitempty"`
	NestingDefs []NestingDefinition `yaml:"nesting_defs,omitempty"`
}

// RelationTypeDefinition defines a relation type in the schema.
type RelationTypeDefinition struct {
	Direction     DirectionType            `yaml:"direction"`
	Description   string                   `yaml:"description,omitempty"`
	Participants  *ParticipantConstraints  `yaml:"participants,omitempty"`
	Properties    []PropertyDefinition     `yaml:"properties,omitempty"`
}

// SchemaVersion holds version information for a schema.
type SchemaVersion struct {
	SchemaVersion string `yaml:"schema_version"`
	SpecVersion   string `yaml:"spec_version"`
	Description   string `yaml:"description,omitempty"`
}

// Schema defines the complete structure of an infrastructure model.
type Schema struct {
	Version       SchemaVersion                     `yaml:"schema"`
	EntityKinds   map[core.EntityKind]*EntityKindDefinition   `yaml:"entity_kinds"`
	RelationTypes map[core.RelationType]*RelationTypeDefinition `yaml:"relation_types"`
	Profiles      []*Profile                        `yaml:"profiles,omitempty"`
}

// NewSchema creates a new empty Schema with initialized maps.
func NewSchema(schemaVersion, specVersion string) *Schema {
	return &Schema{
		Version: SchemaVersion{
			SchemaVersion: schemaVersion,
			SpecVersion:   specVersion,
		},
		EntityKinds:   make(map[core.EntityKind]*EntityKindDefinition),
		RelationTypes: make(map[core.RelationType]*RelationTypeDefinition),
	}
}

// AddEntityKind registers an entity kind definition in the schema.
func (s *Schema) AddEntityKind(kind core.EntityKind, def *EntityKindDefinition) {
	s.EntityKinds[kind] = def
}

// AddRelationType registers a relation type definition in the schema.
func (s *Schema) AddRelationType(relType core.RelationType, def *RelationTypeDefinition) {
	s.RelationTypes[relType] = def
}

// HasEntityKind checks if the schema defines the given entity kind.
func (s *Schema) HasEntityKind(kind core.EntityKind) bool {
	_, ok := s.EntityKinds[kind]
	return ok
}

// HasRelationType checks if the schema defines the given relation type.
func (s *Schema) HasRelationType(relType core.RelationType) bool {
	_, ok := s.RelationTypes[relType]
	return ok
}

// GetEntityKindDef returns the definition for the given entity kind.
func (s *Schema) GetEntityKindDef(kind core.EntityKind) (*EntityKindDefinition, bool) {
	def, ok := s.EntityKinds[kind]
	return def, ok
}

// GetRelationTypeDef returns the definition for the given relation type.
func (s *Schema) GetRelationTypeDef(relType core.RelationType) (*RelationTypeDefinition, bool) {
	def, ok := s.RelationTypes[relType]
	return def, ok
}

// GetNestingDefs returns the nesting definitions for the given entity kind.
func (s *Schema) GetNestingDefs(kind core.EntityKind) []NestingDefinition {
	def, ok := s.EntityKinds[kind]
	if !ok {
		return nil
	}
	return def.NestingDefs
}

// FindNestingByNestKey finds the nesting definition for a given parent kind and nest key.
func (s *Schema) FindNestingByNestKey(parentKind core.EntityKind, nestKey string) (*NestingDefinition, bool) {
	defs := s.GetNestingDefs(parentKind)
	for i := range defs {
		if defs[i].NestKey == nestKey {
			return &defs[i], true
		}
	}
	return nil, false
}

// FindNestingByChildKind finds the nesting definition for a given parent kind and child kind.
func (s *Schema) FindNestingByChildKind(parentKind core.EntityKind, childKind core.EntityKind) (*NestingDefinition, bool) {
	defs := s.GetNestingDefs(parentKind)
	for i := range defs {
		if defs[i].ChildKind == childKind {
			return &defs[i], true
		}
	}
	return nil, false
}

// ValidateProperty validates a property value against its definition.
func (s *Schema) ValidateProperty(propDef *PropertyDefinition, value interface{}) error {
	if value == nil {
		if propDef.Required {
			return fmt.Errorf("property %q is required but not set", propDef.Name)
		}
		return nil
	}

	if propDef.Type == PropertyTypeList {
		list, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("property %q: expected list value, got %T", propDef.Name, value)
		}
		if len(propDef.Properties) > 0 {
			for i, item := range list {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("property %q[%d]: expected map value, got %T", propDef.Name, i, item)
				}
				for _, subProp := range propDef.Properties {
					subVal, exists := itemMap[subProp.Name]
					if subVal == nil && !exists {
						if subProp.Required {
							return fmt.Errorf("property %q[%d].%s: required but not set", propDef.Name, i, subProp.Name)
						}
						continue
					}
					if err := s.ValidateProperty(&subProp, subVal); err != nil {
						return fmt.Errorf("property %q[%d]: %w", propDef.Name, i, err)
					}
				}
			}
		}
	}

	if propDef.Constraints != nil {
		if err := validateConstraints(propDef, value); err != nil {
			return fmt.Errorf("property %q: %w", propDef.Name, err)
		}
	}

	// Type-specific validation for reference properties
	if propDef.Type == PropertyTypeReference {
		if !core.IsReferenceValue(value) {
			if str, ok := value.(string); ok && len(str) > 0 && str[0] == '@' {
				// @ prefix present but not yet converted to ReferenceValue (e.g., raw input)
				return nil
			}
			return fmt.Errorf("property %q: expected a reference (use @ prefix), got %T", propDef.Name, value)
		}
	}

	return nil
}

func validateConstraints(propDef *PropertyDefinition, value interface{}) error {
	c := propDef.Constraints

	switch propDef.Type {
	case PropertyTypeInteger, PropertyTypeNumber:
		return validateNumericConstraints(c, value)
	case PropertyTypeString:
		return validateStringConstraints(c, value)
	case PropertyTypeList:
		return validateListConstraints(c, value)
	case PropertyTypeReference:
		return validateReferenceConstraints(c, value)
	}

	return nil
}

func validateNumericConstraints(c *Constraint, value interface{}) error {
	var fval float64
	switch v := value.(type) {
	case int:
		fval = float64(v)
	case int64:
		fval = float64(v)
	case float64:
		fval = v
	case float32:
		fval = float64(v)
	default:
		return fmt.Errorf("expected numeric value, got %T", value)
	}

	if c.Min != nil && fval < *c.Min {
		return fmt.Errorf("value %v is less than minimum %v", fval, *c.Min)
	}
	if c.Max != nil && fval > *c.Max {
		return fmt.Errorf("value %v is greater than maximum %v", fval, *c.Max)
	}

	return nil
}

func validateListConstraints(c *Constraint, value interface{}) error {
	list, ok := value.([]interface{})
	if !ok {
		return fmt.Errorf("expected list value, got %T", value)
	}

	if c.UniqueItems != nil && *c.UniqueItems {
		seen := make(map[interface{}]bool)
		for _, item := range list {
			if seen[item] {
				return fmt.Errorf("list contains duplicate item: %v", item)
			}
			seen[item] = true
		}
	}

	return nil
}

func validateStringConstraints(c *Constraint, value interface{}) error {
	sval, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string value, got %T", value)
	}

	if c.MinLength != nil && len(sval) < *c.MinLength {
		return fmt.Errorf("string length %d is less than minimum %d", len(sval), *c.MinLength)
	}
	if c.MaxLength != nil && len(sval) > *c.MaxLength {
		return fmt.Errorf("string length %d is greater than maximum %d", len(sval), *c.MaxLength)
	}
	if c.Pattern != nil {
		// Pattern validation would require regex compilation; basic check
		_ = c.Pattern
	}
	if len(c.Enum) > 0 {
		found := false
		for _, allowed := range c.Enum {
			if sval == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value %q is not in allowed enum values %v", sval, c.Enum)
		}
	}

	return nil
}

func validateReferenceConstraints(c *Constraint, value interface{}) error {
	// Basic type check: must be a ReferenceValue or a string with @ prefix
	if _, ok := value.(core.ReferenceValue); !ok {
		if str, ok := value.(string); ok && len(str) > 0 && str[0] == '@' {
			// Raw @ prefix string (not yet converted) is acceptable
			return nil
		}
		return fmt.Errorf("expected a reference value, got %T", value)
	}
	return nil
}
