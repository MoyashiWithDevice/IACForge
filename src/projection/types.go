package projection

import (
	"IACForge/src/core"
)

// Projection represents a complete graph transformation definition.
type Projection struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description,omitempty"`
	Input       *InputClause   `yaml:"input"`
	Operations  []*Operation   `yaml:"operations"`
	Output      *OutputClause  `yaml:"output,omitempty"`
}

// InputClause defines what enters the Projection.
type InputClause struct {
	Type          InputType `yaml:"type"`
	QueryID       string    `yaml:"query_id,omitempty"`
	ProjectionID  string    `yaml:"projection_id,omitempty"`
}

// InputType represents the type of input source.
type InputType string

const (
	InputTypeGraph      InputType = "graph"
	InputTypeQuery      InputType = "query"
	InputTypeProjection InputType = "projection"
)

// OutputClause configures Projection output.
type OutputClause struct {
	Format           OutputFormat `yaml:"format,omitempty"`
	IncludeProvenance *bool       `yaml:"include_provenance,omitempty"`
	IncludeDerived   *bool       `yaml:"include_derived,omitempty"`
}

// OutputFormat represents the output format.
type OutputFormat string

const (
	OutputFormatGraph    OutputFormat = "graph"
	OutputFormatSummary  OutputFormat = "summary"
	OutputFormatFiltered OutputFormat = "filtered"
)

// Operation represents a transformation operation.
type Operation struct {
	Type             OperationType      `yaml:"type"`
	Entities         []*EntitySelector  `yaml:"entities,omitempty"`
	Relations        []*RelationSelector `yaml:"relations,omitempty"`
	Action           FilterAction       `yaml:"action,omitempty"`
	Target           FilterTarget       `yaml:"target,omitempty"`
	Where            *WhereClause       `yaml:"where,omitempty"`
	Direction        TraverseDirection  `yaml:"direction,omitempty"`
	RelationType     core.RelationType  `yaml:"relation_type,omitempty"`
	Owner            *bool             `yaml:"owner,omitempty"`
	Depth            int               `yaml:"depth,omitempty"`
	IncludeOrigin    *bool             `yaml:"include_origin,omitempty"`
	SourceSelector   *Selector         `yaml:"source_selector,omitempty"`
	TargetSelector   *Selector         `yaml:"target_selector,omitempty"`
	TargetKind       string            `yaml:"target_kind,omitempty"`
	GroupBy          []string          `yaml:"group_by,omitempty"`
	Aggregations     []*Aggregation    `yaml:"aggregations,omitempty"`
	Expansion        *ExpansionConfig  `yaml:"expansion,omitempty"`
	Annotations      []*Annotation     `yaml:"annotations,omitempty"`
	GroupKind        string            `yaml:"group_kind,omitempty"`
	PreserveRelations *bool            `yaml:"preserve_relations,omitempty"`
	Properties       []*ComputedProperty `yaml:"properties,omitempty"`
	Transformations  []*Transformation `yaml:"transformations,omitempty"`
}

// OperationType represents the type of operation.
type OperationType string

const (
	OperationSelect    OperationType = "select"
	OperationFilter    OperationType = "filter"
	OperationTraverse  OperationType = "traverse"
	OperationAggregate OperationType = "aggregate"
	OperationExpand    OperationType = "expand"
	OperationAnnotate  OperationType = "annotate"
	OperationGroup     OperationType = "group"
	OperationFlatten   OperationType = "flatten"
	OperationEnrich    OperationType = "enrich"
	OperationTransform OperationType = "transform"
)

// EntitySelector defines entity selection criteria.
type EntitySelector struct {
	Kind  core.EntityKind `yaml:"kind"`
	Where *WhereClause    `yaml:"where,omitempty"`
}

// RelationSelector defines relation selection criteria.
type RelationSelector struct {
	Type  core.RelationType `yaml:"type"`
	Where *WhereClause      `yaml:"where,omitempty"`
}

// FilterAction represents include or exclude.
type FilterAction string

const (
	FilterActionInclude FilterAction = "include"
	FilterActionExclude FilterAction = "exclude"
)

// FilterTarget represents entities or relations.
type FilterTarget string

const (
	FilterTargetEntities  FilterTarget = "entities"
	FilterTargetRelations FilterTarget = "relations"
)

// TraverseDirection represents traversal direction.
type TraverseDirection string

const (
	TraverseDirectionForward  TraverseDirection = "forward"
	TraverseDirectionBackward TraverseDirection = "backward"
	TraverseDirectionBoth     TraverseDirection = "both"
)

// Selector defines a source selection.
type Selector struct {
	Kind         core.EntityKind `yaml:"kind,omitempty"`
	RelationType core.RelationType `yaml:"relation_type,omitempty"`
}

// WhereClause defines filtering conditions.
type WhereClause struct {
	Conditions []*Condition `yaml:"conditions,omitempty"`
}

// Condition represents a single filter condition.
type Condition struct {
	Field    string      `yaml:"field"`
	Operator Operator    `yaml:"operator"`
	Value    interface{} `yaml:"value"`
}

// Operator represents a comparison operator.
type Operator string

const (
	OperatorEq    Operator = "eq"
	OperatorNe    Operator = "ne"
	OperatorIn    Operator = "in"
	OperatorGt    Operator = "gt"
	OperatorLt    Operator = "lt"
	OperatorGe    Operator = "ge"
	OperatorLe    Operator = "le"
	OperatorContains Operator = "contains"
)

