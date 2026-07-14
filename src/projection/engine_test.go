package projection

import (
	"fmt"
	"testing"

	"IACForge/src/core"
)

// createTestGraph creates a test graph with entities and relations for testing.
func createTestGraph() *core.Graph {
	g := core.NewGraph()

	// Site
	site := core.NewEntity("site-01", "site", "Data Center 01")
	site.SetStatus(core.StatusActive)
	g.AddEntity(site)

	// Rack
	rack := core.NewEntity("rack-01", "rack", "Rack A01")
	rack.SetOwner("site-01")
	rack.SetStatus(core.StatusActive)
	rack.SetLabel("zone", "A")
	g.AddEntity(rack)

	// Servers
	srv1 := core.NewEntity("srv-01", "server", "Server 01")
	srv1.SetOwner("rack-01")
	srv1.SetStatus(core.StatusActive)
	srv1.SetProperty("cpu_cores", 16)
	srv1.SetProperty("memory_gb", 64)
	g.AddEntity(srv1)

	srv2 := core.NewEntity("srv-02", "server", "Server 02")
	srv2.SetOwner("rack-01")
	srv2.SetStatus(core.StatusActive)
	srv2.SetProperty("cpu_cores", 32)
	srv2.SetProperty("memory_gb", 128)
	g.AddEntity(srv2)

	srv3 := core.NewEntity("srv-03", "server", "Server 03")
	srv3.SetOwner("rack-01")
	srv3.SetStatus(core.StatusMaintenance)
	srv3.SetProperty("cpu_cores", 8)
	srv3.SetProperty("memory_gb", 32)
	g.AddEntity(srv3)

	// VMs
	vm1 := core.NewEntity("vm-01", "vm", "VM 01")
	vm1.SetOwner("srv-01")
	vm1.SetStatus(core.StatusActive)
	vm1.SetProperty("cpu_cores", 4)
	vm1.SetProperty("memory_gb", 8)
	g.AddEntity(vm1)

	vm2 := core.NewEntity("vm-02", "vm", "VM 02")
	vm2.SetOwner("srv-02")
	vm2.SetStatus(core.StatusActive)
	vm2.SetProperty("cpu_cores", 8)
	vm2.SetProperty("memory_gb", 16)
	g.AddEntity(vm2)

	// Network
	net := core.NewEntity("net-01", "network", "Production Network")
	net.SetOwner("site-01")
	net.SetStatus(core.StatusActive)
	g.AddEntity(net)

	// Switch
	sw := core.NewEntity("sw-01", "switch", "Switch 01")
	sw.SetOwner("rack-01")
	sw.SetStatus(core.StatusActive)
	g.AddEntity(sw)

	// Relations
	r1 := core.NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	g.AddRelation(r1)

	r2 := core.NewDirectedRelation("rel-02", "hosts", "srv-02", "vm-02")
	g.AddRelation(r2)

	r3 := core.NewDirectedRelation("rel-03", "connects", "sw-01", "srv-01")
	g.AddRelation(r3)

	r4 := core.NewDirectedRelation("rel-04", "connects", "sw-01", "srv-02")
	g.AddRelation(r4)

	r5 := core.NewDirectedRelation("rel-05", "belongs_to", "sw-01", "net-01")
	g.AddRelation(r5)

	return g
}

// Test types creation
func TestNewProjection(t *testing.T) {
	proj := NewProjection("test-proj", "Test Projection")
	if proj.ID != "test-proj" {
		t.Errorf("expected ID 'test-proj', got '%s'", proj.ID)
	}
	if proj.Name != "Test Projection" {
		t.Errorf("expected Name 'Test Projection', got '%s'", proj.Name)
	}
	if len(proj.Operations) != 0 {
		t.Errorf("expected empty operations, got %d", len(proj.Operations))
	}
}

func TestNewInputClause(t *testing.T) {
	input := NewInputClause(InputTypeGraph)
	if input.Type != InputTypeGraph {
		t.Errorf("expected type 'graph', got '%s'", input.Type)
	}
}

func TestNewOutputClause(t *testing.T) {
	output := NewOutputClause()
	if output == nil {
		t.Error("expected non-nil output")
	}
}

func TestNewOperation(t *testing.T) {
	op := NewOperation(OperationSelect)
	if op.Type != OperationSelect {
		t.Errorf("expected type 'select', got '%s'", op.Type)
	}
}

