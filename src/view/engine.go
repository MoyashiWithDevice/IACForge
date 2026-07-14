package view

import (
	"fmt"

	"IACForge/src/core"
)

// Engine applies a View to a Graph and produces a ViewResult.
type Engine struct {
	source *core.Graph
}

// NewEngine creates a new view engine.
func NewEngine(source *core.Graph) *Engine {
	return &Engine{
		source: source,
	}
}

// Apply applies a View to the source Graph and returns the result.
func (e *Engine) Apply(v *View) (*ViewResult, error) {
	result := &ViewResult{
		ViewID:      v.ID,
		Title:       v.Name,
		Description: v.Description,
		Annotations: make(map[string]map[string]interface{}),
	}

	visibleEntities := e.applyVisibilityEntities(v, e.source.Entities())
	visibleRelations := e.applyVisibilityRelations(v, e.source.Relations())

	result.VisibleEntities = visibleEntities
	result.VisibleRelations = visibleRelations

	for _, gr := range v.Grouping {
		groups, err := e.applyGrouping(gr, visibleEntities)
		if err != nil {
			return nil, fmt.Errorf("grouping failed: %w", err)
		}
		result.Groups = append(result.Groups, groups...)
	}

	for _, ar := range v.Annotations {
		e.applyAnnotations(ar, visibleEntities, result.Annotations)
	}

	return result, nil
}

// applyVisibilityEntities filters entities based on visibility rules.
func (e *Engine) applyVisibilityEntities(v *View, entities []*core.Entity) []*core.Entity {
	if len(v.Visibility) == 0 {
		result := make([]*core.Entity, len(entities))
		copy(result, entities)
		return result
	}

	hasShowRule := false
	for _, rule := range v.Visibility {
		if rule.Target == VisibilityTargetEntities && rule.Action == VisibilityActionShow {
			hasShowRule = true
			break
		}
	}

	visible := make(map[string]bool)
	if hasShowRule {
		for _, entity := range entities {
			visible[entity.ID] = false
		}
	} else {
		for _, entity := range entities {
			visible[entity.ID] = true
		}
	}

	for _, rule := range v.Visibility {
		if rule.Target != VisibilityTargetEntities {
			continue
		}
		for _, entity := range entities {
			matches := e.matchEntityRule(rule, entity)
			if matches && rule.Action == VisibilityActionShow {
				visible[entity.ID] = true
			} else if matches && rule.Action == VisibilityActionHide {
				visible[entity.ID] = false
			}
		}
	}

	var result []*core.Entity
	for _, entity := range entities {
		if visible[entity.ID] {
			result = append(result, entity)
		}
	}
	return result
}

// applyVisibilityRelations filters relations based on visibility rules.
func (e *Engine) applyVisibilityRelations(v *View, relations []*core.Relation) []*core.Relation {
	if len(v.Visibility) == 0 {
		result := make([]*core.Relation, len(relations))
		copy(result, relations)
		return result
	}

	hidden := make(map[string]bool)

	for _, rule := range v.Visibility {
		if rule.Target != VisibilityTargetRelations {
			continue
		}
		for _, rel := range relations {
			matches := e.matchRelationRule(rule, rel)
			if matches && rule.Action == VisibilityActionHide {
				hidden[rel.ID] = true
			} else if matches && rule.Action == VisibilityActionShow {
				delete(hidden, rel.ID)
			}
		}
	}

	var result []*core.Relation
	for _, rel := range relations {
		if !hidden[rel.ID] {
			result = append(result, rel)
		}
	}
	return result
}

// matchEntityRule checks if an entity matches a visibility rule.
func (e *Engine) matchEntityRule(rule *VisibilityRule, entity *core.Entity) bool {
	if rule.Kind != "" && string(entity.Kind) != rule.Kind {
		return false
	}
	if rule.Where != nil {
		return matchWhereOnEntity(entity, rule.Where)
	}
	return true
}

// matchRelationRule checks if a relation matches a visibility rule.
func (e *Engine) matchRelationRule(rule *VisibilityRule, rel *core.Relation) bool {
	if rule.Relation != "" && string(rel.Type) != rule.Relation {
		return false
	}
	return true
}

