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
	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	g.AddEntity(region)
	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 01")
	rack.SetOwner("region-01")
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
		"dangling-reference", "invalid-path",
		"valid-ip-format", "ip-requires-network", "network-reference-kind",
		"ip-in-cidr", "network-cidr-required", "gateway-in-cidr",
		"ip-unique-in-network",
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

	region := core.NewEntity("dup-01", kinds.Region, "Region 1")
	graph.AddEntity(region)

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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	graph.AddEntity(server)

	r1 := core.NewDirectedRelation("rel-01", types.Hosts, "srv-01", "region-01")
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

	region := core.NewEntity("region-01", "", "Region 1")
	graph.ForceAddEntity(region)

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

	region := core.NewEntity("region-01", kinds.Region, "")
	graph.ForceAddEntity(region)

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

	region := core.NewEntity("region-01", "nonexistent_kind", "Region 1")
	graph.ForceAddEntity(region)

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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	region.SetStatus("invalid_status")
	graph.AddEntity(region)

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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	acl := core.NewEntity("acl-01", kinds.ACL, "ACL 1")
	acl.SetOwner("region-01")
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	graph.AddEntity(server)

	r := core.NewRelation("rel-01", "", core.DirectionDirected)
	r.Participants.Source = "srv-01"
	r.Participants.Target = "region-01"
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	graph.AddEntity(server)

	r := core.NewDirectedRelation("rel-01", "nonexistent_type", "srv-01", "region-01")
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	graph.AddEntity(server)

	// connects should have interface participants, not server
	r := core.NewDirectedRelation("rel-01", types.Connects, "srv-01", "region-01")
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

	root1 := core.NewEntity("root-1", kinds.Region, "Root 1")
	graph.AddEntity(root1)
	root2 := core.NewEntity("root-2", kinds.Region, "Root 2")
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

func TestValidatePathReferenceParticipant(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	graph.AddEntity(server)
	network := core.NewEntity("net-01", kinds.Network, "Network 1")
	network.SetOwner("region-01")
	graph.AddEntity(network)
	intf := core.NewEntity("eth0", kinds.Interface, "eth0")
	intf.SetOwner("srv-01")
	graph.AddEntity(intf)

	r := core.NewDirectedRelation("rel-01", types.BelongsTo, "srv-01/eth0", "net-01")
	graph.AddRelation(r)

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if (f.RuleID == "valid-reference" || f.RuleID == "dangling-reference") && f.Severity == SeverityError {
			t.Errorf("unexpected %s error: %s", f.RuleID, f.Message)
		}
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 1")
	server.SetOwner("region-01")
	graph.AddEntity(server)

	r := core.NewDirectedRelation("rel-01", types.Hosts, "srv-01", "region-01")
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

	region := core.NewEntity("region-01", kinds.Region, "Region 1")
	graph.AddEntity(region)

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
	graph.AddEntity(core.NewEntity("region-01", kinds.Region, "Region 1"))

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

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	graph.AddEntity(region)

	net := core.NewEntity("net-mgmt", kinds.Network, "Management Network")
	net.SetOwner("region-01")
	graph.AddEntity(net)

	vlan := core.NewEntity("vlan-100", kinds.VLAN, "VLAN 100")
	vlan.SetOwner("region-01")
	vlan.SetProperty("associated_network", core.NewReferenceValue("@net-mgmt"))
	graph.AddEntity(vlan)

	vlan2 := core.NewEntity("vlan-200", kinds.VLAN, "VLAN 200")
	vlan2.SetOwner("region-01")
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

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	graph.AddEntity(region)

	net := core.NewEntity("net-mgmt", kinds.Network, "Management Network")
	net.SetOwner("region-01")
	graph.AddEntity(net)

	vlan := core.NewEntity("vlan-100", kinds.VLAN, "VLAN 100")
	vlan.SetOwner("region-01")
	vlan.SetProperty("associated_network", core.NewReferenceValue("@net-mgmt"))
	graph.AddEntity(vlan)

	result := e.Validate(graph, nil)

	for _, f := range result.Findings {
		if f.RuleID == "dangling-reference" && f.ObjectID == "vlan-100" {
			t.Error("valid property reference should not cause dangling-reference error")
		}
	}
}

func TestInvalidPathNonExistentEntity(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	region.SetPath("/region-01")
	graph.AddEntity(region)

	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("region-01")
	server.SetPath("/region-01/nonexistent/srv-01")
	graph.AddEntity(server)

	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected invalid-path error for path referencing non-existent entity")
	}
}

func TestInvalidPathWrongOwnership(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	region.SetPath("/region-01")
	graph.AddEntity(region)

	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 01")
	rack.SetOwner("region-01")
	rack.SetPath("/region-01/rack-01")
	graph.AddEntity(rack)

	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("rack-01")
	server.SetPath("/region-01/rack-01/srv-01")
	graph.AddEntity(server)

	// Manually set wrong path (not matching ownership)
	server.SetPath("/region-01/srv-01")
	result := e.Validate(graph, nil)
	found := false
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected invalid-path error for path with wrong ownership")
	}
}