func TestNewEntitySelector(t *testing.T) {
	sel := NewEntitySelector("server")
	if sel.Kind != "server" {
		t.Errorf("expected kind 'server', got '%s'", sel.Kind)
	}
}

func TestNewRelationSelector(t *testing.T) {
	sel := NewRelationSelector("hosts")
	if sel.Type != "hosts" {
		t.Errorf("expected type 'hosts', got '%s'", sel.Type)
	}
}

func TestNewWhereClause(t *testing.T) {
	w := NewWhereClause()
	if w == nil {
		t.Error("expected non-nil where clause")
	}
}

func TestNewCondition(t *testing.T) {
	c := NewCondition("status", OperatorEq, "active")
	if c.Field != "status" {
		t.Errorf("expected field 'status', got '%s'", c.Field)
	}
	if c.Operator != OperatorEq {
		t.Errorf("expected operator 'eq', got '%s'", c.Operator)
	}
	if c.Value != "active" {
		t.Errorf("expected value 'active', got '%v'", c.Value)
	}
}

func TestWhereClauseAddCondition(t *testing.T) {
	w := NewWhereClause()
	c := w.AddCondition("status", OperatorEq, "active")
	if len(w.Conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(w.Conditions))
	}
	if c.Field != "status" {
		t.Errorf("expected field 'status', got '%s'", c.Field)
	}
}

func TestNewAggregation(t *testing.T) {
	agg := NewAggregation("cpu_cores", AggSum, "total_cpu")
	if agg.Property != "cpu_cores" {
		t.Errorf("expected property 'cpu_cores', got '%s'", agg.Property)
	}
	if agg.Function != AggSum {
		t.Errorf("expected function 'sum', got '%s'", agg.Function)
	}
	if agg.TargetProperty != "total_cpu" {
		t.Errorf("expected target_property 'total_cpu', got '%s'", agg.TargetProperty)
	}
}

func TestNewExpansionConfig(t *testing.T) {
	exp := NewExpansionConfig("vlan")
	if exp.TargetKind != "vlan" {
		t.Errorf("expected target_kind 'vlan', got '%s'", exp.TargetKind)
	}
	if exp.PropertyMapping == nil {
		t.Error("expected non-nil property mapping")
	}
}

func TestNewAnnotation(t *testing.T) {
	ann := NewAnnotation("server_count")
	if ann.Property != "server_count" {
		t.Errorf("expected property 'server_count', got '%s'", ann.Property)
	}
}

func TestNewComputedProperty(t *testing.T) {
	cp := NewComputedProperty("cpu_util", "number", "(used / total) * 100")
	if cp.Name != "cpu_util" {
		t.Errorf("expected name 'cpu_util', got '%s'", cp.Name)
	}
	if cp.Type != "number" {
		t.Errorf("expected type 'number', got '%s'", cp.Type)
	}
}

func TestNewTransformation(t *testing.T) {
	tr := NewTransformation("name", TransfRename)
	if tr.Property != "name" {
		t.Errorf("expected property 'name', got '%s'", tr.Property)
	}
	if tr.Operation != TransfRename {
		t.Errorf("expected operation 'rename', got '%s'", tr.Operation)
	}
}

func TestNewProvenance(t *testing.T) {
	prov := NewProvenance([]string{"srv-01"}, "proj-01", "aggregate")
	if len(prov.SourceIDs) != 1 {
		t.Errorf("expected 1 source ID, got %d", len(prov.SourceIDs))
	}
	if prov.ProjectionID != "proj-01" {
		t.Errorf("expected projection_id 'proj-01', got '%s'", prov.ProjectionID)
	}
}

func TestOperationAddEntity(t *testing.T) {
	op := NewOperation(OperationSelect)
	sel := op.AddEntity("server")
	if len(op.Entities) != 1 {
		t.Errorf("expected 1 entity selector, got %d", len(op.Entities))
	}
	if sel.Kind != "server" {
		t.Errorf("expected kind 'server', got '%s'", sel.Kind)
	}
}

func TestOperationAddRelation(t *testing.T) {
	op := NewOperation(OperationSelect)
	sel := op.AddRelation("hosts")
	if len(op.Relations) != 1 {
		t.Errorf("expected 1 relation selector, got %d", len(op.Relations))
	}
	if sel.Type != "hosts" {
		t.Errorf("expected type 'hosts', got '%s'", sel.Type)
	}
}

