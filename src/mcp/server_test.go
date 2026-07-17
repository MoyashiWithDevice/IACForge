package mcp

import (
	"testing"

	"IACForge/src/core"
)

func TestSessionManager(t *testing.T) {
	sm := NewSessionManager()

	sd1 := sm.GetOrCreate("session-1")
	sd2 := sm.GetOrCreate("session-1")
	if sd1 != sd2 {
		t.Fatal("expected same session data for same ID")
	}

	sd3 := sm.GetOrCreate("session-2")
	if sd1 == sd3 {
		t.Fatal("expected different session data for different IDs")
	}

	sm.Remove("session-1")
	sd4 := sm.GetOrCreate("session-1")
	if sd1 == sd4 {
		t.Fatal("expected new session data after removal")
	}
}

func TestSessionManagerGraph(t *testing.T) {
	sm := NewSessionManager()
	sd := sm.GetOrCreate("test")

	if sd.Graph == nil {
		t.Fatal("expected non-nil graph")
	}

	e := core.NewEntity("srv-01", "server", "Server 01")
	if err := sd.Graph.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	entities := sd.Graph.Entities()
	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}
	if entities[0].ID != "srv-01" {
		t.Errorf("expected entity ID srv-01, got %s", entities[0].ID)
	}
}

func TestNewMCPServer(t *testing.T) {
	sm := NewSessionManager()
	s := NewMCPServer(sm)
	if s == nil {
		t.Fatal("expected non-nil MCP server")
	}
}
