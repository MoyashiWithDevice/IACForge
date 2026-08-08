package extension

import (
	"fmt"
	"os"

	"IACForge/src/schema"
	"IACForge/src/validation"
)

// Setup bundles the Schema, Validation engine, and Extension Manager for a runtime.
// It is the single construction path shared by the CLI and the MCP server so that
// extension-provided kinds, relation types, validation rules, and renderers are
// consistently available everywhere.
type Setup struct {
	Schema     *schema.Schema
	Validation *validation.Engine
	Manager    *Manager
}

// DefaultExtensionDir returns the extension directory from the IACFORGE_EXTENSIONS
// environment variable, or "" if it is not set.
func DefaultExtensionDir() string {
	return os.Getenv("IACFORGE_EXTENSIONS")
}

// NewSetup constructs the default core schema, validation engine, and extension manager.
// It registers the four core extension points and the built-in extensions, then loads
// additional Go plugin (.so) extensions from extDir (when non-empty) and applies all
// registered extensions exactly once.
func NewSetup(extDir string) (*Setup, error) {
	s := schema.CoreSchema()
	v := validation.NewEngine(s)
	validation.RegisterCoreRules(v)

	m := NewManager()
	m.RegisterExtensionPoint(NewEntityKindsExtensionPoint(s))
	m.RegisterExtensionPoint(NewRelationTypesExtensionPoint(s))
	m.RegisterExtensionPoint(NewValidationRulesExtensionPoint(v))
	m.RegisterExtensionPoint(NewRendererExtensionPoint(NewRendererRegistry()))

	for _, ext := range BuiltinExtensions() {
		if err := m.Register(ext); err != nil {
			return nil, fmt.Errorf("failed to register built-in extension %s: %w", ext.Manifest.ID, err)
		}
	}

	if extDir != "" {
		if err := m.LoadFromDir(extDir); err != nil {
			return nil, fmt.Errorf("failed to load extensions from %s: %w", extDir, err)
		}
	} else {
		if err := m.LoadAll(); err != nil {
			return nil, fmt.Errorf("failed to load extensions: %w", err)
		}
	}

	return &Setup{Schema: s, Validation: v, Manager: m}, nil
}

// builtinExts holds the built-in extensions compiled into the binary.
var builtinExts []*Extension

// BuiltinExtensions returns the built-in extensions available in this build.
func BuiltinExtensions() []*Extension {
	result := make([]*Extension, len(builtinExts))
	copy(result, builtinExts)
	return result
}

// RegisterBuiltin makes an extension available as a built-in extension.
// It is intended for extensions compiled into the binary (e.g. the AWS extension)
// so they are available in every CLI and MCP session without a plugin build.
// Registering an extension with an already-registered ID is a no-op.
func RegisterBuiltin(ext *Extension) {
	for _, e := range builtinExts {
		if e.Manifest.ID == ext.Manifest.ID {
			return
		}
	}
	builtinExts = append(builtinExts, ext)
}