// Test DerivedObjectManager
func TestDerivedObjectManager(t *testing.T) {
	mgr := NewDerivedObjectManager()

	e := mgr.CreateDerivedEntity("rack_summary", "Rack Summary", []string{"srv-01", "srv-02"}, "proj-01", "aggregate")

	if e == nil {
		t.Fatal("expected non-nil entity")
	}
	if e.Kind != "rack_summary" {
		t.Errorf("expected kind 'rack_summary', got '%s'", e.Kind)
	}
	if mgr.Count() != 1 {
		t.Errorf("expected 1 derived object, got %d", mgr.Count())
	}

	if !IsDerived(e) {
		t.Error("expected entity to be derived")
	}

	prov, ok := GetProvenance(e)
	if !ok {
		t.Fatal("expected provenance to exist")
	}
	if len(prov.SourceIDs) != 2 {
		t.Errorf("expected 2 source IDs, got %d", len(prov.SourceIDs))
	}
	if prov.ProjectionID != "proj-01" {
		t.Errorf("expected projection_id 'proj-01', got '%s'", prov.ProjectionID)
	}
}

func TestDerivedObjectManagerWithProperties(t *testing.T) {
	mgr := NewDerivedObjectManager()

	props := map[string]interface{}{
		"total_cpu": 48,
		"count":     3,
	}
	e := mgr.CreateDerivedEntityWithProperties("rack_summary", "Rack Summary", props, []string{"srv-01"}, "proj-01", "aggregate")

	if e == nil {
		t.Fatal("expected non-nil entity")
	}
	if v, ok := e.GetProperty("total_cpu"); !ok || v != 48 {
		t.Errorf("expected total_cpu 48, got %v", v)
	}
}

func TestGetProvenanceNoMetadata(t *testing.T) {
	e := core.NewEntity("test", "server", "Test")
	_, ok := GetProvenance(e)
	if ok {
		t.Error("expected no provenance for entity without metadata")
	}
}

func TestIsDerivedNotDerived(t *testing.T) {
	e := core.NewEntity("test", "server", "Test")
	if IsDerived(e) {
		t.Error("expected entity to not be derived")
	}
}

// Test Engine
func TestEngineRegister(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("proj-01", "Test Projection")
	proj.Input = NewInputClause(InputTypeGraph)
	proj.Operations = append(proj.Operations, NewOperation(OperationSelect))

	engine.Register(proj)

	projs := engine.ListProjections()
	if len(projs) != 1 {
		t.Errorf("expected 1 projection, got %d", len(projs))
	}

	p, ok := engine.GetProjection("proj-01")
	if !ok {
		t.Error("expected to find projection")
	}
	if p.Name != "Test Projection" {
		t.Errorf("expected name 'Test Projection', got '%s'", p.Name)
	}
}

func TestEngineExecuteNoInput(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("proj-01", "Test")
	_, err := engine.Execute(proj)
	if err == nil {
		t.Error("expected error for projection without input")
	}
}

// Test Select Operation
func TestSelectOperation(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("select-servers", "Select Servers")
	proj.Input = NewInputClause(InputTypeGraph)
	op := NewOperation(OperationSelect)
	op.AddEntity("server")
	proj.Operations = append(proj.Operations, op)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Graph.EntityCount() != 3 {
		t.Errorf("expected 3 servers, got %d", result.Graph.EntityCount())
	}
}

func TestSelectOperationWithWhere(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("select-active-servers", "Select Active Servers")
	proj.Input = NewInputClause(InputTypeGraph)
	op := NewOperation(OperationSelect)
	sel := op.AddEntity("server")
	sel.Where = NewWhereClause()
	sel.Where.AddCondition("status", OperatorEq, "active")
	proj.Operations = append(proj.Operations, op)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Graph.EntityCount() != 2 {
		t.Errorf("expected 2 active servers, got %d", result.Graph.EntityCount())
	}
}

func TestSelectOperationRelations(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("select-hosts", "Select Hosts Relations")
	proj.Input = NewInputClause(InputTypeGraph)
	op := NewOperation(OperationSelect)
	op.AddRelation("hosts")
	proj.Operations = append(proj.Operations, op)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Graph.RelationCount() != 2 {
		t.Errorf("expected 2 hosts relations, got %d", result.Graph.RelationCount())
	}
}

