package projection

import (
	"fmt"
	"time"

	"IACForge/src/core"
)

// DerivedObjectManager manages derived objects created during projections.
type DerivedObjectManager struct {
	objects map[string]*core.Entity
	counter int
}

// NewDerivedObjectManager creates a new DerivedObjectManager.
func NewDerivedObjectManager() *DerivedObjectManager {
	return &DerivedObjectManager{
		objects: make(map[string]*core.Entity),
	}
}

// CreateDerivedEntity creates a new derived entity with provenance.
func (m *DerivedObjectManager) CreateDerivedEntity(
	kind core.EntityKind,
	name string,
	sourceIDs []string,
	projectionID string,
	operation string,
) *core.Entity {
	m.counter++
	id := m.generateID(kind)

	e := core.NewEntity(id, kind, name)
	e.SetProperty("derived", true)

	if e.Extensions == nil {
		e.Extensions = make(map[string]interface{})
	}
	e.Extensions["provenance"] = &Provenance{
		SourceIDs:    sourceIDs,
		ProjectionID: projectionID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Operation:    operation,
	}

	m.objects[id] = e
	return e
}

// CreateDerivedEntityWithProperties creates a derived entity with specific properties.
func (m *DerivedObjectManager) CreateDerivedEntityWithProperties(
	kind core.EntityKind,
	name string,
	properties map[string]interface{},
	sourceIDs []string,
	projectionID string,
	operation string,
) *core.Entity {
	e := m.CreateDerivedEntity(kind, name, sourceIDs, projectionID, operation)
	for k, v := range properties {
		e.SetProperty(k, v)
	}
	return e
}

// generateID generates a unique ID for a derived entity.
func (m *DerivedObjectManager) generateID(kind core.EntityKind) string {
	return fmt.Sprintf("derived-%s-%d", string(kind), m.counter)
}

// GetAll returns all derived objects.
func (m *DerivedObjectManager) GetAll() map[string]*core.Entity {
	return m.objects
}

// Count returns the number of derived objects.
func (m *DerivedObjectManager) Count() int {
	return len(m.objects)
}

// GetProvenance extracts provenance from an entity's extensions.
func GetProvenance(e *core.Entity) (*Provenance, bool) {
	if e.Extensions == nil {
		return nil, false
	}
	p, ok := e.Extensions["provenance"]
	if !ok {
		return nil, false
	}
	prov, ok := p.(*Provenance)
	return prov, ok
}

// IsDerived checks if an entity is a derived object.
func IsDerived(e *core.Entity) bool {
	if v, ok := e.GetProperty("derived"); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
