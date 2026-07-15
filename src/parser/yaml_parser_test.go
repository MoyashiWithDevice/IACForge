package parser

import (
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
)

func TestParseBasicEntity(t *testing.T) {
	yaml := `
objects:
  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter 1
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if g.EntityCount() != 1 {
		t.Fatalf("expected 1 entity, got %d", g.EntityCount())
	}

	e, ok := g.GetEntity("site-tokyo-01")
	if !ok {
		t.Fatal("entity site-tokyo-01 not found")
	}

	if e.ID != "site-tokyo-01" {
		t.Errorf("expected ID site-tokyo-01, got %s", e.ID)
	}
	if e.Kind != kinds.Site {
		t.Errorf("expected kind site, got %s", e.Kind)
	}
	if e.Name != "Tokyo Datacenter 1" {
		t.Errorf("expected name 'Tokyo Datacenter 1', got %s", e.Name)
	}
}

func TestParseEntityWithAllProperties(t *testing.T) {
	yaml := `
objects:
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    attributes:
      description: "Primary Proxmox server"
      status: active
      tags:
        - production
        - compute
      labels:
        region: ap-northeast-1
        environment: production
      extensions:
        vendor: dell
        model: r740xd
    spec:
      platform: proxmox
      cpu_cores: 32
      memory_gb: 128
      storage_gb: 2000
      ip_address: 10.0.1.10
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	e, ok := g.GetEntity("srv-proxmox-01")
	if !ok {
		t.Fatal("entity srv-proxmox-01 not found")
	}

	if e.Description != "Primary Proxmox server" {
		t.Errorf("expected description 'Primary Proxmox server', got %s", e.Description)
	}
	if e.Status != core.StatusActive {
		t.Errorf("expected status active, got %s", e.Status)
	}
	if len(e.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(e.Tags))
	}
	if e.Labels["region"] != "ap-northeast-1" {
		t.Errorf("expected label region=ap-northeast-1, got %s", e.Labels["region"])
	}
	if e.Extensions["vendor"] != "dell" {
		t.Errorf("expected extensions vendor=dell, got %v", e.Extensions["vendor"])
	}

	// Check kind-specific properties
	if platform, ok := e.GetProperty("platform"); !ok || platform != "proxmox" {
		t.Errorf("expected property platform=proxmox, got %v", platform)
	}
	if cpuCores, ok := e.GetProperty("cpu_cores"); !ok || cpuCores != 32 {
		t.Errorf("expected property cpu_cores=32, got %v", cpuCores)
	}
}

func TestParseEntityWithOwnership(t *testing.T) {
	yaml := `
objects:
  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter 1

  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: site-tokyo-01

  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    attributes:
      owner: rack-a01
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if g.EntityCount() != 3 {
		t.Fatalf("expected 3 entities, got %d", g.EntityCount())
	}

	// Check ownership hierarchy
	site, _ := g.GetEntity("site-tokyo-01")
	if !site.IsRoot() {
		t.Error("site should be root")
	}

	rack, _ := g.GetEntity("rack-a01")
	if rack.Owner != "site-tokyo-01" {
		t.Errorf("expected rack owner site-tokyo-01, got %s", rack.Owner)
	}

	server, _ := g.GetEntity("srv-proxmox-01")
	if server.Owner != "rack-a01" {
		t.Errorf("expected server owner rack-a01, got %s", server.Owner)
	}

	// Check paths were built
	if site.Path() != "/site-tokyo-01" {
		t.Errorf("expected site path /site-tokyo-01, got %s", site.Path())
	}
	if rack.Path() != "/site-tokyo-01/rack-a01" {
		t.Errorf("expected rack path /site-tokyo-01/rack-a01, got %s", rack.Path())
	}
	if server.Path() != "/site-tokyo-01/rack-a01/srv-proxmox-01" {
		t.Errorf("expected server path /site-tokyo-01/rack-a01/srv-proxmox-01, got %s", server.Path())
	}
}

func TestParseDirectedRelation(t *testing.T) {
	yaml := `
objects:
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01

  - id: vm-web-01
    kind: vm
    name: Web Server 01

  - id: rel-hosts-server-vm
    type: hosts
    participants:
      source: srv-proxmox-01
      target: vm-web-01
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if g.RelationCount() != 1 {
		t.Fatalf("expected 1 relation, got %d", g.RelationCount())
	}

	r, ok := g.GetRelation("rel-hosts-server-vm")
	if !ok {
		t.Fatal("relation rel-hosts-server-vm not found")
	}

	if r.Type != types.Hosts {
		t.Errorf("expected type hosts, got %s", r.Type)
	}
	if r.Direction != core.DirectionDirected {
		t.Errorf("expected direction directed, got %s", r.Direction)
	}
	if r.Source() != "srv-proxmox-01" {
		t.Errorf("expected source srv-proxmox-01, got %s", r.Source())
	}
	if r.Target() != "vm-web-01" {
		t.Errorf("expected target vm-web-01, got %s", r.Target())
	}
}