// Test Filter Operation
func TestFilterOperationInclude(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("filter-servers", "Filter Servers")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	op1.AddEntity("rack")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationFilter)
	op2.Action = FilterActionInclude
	op2.Target = FilterTargetEntities
	op2.Where = NewWhereClause()
	op2.Where.AddCondition("kind", OperatorEq, "server")
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Graph.EntityCount() != 3 {
		t.Errorf("expected 3 servers after filter, got %d", result.Graph.EntityCount())
	}
}

func TestFilterOperationExclude(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("filter-exclude-maintenance", "Exclude Maintenance")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationFilter)
	op2.Action = FilterActionExclude
	op2.Target = FilterTargetEntities
	op2.Where = NewWhereClause()
	op2.Where.AddCondition("status", OperatorEq, "maintenance")
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Graph.EntityCount() != 2 {
		t.Errorf("expected 2 active servers after exclude, got %d", result.Graph.EntityCount())
	}
}

// Test Traverse Operation
func TestTraverseOperation(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("traverse-vm", "Traverse VMs")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	sel := op1.AddEntity("server")
	sel.Where = NewWhereClause()
	sel.Where.AddCondition("id", OperatorEq, "srv-01")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationTraverse)
	op2.Direction = TraverseDirectionForward
	op2.RelationType = "hosts"
	op2.IncludeOrigin = boolPtr(true)
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Graph.EntityCount() < 2 {
		t.Errorf("expected at least 2 entities after traverse, got %d", result.Graph.EntityCount())
	}
}

func TestTraverseOperationOwnership(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("traverse-ownership", "Traverse Ownership")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	sel := op1.AddEntity("rack")
	sel.Where = NewWhereClause()
	sel.Where.AddCondition("id", OperatorEq, "rack-01")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationTraverse)
	op2.Direction = TraverseDirectionForward
	ownerTrue := true
	op2.Owner = &ownerTrue
	op2.Depth = 2
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Graph.EntityCount() < 2 {
		t.Errorf("expected at least 2 entities after ownership traverse, got %d", result.Graph.EntityCount())
	}
}

// Test Aggregate Operation
func TestAggregateOperation(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("aggregate-servers", "Aggregate Servers")
	proj.Input = NewInputClause(InputTypeGraph)

	op := NewOperation(OperationAggregate)
	op.SourceSelector = NewSelector()
	op.SourceSelector.Kind = "server"
	op.TargetKind = "rack_summary"
	op.GroupBy = []string{"owner"}
	op.Aggregations = []*Aggregation{
		NewAggregation("cpu_cores", AggSum, "total_cpu_cores"),
		NewAggregation("memory_gb", AggSum, "total_memory_gb"),
		NewAggregation("id", AggCount, "server_count"),
	}
	proj.Operations = append(proj.Operations, op)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DerivedCount != 1 {
		t.Errorf("expected 1 derived object, got %d", result.DerivedCount)
	}

	rackSummary, ok := result.Graph.GetEntity("derived-rack_summary-1")
	if !ok {
		t.Fatal("expected to find rack_summary entity")
	}

	if v, ok := rackSummary.GetProperty("total_cpu_cores"); !ok {
		t.Error("expected total_cpu_cores property")
	} else if v != float64(56) {
		t.Errorf("expected total_cpu_cores 56, got %v", v)
	}
}

// Test Expand Operation
func TestExpandOperation(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("expand-network", "Expand Network")
	proj.Input = NewInputClause(InputTypeGraph)

	op := NewOperation(OperationExpand)
	op.SourceSelector = NewSelector()
	op.SourceSelector.Kind = "network"
	op.Expansion = NewExpansionConfig("vlan")
	op.Expansion.PropertyMapping = map[string]string{
		"name": "vlan_name",
		"id":   "vlan_id",
	}
	proj.Operations = append(proj.Operations, op)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DerivedCount != 1 {
		t.Errorf("expected 1 derived object, got %d", result.DerivedCount)
	}
}

