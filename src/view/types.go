package view

import (
	"IACForge/src/core"
)

// View defines how a projected Graph is presented to a consumer.
type View struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Audience    string            `yaml:"audience,omitempty"`
	Visibility  []*VisibilityRule `yaml:"visibility,omitempty"`
	Grouping    []*GroupingRule   `yaml:"grouping,omitempty"`
	Annotations []*AnnotationRule `yaml:"annotations,omitempty"`
	Layout      *LayoutHint       `yaml:"layout,omitempty"`
}

// VisibilityRule determines which objects are shown or hidden.
type VisibilityRule struct {
	Target   VisibilityTarget `yaml:"target"`
	Kind     string           `yaml:"kind,omitempty"`
	Relation string           `yaml:"relation,omitempty"`
	Action   VisibilityAction `yaml:"action"`
	Where    *WhereClause     `yaml:"where,omitempty"`
}

// VisibilityTarget specifies whether the rule applies to entities or relations.
type VisibilityTarget string

const (
	VisibilityTargetEntities  VisibilityTarget = "entities"
	VisibilityTargetRelations VisibilityTarget = "relations"
)

// VisibilityAction specifies whether to show or hide matching objects.
type VisibilityAction string

const (
	VisibilityActionShow VisibilityAction = "show"
	VisibilityActionHide VisibilityAction = "hide"
)

// GroupingRule organizes objects into logical groups.
type GroupingRule struct {
	TargetKind string          `yaml:"target_kind"`
	GroupKind  string          `yaml:"group_kind"`
	GroupBy    []string        `yaml:"group_by"`
	Owner      string          `yaml:"owner,omitempty"`
	Where      *WhereClause    `yaml:"where,omitempty"`
}

// AnnotationRule attaches computed metadata to objects.
type AnnotationRule struct {
	TargetSelector *EntitySelector `yaml:"target_selector"`
	Annotations    []*Annotation   `yaml:"annotations"`
}

// LayoutHint provides spatial arrangement hints.
type LayoutHint struct {
	Type     string  `yaml:"type,omitempty"`
	Direction string `yaml:"direction,omitempty"`
	Spacing  float64 `yaml:"spacing,omitempty"`
	Padding  float64 `yaml:"padding,omitempty"`
}

// EntitySelector defines entity selection criteria.
type EntitySelector struct {
	Kind  string       `yaml:"kind,omitempty"`
	Where *WhereClause `yaml:"where,omitempty"`
}

// WhereClause defines filtering conditions.
type WhereClause struct {
	Conditions []*Condition `yaml:"conditions,omitempty"`
}

// Condition represents a single filter condition.
type Condition struct {
	Field    string      `yaml:"field"`
	Operator string      `yaml:"operator"`
	Value    interface{} `yaml:"value"`
}

// Annotation defines an annotation to attach.
type Annotation struct {
	Property       string      `yaml:"property"`
	Value          interface{} `yaml:"value,omitempty"`
	Expression     string      `yaml:"expression,omitempty"`
	SourceProperty string      `yaml:"source_property,omitempty"`
}

// ViewResult represents the result of applying a View to a Graph.
type ViewResult struct {
	ViewID        string
	Title         string
	Description   string
	VisibleEntities  []*core.Entity
	VisibleRelations []*core.Relation
	Groups        []*Group
	Annotations   map[string]map[string]interface{}
}

// Group represents a collection of objects.
type Group struct {
	ID       string
	Kind     string
	Name     string
	Members  []string
	Owner    string
	Properties map[string]interface{}
}

// NewView creates a new View.
func NewView(id, name string) *View {
	return &View{
		ID:   id,
		Name: name,
	}
}

// NewVisibilityRule creates a new VisibilityRule.
func NewVisibilityRule(target VisibilityTarget, action VisibilityAction) *VisibilityRule {
	return &VisibilityRule{
		Target: target,
		Action: action,
	}
}

// NewGroupingRule creates a new GroupingRule.
func NewGroupingRule(targetKind, groupKind string) *GroupingRule {
	return &GroupingRule{
		TargetKind: targetKind,
		GroupKind:  groupKind,
		GroupBy:    make([]string, 0),
	}
}

// NewAnnotationRule creates a new AnnotationRule.
func NewAnnotationRule() *AnnotationRule {
	return &AnnotationRule{
		Annotations: make([]*Annotation, 0),
	}
}

// NewLayoutHint creates a new LayoutHint.
func NewLayoutHint() *LayoutHint {
	return &LayoutHint{}
}

// NewEntitySelector creates a new EntitySelector.
func NewEntitySelector() *EntitySelector {
	return &EntitySelector{}
}

// NewWhereClause creates a new WhereClause.
func NewWhereClause() *WhereClause {
	return &WhereClause{}
}

// NewCondition creates a new Condition.
func NewCondition(field, operator string, value interface{}) *Condition {
	return &Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

// NewAnnotation creates a new Annotation.
func NewAnnotation(property string) *Annotation {
	return &Annotation{
		Property: property,
	}
}

// AddCondition adds a condition to the where clause.
func (w *WhereClause) AddCondition(field, operator string, value interface{}) *Condition {
	cond := NewCondition(field, operator, value)
	w.Conditions = append(w.Conditions, cond)
	return cond
}

// AddVisibility adds a visibility rule to the view.
func (v *View) AddVisibility(rule *VisibilityRule) {
	v.Visibility = append(v.Visibility, rule)
}

// AddGrouping adds a grouping rule to the view.
func (v *View) AddGrouping(rule *GroupingRule) {
	v.Grouping = append(v.Grouping, rule)
}

// AddAnnotationRule adds an annotation rule to the view.
func (v *View) AddAnnotationRule(rule *AnnotationRule) {
	v.Annotations = append(v.Annotations, rule)
}
