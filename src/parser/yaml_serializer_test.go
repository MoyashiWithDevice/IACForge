package parser

import (
	"strings"
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
)

func TestSerializeBasicEntity(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("site-tokyo-01", kinds.Site, "Tokyo Datacenter 1")
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	e2, ok := g2.GetEntity("site-tokyo-01")
	if !ok {
		t.Fatal("entity site-tokyo-01 not found in parsed data")
	}

	if e2.ID != e.ID {
		t.Errorf("expected ID %s, got %s", e.ID, e2.ID)
	}
	if e2.Kind != e.Kind {
		t.Errorf("expected kind %s, got %s", e.Kind, e2.Kind)
	}
	if e2.Name != e.Name {
		t.Errorf("expected name %s, got %s", e.Name, e2.Name)
	}
}

func TestSerializeEntityWithAllProperties(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("srv-proxmox-01", kinds.Server, "Proxmox Node 01")
	e.Description = "Primary Proxmox server"
	e.SetStatus(core.StatusActive)
	e.AddTag("production")
	e.AddTag("compute")
	e.SetLabel("region", "ap-northeast-1")
	e.SetLabel("environment", "production")
	e.Extensions = map[string]interface{}{"vendor": "dell"}
	e.SetProperty("platform", "proxmox")
	e.SetProperty("cpu_cores", 32)
	e.SetProperty("memory_gb", 128)

	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	e2, ok := g2.GetEntity("srv-proxmox-01")
	if !ok {
		t.Fatal("entity srv-proxmox-01 not found in parsed data")
	}

	if e2.Description != e.Description {
		t.Errorf("expected description %s, got %s", e.Description, e2.Description)
	}
	if e2.Status != e.Status {
		t.Errorf("expected status %s, got %s", e.Status, e2.Status)
	}
	if len(e2.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(e2.Tags))
	}
	if e2.Labels["region"] != "ap-northeast-1" {
		t.Errorf("expected label region=ap-northeast-1, got %s", e2.Labels["region"])
	}
	if platform, ok := e2.GetProperty("platform"); !ok || platform != "proxmox" {
		t.Errorf("expected property platform=proxmox, got %v", platform)
	}
}

func TestSerializeDirectedRelation(t *testing.T) {
	g := core.NewGraph()

	srv := core.NewEntity("srv-01", kinds.Server, "Server 01")
	vm := core.NewEntity("vm-01", kinds.VM, "VM 01")

	if err := g.AddEntity(srv); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(vm); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	r := core.NewDirectedRelation("rel-hosts-vm", types.Hosts, "srv-01", "vm-01")
	if err := g.AddRelation(r); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	r2, ok := g2.GetRelation("rel-hosts-vm")
	if !ok {
		t.Fatal("relation rel-hosts-vm not found in parsed data")
	}

	if r2.Type != types.Hosts {
		t.Errorf("expected type hosts, got %s", r2.Type)
	}
	if r2.Direction != core.DirectionDirected {
		t.Errorf("expected direction directed, got %s", r2.Direction)
	}
	if r2.Source() != "srv-01" {
		t.Errorf("expected source srv-01, got %s", r2.Source())
	}
	if r2.Target() != "vm-01" {
		t.Errorf("expected target vm-01, got %s", r2.Target())
	}
}

func TestSerializeSymmetricRelation(t *testing.T) {
	g := core.NewGraph()

	// Create parent entities first
	srv := core.NewEntity("srv-01", kinds.Server, "Server 01")
	sw := core.NewEntity("sw-01", kinds.Switch, "Switch 01")
	if err := g.AddEntity(srv); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(sw); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	iface1 := core.NewEntity("eno1", kinds.Interface, "eno1")
	iface1.SetOwner("srv-01")
	iface2 := core.NewEntity("port1", kinds.Interface, "port1")
	iface2.SetOwner("sw-01")

	if err := g.AddEntity(iface1); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}
	if err := g.AddEntity(iface2); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	r := core.NewSymmetricRelation("rel-connects", types.Connects, []string{"eno1", "port1"})
	if err := g.AddRelation(r); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	r2, ok := g2.GetRelation("rel-connects")
	if !ok {
		t.Fatal("relation rel-connects not found in parsed data")
	}

	if r2.Type != types.Connects {
		t.Errorf("expected type connects, got %s", r2.Type)
	}
	if r2.Direction != core.DirectionSymmetric {
		t.Errorf("expected direction symmetric, got %s", r2.Direction)
	}
	if len(r2.Participants.List) != 2 {
		t.Errorf("expected 2 participants, got %d", len(r2.Participants.List))
	}
}