// Aggregation defines an aggregation function.
type Aggregation struct {
	Property       string           `yaml:"property"`
	Function       AggFunction      `yaml:"function"`
	TargetProperty string           `yaml:"target_property"`
}

// AggFunction represents an aggregation function.
type AggFunction string

const (
	AggCount AggFunction = "count"
	AggSum   AggFunction = "sum"
	AggAvg   AggFunction = "avg"
	AggMin   AggFunction = "min"
	AggMax   AggFunction = "max"
	AggList  AggFunction = "list"
	AggFirst AggFunction = "first"
	AggLast  AggFunction = "last"
)

// ExpansionConfig defines expansion configuration.
type ExpansionConfig struct {
	TargetKind     string            `yaml:"target_kind"`
	PropertyMapping map[string]string `yaml:"property_mapping,omitempty"`
	Owner          string            `yaml:"owner,omitempty"`
}

// Annotation defines an annotation to attach.
type Annotation struct {
	Property       string `yaml:"property"`
	Value          interface{} `yaml:"value,omitempty"`
	Expression     string `yaml:"expression,omitempty"`
	SourceProperty string `yaml:"source_property,omitempty"`
}

// ComputedProperty defines a property to add.
type ComputedProperty struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
	Expression string `yaml:"expression"`
}

// Transformation defines a property transformation.
type Transformation struct {
	Property  string      `yaml:"property"`
	Operation TransfOp    `yaml:"operation"`
	Value     interface{} `yaml:"value,omitempty"`
}

// TransfOp represents a transformation operation.
type TransfOp string

const (
	TransfRename TransfOp = "rename"
	TransfCast   TransfOp = "cast"
	TransfSet    TransfOp = "set"
	TransfRemove TransfOp = "remove"
	TransfDefault TransfOp = "default"
)

// Provenance tracks the source of a derived object.
type Provenance struct {
	SourceIDs    []string `yaml:"source_ids"`
	ProjectionID string   `yaml:"projection_id"`
	Timestamp    string   `yaml:"timestamp"`
	Operation    string   `yaml:"operation"`
}

// ProjectionResult represents the result of a projection execution.
type ProjectionResult struct {
	ProjectionID string         `yaml:"projection_id"`
	Graph        *core.Graph    `yaml:"graph"`
	DerivedCount int            `yaml:"derived_count"`
}

// NewProjection creates a new Projection.
func NewProjection(id, name string) *Projection {
	return &Projection{
		ID:         id,
		Name:       name,
		Operations: make([]*Operation, 0),
	}
}

// NewInputClause creates a new InputClause.
func NewInputClause(inputType InputType) *InputClause {
	return &InputClause{Type: inputType}
}

// NewOutputClause creates a new OutputClause.
func NewOutputClause() *OutputClause {
	return &OutputClause{}
}

// NewOperation creates a new Operation.
func NewOperation(opType OperationType) *Operation {
	return &Operation{Type: opType}
}

// NewEntitySelector creates a new EntitySelector.
func NewEntitySelector(kind core.EntityKind) *EntitySelector {
	return &EntitySelector{Kind: kind}
}

// NewRelationSelector creates a new RelationSelector.
func NewRelationSelector(relType core.RelationType) *RelationSelector {
	return &RelationSelector{Type: relType}
}

// NewSelector creates a new Selector.
func NewSelector() *Selector {
	return &Selector{}
}

// NewWhereClause creates a new WhereClause.
func NewWhereClause() *WhereClause {
	return &WhereClause{}
}

// NewCondition creates a new Condition.
func NewCondition(field string, op Operator, value interface{}) *Condition {
	return &Condition{Field: field, Operator: op, Value: value}
}

// AddCondition adds a condition to the where clause.
func (w *WhereClause) AddCondition(field string, op Operator, value interface{}) *Condition {
	cond := NewCondition(field, op, value)
	w.Conditions = append(w.Conditions, cond)
	return cond
}

// NewAggregation creates a new Aggregation.
func NewAggregation(property string, function AggFunction, targetProperty string) *Aggregation {
	return &Aggregation{
		Property:       property,
		Function:       function,
		TargetProperty: targetProperty,
	}
}

// NewExpansionConfig creates a new ExpansionConfig.
func NewExpansionConfig(targetKind string) *ExpansionConfig {
	return &ExpansionConfig{
		TargetKind:      targetKind,
		PropertyMapping: make(map[string]string),
	}
}

// NewAnnotation creates a new Annotation.
func NewAnnotation(property string) *Annotation {
	return &Annotation{Property: property}
}

// NewComputedProperty creates a new ComputedProperty.
func NewComputedProperty(name, propType, expression string) *ComputedProperty {
	return &ComputedProperty{
		Name:       name,
		Type:       propType,
		Expression: expression,
	}
}

// NewTransformation creates a new Transformation.
func NewTransformation(property string, op TransfOp) *Transformation {
	return &Transformation{Property: property, Operation: op}
}

// NewProvenance creates a new Provenance.
func NewProvenance(sourceIDs []string, projectionID, operation string) *Provenance {
	return &Provenance{
		SourceIDs:    sourceIDs,
		ProjectionID: projectionID,
		Operation:    operation,
	}
}

// AddEntity adds an entity selector to the operation.
func (o *Operation) AddEntity(kind core.EntityKind) *EntitySelector {
	sel := NewEntitySelector(kind)
	o.Entities = append(o.Entities, sel)
	return sel
}

// AddRelation adds a relation selector to the operation.
func (o *Operation) AddRelation(relType core.RelationType) *RelationSelector {
	sel := NewRelationSelector(relType)
	o.Relations = append(o.Relations, sel)
	return sel
}