// Test Annotate Operation
func TestAnnotateOperation(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("annotate-servers", "Annotate Servers")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationAnnotate)
	op2.TargetSelector = NewSelector()
	op2.TargetSelector.Kind = "server"
	op2.Annotations = []*Annotation{
		{Property: "annotation_type", Value: "server_annotation"},
	}
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range result.Graph.Entities() {
		if v, ok := e.GetProperty("annotation_type"); !ok || v != "server_annotation" {
			t.Errorf("expected annotation_type 'server_annotation' on %s", e.ID)
		}
	}
}

// Test Group Operation
func TestGroupOperation(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("group-servers", "Group Servers")
	proj.Input = NewInputClause(InputTypeGraph)

	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationGroup)
	op2.SourceSelector = NewSelector()
	op2.SourceSelector.Kind = "server"
	op2.GroupKind = "server_group"
	op2.GroupBy = []string{"owner"}
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DerivedCount != 1 {
		t.Errorf("expected 1 group, got %d", result.DerivedCount)
	}
}

// Test Flatten Operation
func TestFlattenOperation(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("flatten-hierarchy", "Flatten Hierarchy")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationFlatten)
	op2.TargetSelector = NewSelector()
	op2.TargetSelector.Kind = "server"
	preserveFalse := false
	op2.PreserveRelations = &preserveFalse
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range result.Graph.Entities() {
		if e.Owner != "" {
			t.Errorf("expected entity %s to have empty owner after flatten", e.ID)
		}
	}
}

// Test Enrich Operation
func TestEnrichOperation(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("enrich-servers", "Enrich Servers")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationEnrich)
	op2.TargetSelector = NewSelector()
	op2.TargetSelector.Kind = "server"
	op2.Properties = []*ComputedProperty{
		NewComputedProperty("server_info", "string", "name"),
	}
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range result.Graph.Entities() {
		if _, ok := e.GetProperty("server_info"); !ok {
			t.Errorf("expected server_info property on %s", e.ID)
		}
	}
}

// Test Transform Operation
func TestTransformOperationSet(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("transform-servers", "Transform Servers")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationTransform)
	op2.TargetSelector = NewSelector()
	op2.TargetSelector.Kind = "server"
	op2.Transformations = []*Transformation{
		NewTransformation("status", TransfSet),
		{Property: "status", Operation: TransfSet, Value: "updated"},
	}
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range result.Graph.Entities() {
		if v, ok := e.GetProperty("status"); !ok || v != "updated" {
			t.Errorf("expected status 'updated' on %s, got %v", e.ID, v)
		}
	}
}

func TestTransformOperationDefault(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("transform-default", "Transform Default")
	proj.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	proj.Operations = append(proj.Operations, op1)

	op2 := NewOperation(OperationTransform)
	op2.TargetSelector = NewSelector()
	op2.TargetSelector.Kind = "server"
	op2.Transformations = []*Transformation{
		{Property: "nonexistent", Operation: TransfDefault, Value: "default_value"},
	}
	proj.Operations = append(proj.Operations, op2)

	result, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, e := range result.Graph.Entities() {
		if v, ok := e.GetProperty("nonexistent"); !ok || v != "default_value" {
			t.Errorf("expected nonexistent 'default_value' on %s, got %v", e.ID, v)
		}
	}
}

