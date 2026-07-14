package query

import (
	"fmt"
	"regexp"
	"strings"

	"IACForge/src/core"
)

// Result represents the result of a query execution.
type Result struct {
	QueryID   string        `yaml:"query_id"`
	Results   []*ResultItem `yaml:"results"`
	Count     int           `yaml:"count"`
	Truncated bool          `yaml:"truncated"`
	Metadata  map[string]interface{} `yaml:"metadata,omitempty"`
}

// ResultItem represents a single result item.
type ResultItem struct {
	ID     string      `yaml:"id"`
	Type   string      `yaml:"type"`
	Path   string      `yaml:"path"`
	Object interface{} `yaml:"object"`
}

// Engine executes queries against a Graph.
type Engine struct {
	graph *core.Graph
}

// NewEngine creates a new query engine.
func NewEngine(graph *core.Graph) *Engine {
	return &Engine{graph: graph}
}

// Execute executes a query and returns the result.
func (e *Engine) Execute(q *Query) (*Result, error) {
	if q.Select == nil {
		return nil, fmt.Errorf("query must have a select clause")
	}

	var results []*ResultItem

	// Execute select
	if q.Select.Entities != nil || q.Select.Relations != nil {
		selected, err := e.executeSelect(q.Select)
		if err != nil {
			return nil, fmt.Errorf("select execution failed: %w", err)
		}
		results = append(results, selected...)
	}

	// Apply where filter
	if q.Where != nil {
		filtered, err := e.applyWhere(results, q.Where)
		if err != nil {
			return nil, fmt.Errorf("where filter failed: %w", err)
		}
		results = filtered
	}

	// Apply traverse
	if q.Traverse != nil {
		traversed, err := e.applyTraverse(results, q.Traverse)
		if err != nil {
			return nil, fmt.Errorf("traverse failed: %w", err)
		}
		results = traversed
	}

	// Apply project
	if q.Project != nil {
		projected, err := e.applyProject(results, q.Project)
		if err != nil {
			return nil, fmt.Errorf("project failed: %w", err)
		}
		results = projected
	}

	// Apply limit and offset
	totalCount := len(results)
	if q.Offset > 0 && q.Offset < len(results) {
		results = results[q.Offset:]
	} else if q.Offset >= len(results) {
		results = []*ResultItem{}
	}

	truncated := false
	if q.Limit > 0 && q.Limit < len(results) {
		results = results[:q.Limit]
		truncated = true
	}

	return &Result{
		QueryID:   q.ID,
		Results:   results,
		Count:     totalCount,
		Truncated: truncated,
	}, nil
}

// executeSelect executes the select clause.
func (e *Engine) executeSelect(sel *SelectClause) ([]*ResultItem, error) {
	var results []*ResultItem

	// Select entities
	if sel.Entities != nil {
		for _, entitySel := range sel.Entities {
			items, err := e.selectEntities(entitySel)
			if err != nil {
				return nil, err
			}
			results = append(results, items...)
		}
	}

	// Select relations
	if sel.Relations != nil {
		for _, relSel := range sel.Relations {
			items, err := e.selectRelations(relSel)
			if err != nil {
				return nil, err
			}
			results = append(results, items...)
		}
	}

	return results, nil
}

// selectEntities selects entities matching the selection criteria.
func (e *Engine) selectEntities(sel *EntitySelection) ([]*ResultItem, error) {
	var results []*ResultItem

	// Get all entities of the specified kind
	entities := e.graph.EntitiesByKind(sel.Kind)

	// Apply selection filter if present
	for _, entity := range entities {
		if sel.Where != nil {
			matches, err := e.evaluateWhereOnObject(entity, sel.Where)
			if err != nil {
				return nil, err
			}
			if !matches {
				continue
			}
		}

		results = append(results, &ResultItem{
			ID:     entity.ID,
			Type:   "entity",
			Path:   entity.Path(),
			Object: entity,
		})
	}

	return results, nil
}

