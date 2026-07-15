package projection

import (
	"fmt"

	"IACForge/src/core"
)

// OperationContext holds the state for operation execution.
type OperationContext struct {
	Graph    *core.Graph
	Source   *core.Graph
	Derived  *DerivedObjectManager
	ProjID   string
}

// NewOperationContext creates a new OperationContext.
func NewOperationContext(source *core.Graph, projID string) *OperationContext {
	return &OperationContext{
		Graph:   core.NewGraph(),
		Source:  source,
		Derived: NewDerivedObjectManager(),
		ProjID:  projID,
	}
}

// executeOperation dispatches to the appropriate operation handler.
func (ctx *OperationContext) executeOperation(op *Operation) error {
	switch op.Type {
	case OperationSelect:
		return ctx.executeSelect(op)
	case OperationFilter:
		return ctx.executeFilter(op)
	case OperationTraverse:
		return ctx.executeTraverse(op)
	case OperationAggregate:
		return ctx.executeAggregate(op)
	case OperationExpand:
		return ctx.executeExpand(op)
	case OperationAnnotate:
		return ctx.executeAnnotate(op)
	case OperationGroup:
		return ctx.executeGroup(op)
	case OperationFlatten:
		return ctx.executeFlatten(op)
	case OperationEnrich:
		return ctx.executeEnrich(op)
	case OperationTransform:
		return ctx.executeTransform(op)
	default:
		return fmt.Errorf("unsupported operation type: %s", op.Type)
	}
}

// executeSelect selects a subset of objects from the source graph.
func (ctx *OperationContext) executeSelect(op *Operation) error {
	for _, entitySel := range op.Entities {
		entities := ctx.Source.EntitiesByKind(entitySel.Kind)
		for _, e := range entities {
			if entitySel.Where != nil {
				if !matchWhereOnEntity(e, entitySel.Where) {
					continue
				}
			}
			if err := ctx.copyEntity(e); err != nil {
				return err
			}
		}
	}

	for _, relSel := range op.Relations {
		relations := ctx.Source.RelationsByType(relSel.Type)
		for _, r := range relations {
			if relSel.Where != nil {
				if !matchWhereOnRelation(r, relSel.Where) {
					continue
				}
			}
			if err := ctx.copyRelation(r); err != nil {
				return err
			}
		}
	}

	return nil
}

// executeFilter includes or excludes objects matching conditions.
func (ctx *OperationContext) executeFilter(op *Operation) error {
	if op.Target == FilterTargetEntities {
		return ctx.filterEntities(op)
	}
	return ctx.filterRelations(op)
}

func (ctx *OperationContext) filterEntities(op *Operation) error {
	entities := ctx.Graph.Entities()
	for _, e := range entities {
		matches := matchWhereOnEntity(e, op.Where)
		shouldInclude := (op.Action == FilterActionInclude && matches) ||
			(op.Action == FilterActionExclude && !matches)
		if !shouldInclude {
			ctx.Graph.RemoveEntity(e.ID)
		}
	}
	return nil
}

func (ctx *OperationContext) filterRelations(op *Operation) error {
	relations := ctx.Graph.Relations()
	for _, r := range relations {
		matches := matchWhereOnRelation(r, op.Where)
		shouldInclude := (op.Action == FilterActionInclude && matches) ||
			(op.Action == FilterActionExclude && !matches)
		if !shouldInclude {
			ctx.Graph.RemoveRelation(r.ID)
		}
	}
	return nil
}

