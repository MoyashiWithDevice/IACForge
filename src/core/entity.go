package core

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEntityMissingID   = errors.New("entity missing required property: id")
	ErrEntityMissingKind = errors.New("entity missing required property: kind")
	ErrEntityMissingName = errors.New("entity missing required property: name")
)

// ReferenceValue represents a reference to another Entity in a property value.
// In YAML, references are denoted with the @ prefix (e.g., "@net-mgmt").
// At runtime, the @ prefix is stripped and the value is stored as a ReferenceValue.
type ReferenceValue string

// RefTargetID returns the ID of the referenced Entity.
func (r ReferenceValue) RefTargetID() string {
	return string(r)
}

// NewReferenceValue creates a ReferenceValue from a raw string, stripping the @ prefix if present.
func NewReferenceValue(raw string) ReferenceValue {
	return ReferenceValue(strings.TrimPrefix(raw, "@"))
}

// String returns the reference with the @ prefix for serialization.
func (r ReferenceValue) String() string {
	return "@" + string(r)
}

// IsReferenceValue checks if a property value is a ReferenceValue.
func IsReferenceValue(v interface{}) bool {
	_, ok := v.(ReferenceValue)
	return ok
}

// ExtractReferenceValue extracts the target ID from a property value if it is a ReferenceValue.
// Returns the target ID and true if the value is a reference, empty string and false otherwise.
func ExtractReferenceValue(v interface{}) (string, bool) {
	if ref, ok := v.(ReferenceValue); ok {
		return string(ref), true
	}
	return "", false
}

type Status string

const (
	StatusPlanned    Status = "planned"
	StatusActive     Status = "active"
	StatusMaintenance Status = "maintenance"
	StatusDeprecated Status = "deprecated"
	StatusOffline    Status = "offline"
)

type EntityKind string

type Entity struct {
	ID          string                 `yaml:"id"`
	Kind        EntityKind             `yaml:"kind"`
	Name        string                 `yaml:"name"`
	Owner       string                 `yaml:"owner,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Status      Status                 `yaml:"status,omitempty"`
	Tags        []string               `yaml:"tags,omitempty"`
	Labels      map[string]string      `yaml:"labels,omitempty"`
	Extensions  map[string]interface{} `yaml:"extensions,omitempty"`
	Properties  map[string]interface{} `yaml:"spec,omitempty"`
	path        string
	internal    bool
}

func NewEntity(id string, kind EntityKind, name string) *Entity {
	return &Entity{
		ID:         id,
		Kind:       kind,
		Name:       name,
		Properties: make(map[string]interface{}),
	}
}

func (e *Entity) SetOwner(ownerID string) {
	e.Owner = ownerID
}

func (e *Entity) SetStatus(status Status) {
	e.Status = status
}

func (e *Entity) AddTag(tag string) {
	e.Tags = append(e.Tags, tag)
}

func (e *Entity) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (e *Entity) SetLabel(key, value string) {
	if e.Labels == nil {
		e.Labels = make(map[string]string)
	}
	e.Labels[key] = value
}

func (e *Entity) GetLabel(key string) (string, bool) {
	if e.Labels == nil {
		return "", false
	}
	v, ok := e.Labels[key]
	return v, ok
}

func (e *Entity) SetProperty(key string, value interface{}) {
	if e.Properties == nil {
		e.Properties = make(map[string]interface{})
	}
	e.Properties[key] = value
}

func (e *Entity) GetProperty(key string) (interface{}, bool) {
	if e.Properties == nil {
		return nil, false
	}
	v, ok := e.Properties[key]
	return v, ok
}

func (e *Entity) Path() string {
	return e.path
}

func (e *Entity) SetPath(path string) {
	e.path = path
}

func (e *Entity) IsRoot() bool {
	return e.Owner == ""
}

func (e *Entity) SetInternal(internal bool) {
	e.internal = internal
}

func (e *Entity) IsInternal() bool {
	return e.internal
}

func (e *Entity) Validate() error {
	if e.ID == "" {
		return ErrEntityMissingID
	}
	if e.Kind == "" {
		return ErrEntityMissingKind
	}
	if e.Name == "" {
		return ErrEntityMissingName
	}
	return nil
}

func (e *Entity) FullPath() string {
	if e.path != "" {
		return e.path
	}
	return "/" + e.ID
}

func (e *Entity) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("id=%s", e.ID))
	parts = append(parts, fmt.Sprintf("kind=%s", e.Kind))
	parts = append(parts, fmt.Sprintf("name=%s", e.Name))
	if e.Owner != "" {
		parts = append(parts, fmt.Sprintf("owner=%s", e.Owner))
	}
	if e.Status != "" {
		parts = append(parts, fmt.Sprintf("status=%s", e.Status))
	}
	return fmt.Sprintf("Entity{%s}", strings.Join(parts, ", "))
}