// applyGrouping creates groups based on grouping rules.
func (e *Engine) applyGrouping(rule *GroupingRule, entities []*core.Entity) ([]*Group, error) {
	groups := make(map[string][]string)

	for _, entity := range entities {
		if rule.TargetKind != "" && string(entity.Kind) != rule.TargetKind {
			continue
		}

		if rule.Where != nil {
			if !matchWhereOnEntity(entity, rule.Where) {
				continue
			}
		}

		key := e.buildGroupKey(entity, rule.GroupBy)
		groups[key] = append(groups[key], entity.ID)
	}

	var result []*Group
	for key, members := range groups {
		group := &Group{
			ID:       fmt.Sprintf("group-%s-%s", rule.GroupKind, key),
			Kind:     rule.GroupKind,
			Name:     fmt.Sprintf("Group %s", key),
			Members:  members,
			Owner:    rule.Owner,
			Properties: map[string]interface{}{
				"group_key": key,
				"count":     len(members),
			},
		}
		result = append(result, group)
	}

	return result, nil
}

// buildGroupKey builds a group key from entity fields.
func (e *Engine) buildGroupKey(entity *core.Entity, groupBy []string) string {
	if len(groupBy) == 0 {
		return "all"
	}
	key := ""
	for _, field := range groupBy {
		value := getEntityField(entity, field)
		key += fmt.Sprintf("%v:", value)
	}
	return key
}

// applyAnnotations attaches annotations to entities.
func (e *Engine) applyAnnotations(rule *AnnotationRule, entities []*core.Entity, annotations map[string]map[string]interface{}) {
	for _, entity := range entities {
		if rule.TargetSelector != nil && rule.TargetSelector.Kind != "" {
			if string(entity.Kind) != rule.TargetSelector.Kind {
				continue
			}
		}

		if rule.TargetSelector != nil && rule.TargetSelector.Where != nil {
			if !matchWhereOnEntity(entity, rule.TargetSelector.Where) {
				continue
			}
		}

		if _, ok := annotations[entity.ID]; !ok {
			annotations[entity.ID] = make(map[string]interface{})
		}

		for _, ann := range rule.Annotations {
			if ann.Expression != "" {
				annotations[entity.ID][ann.Property] = ann.Expression
			} else if ann.SourceProperty != "" {
				v := getEntityField(entity, ann.SourceProperty)
				annotations[entity.ID][ann.Property] = v
			} else {
				annotations[entity.ID][ann.Property] = ann.Value
			}
		}
	}
}

// matchWhereOnEntity evaluates a where clause on an entity.
func matchWhereOnEntity(entity *core.Entity, where *WhereClause) bool {
	if where == nil || len(where.Conditions) == 0 {
		return true
	}
	for _, cond := range where.Conditions {
		value := getEntityField(entity, cond.Field)
		if !evaluateCondition(value, cond) {
			return false
		}
	}
	return true
}

// getEntityField gets a field value from an entity.
func getEntityField(entity *core.Entity, field string) interface{} {
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
	default:
		if v, ok := entity.GetProperty(field); ok {
			return v
		}
		if v, ok := entity.GetLabel(field); ok {
			return v
		}
		return nil
	}
}

// evaluateCondition evaluates a condition against a value.
func evaluateCondition(value interface{}, cond *Condition) bool {
	switch cond.Operator {
	case "eq":
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", cond.Value)
	case "ne":
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", cond.Value)
	case "contains":
		if str, ok := value.(string); ok {
			if expected, ok := cond.Value.(string); ok {
				return len(str) >= len(expected) && containsSubstring(str, expected)
			}
		}
		return false
	case "gt":
		v, vok := toFloat64(value)
		c, cok := toFloat64(cond.Value)
		if vok && cok {
			return v > c
		}
		return false
	case "lt":
		v, vok := toFloat64(value)
		c, cok := toFloat64(cond.Value)
		if vok && cok {
			return v < c
		}
		return false
	case "ge":
		v, vok := toFloat64(value)
		c, cok := toFloat64(cond.Value)
		if vok && cok {
			return v >= c
		}
		return false
	case "le":
		v, vok := toFloat64(value)
		c, cok := toFloat64(cond.Value)
		if vok && cok {
			return v <= c
		}
		return false
	case "in":
		if list, ok := cond.Value.([]interface{}); ok {
			for _, item := range list {
				if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", item) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

// toFloat64 converts a value to float64.
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

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ValidateView validates a view definition.
func ValidateView(v *View) error {
	if v.ID == "" {
		return fmt.Errorf("view must have an id")
	}
	if v.Name == "" {
		return fmt.Errorf("view must have a name")
	}
	return nil
}