// executeTraverse follows relations to discover additional objects.
func (ctx *OperationContext) executeTraverse(op *Operation) error {
	currentEntities := ctx.Graph.Entities()
	visited := make(map[string]bool)

	for _, e := range currentEntities {
		visited[e.ID] = true
	}

	depth := op.Depth
	if depth == 0 {
		depth = 1
	}

	for _, e := range currentEntities {
		discovered, err := ctx.traverseFromEntity(e, op, visited, 0, depth)
		if err != nil {
			return err
		}
		for _, d := range discovered {
			if !visited[d.ID] {
				visited[d.ID] = true
				if err := ctx.copyEntity(d); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (ctx *OperationContext) traverseFromEntity(
	e *core.Entity,
	op *Operation,
	visited map[string]bool,
	currentDepth, maxDepth int,
) ([]*core.Entity, error) {
	if currentDepth >= maxDepth {
		return nil, nil
	}

	var result []*core.Entity

	includeOrigin := false
	if op.IncludeOrigin != nil {
		includeOrigin = *op.IncludeOrigin
	}
	if includeOrigin && currentDepth == 0 {
		result = append(result, e)
	}

	useOwner := false
	if op.Owner != nil {
		useOwner = *op.Owner
	}

	if useOwner {
		result = append(result, ctx.traverseOwnership(e, op, visited, currentDepth, maxDepth)...)
	} else {
		result = append(result, ctx.traverseRelations(e, op, visited, currentDepth, maxDepth)...)
	}

	return result, nil
}

func (ctx *OperationContext) traverseOwnership(
	e *core.Entity,
	op *Operation,
	visited map[string]bool,
	currentDepth, maxDepth int,
) []*core.Entity {
	var result []*core.Entity

	switch op.Direction {
	case TraverseDirectionForward:
		for _, child := range ctx.Source.Children(e.ID) {
			if !visited[child.ID] {
				result = append(result, child)
				result = append(result, ctx.traverseOwnership(child, op, visited, currentDepth+1, maxDepth)...)
			}
		}
	case TraverseDirectionBackward:
		if parent, ok := ctx.Source.Parent(e.ID); ok {
			if !visited[parent.ID] {
				result = append(result, parent)
				result = append(result, ctx.traverseOwnership(parent, op, visited, currentDepth+1, maxDepth)...)
			}
		}
	case TraverseDirectionBoth:
		for _, child := range ctx.Source.Children(e.ID) {
			if !visited[child.ID] {
				result = append(result, child)
				result = append(result, ctx.traverseOwnership(child, op, visited, currentDepth+1, maxDepth)...)
			}
		}
		if parent, ok := ctx.Source.Parent(e.ID); ok {
			if !visited[parent.ID] {
				result = append(result, parent)
				result = append(result, ctx.traverseOwnership(parent, op, visited, currentDepth+1, maxDepth)...)
			}
		}
	}

	return result
}

func (ctx *OperationContext) traverseRelations(
	e *core.Entity,
	op *Operation,
	visited map[string]bool,
	currentDepth, maxDepth int,
) []*core.Entity {
	var result []*core.Entity
	relations := ctx.Source.RelationsForEntity(e.ID)

	for _, r := range relations {
		if op.RelationType != "" && r.Type != op.RelationType {
			continue
		}

		otherIDs := getOtherParticipantIDs(r, e.ID)
		for _, otherID := range otherIDs {
			if visited[otherID] {
				continue
			}
			other, ok := ctx.Source.GetEntity(otherID)
			if !ok {
				continue
			}

			shouldTraverse := false
			switch op.Direction {
			case TraverseDirectionForward:
				shouldTraverse = r.Source() == e.ID || r.Source() == ""
			case TraverseDirectionBackward:
				shouldTraverse = r.Target() == e.ID || r.Target() == ""
			case TraverseDirectionBoth:
				shouldTraverse = true
			}

			if shouldTraverse {
				result = append(result, other)
				subResults, _ := ctx.traverseFromEntity(other, op, visited, currentDepth+1, maxDepth)
				result = append(result, subResults...)
			}
		}
	}

	return result
}

// executeAggregate combines multiple objects into one derived object.
func (ctx *OperationContext) executeAggregate(op *Operation) error {
	if op.SourceSelector == nil || op.TargetKind == "" {
		return fmt.Errorf("aggregate requires source_selector and target_kind")
	}

	var sourceEntities []*core.Entity
	if op.SourceSelector.Kind != "" {
		sourceEntities = ctx.Source.EntitiesByKind(op.SourceSelector.Kind)
	} else {
		sourceEntities = ctx.Source.Entities()
	}

	groups := make(map[string][]*core.Entity)
	for _, e := range sourceEntities {
		key := ctx.buildGroupKey(e, op.GroupBy)
		groups[key] = append(groups[key], e)
	}

	for key, group := range groups {
		props := make(map[string]interface{})
		var sourceIDs []string
		for _, e := range group {
			sourceIDs = append(sourceIDs, e.ID)
		}

		for _, agg := range op.Aggregations {
			value := ctx.computeAggregation(group, agg)
			props[agg.TargetProperty] = value
		}

		props["group_key"] = key
		props["count"] = len(group)

		derived := ctx.Derived.CreateDerivedEntityWithProperties(
			core.EntityKind(op.TargetKind),
			fmt.Sprintf("%s Summary %s", op.TargetKind, key),
			props,
			sourceIDs,
			ctx.ProjID,
			"aggregate",
		)
		if err := ctx.Graph.AddEntity(derived); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *OperationContext) buildGroupKey(e *core.Entity, groupBy []string) string {
	if len(groupBy) == 0 {
		return "all"
	}
	key := ""
	for _, field := range groupBy {
		value := getEntityField(e, field)
		key += fmt.Sprintf("%v:", value)
	}
	return key
}

func (ctx *OperationContext) computeAggregation(entities []*core.Entity, agg *Aggregation) interface{} {
	switch agg.Function {
	case AggCount:
		return len(entities)
	case AggSum:
		return ctx.sumProperty(entities, agg.Property)
	case AggAvg:
		return ctx.avgProperty(entities, agg.Property)
	case AggMin:
		return ctx.minProperty(entities, agg.Property)
	case AggMax:
		return ctx.maxProperty(entities, agg.Property)
	case AggList:
		return ctx.listProperty(entities, agg.Property)
	case AggFirst:
		if len(entities) > 0 {
			return getEntityField(entities[0], agg.Property)
		}
		return nil
	case AggLast:
		if len(entities) > 0 {
			return getEntityField(entities[len(entities)-1], agg.Property)
		}
		return nil
	default:
		return nil
	}
}

func (ctx *OperationContext) sumProperty(entities []*core.Entity, property string) float64 {
	var sum float64
	for _, e := range entities {
		v := getEntityField(e, property)
		if f, ok := toFloat64(v); ok {
			sum += f
		}
	}
	return sum
}

func (ctx *OperationContext) avgProperty(entities []*core.Entity, property string) float64 {
	if len(entities) == 0 {
		return 0
	}
	return ctx.sumProperty(entities, property) / float64(len(entities))
}

func (ctx *OperationContext) minProperty(entities []*core.Entity, property string) interface{} {
	var min interface{}
	for _, e := range entities {
		v := getEntityField(e, property)
		if v == nil {
			continue
		}
		if min == nil {
			min = v
			continue
		}
		f1, ok1 := toFloat64(v)
		f2, ok2 := toFloat64(min)
		if ok1 && ok2 && f1 < f2 {
			min = v
		}
	}
	return min
}

func (ctx *OperationContext) maxProperty(entities []*core.Entity, property string) interface{} {
	var max interface{}
	for _, e := range entities {
		v := getEntityField(e, property)
		if v == nil {
			continue
		}
		if max == nil {
			max = v
			continue
		}
		f1, ok1 := toFloat64(v)
		f2, ok2 := toFloat64(max)
		if ok1 && ok2 && f1 > f2 {
			max = v
		}
	}
	return max
}

func (ctx *OperationContext) listProperty(entities []*core.Entity, property string) []interface{} {
	var result []interface{}
	for _, e := range entities {
		v := getEntityField(e, property)
		if v != nil {
			result = append(result, v)
		}
	}
	return result
}

// executeExpand replaces abstract objects with detailed objects.
func (ctx *OperationContext) executeExpand(op *Operation) error {
	if op.SourceSelector == nil || op.Expansion == nil {
		return fmt.Errorf("expand requires source_selector and expansion")
	}

	var sourceEntities []*core.Entity
	if op.SourceSelector.Kind != "" {
		sourceEntities = ctx.Source.EntitiesByKind(op.SourceSelector.Kind)
	} else {
		sourceEntities = ctx.Source.Entities()
	}

	for _, e := range sourceEntities {
		derived := ctx.Derived.CreateDerivedEntity(
			core.EntityKind(op.Expansion.TargetKind),
			fmt.Sprintf("%s from %s", op.Expansion.TargetKind, e.Name),
			[]string{e.ID},
			ctx.ProjID,
			"expand",
		)

		if op.Expansion.PropertyMapping != nil {
			for targetProp, sourceProp := range op.Expansion.PropertyMapping {
				v := getEntityField(e, sourceProp)
				derived.SetProperty(targetProp, v)
			}
		}

		if op.Expansion.Owner != "" {
			derived.SetOwner(op.Expansion.Owner)
		}

		if err := ctx.Graph.AddEntity(derived); err != nil {
			return err
		}
	}

	return nil
}

// executeAnnotate attaches computed metadata to objects.
func (ctx *OperationContext) executeAnnotate(op *Operation) error {
	if op.TargetSelector == nil {
		return fmt.Errorf("annotate requires target_selector")
	}

	var targetEntities []*core.Entity
	if op.TargetSelector.Kind != "" {
		targetEntities = ctx.Graph.EntitiesByKind(op.TargetSelector.Kind)
	} else {
		targetEntities = ctx.Graph.Entities()
	}

	for _, e := range targetEntities {
		for _, ann := range op.Annotations {
			if ann.Expression != "" {
				value := ctx.evaluateAnnotationExpression(e, ann)
				e.SetProperty(ann.Property, value)
			} else if ann.SourceProperty != "" {
				v := getEntityField(e, ann.SourceProperty)
				e.SetProperty(ann.Property, v)
			} else {
				e.SetProperty(ann.Property, ann.Value)
			}
		}
	}

	return nil
}

func (ctx *OperationContext) evaluateAnnotationExpression(e *core.Entity, ann *Annotation) interface{} {
	expr := ann.Expression
	switch {
	case expr == "timestamp":
		return fmt.Sprintf("%d", 0)
	case len(expr) > 6 && expr[:6] == "count(":
		return len(ctx.Source.Children(e.ID))
	default:
		return expr
	}
}

// executeGroup organizes objects into logical collections.
func (ctx *OperationContext) executeGroup(op *Operation) error {
	if op.SourceSelector == nil || op.GroupKind == "" || len(op.GroupBy) == 0 {
		return fmt.Errorf("group requires source_selector, group_kind, and group_by")
	}

	var sourceEntities []*core.Entity
	if op.SourceSelector.Kind != "" {
		sourceEntities = ctx.Graph.EntitiesByKind(op.SourceSelector.Kind)
	} else {
		sourceEntities = ctx.Graph.Entities()
	}

	groups := make(map[string][]*core.Entity)
	for _, e := range sourceEntities {
		key := ctx.buildGroupKey(e, op.GroupBy)
		groups[key] = append(groups[key], e)
	}

	for key, group := range groups {
		groupEntity := ctx.Derived.CreateDerivedEntity(
			core.EntityKind(op.GroupKind),
			fmt.Sprintf("Group %s", key),
			nil,
			ctx.ProjID,
			"group",
		)
		groupEntity.SetProperty("group_key", key)
		groupEntity.SetProperty("members", ctx.extractEntityIDs(group))

		if err := ctx.Graph.AddEntity(groupEntity); err != nil {
			return err
		}
	}

	return nil
}

// executeFlatten simplifies hierarchical structure.
func (ctx *OperationContext) executeFlatten(op *Operation) error {
	if op.TargetSelector == nil {
		return fmt.Errorf("flatten requires target_selector")
	}

	var targetEntities []*core.Entity
	if op.TargetSelector.Kind != "" {
		targetEntities = ctx.Graph.EntitiesByKind(op.TargetSelector.Kind)
	} else {
		targetEntities = ctx.Graph.Entities()
	}

	preserveRelations := true
	if op.PreserveRelations != nil {
		preserveRelations = *op.PreserveRelations
	}

	for _, e := range targetEntities {
		if !e.IsRoot() {
			e.SetOwner("")
		}
	}

	if !preserveRelations {
		for _, r := range ctx.Graph.Relations() {
			ctx.Graph.RemoveRelation(r.ID)
		}
	}

	return nil
}

// executeEnrich adds computed properties to objects.
func (ctx *OperationContext) executeEnrich(op *Operation) error {
	if op.TargetSelector == nil {
		return fmt.Errorf("enrich requires target_selector")
	}

	var targetEntities []*core.Entity
	if op.TargetSelector.Kind != "" {
		targetEntities = ctx.Graph.EntitiesByKind(op.TargetSelector.Kind)
	} else {
		targetEntities = ctx.Graph.Entities()
	}

	for _, e := range targetEntities {
		for _, prop := range op.Properties {
			value := ctx.evaluateExpression(e, prop.Expression)
			e.SetProperty(prop.Name, value)
		}
	}

	return nil
}

func (ctx *OperationContext) evaluateExpression(e *core.Entity, expression string) interface{} {
	if len(expression) > 1 && expression[0] == '(' && expression[len(expression)-1] == ')' {
		return ctx.evaluateArithmetic(e, expression[1:len(expression)-1])
	}
	return expression
}

func (ctx *OperationContext) evaluateArithmetic(e *core.Entity, expr string) interface{} {
	for _, op := range []string{"+", "-", "*", "/"} {
		if idx := findOperator(expr, op); idx > 0 {
			left := ctx.evaluateArithmetic(e, expr[:idx])
			right := ctx.evaluateArithmetic(e, expr[idx+1:])
			l, lok := toFloat64(left)
			r, rok := toFloat64(right)
			if lok && rok {
				switch op {
				case "+":
					return l + r
				case "-":
					return l - r
				case "*":
					return l * r
				case "/":
					if r != 0 {
						return l / r
					}
				}
			}
			return nil
		}
	}

	if isNumeric(expr) {
		if f, ok := toFloat64(expr); ok {
			return f
		}
	}

	v := getEntityField(e, expr)
	if v != nil {
		return v
	}

	return expr
}

// executeTransform transforms object properties.
func (ctx *OperationContext) executeTransform(op *Operation) error {
	if op.TargetSelector == nil {
		return fmt.Errorf("transform requires target_selector")
	}

	var targetEntities []*core.Entity
	if op.TargetSelector.Kind != "" {
		targetEntities = ctx.Graph.EntitiesByKind(op.TargetSelector.Kind)
	} else {
		targetEntities = ctx.Graph.Entities()
	}

	for _, e := range targetEntities {
		for _, t := range op.Transformations {
			switch t.Operation {
			case TransfRename:
				v := getEntityField(e, t.Property)
				if t.Value != nil {
					if newProp, ok := t.Value.(string); ok {
						e.SetProperty(newProp, v)
						e.SetProperty(t.Property, nil)
					}
				}
			case TransfCast:
				v := getEntityField(e, t.Property)
				if t.Value != nil {
					castType, ok := t.Value.(string)
					if ok {
						castValue := castProperty(v, castType)
						e.SetProperty(t.Property, castValue)
					}
				}
			case TransfSet:
				e.SetProperty(t.Property, t.Value)
			case TransfRemove:
				e.SetProperty(t.Property, nil)
			case TransfDefault:
				v := getEntityField(e, t.Property)
				if v == nil || v == "" {
					e.SetProperty(t.Property, t.Value)
				}
			}
		}
	}

	return nil
}

// copyEntity copies an entity from the source graph to the projection graph.
func (ctx *OperationContext) copyEntity(e *core.Entity) error {
	if _, exists := ctx.Graph.GetEntity(e.ID); exists {
		return nil
	}

	copy := core.NewEntity(e.ID, e.Kind, e.Name)
	copy.Owner = e.Owner
	copy.Description = e.Description
	copy.Status = e.Status
	copy.Tags = append([]string{}, e.Tags...)

	if e.Labels != nil {
		copy.Labels = make(map[string]string)
		for k, v := range e.Labels {
			copy.Labels[k] = v
		}
	}
	if e.Extensions != nil {
		copy.Extensions = make(map[string]interface{})
		for k, v := range e.Extensions {
			copy.Extensions[k] = v
		}
	}
	if e.Properties != nil {
		copy.Properties = make(map[string]interface{})
		for k, v := range e.Properties {
			copy.Properties[k] = v
		}
	}

	return ctx.Graph.AddEntity(copy)
}

// copyRelation copies a relation from the source graph to the projection graph.
func (ctx *OperationContext) copyRelation(r *core.Relation) error {
	if _, exists := ctx.Graph.GetRelation(r.ID); exists {
		return nil
	}

	copy := core.NewRelation(r.ID, r.Type, r.Direction)
	copy.Participants = r.Participants
	copy.Description = r.Description
	copy.Status = r.Status
	copy.Tags = append([]string{}, r.Tags...)

	if r.Labels != nil {
		copy.Labels = make(map[string]string)
		for k, v := range r.Labels {
			copy.Labels[k] = v
		}
	}
	if r.Extensions != nil {
		copy.Extensions = make(map[string]interface{})
		for k, v := range r.Extensions {
			copy.Extensions[k] = v
		}
	}
	if r.Properties != nil {
		copy.Properties = make(map[string]interface{})
		for k, v := range r.Properties {
			copy.Properties[k] = v
		}
	}

	allParticipantIDs := r.ParticipantIDs()
	for _, pid := range allParticipantIDs {
		if _, ok := ctx.Graph.GetEntity(pid); !ok {
			if sourceEntity, ok := ctx.Source.GetEntity(pid); ok {
				if err := ctx.copyEntity(sourceEntity); err != nil {
					return err
				}
			}
		}
	}

	return ctx.Graph.AddRelation(copy)
}

func (ctx *OperationContext) extractEntityIDs(entities []*core.Entity) []string {
	ids := make([]string, len(entities))
	for i, e := range entities {
		ids[i] = e.ID
	}
	return ids
}

// getOtherParticipantIDs returns participant IDs other than the given entity.
func getOtherParticipantIDs(r *core.Relation, entityID string) []string {
	var others []string
	for _, pid := range r.ParticipantIDs() {
		if pid != entityID {
			others = append(others, pid)
		}
	}
	return others
}

// getEntityField gets a field value from an entity.
func getEntityField(e *core.Entity, field string) interface{} {
	switch field {
	case "id":
		return e.ID
	case "kind":
		return e.Kind
	case "name":
		return e.Name
	case "owner":
		return e.Owner
	case "description":
		return e.Description
	case "status":
		return e.Status
	case "tags":
		return e.Tags
	default:
		if v, ok := e.GetProperty(field); ok {
			return v
		}
		if v, ok := e.GetLabel(field); ok {
			return v
		}
		return nil
	}
}

// matchWhereOnEntity evaluates a where clause on an entity.
func matchWhereOnEntity(e *core.Entity, where *WhereClause) bool {
	if where == nil || len(where.Conditions) == 0 {
		return true
	}
	for _, cond := range where.Conditions {
		value := getEntityField(e, cond.Field)
		if !evaluateCondition(value, cond) {
			return false
		}
	}
	return true
}

// matchWhereOnRelation evaluates a where clause on a relation.
func matchWhereOnRelation(r *core.Relation, where *WhereClause) bool {
	if where == nil || len(where.Conditions) == 0 {
		return true
	}
	for _, cond := range where.Conditions {
		value := getRelationField(r, cond.Field)
		if !evaluateCondition(value, cond) {
			return false
		}
	}
	return true
}

// getRelationField gets a field value from a relation.
func getRelationField(r *core.Relation, field string) interface{} {
	switch field {
	case "id":
		return r.ID
	case "type":
		return r.Type
	case "direction":
		return r.Direction
	case "source":
		return r.Source()
	case "target":
		return r.Target()
	case "description":
		return r.Description
	case "status":
		return r.Status
	default:
		if v, ok := r.GetProperty(field); ok {
			return v
		}
		if v, ok := r.GetLabel(field); ok {
			return v
		}
		return nil
	}
}

// evaluateCondition evaluates a condition against a value.
func evaluateCondition(value interface{}, cond *Condition) bool {
	switch cond.Operator {
	case OperatorEq:
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", cond.Value)
	case OperatorNe:
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", cond.Value)
	case OperatorContains:
		if str, ok := value.(string); ok {
			if expected, ok := cond.Value.(string); ok {
				return contains(str, expected)
			}
		}
		return false
	case OperatorGt:
		v, vok := toFloat64(value)
		c, cok := toFloat64(cond.Value)
		if vok && cok {
			return v > c
		}
		return false
	case OperatorLt:
		v, vok := toFloat64(value)
		c, cok := toFloat64(cond.Value)
		if vok && cok {
			return v < c
		}
		return false
	case OperatorGe:
		v, vok := toFloat64(value)
		c, cok := toFloat64(cond.Value)
		if vok && cok {
			return v >= c
		}
		return false
	case OperatorLe:
		v, vok := toFloat64(value)
		c, cok := toFloat64(cond.Value)
		if vok && cok {
			return v <= c
		}
		return false
	case OperatorIn:
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

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// findOperator finds the position of an operator in an expression.
func findOperator(expr string, op string) int {
	depth := 0
	for i := len(expr) - 1; i >= 0; i-- {
		if expr[i] == ')' {
			depth++
		} else if expr[i] == '(' {
			depth--
		} else if depth == 0 && string(expr[i]) == op {
			return i
		}
	}
	return -1
}

// isNumeric checks if a string is numeric.
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if c == '-' && i == 0 {
			continue
		}
		if c < '0' || c > '9' {
			if c != '.' {
				return false
			}
		}
	}
	return true
}

// castProperty casts a property value to a different type.
func castProperty(v interface{}, targetType string) interface{} {
	switch targetType {
	case "string":
		return fmt.Sprintf("%v", v)
	case "integer":
		if f, ok := toFloat64(v); ok {
			return int(f)
		}
		return v
	case "number":
		if f, ok := toFloat64(v); ok {
			return f
		}
		return v
	default:
		return v
	}
}