func TestValidPath(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()

	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	region.SetPath("/region-01")
	graph.AddEntity(region)

	rack := core.NewEntity("rack-01", kinds.Rack, "Rack 01")
	rack.SetOwner("region-01")
	rack.SetPath("/region-01/rack-01")
	graph.AddEntity(rack)

	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("rack-01")
	server.SetPath("/region-01/rack-01/srv-01")
	graph.AddEntity(server)

	result := e.Validate(graph, nil)
	for _, f := range result.Findings {
		if f.RuleID == "invalid-path" && f.Severity == SeverityError {
			t.Errorf("unexpected invalid-path error: %s", f.Message)
		}
	}
}

// --- Network Rules ---

func newNetworkGraph() *core.Graph {
	g := core.NewGraph()
	region := core.NewEntity("region-01", kinds.Region, "Region 01")
	g.AddEntity(region)
	server := core.NewEntity("srv-01", kinds.Server, "Server 01")
	server.SetOwner("region-01")
	g.AddEntity(server)
	net := core.NewEntity("net-01", kinds.Network, "Network 01")
	net.SetOwner("region-01")
	net.SetProperty("cidr", "10.0.1.0/24")
	g.AddEntity(net)
	return g
}

func newServerInterface(graph *core.Graph, id string) *core.Entity {
	intf := core.NewEntity(id, kinds.Interface, id)
	intf.SetOwner("srv-01")
	graph.AddEntity(intf)
	return intf
}

func hasFindingByRule(result *Result, ruleID string) bool {
	for _, f := range result.Findings {
		if f.RuleID == ruleID {
			return true
		}
	}
	return false
}

func TestRuleValidIPFormat(t *testing.T) {
	e := newTestEngine()
	graph := core.NewGraph()
	graph.AddEntity(core.NewEntity("region-01", kinds.Region, "Region 01"))

	valid := core.NewEntity("eth0", kinds.Interface, "eth0")
	valid.SetOwner("region-01")
	valid.SetProperty("ip_address", "10.0.1.10")
	graph.AddEntity(valid)

	invalid := core.NewEntity("eth1", kinds.Interface, "eth1")
	invalid.SetOwner("region-01")
	invalid.SetProperty("ip_address", []interface{}{"10.0.1.10", "not-an-ip"})
	graph.AddEntity(invalid)

	result := e.Validate(graph, nil)
	if hasFindingByRule(result, "valid-ip-format") == false {
		t.Error("expected valid-ip-format warning for invalid IP address")
	}
	for _, f := range result.Findings {
		if f.RuleID == "valid-ip-format" && f.ObjectID == "eth0" {
			t.Error("eth0 should not have valid-ip-format warning")
		}
		if f.RuleID == "valid-ip-format" && f.Severity != SeverityWarning {
			t.Errorf("valid-ip-format should be warning severity, got %s", f.Severity)
		}
	}
}

func TestRuleIPRequiresNetwork(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	withNetwork := newServerInterface(graph, "eth0")
	withNetwork.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	withNetwork.SetProperty("network", core.NewReferenceValue("@net-01"))

	withoutNetwork := newServerInterface(graph, "eth1")
	withoutNetwork.SetProperty("ip_address", []interface{}{"10.0.1.11"})

	newServerInterface(graph, "eth2")

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "ip-requires-network") {
		t.Error("expected ip-requires-network warning for interface without network")
	}
	for _, f := range result.Findings {
		if f.RuleID == "ip-requires-network" && f.ObjectID == "eth0" {
			t.Error("eth0 references a network and should not have ip-requires-network warning")
		}
		if f.RuleID == "ip-requires-network" && f.ObjectID == "eth2" {
			t.Error("eth2 has no IP address and should not have ip-requires-network warning")
		}
	}
}

func TestRuleIPRequiresNetworkViaBelongsTo(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	intf := newServerInterface(graph, "eth0")
	intf.SetProperty("ip_address", []interface{}{"10.0.1.10"})

	rel := core.NewDirectedRelation("rel-intf-net", types.BelongsTo, "eth0", "net-01")
	graph.AddRelation(rel)

	result := e.Validate(graph, nil)
	if hasFindingByRule(result, "ip-requires-network") {
		t.Error("interface with belongs_to relation to a network should not have ip-requires-network warning")
	}
}

func TestRuleNetworkReferenceKind(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	// net-01 is referenced correctly (kind network); create a server reference too
	intf := newServerInterface(graph, "eth0")
	intf.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	intf.SetProperty("network", core.NewReferenceValue("@net-01"))

	bad := newServerInterface(graph, "eth1")
	bad.SetProperty("ip_address", []interface{}{"10.0.1.11"})
	bad.SetProperty("network", core.NewReferenceValue("@srv-01"))

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "network-reference-kind") {
		t.Error("expected network-reference-kind warning for network property referencing non-network entity")
	}
	for _, f := range result.Findings {
		if f.RuleID == "network-reference-kind" && f.ObjectID == "eth0" {
			t.Error("eth0 references a network and should not have network-reference-kind warning")
		}
	}
}