func TestParseSymmetricRelation(t *testing.T) {
	yaml := `
objects:
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01

  - id: eno1
    kind: interface
    name: eno1
    attributes:
      owner: srv-proxmox-01

  - id: sw-core-01
    kind: switch
    name: Core Switch 01

  - id: sw-port1
    kind: interface
    name: port1
    attributes:
      owner: sw-core-01

  - id: rel-connects-srv-sw
    type: connects
    participants:
      - srv-proxmox-01/eno1
      - sw-core-01/port1
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	r, ok := g.GetRelation("rel-connects-srv-sw")
	if !ok {
		t.Fatal("relation rel-connects-srv-sw not found")
	}

	if r.Type != types.Connects {
		t.Errorf("expected type connects, got %s", r.Type)
	}
	if r.Direction != core.DirectionSymmetric {
		t.Errorf("expected direction symmetric, got %s", r.Direction)
	}
	if len(r.Participants.List) != 2 {
		t.Errorf("expected 2 participants, got %d", len(r.Participants.List))
	}
}

func TestParseRelationWithAllProperties(t *testing.T) {
	yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01

  - id: vm-01
    kind: vm
    name: VM 01

  - id: rel-hosts-vm
    type: hosts
    participants:
      source: srv-01
      target: vm-01
    attributes:
      description: "Server hosts VM"
      status: active
      tags:
        - hosting
      labels:
        source_type: server
        target_type: vm
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	r, ok := g.GetRelation("rel-hosts-vm")
	if !ok {
		t.Fatal("relation rel-hosts-vm not found")
	}

	if r.Description != "Server hosts VM" {
		t.Errorf("expected description 'Server hosts VM', got %s", r.Description)
	}
	if r.Status != core.StatusActive {
		t.Errorf("expected status active, got %s", r.Status)
	}
	if len(r.Tags) != 1 || r.Tags[0] != "hosting" {
		t.Errorf("expected tags [hosting], got %v", r.Tags)
	}
	if r.Labels["source_type"] != "server" {
		t.Errorf("expected label source_type=server, got %s", r.Labels["source_type"])
	}
}

func TestParseCompleteExample(t *testing.T) {
	yaml := `
objects:
  # Sites
  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter 1
    attributes:
      status: active
      labels:
        region: ap-northeast-1

  # Racks
  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: site-tokyo-01
      status: active
      labels:
        row: A
    spec:
      height_units: 42

  # Servers
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
      storage_gb: 2000
      ip_address: 10.0.1.10

  # VMs
  - id: vm-web-01
    kind: vm
    name: Web Server 01
    attributes:
      owner: srv-proxmox-01
      status: active
    spec:
      cpu_cores: 4
      memory_gb: 8
      disk_gb: 100
      os: ubuntu
      os_version: "22.04"
      ip_address: 10.0.2.10

  # Applications
  - id: app-web-server
    kind: application
    name: Nginx Web Server
    attributes:
      owner: vm-web-01
      status: active
    spec:
      version: "1.24.0"
      port: 443
      protocol: https

  # Hosting Relations
  - id: rel-hosts-server-vm
    type: hosts
    participants:
      source: srv-proxmox-01
      target: vm-web-01

  - id: rel-hosts-vm-app
    type: hosts
    participants:
      source: vm-web-01
      target: app-web-server
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if g.EntityCount() != 5 {
		t.Fatalf("expected 5 entities, got %d", g.EntityCount())
	}
	if g.RelationCount() != 2 {
		t.Fatalf("expected 2 relations, got %d", g.RelationCount())
	}

	// Verify ownership hierarchy
	server, _ := g.GetEntity("srv-proxmox-01")
	if server.Owner != "rack-a01" {
		t.Errorf("expected server owner rack-a01, got %s", server.Owner)
	}

	vm, _ := g.GetEntity("vm-web-01")
	if vm.Owner != "srv-proxmox-01" {
		t.Errorf("expected vm owner srv-proxmox-01, got %s", vm.Owner)
	}
}

func TestParseInvalidEntity(t *testing.T) {
	yaml := `
objects:
  - id: test-entity
    kind: server
`

	parser := NewParser()
	_, err := parser.Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing name field")
	}
}

func TestParseInvalidRelation(t *testing.T) {
	yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01

  - id: test-relation
    type: hosts
`

	parser := NewParser()
	_, err := parser.Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for missing participants")
	}
}

func TestParseInvalidYAML(t *testing.T) {
	yaml := `
objects:
  - id: test
    kind: server
    name: Test
    invalid: [unclosed
`

	parser := NewParser()
	_, err := parser.Parse([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseEmptyDocument(t *testing.T) {
	yaml := `
objects: []
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse empty document: %v", err)
	}

	if g.EntityCount() != 0 {
		t.Errorf("expected 0 entities, got %d", g.EntityCount())
	}
	if g.RelationCount() != 0 {
		t.Errorf("expected 0 relations, got %d", g.RelationCount())
	}
}