// selectRelations selects relations matching the selection criteria.
func (e *Engine) selectRelations(sel *RelationSelection) ([]*ResultItem, error) {
	var results []*ResultItem

	// Get all relations of the specified type
	relations := e.graph.RelationsByType(sel.Type)

	// Apply selection filter if present
	for _, rel := range relations {
		if sel.Where != nil {
			matches, err := e.evaluateWhereOnRelation(rel, sel.Where)
			if err != nil {
				return nil, err
			}
			if !matches {
				continue
			}
		}

		results = append(results, &ResultItem{
			ID:     rel.ID,
			Type:   "relation",
			Path:   "",
			Object: rel,
		})
	}

	return results, nil
}

// applyWhere applies a where clause to filter results.
func (e *Engine) applyWhere(items []*ResultItem, where *WhereClause) ([]*ResultItem, error) {
	var filtered []*ResultItem

	for _, item := range items {
		matches, err := e.evaluateWhereOnItem(item, where)
		if err != nil {
			return nil, err
		}
		if matches {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// evaluateWhereOnItem evaluates a where clause on a result item.
func (e *Engine) evaluateWhereOnItem(item *ResultItem, where *WhereClause) (bool, error) {
	switch obj := item.Object.(type) {
	case *core.Entity:
		return e.evaluateWhereOnObject(obj, where)
	case *core.Relation:
		return e.evaluateWhereOnRelation(obj, where)
	default:
		return false, fmt.Errorf("unsupported object type: %T", item.Object)
	}
}

// evaluateWhereOnObject evaluates a where clause on an entity.
func (e *Engine) evaluateWhereOnObject(entity *core.Entity, where *WhereClause) (bool, error) {
	if where.Logical != nil {
		return e.evaluateLogicalOpOnObject(entity, where.Logical)
	}

	if len(where.Conditions) == 0 {
		return true, nil
	}

	for _, cond := range where.Conditions {
		matches, err := e.evaluateConditionOnObject(entity, cond)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

// evaluateWhereOnRelation evaluates a where clause on a relation.
func (e *Engine) evaluateWhereOnRelation(rel *core.Relation, where *WhereClause) (bool, error) {
	if where.Logical != nil {
		return e.evaluateLogicalOpOnRelation(rel, where.Logical)
	}

	if len(where.Conditions) == 0 {
		return true, nil
	}

	for _, cond := range where.Conditions {
		matches, err := e.evaluateConditionOnRelation(rel, cond)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

// evaluateConditionOnObject evaluates a condition on an entity.
func (e *Engine) evaluateConditionOnObject(entity *core.Entity, cond *Condition) (bool, error) {
	value := e.getObjectField(entity, cond.Field)
	return e.evaluateCondition(value, cond)
}

// evaluateConditionOnRelation evaluates a condition on a relation.
func (e *Engine) evaluateConditionOnRelation(rel *core.Relation, cond *Condition) (bool, error) {
	value := e.getRelationField(rel, cond.Field)
	return e.evaluateCondition(value, cond)
}

// getObjectField gets a field value from an entity.
func (e *Engine) getObjectField(entity *core.Entity, field string) interface{} {
	switch field {
	case "id":
		return entity.ID
	case "kind":
		return entity.Kind
	case "name":
		return entity.Name
	case "owner":
		return entity.Owner
	case "description":
		return entity.Description
	case "status":
		return entity.Status
	case "tags":
		return entity.Tags
	case "labels":
		return entity.Labels
	case "metadata":
		return entity.Metadata
	default:
		// Check properties
		if val, ok := entity.GetProperty(field); ok {
			return val
		}
		// Check labels
		if val, ok := entity.GetLabel(field); ok {
			return val
		}
		return nil
	}
}

// getRelationField gets a field value from a relation.
func (e *Engine) getRelationField(rel *core.Relation, field string) interface{} {
	switch field {
	case "id":
		return rel.ID
	case "type":
		return rel.Type
	case "direction":
		return rel.Direction
	case "source":
		return rel.Source()
	case "target":
		return rel.Target()
	case "description":
		return rel.Description
	case "status":
		return rel.Status
	case "tags":
		return rel.Tags
	case "labels":
		return rel.Labels
	case "metadata":
		return rel.Metadata
	default:
		// Check properties
		if val, ok := rel.GetProperty(field); ok {
			return val
		}
		// Check labels
		if val, ok := rel.GetLabel(field); ok {
			return val
		}
		return nil
	}
}

// evaluateCondition evaluates a condition against a value.
func (e *Engine) evaluateCondition(value interface{}, cond *Condition) (bool, error) {
	switch cond.Operator {
	case OperatorEq:
		return e.evaluateEq(value, cond.Value)
	case OperatorNe:
		result, err := e.evaluateEq(value, cond.Value)
		return !result, err
	case OperatorIn:
		return e.evaluateIn(value, cond.Value)
	case OperatorNin:
		result, err := e.evaluateIn(value, cond.Value)
		return !result, err
	case OperatorGt:
		return e.evaluateGt(value, cond.Value)
	case OperatorGe:
		return e.evaluateGe(value, cond.Value)
	case OperatorLt:
		return e.evaluateLt(value, cond.Value)
	case OperatorLe:
		return e.evaluateLe(value, cond.Value)
	case OperatorContains:
		return e.evaluateContains(value, cond.Value)
	case OperatorStartsWith:
		return e.evaluateStartsWith(value, cond.Value)
	case OperatorEndsWith:
		return e.evaluateEndsWith(value, cond.Value)
	case OperatorMatches:
		return e.evaluateMatches(value, cond.Value)
	case OperatorDefined:
		return e.evaluateDefined(value, cond.Value)
	case OperatorUndefined:
		result, err := e.evaluateDefined(value, cond.Value)
		return !result, err
	default:
		return false, fmt.Errorf("unsupported operator: %s", cond.Operator)
	}
}

// evaluateEq evaluates equality.
func (e *Engine) evaluateEq(value, expected interface{}) (bool, error) {
	if value == nil && expected == nil {
		return true, nil
	}
	if value == nil || expected == nil {
		return false, nil
	}
	return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", expected), nil
}

// evaluateIn evaluates if value is in a list.
func (e *Engine) evaluateIn(value interface{}, list interface{}) (bool, error) {
	if list == nil {
		return false, nil
	}

	listSlice, ok := list.([]interface{})
	if !ok {
		return false, fmt.Errorf("in operator requires a list, got %T", list)
	}

	for _, item := range listSlice {
		matches, err := e.evaluateEq(value, item)
		if err != nil {
			return false, err
		}
		if matches {
			return true, nil
		}
	}

	return false, nil
}

// toFloat64 converts a value to float64 for comparison.
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}

// evaluateGt evaluates greater than.
func (e *Engine) evaluateGt(value, expected interface{}) (bool, error) {
	v, ok := toFloat64(value)
	if !ok {
		return false, fmt.Errorf("cannot compare non-numeric value: %v", value)
	}
	ev, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("cannot compare with non-numeric value: %v", expected)
	}
	return v > ev, nil
}

// evaluateGe evaluates greater than or equal.
func (e *Engine) evaluateGe(value, expected interface{}) (bool, error) {
	v, ok := toFloat64(value)
	if !ok {
		return false, fmt.Errorf("cannot compare non-numeric value: %v", value)
	}
	ev, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("cannot compare with non-numeric value: %v", expected)
	}
	return v >= ev, nil
}

// evaluateLt evaluates less than.
func (e *Engine) evaluateLt(value, expected interface{}) (bool, error) {
	v, ok := toFloat64(value)
	if !ok {
		return false, fmt.Errorf("cannot compare non-numeric value: %v", value)
	}
	ev, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("cannot compare with non-numeric value: %v", expected)
	}
	return v < ev, nil
}

// evaluateLe evaluates less than or equal.
func (e *Engine) evaluateLe(value, expected interface{}) (bool, error) {
	v, ok := toFloat64(value)
	if !ok {
		return false, fmt.Errorf("cannot compare non-numeric value: %v", value)
	}
	ev, ok := toFloat64(expected)
	if !ok {
		return false, fmt.Errorf("cannot compare with non-numeric value: %v", expected)
	}
	return v <= ev, nil
}

// evaluateContains evaluates string contains or slice contains.
func (e *Engine) evaluateContains(value, expected interface{}) (bool, error) {
	// Handle string contains
	if str, ok := value.(string); ok {
		if expectedStr, ok := expected.(string); ok {
			return strings.Contains(str, expectedStr), nil
		}
	}

	// Handle slice contains
	if slice, ok := value.([]interface{}); ok {
		for _, item := range slice {
			matches, err := e.evaluateEq(item, expected)
			if err != nil {
				return false, err
			}
			if matches {
				return true, nil
			}
		}
		return false, nil
	}

	// Handle []string
	if strSlice, ok := value.([]string); ok {
		expectedStr, ok := expected.(string)
		if !ok {
			return false, nil
		}
		for _, s := range strSlice {
			if s == expectedStr {
				return true, nil
			}
		}
		return false, nil
	}

	return false, nil
}

// evaluateStartsWith evaluates string starts with.
func (e *Engine) evaluateStartsWith(value, expected interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, nil
	}
	expectedStr, ok := expected.(string)
	if !ok {
		return false, nil
	}
	return strings.HasPrefix(str, expectedStr), nil
}

// evaluateEndsWith evaluates string ends with.
func (e *Engine) evaluateEndsWith(value, expected interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, nil
	}
	expectedStr, ok := expected.(string)
	if !ok {
		return false, nil
	}
	return strings.HasSuffix(str, expectedStr), nil
}

// evaluateMatches evaluates regex match.
func (e *Engine) evaluateMatches(value, expected interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, nil
	}
	pattern, ok := expected.(string)
	if !ok {
		return false, nil
	}
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}
	return matched, nil
}