// Test Composition (chaining)
func TestProjectionComposition(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj1 := NewProjection("step-1", "Step 1 - Select Servers")
	proj1.Input = NewInputClause(InputTypeGraph)
	op1 := NewOperation(OperationSelect)
	op1.AddEntity("server")
	proj1.Operations = append(proj1.Operations, op1)
	engine.Register(proj1)

	proj2 := NewProjection("step-2", "Step 2 - Filter Active")
	proj2.Input = NewInputClause(InputTypeProjection)
	proj2.Input.ProjectionID = "step-1"
	op2 := NewOperation(OperationFilter)
	op2.Action = FilterActionInclude
	op2.Target = FilterTargetEntities
	op2.Where = NewWhereClause()
	op2.Where.AddCondition("status", OperatorEq, "active")
	proj2.Operations = append(proj2.Operations, op2)
	engine.Register(proj2)

	result, err := engine.Execute(proj2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Graph.EntityCount() != 2 {
		t.Errorf("expected 2 active servers after composition, got %d", result.Graph.EntityCount())
	}
}

// Test Determinism
func TestProjectionDeterminism(t *testing.T) {
	g := createTestGraph()
	engine := NewEngine(g)

	proj := NewProjection("deterministic", "Deterministic Projection")
	proj.Input = NewInputClause(InputTypeGraph)
	op := NewOperation(OperationSelect)
	op.AddEntity("server")
	proj.Operations = append(proj.Operations, op)

	result1, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result2, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result1.Graph.EntityCount() != result2.Graph.EntityCount() {
		t.Errorf("expected same entity count, got %d and %d",
			result1.Graph.EntityCount(), result2.Graph.EntityCount())
	}
}

// Test Side Effects (source graph unchanged)
func TestProjectionNoSideEffects(t *testing.T) {
	g := createTestGraph()
	originalEntityCount := g.EntityCount()
	originalRelationCount := g.RelationCount()

	engine := NewEngine(g)

	proj := NewProjection("no-effects", "No Side Effects")
	proj.Input = NewInputClause(InputTypeGraph)
	op := NewOperation(OperationSelect)
	op.AddEntity("server")
	proj.Operations = append(proj.Operations, op)

	_, err := engine.Execute(proj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.EntityCount() != originalEntityCount {
		t.Errorf("source graph entity count changed: expected %d, got %d",
			originalEntityCount, g.EntityCount())
	}
	if g.RelationCount() != originalRelationCount {
		t.Errorf("source graph relation count changed: expected %d, got %d",
			originalRelationCount, g.RelationCount())
	}
}

// Test Validation
func TestValidateProjection(t *testing.T) {
	proj := NewProjection("valid", "Valid Projection")
	proj.Input = NewInputClause(InputTypeGraph)
	op := NewOperation(OperationSelect)
	op.AddEntity("server")
	proj.Operations = append(proj.Operations, op)

	err := ValidateProjection(proj)
	if err != nil {
		t.Errorf("expected valid projection, got error: %v", err)
	}
}

func TestValidateProjectionNoID(t *testing.T) {
	proj := NewProjection("", "No ID")
	proj.Input = NewInputClause(InputTypeGraph)
	op := NewOperation(OperationSelect)
	op.AddEntity("server")
	proj.Operations = append(proj.Operations, op)

	err := ValidateProjection(proj)
	if err == nil {
		t.Error("expected error for projection without ID")
	}
}

func TestValidateProjectionNoName(t *testing.T) {
	proj := NewProjection("test", "")
	proj.Input = NewInputClause(InputTypeGraph)
	op := NewOperation(OperationSelect)
	op.AddEntity("server")
	proj.Operations = append(proj.Operations, op)

	err := ValidateProjection(proj)
	if err == nil {
		t.Error("expected error for projection without name")
	}
}

func TestValidateProjectionNoOperations(t *testing.T) {
	proj := NewProjection("test", "Test")
	proj.Input = NewInputClause(InputTypeGraph)

	err := ValidateProjection(proj)
	if err == nil {
		t.Error("expected error for projection without operations")
	}
}

func TestValidateSelectNoEntitiesOrRelations(t *testing.T) {
	op := NewOperation(OperationSelect)
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for select without entities or relations")
	}
}

func TestValidateFilterNoAction(t *testing.T) {
	op := NewOperation(OperationFilter)
	op.Target = FilterTargetEntities
	op.Where = NewWhereClause()
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for filter without action")
	}
}

func TestValidateFilterNoTarget(t *testing.T) {
	op := NewOperation(OperationFilter)
	op.Action = FilterActionInclude
	op.Where = NewWhereClause()
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for filter without target")
	}
}

func TestValidateFilterNoWhere(t *testing.T) {
	op := NewOperation(OperationFilter)
	op.Action = FilterActionInclude
	op.Target = FilterTargetEntities
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for filter without where")
	}
}

func TestValidateTraverseNoDirection(t *testing.T) {
	op := NewOperation(OperationTraverse)
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for traverse without direction")
	}
}

func TestValidateAggregateNoSourceSelector(t *testing.T) {
	op := NewOperation(OperationAggregate)
	op.TargetKind = "summary"
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for aggregate without source_selector")
	}
}

func TestValidateAggregateNoTargetKind(t *testing.T) {
	op := NewOperation(OperationAggregate)
	op.SourceSelector = NewSelector()
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for aggregate without target_kind")
	}
}

