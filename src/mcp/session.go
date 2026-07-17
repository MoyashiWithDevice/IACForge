package mcp

import (
	"sync"

	"IACForge/src/core"
	"IACForge/src/schema"
	"IACForge/src/validation"
)

// SessionManager manages per-session Graph instances.
type SessionManager struct {
	mu      sync.RWMutex
	sessions map[string]*SessionData
}

// SessionData holds the Graph and related engines for a single MCP session.
type SessionData struct {
	Graph      *core.Graph
	Schema     *schema.Schema
	Validation *validation.Engine
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SessionData),
	}
}

// GetOrCreate returns the session data for the given ID, creating if needed.
func (m *SessionManager) GetOrCreate(sessionID string) *SessionData {
	m.mu.RLock()
	sd, ok := m.sessions[sessionID]
	m.mu.RUnlock()
	if ok {
		return sd
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sd, ok = m.sessions[sessionID]
	if ok {
		return sd
	}

	s := schema.CoreSchema()
	v := validation.NewEngine(s)
	validation.RegisterCoreRules(v)

	sd = &SessionData{
		Graph:      core.NewGraph(),
		Schema:     s,
		Validation: v,
	}
	m.sessions[sessionID] = sd
	return sd
}

// Remove deletes session data for the given ID.
func (m *SessionManager) Remove(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}
