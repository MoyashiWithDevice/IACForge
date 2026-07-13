package core

import (
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.EntityCount() != 0 {
		t.Errorf("expected 0 entities, got %d", g.EntityCount())
	}
	if g.RelationCount() != 0 {
		t.Errorf("expected 0 relations, got %d", g.RelationCount())
	}
}

func TestGraphAddEntity(t *testing.T) {
	g := NewGraph()
	e := NewEntity("srv-01", "server", "Server 01")
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.EntityCount() != 1 {
		t.Errorf("expected 1 entity, got %d", g.EntityCount())
	}
}

func TestGraphAddEntityDuplicate(t *testing.T) {
	g := NewGraph()
	e1 := NewEntity("srv-01", "server", "Server 01")
	e2 := NewEntity("srv-01", "server", "Server 02")
	g.AddEntity(e1)
	err := g.AddEntity(e2)
	if err == nil {
		t.Error("expected duplicate entity error")
	}
}

func TestGraphAddEntityInvalid(t *testing.T) {
	g := NewGraph()
	e := NewEntity("", "server", "Server 01")
	err := g.AddEntity(e)
	if err == nil {
		t.Error("expected validation error for missing ID")
	}
}

func TestGraphGetEntity(t *testing.T) {
	g := NewGraph()
	e := NewEntity("srv-01", "server", "Server 01")
	g.AddEntity(e)

	found, ok := g.GetEntity("srv-01")
	if !ok || found == nil {
		t.Error("expected to find entity")
	}
	if found.ID != "srv-01" {
		t.Errorf("expected ID srv-01, got %s", found.ID)
	}

	_, ok = g.GetEntity("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent entity")
	}
}

func TestGraphRemoveEntity(t *testing.T) {
	g := NewGraph()
	e := NewEntity("srv-01", "server", "Server 01")
	g.AddEntity(e)

	if !g.RemoveEntity("srv-01") {
		t.Error("expected remove to return true")
	}
	if g.EntityCount() != 0 {
		t.Errorf("expected 0 entities after remove, got %d", g.EntityCount())
	}
	if g.RemoveEntity("srv-01") {
		t.Error("expected remove to return false for non-existent entity")
	}
}

func TestGraphUpdateEntity(t *testing.T) {
	g := NewGraph()
	e := NewEntity("srv-01", "server", "Server 01")
	g.AddEntity(e)

	e.Name = "Updated Server"
	if err := g.UpdateEntity(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found, _ := g.GetEntity("srv-01")
	if found.Name != "Updated Server" {
		t.Errorf("expected name Updated Server, got %s", found.Name)
	}
}

func TestGraphEntities(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))
	g.AddEntity(NewEntity("vm-01", "vm", "VM 01"))
	g.AddEntity(NewEntity("rack-01", "rack", "Rack 01"))

	entities := g.Entities()
	if len(entities) != 3 {
		t.Errorf("expected 3 entities, got %d", len(entities))
	}
}

func TestGraphEntitiesByKind(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))
	g.AddEntity(NewEntity("srv-02", "server", "Server 02"))
	g.AddEntity(NewEntity("vm-01", "vm", "VM 01"))

	servers := g.EntitiesByKind("server")
	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

func TestGraphAddRelation(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))
	g.AddEntity(NewEntity("vm-01", "vm", "VM 01"))

	r := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	if err := g.AddRelation(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.RelationCount() != 1 {
		t.Errorf("expected 1 relation, got %d", g.RelationCount())
	}
}

func TestGraphAddRelationInvalidReference(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))

	r := NewDirectedRelation("rel-01", "hosts", "srv-01", "nonexistent")
	err := g.AddRelation(r)
	if err == nil {
		t.Error("expected invalid reference error")
	}
}

func TestGraphAddRelationDuplicate(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))
	g.AddEntity(NewEntity("vm-01", "vm", "VM 01"))

	r1 := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	r2 := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	g.AddRelation(r1)
	err := g.AddRelation(r2)
	if err == nil {
		t.Error("expected duplicate relation error")
	}
}

func TestGraphGetRelation(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))
	g.AddEntity(NewEntity("vm-01", "vm", "VM 01"))
	r := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	g.AddRelation(r)

	found, ok := g.GetRelation("rel-01")
	if !ok || found == nil {
		t.Error("expected to find relation")
	}
	if found.Type != "hosts" {
		t.Errorf("expected type hosts, got %s", found.Type)
	}

	_, ok = g.GetRelation("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent relation")
	}
}

func TestGraphRemoveRelation(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))
	g.AddEntity(NewEntity("vm-01", "vm", "VM 01"))
	r := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	g.AddRelation(r)

	if !g.RemoveRelation("rel-01") {
		t.Error("expected remove to return true")
	}
	if g.RelationCount() != 0 {
		t.Errorf("expected 0 relations after remove, got %d", g.RelationCount())
	}
}

