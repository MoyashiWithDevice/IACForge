package extension

import (
	"IACForge/src/core"
	"IACForge/src/renderer"
	"IACForge/src/schema"
	"IACForge/src/validation"
)

// Manifest holds machine-readable metadata for an extension.
type Manifest struct {
	ID             string   `yaml:"id"`
	Name           string   `yaml:"name"`
	Version        string   `yaml:"version"`
	Author         string   `yaml:"author,omitempty"`
	Description    string   `yaml:"description,omitempty"`
	SpecVersion    string   `yaml:"spec_version,omitempty"`
	SchemaVersion  string   `yaml:"schema_version,omitempty"`
	Namespace      string   `yaml:"namespace"`
	Dependencies   []string `yaml:"dependencies,omitempty"`
	ExtensionPoints []string `yaml:"extension_points"`
}

// ExtensionPointType identifies the type of an extension point.
type ExtensionPointType string

const (
	ExtensionPointEntityKinds    ExtensionPointType = "entity_kinds"
	ExtensionPointRelationTypes  ExtensionPointType = "relation_types"
	ExtensionPointValidationRules ExtensionPointType = "validation_rules"
	ExtensionPointRenderers      ExtensionPointType = "renderers"
)

// ExtensionPoint defines the interface that all extension points must implement.
type ExtensionPoint interface {
	Type() ExtensionPointType
	Register(ext *Extension) error
}

// EntityKindContribution represents a single entity kind contributed by an extension.
type EntityKindContribution struct {
	Kind        core.EntityKind
	Definition  *schema.EntityKindDefinition
}

// RelationTypeContribution represents a single relation type contributed by an extension.
type RelationTypeContribution struct {
	Type       core.RelationType
	Definition *schema.RelationTypeDefinition
}

// ValidationRuleContribution represents a single validation rule contributed by an extension.
type ValidationRuleContribution struct {
	Rule *validation.Rule
	Fn   validation.RuleFunc
}

// RendererContribution represents a single renderer contributed by an extension.
type RendererContribution struct {
	Renderer renderer.Renderer
}

// Extension represents a loaded extension with its contributions.
type Extension struct {
	Manifest        *Manifest
	EntityKinds     []EntityKindContribution
	RelationTypes   []RelationTypeContribution
	ValidationRules []ValidationRuleContribution
	Renderers       []RendererContribution
}