func TestRoundTrip(t *testing.T) {
	yaml := `
objects:
  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter 1
    attributes:
      status: active
      labels:
        region: ap-northeast-1

  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: site-tokyo-01
      status: active
    spec:
      height_units: 42

  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    attributes:
      owner: rack-a01
      status: active
    spec:
      platform: proxmox
      cpu_cores: 32
      memory_gb: 128

  - id: vm-web-01
    kind: vm
    name: Web Server 01
    attributes:
      owner: srv-proxmox-01
      status: active
    spec:
      cpu_cores: 4
      memory_gb: 8
      os: ubuntu

  - id: rel-hosts-server-vm
    type: hosts
    participants:
      source: srv-proxmox-01
      target: vm-web-01
`

	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse serialized data
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized data: %v", err)
	}

	// Verify entities
	if g1.EntityCount() != g2.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	for _, e1 := range g1.Entities() {
		e2, ok := g2.GetEntity(e1.ID)
		if !ok {
			t.Errorf("entity %s not found in round-trip", e1.ID)
			continue
		}
		if e1.ID != e2.ID {
			t.Errorf("entity ID mismatch: %s vs %s", e1.ID, e2.ID)
		}
		if e1.Kind != e2.Kind {
			t.Errorf("entity kind mismatch for %s: %s vs %s", e1.ID, e1.Kind, e2.Kind)
		}
		if e1.Name != e2.Name {
			t.Errorf("entity name mismatch for %s: %s vs %s", e1.ID, e1.Name, e2.Name)
		}
		if e1.Owner != e2.Owner {
			t.Errorf("entity owner mismatch for %s: %s vs %s", e1.ID, e1.Owner, e2.Owner)
		}
	}

	// Verify relations
	if g1.RelationCount() != g2.RelationCount() {
		t.Errorf("relation count mismatch: %d vs %d", g1.RelationCount(), g2.RelationCount())
	}

	for _, r1 := range g1.Relations() {
		r2, ok := g2.GetRelation(r1.ID)
		if !ok {
			t.Errorf("relation %s not found in round-trip", r1.ID)
			continue
		}
		if r1.ID != r2.ID {
			t.Errorf("relation ID mismatch: %s vs %s", r1.ID, r2.ID)
		}
		if r1.Type != r2.Type {
			t.Errorf("relation type mismatch for %s: %s vs %s", r1.ID, r1.Type, r2.Type)
		}
		if r1.Direction != r2.Direction {
			t.Errorf("relation direction mismatch for %s: %s vs %s", r1.ID, r1.Direction, r2.Direction)
		}
		if r1.Source() != r2.Source() {
			t.Errorf("relation source mismatch for %s: %s vs %s", r1.ID, r1.Source(), r2.Source())
		}
		if r1.Target() != r2.Target() {
			t.Errorf("relation target mismatch for %s: %s vs %s", r1.ID, r1.Target(), r2.Target())
		}
	}
}

func TestSerializeFile(t *testing.T) {
	g := core.NewGraph()
	e := core.NewEntity("test-entity", kinds.Server, "Test Entity")
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("failed to add entity: %v", err)
	}

	serializer := NewSerializer()
	err := serializer.SerializeFile(g, "/tmp/test_serialize.yaml")
	if err != nil {
		t.Fatalf("failed to serialize to file: %v", err)
	}

	// Parse back and verify
	parser := NewParser()
	g2, err := parser.ParseFile("/tmp/test_serialize.yaml")
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	e2, ok := g2.GetEntity("test-entity")
	if !ok {
		t.Fatal("entity test-entity not found in file")
	}

	if e2.ID != e.ID {
		t.Errorf("expected ID %s, got %s", e.ID, e2.ID)
	}
}

func TestRoundTripPropertyReference(t *testing.T) {
	input := `
objects:
  - id: net-mgmt
    kind: network
    name: Management Network
    spec:
      cidr: 10.0.0.0/24

  - id: vlan-100
    kind: vlan
    name: VLAN 100
    spec:
      vlan_id: 100
      associated_network: "@net-mgmt"
`
	parser := NewParser()
	g, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	serializer := NewSerializer()
	data, err := serializer.Serialize(g)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Verify @ prefix is preserved in output (YAML may use single or double quotes)
	output := string(data)
	if !strings.Contains(output, "@net-mgmt") {
		t.Errorf("serialized output should contain @net-mgmt reference, got:\n%s", output)
	}

	// Parse back and verify round-trip
	g2, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("failed to re-parse: %v", err)
	}

	vlan, ok := g2.GetEntity("vlan-100")
	if !ok {
		t.Fatal("expected entity vlan-100")
	}

	v, ok := vlan.GetProperty("associated_network")
	if !ok {
		t.Fatal("expected property associated_network")
	}
	ref, ok := v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue after round-trip, got %T", v)
	}
	if ref.RefTargetID() != "net-mgmt" {
		t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
	}
}