func TestGraphRelationsForEntity(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))
	g.AddEntity(NewEntity("vm-01", "vm", "VM 01"))
	g.AddEntity(NewEntity("vm-02", "vm", "VM 02"))

	g.AddRelation(NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01"))
	g.AddRelation(NewDirectedRelation("rel-02", "hosts", "srv-01", "vm-02"))

	rels := g.RelationsForEntity("srv-01")
	if len(rels) != 2 {
		t.Errorf("expected 2 relations for srv-01, got %d", len(rels))
	}

	rels = g.RelationsForEntity("vm-01")
	if len(rels) != 1 {
		t.Errorf("expected 1 relation for vm-01, got %d", len(rels))
	}
}

func TestGraphRelationsByType(t *testing.T) {
	g := NewGraph()
	g.AddEntity(NewEntity("srv-01", "server", "Server 01"))
	g.AddEntity(NewEntity("vm-01", "vm", "VM 01"))
	g.AddEntity(NewEntity("intf-01", "interface", "intf-01"))
	g.AddEntity(NewEntity("intf-02", "interface", "intf-02"))

	g.AddRelation(NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01"))
	g.AddRelation(NewSymmetricRelation("rel-02", "connects", []string{"intf-01", "intf-02"}))

	hostsRels := g.RelationsByType("hosts")
	if len(hostsRels) != 1 {
		t.Errorf("expected 1 hosts relation, got %d", len(hostsRels))
	}

	connectsRels := g.RelationsByType("connects")
	if len(connectsRels) != 1 {
		t.Errorf("expected 1 connects relation, got %d", len(connectsRels))
	}
}

func TestGraphResolveReference(t *testing.T) {
	g := NewGraph()
	e := NewEntity("srv-01", "server", "Server 01")
	g.AddEntity(e)
	r := NewDirectedRelation("rel-01", "hosts", "srv-01", "srv-01")
	if err := g.AddRelation(r); err != nil {
		t.Fatalf("unexpected error adding relation: %v", err)
	}

	obj, ok := g.ResolveReference("srv-01")
	if !ok {
		t.Error("expected to resolve entity reference")
	}
	if _, ok := obj.(*Entity); !ok {
		t.Error("expected resolved reference to be Entity")
	}

	obj, ok = g.ResolveReference("rel-01")
	if !ok {
		t.Error("expected to resolve relation reference")
	}
	if _, ok := obj.(*Relation); !ok {
		t.Error("expected resolved reference to be Relation")
	}

	_, ok = g.ResolveReference("nonexistent")
	if ok {
		t.Error("expected not to resolve nonexistent reference")
	}
}

func TestGraphOwnershipPaths(t *testing.T) {
	g := NewGraph()
	site := NewEntity("site-01", "site", "Site 01")
	rack := NewEntity("rack-01", "rack", "Rack 01")
	server := NewEntity("srv-01", "server", "Server 01")
	intf := NewEntity("eno1", "interface", "eno1")

	rack.SetOwner("site-01")
	server.SetOwner("rack-01")
	intf.SetOwner("srv-01")

	g.AddEntity(site)
	g.AddEntity(rack)
	g.AddEntity(server)
	g.AddEntity(intf)

	if err := g.BuildOwnershipPaths(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if site.Path() != "/site-01" {
		t.Errorf("expected path /site-01, got %s", site.Path())
	}
	if rack.Path() != "/site-01/rack-01" {
		t.Errorf("expected path /site-01/rack-01, got %s", rack.Path())
	}
	if server.Path() != "/site-01/rack-01/srv-01" {
		t.Errorf("expected path /site-01/rack-01/srv-01, got %s", server.Path())
	}
	if intf.Path() != "/site-01/rack-01/srv-01/eno1" {
		t.Errorf("expected path /site-01/rack-01/srv-01/eno1, got %s", intf.Path())
	}
}

func TestGraphOwnershipCycle(t *testing.T) {
	g := NewGraph()
	e1 := NewEntity("a", "server", "A")
	e2 := NewEntity("b", "server", "B")
	e1.SetOwner("b")
	e2.SetOwner("a")

	g.AddEntity(e1)
	g.AddEntity(e2)

	err := g.BuildOwnershipPaths()
	if err == nil {
		t.Error("expected ownership cycle error")
	}
}

func TestGraphOwnershipMissingOwner(t *testing.T) {
	g := NewGraph()
	e := NewEntity("srv-01", "server", "Server 01")
	e.SetOwner("nonexistent")
	g.AddEntity(e)

	err := g.BuildOwnershipPaths()
	if err == nil {
		t.Error("expected owner not found error")
	}
}

func TestGraphChildren(t *testing.T) {
	g := NewGraph()
	site := NewEntity("site-01", "site", "Site 01")
	rack1 := NewEntity("rack-01", "rack", "Rack 01")
	rack2 := NewEntity("rack-02", "rack", "Rack 02")
	rack1.SetOwner("site-01")
	rack2.SetOwner("site-01")

	g.AddEntity(site)
	g.AddEntity(rack1)
	g.AddEntity(rack2)

	children := g.Children("site-01")
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestGraphParent(t *testing.T) {
	g := NewGraph()
	site := NewEntity("site-01", "site", "Site 01")
	rack := NewEntity("rack-01", "rack", "Rack 01")
	rack.SetOwner("site-01")

	g.AddEntity(site)
	g.AddEntity(rack)

	parent, ok := g.Parent("rack-01")
	if !ok || parent == nil {
		t.Error("expected to find parent")
	}
	if parent.ID != "site-01" {
		t.Errorf("expected parent ID site-01, got %s", parent.ID)
	}

	_, ok = g.Parent("site-01")
	if ok {
		t.Error("expected root entity to have no parent")
	}
}

func TestGraphAncestors(t *testing.T) {
	g := NewGraph()
	site := NewEntity("site-01", "site", "Site 01")
	rack := NewEntity("rack-01", "rack", "Rack 01")
	server := NewEntity("srv-01", "server", "Server 01")
	rack.SetOwner("site-01")
	server.SetOwner("rack-01")

	g.AddEntity(site)
	g.AddEntity(rack)
	g.AddEntity(server)

	ancestors := g.Ancestors("srv-01")
	if len(ancestors) != 2 {
		t.Errorf("expected 2 ancestors, got %d", len(ancestors))
	}
	if ancestors[0].ID != "rack-01" {
		t.Errorf("expected first ancestor rack-01, got %s", ancestors[0].ID)
	}
	if ancestors[1].ID != "site-01" {
		t.Errorf("expected second ancestor site-01, got %s", ancestors[1].ID)
	}
}

func TestGraphDescendants(t *testing.T) {
	g := NewGraph()
	site := NewEntity("site-01", "site", "Site 01")
	rack := NewEntity("rack-01", "rack", "Rack 01")
	server := NewEntity("srv-01", "server", "Server 01")
	rack.SetOwner("site-01")
	server.SetOwner("rack-01")

	g.AddEntity(site)
	g.AddEntity(rack)
	g.AddEntity(server)

	descendants := g.Descendants("site-01")
	if len(descendants) != 2 {
		t.Errorf("expected 2 descendants, got %d", len(descendants))
	}
}

func TestGraphValidateIntegrity(t *testing.T) {
	g := NewGraph()
	site := NewEntity("site-01", "site", "Site 01")
	rack := NewEntity("rack-01", "rack", "Rack 01")
	server := NewEntity("srv-01", "server", "Server 01")
	vm := NewEntity("vm-01", "vm", "VM 01")

	rack.SetOwner("site-01")
	server.SetOwner("rack-01")
	vm.SetOwner("srv-01")

	g.AddEntity(site)
	g.AddEntity(rack)
	g.AddEntity(server)
	g.AddEntity(vm)

	rel := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	if err := g.AddRelation(rel); err != nil {
		t.Fatalf("unexpected error adding relation: %v", err)
	}

	g.RemoveEntity("vm-01")

	errs := g.ValidateIntegrity()
	foundInvalidRef := false
	for _, err := range errs {
		if containsString(err.Error(), "non-existent entity vm-01") {
			foundInvalidRef = true
		}
	}
	if !foundInvalidRef {
		t.Error("expected invalid reference error for vm-01")
	}
}

func TestGraphValidateIntegrityValid(t *testing.T) {
	g := NewGraph()
	site := NewEntity("site-01", "site", "Site 01")
	server := NewEntity("srv-01", "server", "Server 01")
	vm := NewEntity("vm-01", "vm", "VM 01")
	server.SetOwner("site-01")
	vm.SetOwner("srv-01")

	g.AddEntity(site)
	g.AddEntity(server)
	g.AddEntity(vm)

	rel := NewDirectedRelation("rel-01", "hosts", "srv-01", "vm-01")
	if err := g.AddRelation(rel); err != nil {
		t.Fatalf("unexpected error adding relation: %v", err)
	}

	errs := g.ValidateIntegrity()
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestGraphOwnershipTreeBroken(t *testing.T) {
	g := NewGraph()
	site1 := NewEntity("site-01", "site", "Site 01")
	site2 := NewEntity("site-02", "site", "Site 02")

	g.AddEntity(site1)
	g.AddEntity(site2)

	errs := g.ValidateIntegrity()
	foundBrokenTree := false
	for _, err := range errs {
		if err == ErrOwnershipTreeBroken {
			foundBrokenTree = true
		}
	}
	if !foundBrokenTree {
		t.Error("expected ownership tree broken error for multiple roots")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
