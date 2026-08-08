package mcp

import (
	"testing"

	"IACForge/src/extension"
)

func TestSessionExtensionManager(t *testing.T) {
	sm := NewSessionManager()
	sd := sm.GetOrCreate("ext-session")

	if sd.Extensions == nil {
		t.Fatal("expected non-nil extension manager in session data")
	}
	if !sd.Extensions.IsLoaded() {
		t.Error("expected extension manager to be loaded")
	}
	if sd.Schema == nil {
		t.Fatal("expected non-nil schema in session data")
	}
	if !sd.Schema.HasEntityKind("server") {
		t.Error("expected core kind 'server' in session schema")
	}
	if sd.Validation == nil {
		t.Fatal("expected non-nil validation engine in session data")
	}
	if _, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointEntityKinds); !ok {
		t.Error("expected entity kinds extension point to be registered")
	}
	if _, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointRelationTypes); !ok {
		t.Error("expected relation types extension point to be registered")
	}
	if _, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointValidationRules); !ok {
		t.Error("expected validation rules extension point to be registered")
	}
	if _, ok := sd.Extensions.GetExtensionPoint(extension.ExtensionPointRenderers); !ok {
		t.Error("expected renderers extension point to be registered")
	}
}
