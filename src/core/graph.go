package core

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEntityNotFound      = errors.New("entity not found")
	ErrRelationNotFound    = errors.New("relation not found")
	ErrDuplicateEntityID   = errors.New("duplicate entity id")
	ErrDuplicateRelationID = errors.New("duplicate relation id")
	ErrOwnerNotFound       = errors.New("owner entity not found")
	ErrOwnershipCycle      = errors.New("ownership cycle detected")
	ErrInvalidReference    = errors.New("invalid reference to non-existent object")
	ErrOwnershipTreeBroken = errors.New("ownership tree is broken")
)

type Graph struct {
	entities  map[string]*Entity
	relations map[string]*Relation
}

func NewGraph() *Graph {
	return &Graph{
		entities:  make(map[string]*Entity),
		relations: make(map[string]*Relation),
	}
}

func (g *Graph) AddEntity(e *Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if _, exists := g.entities[e.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateEntityID, e.ID)
	}
	g.entities[e.ID] = e
	return nil
}

// ForceAddEntity adds an entity to the graph without validation.
// Useful for loading partially-valid graphs for validation engine testing.
func (g *Graph) ForceAddEntity(e *Entity) {
	g.entities[e.ID] = e
}

// ForceAddRelation adds a relation to the graph without reference validation.
// Useful for loading partially-valid graphs for validation engine testing.
func (g *Graph) ForceAddRelation(r *Relation) {
	g.relations[r.ID] = r
}

func (g *Graph) GetEntity(id string) (*Entity, bool) {
	e, ok := g.entities[id]
	return e, ok
}

func (g *Graph) MustGetEntity(id string) *Entity {
	e, ok := g.entities[id]
	if !ok {
		panic(fmt.Sprintf("entity not found: %s", id))
	}
	return e
}

func (g *Graph) RemoveEntity(id string) bool {
	_, exists := g.entities[id]
	if exists {
		delete(g.entities, id)
	}
	return exists
}

func (g *Graph) UpdateEntity(e *Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}
	g.entities[e.ID] = e
	return nil
}

func (g *Graph) Entities() []*Entity {
	result := make([]*Entity, 0, len(g.entities))
	for _, e := range g.entities {
		result = append(result, e)
	}
	return result
}

func (g *Graph) EntitiesByKind(kind EntityKind) []*Entity {
	var result []*Entity
	for _, e := range g.entities {
		if e.Kind == kind {
			result = append(result, e)
		}
	}
	return result
}

func (g *Graph) EntityCount() int {
	return len(g.entities)
}

func (g *Graph) AddRelation(r *Relation) error {
	if err := r.Validate(); err != nil {
		return err
	}
	if _, exists := g.relations[r.ID]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateRelationID, r.ID)
	}
	for _, pid := range r.ParticipantIDs() {
		if !g.isValidReference(pid) {
			return fmt.Errorf("%w: relation %s references entity %s", ErrInvalidReference, r.ID, pid)
		}
	}
	g.relations[r.ID] = r
	return nil
}

// isValidReference checks if a reference string points to a valid entity.
// Supports both simple entity IDs and interface references (entity/interface format).
func (g *Graph) isValidReference(ref string) bool {
	// Direct entity reference
	if _, ok := g.entities[ref]; ok {
		return true
	}
	// Interface reference (entity/interface format)
	if idx := strings.Index(ref, "/"); idx != -1 {
		entityID := ref[:idx]
		if _, ok := g.entities[entityID]; ok {
			return true
		}
	}
	return false
}

func (g *Graph) GetRelation(id string) (*Relation, bool) {
	r, ok := g.relations[id]
	return r, ok
}

func (g *Graph) MustGetRelation(id string) *Relation {
	r, ok := g.relations[id]
	if !ok {
		panic(fmt.Sprintf("relation not found: %s", id))
	}
	return r
}

func (g *Graph) RemoveRelation(id string) bool {
	_, exists := g.relations[id]
	if exists {
		delete(g.relations, id)
	}
	return exists
}

func (g *Graph) UpdateRelation(r *Relation) error {
	if err := r.Validate(); err != nil {
		return err
	}
	for _, pid := range r.ParticipantIDs() {
		if _, ok := g.entities[pid]; !ok {
			return fmt.Errorf("%w: relation %s references entity %s", ErrInvalidReference, r.ID, pid)
		}
	}
	g.relations[r.ID] = r
	return nil
}

func (g *Graph) Relations() []*Relation {
	result := make([]*Relation, 0, len(g.relations))
	for _, r := range g.relations {
		result = append(result, r)
	}
	return result
}

func (g *Graph) RelationsByType(relType RelationType) []*Relation {
	var result []*Relation
	for _, r := range g.relations {
		if r.Type == relType {
			result = append(result, r)
		}
	}
	return result
}

