package validation

import (
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
	"IACForge/src/schema"
)

func newTestGraph() *core.Graph {
	g := core.NewGraph()
	site := core.NewEntity("site-01", kinds.Site, "Site 01")
	g.AddEntity(site)
	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 01")
	rack.SetOwner("site-01")
	g.AddEntity(rack)
	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("rack-01")
	g.AddEntity(server)
	return g
}

func newTestEngine() *Engine {
	s := schema.CoreSchema()
	e := NewEngine(s)
	RegisterCoreRules(e)
	return e
}

func TestEngineCoreRulesRegistered(t *testing.T) {
	e := newTestEngine()
	expectedRules := []string{
		"unique-id", "valid-reference", "valid-owner", "single-owner",
		"required-kind", "required-name", "valid-kind", "valid-status",
		"valid-port-range", "valid-acl-rule-parent",
		"required-type", "required-participants", "valid-type", "valid-direction",
		"valid-cardinality", "valid-participant-kind",
		"ownership-tree", "no-ownership-cycle", "root-entity",
		"dangling-reference",
	}

	for _, ruleID := range expectedRules {
		if _, ok := e.ruleDefs[ruleID]; !ok {
			t.Errorf("expected rule %q to be registered", ruleID)
		}
	}
}

func TestValidateValidGraph(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	result := e.Validate(graph, nil)
	if !result.Passed {
		t.Errorf("expected validation to pass, but found errors:")
		for _, f := range result.Findings {
			if f.Severity == SeverityError {
				t.Errorf("  %s: %s", f.RuleID, f.Message)
			}
		}
	}
}

func TestValidateDuplicateEntityID(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("dup-01", kinds.Site, "Site 1")
	graph.AddEntity(site)

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "unique-id" && f.Severity == SeverityError {
			t.Errorf("unexpected unique-id error: %s", f.Message)
		}
	}
}

func TestValidateDuplicateRelationID(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	r1 := core.NewDirectedRelation("rel-01", types.Hosts, "srv-01", "site-01")
	graph.AddRelation(r1)

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "unique-id" && f.ObjectType == ObjectTypeRelation {
			t.Errorf("unexpected unique-id error: %s", f.Message)
		}
	}
}

func TestValidateMissingKind(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", "", "Site 1")
	graph.ForceAddEntity(site)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-kind error")
	}
}

func TestValidateMissingName(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "")
	graph.ForceAddEntity(site)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-name error")
	}
}

func TestValidateInvalidKind(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", "nonexistent_kind", "Site 1")
	graph.ForceAddEntity(site)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-kind error")
	}
}

func TestValidateInvalidStatus(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	site.SetStatus("invalid_status")
	graph.AddEntity(site)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-status" && f.Severity == SeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-status warning")
	}
}

func TestValidateInvalidPortRange(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	port := core.NewEntity("port-01", kinds.OpenPort, "Port 1")
	port.SetOwner("srv-01")
	port.SetProperty("port", 70000)
	graph.AddEntity(port)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-port-range" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-port-range error")
	}
}

func TestValidateValidPortRange(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	port := core.NewEntity("port-01", kinds.OpenPort, "Port 1")
	port.SetOwner("srv-01")
	port.SetProperty("port", 443)
	graph.AddEntity(port)

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-port-range" && f.Severity == SeverityError {
			t.Errorf("unexpected valid-port-range error: %s", f.Message)
		}
	}
}

func TestValidateACLRULEParent(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	acl := core.NewEntity("acl-01", kinds.ACL, "ACL 1")
	acl.SetOwner("site-01")
	graph.AddEntity(acl)

	rule := core.NewEntity("rule-01", kinds.ACLRule, "Rule 1")
	rule.SetOwner("acl-01")
	rule.SetProperty("action", "allow")
	graph.AddEntity(rule)

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "valid-acl-rule-parent" && f.Severity == SeverityError {
			t.Errorf("unexpected valid-acl-rule-parent error: %s", f.Message)
		}
	}
}

func TestValidateACLRULEWrongParent(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	rule := core.NewEntity("rule-01", kinds.ACLRule, "Rule 1")
	rule.SetOwner("srv-01")
	rule.SetProperty("action", "allow")
	graph.AddEntity(rule)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-acl-rule-parent" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-acl-rule-parent error for wrong parent kind")
	}
}

func TestValidateMissingRelationType(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	r := core.NewRelation("rel-01", "", core.DirectionDirected)
	r.Participants.Source = "srv-01"
	r.Participants.Target = "site-01"
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "required-type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected required-type error")
	}
}

func TestValidateInvalidRelationType(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	r := core.NewDirectedRelation("rel-01", "nonexistent_type", "srv-01", "site-01")
	graph.AddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-type error")
	}
}

func TestValidateDirectedRelationMissingTarget(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	r := core.NewRelation("rel-01", types.Hosts, core.DirectionDirected)
	r.Participants.Source = "srv-01"
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-direction" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-direction error for directed relation without target")
	}
}

func TestValidateParticipantKindWarning(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	// connects should have interface participants, not server
	r := core.NewDirectedRelation("rel-01", types.Connects, "srv-01", "site-01")
	graph.AddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "valid-participant-kind" && f.Severity == SeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected valid-participant-kind warning for server in connects relation")
	}
}

func TestValidateOwnershipCycle(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	a := core.NewEntity("a", kinds.Server, "A")
	a.SetOwner("b")
	graph.ForceAddEntity(a)
	b := core.NewEntity("b", kinds.Server, "B")
	b.SetOwner("a")
	graph.ForceAddEntity(b)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "no-ownership-cycle" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected no-ownership-cycle error")
	}
}

