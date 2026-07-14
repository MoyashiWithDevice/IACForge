package projection

import (
	"fmt"

	"IACForge/src/core"
)

// Engine executes projections against a Graph.
type Engine struct {
	source       *core.Graph
	projections  map[string]*Projection
	chains       map[string][]string
}

// NewEngine creates a new projection engine.
func NewEngine(source *core.Graph) *Engine {
	return &Engine{
		source:      source,
		projections: make(map[string]*Projection),
		chains:      make(map[string][]string),
	}
}

// Register adds a projection definition to the engine.
func (e *Engine) Register(proj *Projection) {
	e.projections[proj.ID] = proj
}

// Execute runs a projection and returns the result.
func (e *Engine) Execute(proj *Projection) (*ProjectionResult, error) {
	if proj.Input == nil {
		return nil, fmt.Errorf("projection must have an input clause")
	}

	var inputGraph *core.Graph
	var err error

	switch proj.Input.Type {
	case InputTypeGraph:
		inputGraph = e.source
	case InputTypeQuery:
		return nil, fmt.Errorf("query input not yet supported")
	case InputTypeProjection:
		if proj.Input.ProjectionID == "" {
			return nil, fmt.Errorf("projection input requires projection_id")
		}
		inputGraph, err = e.executeChainedProjection(proj.Input.ProjectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve projection input: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %s", proj.Input.Type)
	}

	ctx := NewOperationContext(inputGraph, proj.ID)

	if proj.Input.Type != InputTypeGraph {
		for _, e := range inputGraph.Entities() {
			ctx.copyEntity(e)
		}
		for _, r := range inputGraph.Relations() {
			ctx.copyRelation(r)
		}
	}

	for i, op := range proj.Operations {
		if err := ctx.executeOperation(op); err != nil {
			return nil, fmt.Errorf("operation %d (%s) failed: %w", i, op.Type, err)
		}
	}

	result := &ProjectionResult{
		ProjectionID: proj.ID,
		Graph:        ctx.Graph,
		DerivedCount: ctx.Derived.Count(),
	}

	return result, nil
}

// ExecuteChain runs a chain of projections in order.
func (e *Engine) ExecuteChain(projIDs ...string) (*ProjectionResult, error) {
	if len(projIDs) == 0 {
		return nil, fmt.Errorf("projection chain must have at least one projection")
	}

	var lastResult *ProjectionResult

	for _, projID := range projIDs {
		proj, ok := e.projections[projID]
		if !ok {
			return nil, fmt.Errorf("projection not found: %s", projID)
		}

		if lastResult != nil {
			proj.Input = &InputClause{
				Type:         InputTypeProjection,
				ProjectionID: proj.ID,
			}
		}

		result, err := e.Execute(proj)
		if err != nil {
			return nil, fmt.Errorf("chain execution failed at %s: %w", projID, err)
		}

		lastResult = result
	}

	return lastResult, nil
}

// executeChainedProjection executes a projection and returns its graph.
func (e *Engine) executeChainedProjection(projID string) (*core.Graph, error) {
	proj, ok := e.projections[projID]
	if !ok {
		return nil, fmt.Errorf("projection not found: %s", projID)
	}

	result, err := e.Execute(proj)
	if err != nil {
		return nil, err
	}

	return result.Graph, nil
}

// GetProjection returns a registered projection by ID.
func (e *Engine) GetProjection(id string) (*Projection, bool) {
	p, ok := e.projections[id]
	return p, ok
}

// ListProjections returns all registered projection IDs.
func (e *Engine) ListProjections() []string {
	ids := make([]string, 0, len(e.projections))
	for id := range e.projections {
		ids = append(ids, id)
	}
	return ids
}

// ValidateProjection validates a projection definition.
func ValidateProjection(proj *Projection) error {
	if proj.ID == "" {
		return fmt.Errorf("projection must have an id")
	}
	if proj.Name == "" {
		return fmt.Errorf("projection must have a name")
	}
	if proj.Input == nil {
		return fmt.Errorf("projection must have an input clause")
	}
	if len(proj.Operations) == 0 {
		return fmt.Errorf("projection must have at least one operation")
	}

	for i, op := range proj.Operations {
		if err := validateOperation(op); err != nil {
			return fmt.Errorf("operation %d: %w", i, err)
		}
	}

	return nil
}

func validateOperation(op *Operation) error {
	switch op.Type {
	case OperationSelect:
		if len(op.Entities) == 0 && len(op.Relations) == 0 {
			return fmt.Errorf("select must specify entities or relations")
		}
	case OperationFilter:
		if op.Action == "" {
			return fmt.Errorf("filter must specify action")
		}
		if op.Target == "" {
			return fmt.Errorf("filter must specify target")
		}
		if op.Where == nil {
			return fmt.Errorf("filter must specify where")
		}
	case OperationTraverse:
		if op.Direction == "" {
			return fmt.Errorf("traverse must specify direction")
		}
	case OperationAggregate:
		if op.SourceSelector == nil {
			return fmt.Errorf("aggregate must specify source_selector")
		}
		if op.TargetKind == "" {
			return fmt.Errorf("aggregate must specify target_kind")
		}
	case OperationExpand:
		if op.SourceSelector == nil {
			return fmt.Errorf("expand must specify source_selector")
		}
		if op.Expansion == nil {
			return fmt.Errorf("expand must specify expansion")
		}
	case OperationAnnotate:
		if op.TargetSelector == nil {
			return fmt.Errorf("annotate must specify target_selector")
		}
		if len(op.Annotations) == 0 {
			return fmt.Errorf("annotate must specify annotations")
		}
	case OperationGroup:
		if op.SourceSelector == nil {
			return fmt.Errorf("group must specify source_selector")
		}
		if op.GroupKind == "" {
			return fmt.Errorf("group must specify group_kind")
		}
		if len(op.GroupBy) == 0 {
			return fmt.Errorf("group must specify group_by")
		}
	case OperationFlatten:
		if op.TargetSelector == nil {
			return fmt.Errorf("flatten must specify target_selector")
		}
	case OperationEnrich:
		if op.TargetSelector == nil {
			return fmt.Errorf("enrich must specify target_selector")
		}
		if len(op.Properties) == 0 {
			return fmt.Errorf("enrich must specify properties")
		}
	case OperationTransform:
		if op.TargetSelector == nil {
			return fmt.Errorf("transform must specify target_selector")
		}
		if len(op.Transformations) == 0 {
			return fmt.Errorf("transform must specify transformations")
		}
	default:
		return fmt.Errorf("unsupported operation type: %s", op.Type)
	}

	return nil
}