func TestValidateExpandNoSourceSelector(t *testing.T) {
	op := NewOperation(OperationExpand)
	op.Expansion = NewExpansionConfig("vlan")
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for expand without source_selector")
	}
}

func TestValidateExpandNoExpansion(t *testing.T) {
	op := NewOperation(OperationExpand)
	op.SourceSelector = NewSelector()
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for expand without expansion")
	}
}

func TestValidateAnnotateNoTargetSelector(t *testing.T) {
	op := NewOperation(OperationAnnotate)
	op.Annotations = []*Annotation{{Property: "test"}}
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for annotate without target_selector")
	}
}

func TestValidateAnnotateNoAnnotations(t *testing.T) {
	op := NewOperation(OperationAnnotate)
	op.TargetSelector = NewSelector()
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for annotate without annotations")
	}
}

func TestValidateGroupNoSourceSelector(t *testing.T) {
	op := NewOperation(OperationGroup)
	op.GroupKind = "group"
	op.GroupBy = []string{"kind"}
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for group without source_selector")
	}
}

func TestValidateGroupNoGroupKind(t *testing.T) {
	op := NewOperation(OperationGroup)
	op.SourceSelector = NewSelector()
	op.GroupBy = []string{"kind"}
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for group without group_kind")
	}
}

func TestValidateGroupNoGroupBy(t *testing.T) {
	op := NewOperation(OperationGroup)
	op.SourceSelector = NewSelector()
	op.GroupKind = "group"
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for group without group_by")
	}
}

func TestValidateFlattenNoTargetSelector(t *testing.T) {
	op := NewOperation(OperationFlatten)
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for flatten without target_selector")
	}
}

func TestValidateEnrichNoTargetSelector(t *testing.T) {
	op := NewOperation(OperationEnrich)
	op.Properties = []*ComputedProperty{{Name: "test"}}
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for enrich without target_selector")
	}
}

func TestValidateEnrichNoProperties(t *testing.T) {
	op := NewOperation(OperationEnrich)
	op.TargetSelector = NewSelector()
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for enrich without properties")
	}
}

func TestValidateTransformNoTargetSelector(t *testing.T) {
	op := NewOperation(OperationTransform)
	op.Transformations = []*Transformation{{Property: "test", Operation: TransfSet}}
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for transform without target_selector")
	}
}

func TestValidateTransformNoTransformations(t *testing.T) {
	op := NewOperation(OperationTransform)
	op.TargetSelector = NewSelector()
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for transform without transformations")
	}
}

func TestValidateUnsupportedOperation(t *testing.T) {
	op := &Operation{Type: "unsupported"}
	err := validateOperation(op)
	if err == nil {
		t.Error("expected error for unsupported operation")
	}
}

// Test helper functions
func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
		ok       bool
	}{
		{42, 42.0, true},
		{int64(100), 100.0, true},
		{3.14, 3.14, true},
		{float32(2.5), 2.5, true},
		{"string", 0, false},
		{nil, 0, false},
	}

	for _, tt := range tests {
		result, ok := toFloat64(tt.input)
		if ok != tt.ok {
			t.Errorf("toFloat64(%v): expected ok=%v, got ok=%v", tt.input, tt.ok, ok)
		}
		if ok && result != tt.expected {
			t.Errorf("toFloat64(%v): expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello world", "xyz", false},
		{"hello", "hello world", false},
		{"", "", true},
	}

	for _, tt := range tests {
		result := contains(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q): expected %v, got %v", tt.s, tt.substr, tt.expected, result)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"-123", true},
		{"3.14", true},
		{"abc", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isNumeric(tt.input)
		if result != tt.expected {
			t.Errorf("isNumeric(%q): expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestCastProperty(t *testing.T) {
	tests := []struct {
		value      interface{}
		castType   string
		expectType string
	}{
		{42, "string", "string"},
		{"42", "integer", "int"},
		{"42", "number", "float64"},
		{42, "unknown", "int"},
	}

	for _, tt := range tests {
		result := castProperty(tt.value, tt.castType)
		typeName := fmt.Sprintf("%T", result)
		if tt.castType == "string" && typeName != "string" {
			t.Errorf("castProperty(%v, %q): expected string, got %s", tt.value, tt.castType, typeName)
		}
	}
}

func boolPtr(b bool) *bool {
	return &b}