func TestValidateMultipleRoots(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	root1 := core.NewEntity("root-1", kinds.Site, "Root 1")
	graph.AddEntity(root1)
	root2 := core.NewEntity("root-2", kinds.Site, "Root 2")
	graph.AddEntity(root2)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected root-entity error for multiple roots")
	}
}

func TestValidateNoRoot(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	a := core.NewEntity("a", kinds.Server, "A")
	a.SetOwner("b")
	graph.ForceAddEntity(a)
	b := core.NewEntity("b", kinds.Server, "B")
	b.SetOwner("a")
	graph.ForceAddEntity(b)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "root-entity" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected root-entity error")
	}
}

func TestValidateDanglingReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	graph.AddEntity(server)

	r := core.NewDirectedRelation("rel-01", types.Hosts, "srv-01", "nonexistent")
	graph.ForceAddRelation(r)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected dangling-reference error")
	}
}

func TestValidateValidRelationParticipants(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("site-01")
	graph.AddEntity(server)

	r := core.NewDirectedRelation("rel-01", types.Hosts, "srv-01", "site-01")
	graph.AddRelation(r)

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "required-participants" && f.Severity == SeverityError {
			t.Errorf("unexpected required-participants error: %s", f.Message)
		}
	}
}

func TestValidateSummary(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	result := e.Validate(graph, nil)

	if result.Summary.TotalRules == 0 {
		t.Error("expected non-zero total rules")
	}
	if result.Summary.TotalFindings != len(result.Findings) {
		t.Errorf("expected summary total findings %d to match findings count %d",
			result.Summary.TotalFindings, len(result.Findings))
	}
}

func TestValidateWithProfile(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	profile := schema.NewProfile("minimal")
	profile.AddRule("required-kind")
	profile.AddRule("required-name")

	result := e.Validate(graph, profile)

	if !result.Passed {
		t.Errorf("expected validation to pass with minimal profile, but found errors:")
		for _, f := range result.Findings {
			if f.Severity == SeverityError {
				t.Errorf("  %s: %s", f.RuleID, f.Message)
			}
		}
	}

	for _, f := range result.Findings {
		if f.RuleID != "required-kind" && f.RuleID != "required-name" {
			t.Errorf("unexpected rule %q found in profile-filtered validation", f.RuleID)
		}
	}
}

func TestValidateProfileRequiredKinds(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 1")
	graph.AddEntity(site)

	profile := schema.NewProfile("with-kinds")
	profile.AddRequiredKind("server")

	result := e.Validate(graph, profile)

	found := false
	for _, f := range result.Findings {
		if f.RuleID == "profile-required-kind" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected profile-required-kind error")
	}
}

func TestValidateProfileRequiredRelations(t *testing.T) {
	e := newTestEngine()
	graph := newTestGraph()

	profile := schema.NewProfile("with-relations")
	profile.AddRequiredRelation("connects")

	result := e.Validate(graph, profile)

	found := false
	for _, f := range result.Findings {
		if f.RuleID == "profile-required-relation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected profile-required-relation error")
	}
}

func TestValidateResultPassed(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	graph.AddEntity(core.NewEntity("site-01", kinds.Site, "Site 1"))

	result := e.Validate(graph, nil)

	errorCount := 0
	for _, f := range result.Findings {
		if f.Severity == SeverityError {
			errorCount++
		}
	}

	if errorCount > 0 && result.Passed {
		t.Error("expected Passed=false when there are errors")
	}
	if errorCount == 0 && !result.Passed {
		t.Error("expected Passed=true when there are no errors")
	}
}

func TestDanglingPropertyReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 01")
	graph.AddEntity(site)

	net := core.NewEntity("net-mgmt", kinds.Network, "Management Network")
	net.SetOwner("site-01")
	graph.AddEntity(net)

	vlan := core.NewEntity("vlan-100", kinds.VLAN, "VLAN 100")
	vlan.SetOwner("site-01")
	vlan.SetProperty("associated_network", core.NewReferenceValue("@net-mgmt"))
	graph.AddEntity(vlan)

	vlan2 := core.NewEntity("vlan-200", kinds.VLAN, "VLAN 200")
	vlan2.SetOwner("site-01")
	vlan2.SetProperty("associated_network", core.NewReferenceValue("@nonexistent"))
	graph.AddEntity(vlan2)

	result := e.Validate(graph, nil)

	foundDangling := false
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "vlan-200" {
			foundDangling = true
			break
		}
	}
	if !foundDangling {
		t.Error("expected dangling-reference error for vlan-200 property reference")
	}

	// vlan-100 should NOT have a dangling reference error
	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "vlan-100" {
			t.Error("vlan-100 should not have dangling-reference error")
		}
	}
}

func TestValidPropertyReference(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	site := core.NewEntity("site-01", kinds.Site, "Site 01")
	graph.AddEntity(site)

	net := core.NewEntity("net-mgmt", kinds.Network, "Management Network")
	net.SetOwner("site-01")
	graph.AddEntity(net)

	vlan := core.NewEntity("vlan-100", kinds.VLAN, "VLAN 100")
	vlan.SetOwner("site-01")
	vlan.SetProperty("associated_network", core.NewReferenceValue("@net-mgmt"))
	graph.AddEntity(vlan)

	result := e.Validate(graph, nil)

	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "vlan-100" {
			t.Error("valid property reference should not cause dangling-reference error")
		}
	}
}