func TestRuleIPInCIDR(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	inRange := newServerInterface(graph, "eth0")
	inRange.SetProperty("ip_address", []interface{}{"10.0.1.10/24"})
	inRange.SetProperty("network", core.NewReferenceValue("@net-01"))

	outOfRange := newServerInterface(graph, "eth1")
	outOfRange.SetProperty("ip_address", []interface{}{"192.168.0.10"})
	outOfRange.SetProperty("network", core.NewReferenceValue("@net-01"))

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "ip-in-cidr") {
		t.Error("expected ip-in-cidr warning for IP outside network CIDR")
	}
	for _, f := range result.Findings {
		if f.RuleID == "ip-in-cidr" && f.ObjectID == "eth0" {
			t.Error("eth0 IP is within network CIDR and should not have ip-in-cidr warning")
		}
	}
}

func TestRuleNetworkCIDRRequired(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	net2 := core.NewEntity("net-02", kinds.Network, "Network 02")
	net2.SetOwner("region-01")
	graph.AddEntity(net2)

	withNetwork := newServerInterface(graph, "eth0")
	withNetwork.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	withNetwork.SetProperty("network", core.NewReferenceValue("@net-01"))

	noCidr := newServerInterface(graph, "eth1")
	noCidr.SetProperty("ip_address", []interface{}{"10.0.2.10"})
	noCidr.SetProperty("network", core.NewReferenceValue("@net-02"))

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "network-cidr-required") {
		t.Error("expected network-cidr-required warning for network without cidr that has IP members")
	}
	for _, f := range result.Findings {
		if f.RuleID == "network-cidr-required" && f.ObjectID == "net-01" {
			t.Error("net-01 defines a cidr and should not have network-cidr-required warning")
		}
	}
}

func TestRuleGatewayInCIDR(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	good := core.NewEntity("net-good", kinds.Network, "Good Network")
	good.SetOwner("region-01")
	good.SetProperty("cidr", "10.0.1.0/24")
	good.SetProperty("gateway", "10.0.1.1")
	graph.AddEntity(good)

	bad := core.NewEntity("net-bad", kinds.Network, "Bad Network")
	bad.SetOwner("region-01")
	bad.SetProperty("cidr", "10.0.1.0/24")
	bad.SetProperty("gateway", "192.168.0.1")
	graph.AddEntity(bad)

	invalid := core.NewEntity("net-invalid", kinds.Network, "Invalid Gateway Network")
	invalid.SetOwner("region-01")
	invalid.SetProperty("cidr", "10.0.1.0/24")
	invalid.SetProperty("gateway", "not-an-ip")
	graph.AddEntity(invalid)

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "gateway-in-cidr") {
		t.Error("expected gateway-in-cidr warning for gateway outside CIDR or invalid gateway")
	}
	for _, f := range result.Findings {
		if f.RuleID == "gateway-in-cidr" && f.ObjectID == "net-good" {
			t.Error("net-good gateway is within CIDR and should not have gateway-in-cidr warning")
		}
	}
}

func TestRuleIPUniqueInNetwork(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	intf1 := newServerInterface(graph, "eth0")
	intf1.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	intf1.SetProperty("network", core.NewReferenceValue("@net-01"))

	intf2 := newServerInterface(graph, "eth1")
	intf2.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	intf2.SetProperty("network", core.NewReferenceValue("@net-01"))

	intf3 := newServerInterface(graph, "eth2")
	intf3.SetProperty("ip_address", []interface{}{"10.0.1.11"})
	intf3.SetProperty("network", core.NewReferenceValue("@net-01"))

	result := e.Validate(graph, nil)
	if !hasFindingByRule(result, "ip-unique-in-network") {
		t.Error("expected ip-unique-in-network warning for duplicate IP within a network")
	}
}

func TestNetworkRulesNoWarningsForCompliantGraph(t *testing.T) {
	e := newTestEngine()
	graph := newNetworkGraph()

	intf := newServerInterface(graph, "eth0")
	intf.SetProperty("ip_address", []interface{}{"10.0.1.10"})
	intf.SetProperty("network", core.NewReferenceValue("@net-01"))

	result := e.Validate(graph, nil)
	networkRules := map[string]bool{
		"valid-ip-format": false, "ip-requires-network": false,
		"network-reference-kind": false, "ip-in-cidr": false,
		"network-cidr-required": false, "gateway-in-cidr": false,
		"ip-unique-in-network": false,
	}
	for _, f := range result.Findings {
		if _, ok := networkRules[f.RuleID]; ok {
			t.Errorf("unexpected %s warning: %s", f.RuleID, f.Message)
		}
	}
}