func (g *Graph) RelationsForEntity(entityID string) []*Relation {
	var result []*Relation
	for _, r := range g.relations {
		for _, pid := range r.ParticipantIDs() {
			if pid == entityID {
				result = append(result, r)
				break
			}
		}
	}
	return result
}

func (g *Graph) RelationCount() int {
	return len(g.relations)
}

func (g *Graph) ResolveReference(id string) (interface{}, bool) {
	if e, ok := g.entities[id]; ok {
		return e, true
	}
	if r, ok := g.relations[id]; ok {
		return r, true
	}
	return nil, false
}

func (g *Graph) BuildOwnershipPaths() error {
	for _, e := range g.entities {
		path, err := g.computePath(e)
		if err != nil {
			return err
		}
		e.SetPath(path)
	}
	return nil
}

func (g *Graph) computePath(e *Entity) (string, error) {
	if e.IsRoot() {
		return "/" + e.ID, nil
	}

	visited := make(map[string]bool)
	var pathParts []string
	current := e

	for current != nil {
		if visited[current.ID] {
			return "", fmt.Errorf("%w: %s", ErrOwnershipCycle, current.ID)
		}
		visited[current.ID] = true
		pathParts = append([]string{current.ID}, pathParts...)

		if current.IsRoot() {
			break
		}

		owner, ok := g.entities[current.Owner]
		if !ok {
			return "", fmt.Errorf("%w: entity %s references owner %s", ErrOwnerNotFound, current.ID, current.Owner)
		}
		current = owner
	}

	return "/" + strings.Join(pathParts, "/"), nil
}

func (g *Graph) Children(entityID string) []*Entity {
	var result []*Entity
	for _, e := range g.entities {
		if e.Owner == entityID {
			result = append(result, e)
		}
	}
	return result
}

func (g *Graph) Parent(entityID string) (*Entity, bool) {
	e, ok := g.entities[entityID]
	if !ok || e.IsRoot() {
		return nil, false
	}
	parent, ok := g.entities[e.Owner]
	return parent, ok
}

func (g *Graph) Ancestors(entityID string) []*Entity {
	var result []*Entity
	current := entityID
	visited := make(map[string]bool)

	for {
		visited[current] = true

		e, ok := g.entities[current]
		if !ok || e.IsRoot() {
			break
		}
		parent, ok := g.entities[e.Owner]
		if !ok {
			break
		}
		result = append(result, parent)
		current = parent.ID
	}
	return result
}

func (g *Graph) Descendants(entityID string) []*Entity {
	var result []*Entity
	queue := []string{entityID}
	visited := make(map[string]bool)
	visited[entityID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, child := range g.Children(current) {
			if !visited[child.ID] {
				visited[child.ID] = true
				result = append(result, child)
				queue = append(queue, child.ID)
			}
		}
	}
	return result
}

func (g *Graph) ValidateIntegrity() []error {
	var errs []error

	for _, e := range g.entities {
		if err := e.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("entity %s: %w", e.ID, err))
		}
	}

	for _, r := range g.relations {
		if err := r.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("relation %s: %w", r.ID, err))
		}
		for _, pid := range r.ParticipantIDs() {
			if _, ok := g.entities[pid]; !ok {
				errs = append(errs, fmt.Errorf("%w: relation %s references non-existent entity %s", ErrInvalidReference, r.ID, pid))
			}
		}
	}

	for _, e := range g.entities {
		if !e.IsRoot() {
			if _, ok := g.entities[e.Owner]; !ok {
				errs = append(errs, fmt.Errorf("%w: entity %s references owner %s", ErrOwnerNotFound, e.ID, e.Owner))
			}
		}
	}

	if err := g.checkOwnershipTree(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

func (g *Graph) checkOwnershipTree() error {
	visited := make(map[string]bool)
	path := make(map[string]bool)

	var visit func(id string) error
	visit = func(id string) error {
		if path[id] {
			return fmt.Errorf("%w: %s", ErrOwnershipCycle, id)
		}
		if visited[id] {
			return nil
		}
		visited[id] = true
		path[id] = true
		defer func() { path[id] = false }()

		for _, child := range g.Children(id) {
			if err := visit(child.ID); err != nil {
				return err
			}
		}
		return nil
	}

	for _, e := range g.entities {
		if e.IsRoot() {
			if err := visit(e.ID); err != nil {
				return err
			}
		}
	}

	ownershipCount := 0
	for _, e := range g.entities {
		if e.IsRoot() {
			ownershipCount++
		}
	}
	if ownershipCount > 1 {
		return ErrOwnershipTreeBroken
	}

	return nil
}