func TestParseCompleteExampleFile(t *testing.T) {
	parser := NewParser()
	g, err := parser.ParseFile("../../testdata/complete-example.yaml")
	if err != nil {
		t.Fatalf("failed to parse complete example: %v", err)
	}

	// Verify entity count (now includes switch sw-core-01)
	if g.EntityCount() != 18 {
		t.Errorf("expected 18 entities, got %d", g.EntityCount())
	}

	// Verify relation count
	if g.RelationCount() != 10 {
		t.Errorf("expected 10 relations, got %d", g.RelationCount())
	}

	// Verify specific entities
	site, ok := g.GetEntity("site-tokyo-01")
	if !ok {
		t.Error("entity site-tokyo-01 not found")
	} else {
		if site.Kind != kinds.Site {
			t.Errorf("expected kind site, got %s", site.Kind)
		}
		if site.Labels["region"] != "ap-northeast-1" {
			t.Errorf("expected label region=ap-northeast-1, got %s", site.Labels["region"])
		}
	}

	server, ok := g.GetEntity("srv-proxmox-01")
	if !ok {
		t.Error("entity srv-proxmox-01 not found")
	} else {
		if server.Owner != "rack-a01" {
			t.Errorf("expected owner rack-a01, got %s", server.Owner)
		}
		if cpuCores, ok := server.GetProperty("cpu_cores"); !ok || cpuCores != 32 {
			t.Errorf("expected cpu_cores=32, got %v", cpuCores)
		}
	}

	// Verify relations
	relation, ok := g.GetRelation("rel-hosts-server-vm")
	if !ok {
		t.Error("relation rel-hosts-server-vm not found")
	} else {
		if relation.Type != types.Hosts {
			t.Errorf("expected type hosts, got %s", relation.Type)
		}
		if relation.Source() != "srv-proxmox-01" {
			t.Errorf("expected source srv-proxmox-01, got %s", relation.Source())
		}
		if relation.Target() != "vm-web-01" {
			t.Errorf("expected target vm-web-01, got %s", relation.Target())
		}
	}

	// Verify symmetric relation
	connectsRelation, ok := g.GetRelation("rel-connects-srv-sw")
	if !ok {
		t.Error("relation rel-connects-srv-sw not found")
	} else {
		if connectsRelation.Type != types.Connects {
			t.Errorf("expected type connects, got %s", connectsRelation.Type)
		}
		if connectsRelation.Direction != core.DirectionSymmetric {
			t.Errorf("expected direction symmetric, got %s", connectsRelation.Direction)
		}
	}
}

func TestParseDir(t *testing.T) {
	parser := NewParser()
	g, err := parser.ParseDir("../../testdata/multi-file/infra/")
	if err != nil {
		t.Fatalf("failed to parse directory: %v", err)
	}

	if g.EntityCount() != 11 {
		t.Errorf("expected 11 entities, got %d", g.EntityCount())
	}
	if g.RelationCount() != 2 {
		t.Errorf("expected 2 relations, got %d", g.RelationCount())
	}

	// Verify cross-file ownership: site in site.yaml owns rack in rack.yaml
	rack, ok := g.GetEntity("rack-a01")
	if !ok {
		t.Fatal("entity rack-a01 not found")
	}
	if rack.Owner != "site-tokyo-01" {
		t.Errorf("expected owner site-tokyo-01, got %s", rack.Owner)
	}

	// Verify cross-file ownership: server in servers.yaml owns interface in interfaces.yaml
	eno1, ok := g.GetEntity("eno1")
	if !ok {
		t.Fatal("entity eno1 not found")
	}
	if eno1.Owner != "srv-proxmox-01" {
		t.Errorf("expected owner srv-proxmox-01, got %s", eno1.Owner)
	}

	// Verify paths are built correctly across files
	if rack.Path() != "/site-tokyo-01/rack-a01" {
		t.Errorf("expected path /site-tokyo-01/rack-a01, got %s", rack.Path())
	}

	// Verify cross-file relation: hosts relation between server and VM
	rel, ok := g.GetRelation("rel-hosts-server-vm")
	if !ok {
		t.Fatal("relation rel-hosts-server-vm not found")
	}
	if rel.Source() != "srv-proxmox-01" {
		t.Errorf("expected source srv-proxmox-01, got %s", rel.Source())
	}
	if rel.Target() != "vm-web-01" {
		t.Errorf("expected target vm-web-01, got %s", rel.Target())
	}
}

func TestLoadFile(t *testing.T) {
	parser := NewParser()
	g, err := parser.Load("../../testdata/complete-example.yaml")
	if err != nil {
		t.Fatalf("failed to load file: %v", err)
	}
	if g.EntityCount() != 18 {
		t.Errorf("expected 18 entities, got %d", g.EntityCount())
	}
}

func TestLoadDir(t *testing.T) {
	parser := NewParser()
	g, err := parser.Load("../../testdata/multi-file/infra/")
	if err != nil {
		t.Fatalf("failed to load directory: %v", err)
	}
	if g.EntityCount() != 11 {
		t.Errorf("expected 11 entities, got %d", g.EntityCount())
	}
}