// evaluateDefined evaluates if a property is defined.
func (e *Engine) evaluateDefined(value interface{}, expected interface{}) (bool, error) {
	defined, ok := expected.(bool)
	if !ok {
		return false, nil
	}
	if defined {
		// Check if value is defined (not nil and not empty string)
		if value == nil {
			return false, nil
		}
		if str, ok := value.(string); ok && str == "" {
			return false, nil
		}
		return true, nil
	}
	// Check if value is undefined (nil or empty string)
	if value == nil {
		return true, nil
	}
	if str, ok := value.(string); ok && str == "" {
		return true, nil
	}
	return false, nil
}

// evaluateLogicalOpOnObject evaluates a logical operation on an entity.
func (e *Engine) evaluateLogicalOpOnObject(entity *core.Entity, op *LogicalOp) (bool, error) {
	switch op.Type {
	case LogicalOpAnd:
		for _, rule := range op.Rules {
			matches, err := e.evaluateWhereOnObject(entity, rule)
			if err != nil {
				return false, err
			}
			if !matches {
				return false, nil
			}
		}
		return true, nil
	case LogicalOpOr:
		for _, rule := range op.Rules {
			matches, err := e.evaluateWhereOnObject(entity, rule)
			if err != nil {
				return false, err
			}
			if matches {
				return true, nil
			}
		}
		return false, nil
	case LogicalOpNot:
		if len(op.Rules) == 0 {
			return true, nil
		}
		matches, err := e.evaluateWhereOnObject(entity, op.Rules[0])
		if err != nil {
			return false, err
		}
		return !matches, nil
	default:
		return false, fmt.Errorf("unsupported logical operator: %s", op.Type)
	}
}

// evaluateLogicalOpOnRelation evaluates a logical operation on a relation.
func (e *Engine) evaluateLogicalOpOnRelation(rel *core.Relation, op *LogicalOp) (bool, error) {
	switch op.Type {
	case LogicalOpAnd:
		for _, rule := range op.Rules {
			matches, err := e.evaluateWhereOnRelation(rel, rule)
			if err != nil {
				return false, err
			}
			if !matches {
				return false, nil
			}
		}
		return true, nil
	case LogicalOpOr:
		for _, rule := range op.Rules {
			matches, err := e.evaluateWhereOnRelation(rel, rule)
			if err != nil {
				return false, err
			}
			if matches {
				return true, nil
			}
		}
		return false, nil
	case LogicalOpNot:
		if len(op.Rules) == 0 {
			return true, nil
		}
		matches, err := e.evaluateWhereOnRelation(rel, op.Rules[0])
		if err != nil {
			return false, err
		}
		return !matches, nil
	default:
		return false, fmt.Errorf("unsupported logical operator: %s", op.Type)
	}
}
