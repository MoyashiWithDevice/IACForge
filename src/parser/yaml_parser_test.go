package parser

import (
	"strings"
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/core/types"
	"IACForge/src/schema"
	"IACForge/src/validation"
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
      memory:
        - size_gb: 64
          speed: 3200
          type: ddr4
      storage_gb: 2000
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
      memory:
        - size_gb: 64
          speed: 3200
          type: ddr4
      storage_gb: 2000

  # VMs
  - id: vm-web-01
    kind: vm
    name: Web Server 01
    attributes:
      owner: srv-proxmox-01
      status: active
    spec:
      cpu_cores: 4
      memory:
        - size_gb: 8
          speed: 3200
          type: ddr4
      storage_gb: 100
      os: ubuntu
      os_version: "22.04"

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

func TestParseNestedEntities(t *testing.T) {
	yaml := `
objects:
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    spec:
      cpu_cores: 32
      memory:
        - size_gb: 64
          speed: 3200
          type: ddr4
      networks:
        - id: net-private
          name: private
          spec:
            cidr: 172.31.0.0/24
            gateway: 172.31.0.254
          interfaces:
            - id: eth1
              name: proxmox/eth1
              spec:
                ip_address: 172.31.0.15
                type: ethernet
      vms:
        - id: vm-web-01
          name: Web Server 01
          spec:
            cpu_cores: 4
            memory:
              - size_gb: 8
                speed: 3200
                type: ddr4
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 server + 1 network + 1 interface + 1 vm = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check parent server
	server, ok := g.GetEntity("srv-proxmox-01")
	if !ok {
		t.Fatal("entity srv-proxmox-01 not found")
	}
	if server.Owner != "" {
		t.Errorf("server should be root, got owner %s", server.Owner)
	}

	// Check nested network
	net, ok := g.GetEntity("net-private")
	if !ok {
		t.Fatal("entity net-private not found")
	}
	if net.Owner != "srv-proxmox-01" {
		t.Errorf("expected network owner srv-proxmox-01, got %s", net.Owner)
	}
	if net.Kind != kinds.Network {
		t.Errorf("expected kind network, got %s", net.Kind)
	}

	// Check nested interface - owned by network, not server
	iface, ok := g.GetEntity("eth1")
	if !ok {
		t.Fatal("entity eth1 not found")
	}
	if iface.Owner != "net-private" {
		t.Errorf("expected interface owner net-private, got %s", iface.Owner)
	}
	if iface.Kind != kinds.Interface {
		t.Errorf("expected kind interface, got %s", iface.Kind)
	}

	// Check nested vm
	vm, ok := g.GetEntity("vm-web-01")
	if !ok {
		t.Fatal("entity vm-web-01 not found")
	}
	if vm.Owner != "srv-proxmox-01" {
		t.Errorf("expected vm owner srv-proxmox-01, got %s", vm.Owner)
	}

	// Check network properties
	cidr, ok := net.GetProperty("cidr")
	if !ok || cidr != "172.31.0.0/24" {
		t.Errorf("expected cidr 172.31.0.0/24, got %v", cidr)
	}
}

func TestParseNestedEntityAutoID(t *testing.T) {
	yaml := `
objects:
  - id: rack-a01
    kind: rack
    name: Rack A01
    spec:
      servers:
        - id: srv-01
          name: Server 01
          spec:
            cpu_cores: 16
        - id: srv-02
          name: Server 02
          spec:
            cpu_cores: 32
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 rack + 2 servers = 3
	if g.EntityCount() != 3 {
		t.Fatalf("expected 3 entities, got %d", g.EntityCount())
	}

	// Check explicit IDs
	s1, ok := g.GetEntity("srv-01")
	if !ok {
		t.Fatal("entity srv-01 not found")
	}
	if s1.Owner != "rack-a01" {
		t.Errorf("expected owner rack-a01, got %s", s1.Owner)
	}
	if s1.Name != "Server 01" {
		t.Errorf("expected name Server 01, got %s", s1.Name)
	}

	s2, ok := g.GetEntity("srv-02")
	if !ok {
		t.Fatal("entity srv-02 not found")
	}
	if s2.Owner != "rack-a01" {
		t.Errorf("expected owner rack-a01, got %s", s2.Owner)
	}
}

func TestParseNestedEntitiesWithoutID(t *testing.T) {
	yaml := `
objects:
  - id: acl-web
    kind: acl
    name: Web ACL
    spec:
      default_action: deny
      acl_rules:
        - id: rule-https
          name: Allow HTTPS
          spec:
            action: allow
            protocol: tcp
            destination_port: "443"
        - id: rule-ssh
          name: Allow SSH
          spec:
            action: allow
            protocol: tcp
            destination_port: "22"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 acl + 2 acl_rules = 3
	if g.EntityCount() != 3 {
		t.Fatalf("expected 3 entities, got %d", g.EntityCount())
	}

	// Check first rule
	r1, ok := g.GetEntity("rule-https")
	if !ok {
		t.Fatal("entity rule-https not found")
	}
	if r1.Owner != "acl-web" {
		t.Errorf("expected owner acl-web, got %s", r1.Owner)
	}
	if r1.Kind != kinds.ACLRule {
		t.Errorf("expected kind acl_rule, got %s", r1.Kind)
	}
}

func TestParseFlatAndNestedMixed(t *testing.T) {
	yaml := `
objects:
  # Flat definition
  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: site-tokyo-01

  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter

  # Nested definition
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    attributes:
      owner: rack-a01
    spec:
      cpu_cores: 32
      vms:
        - id: vm-web-01
          name: Web Server 01
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// site + rack + server + vm = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check flat ownership
	rack, ok := g.GetEntity("rack-a01")
	if !ok {
		t.Fatal("rack not found")
	}
	if rack.Owner != "site-tokyo-01" {
		t.Errorf("expected rack owner site-tokyo-01, got %s", rack.Owner)
	}

	// Check nested ownership
	vm, ok := g.GetEntity("vm-web-01")
	if !ok {
		t.Fatal("vm not found")
	}
	if vm.Owner != "srv-proxmox-01" {
		t.Errorf("expected vm owner srv-proxmox-01, got %s", vm.Owner)
	}
}

func TestParseNestedServerNetworkInterface(t *testing.T) {
	yaml := `
objects:
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    spec:
      cpu_cores: 32
      networks:
        - id: net-private
          name: private
          spec:
            cidr: 172.31.0.0/24
          interfaces:
            - id: eth1
              name: proxmox/eth1
              spec:
                ip_address: 172.31.0.15
        - name: mgmt
          spec:
            cidr: 10.0.0.0/24
          interfaces:
            - name: eth0
              spec:
                ip_address: 10.0.0.10
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 server + 2 networks + 2 interfaces = 5
	if g.EntityCount() != 5 {
		t.Fatalf("expected 5 entities, got %d", g.EntityCount())
	}

	// Both networks should be owned by the server
	for _, id := range []string{"net-private", "_srv-proxmox-01-network"} {
		e, ok := g.GetEntity(id)
		if !ok {
			t.Fatalf("entity %s not found", id)
		}
		if e.Owner != "srv-proxmox-01" {
			t.Errorf("entity %s expected owner srv-proxmox-01, got %s", id, e.Owner)
		}
	}

	// eth1 should be owned by net-private (network nests interfaces)
	eth1, ok := g.GetEntity("eth1")
	if !ok {
		t.Fatal("entity eth1 not found")
	}
	if eth1.Owner != "net-private" {
		t.Errorf("expected eth1 owner net-private, got %s", eth1.Owner)
	}

	// eth0 should be owned by _srv-proxmox-01-network (the auto-ID'd network)
	eth0, ok := g.GetEntity("_srv-proxmox-01-network-interface")
	if !ok {
		t.Fatal("entity _srv-proxmox-01-network-interface not found")
	}
	if eth0.Owner != "_srv-proxmox-01-network" {
		t.Errorf("expected eth0 owner _srv-proxmox-01-network, got %s", eth0.Owner)
	}
}

func TestSchemaNestingDefinitions(t *testing.T) {
	s := schema.CoreSchema()

	// Global nesting defs should exist
	if len(s.NestingDefs) != 6 {
		t.Fatalf("expected 6 global nesting defs, got %d", len(s.NestingDefs))
	}

	// Site: 6 global + 2 per-kind (racks, clusters) = 8
	siteNesting := s.GetNestingDefs(kinds.Site)
	if len(siteNesting) != 8 {
		t.Fatalf("expected 8 nesting defs for site, got %d", len(siteNesting))
	}

	// Rack: 6 global + 0 per-kind = 6
	rackNesting := s.GetNestingDefs(kinds.Rack)
	if len(rackNesting) != 6 {
		t.Fatalf("expected 6 nesting defs for rack, got %d", len(rackNesting))
	}

	// Server: 6 global + 2 per-kind (vms, containers) = 8
	serverNesting := s.GetNestingDefs(kinds.Server)
	if len(serverNesting) != 8 {
		t.Fatalf("expected 8 nesting defs for server, got %d", len(serverNesting))
	}

	// Find network nesting under server (now from global)
	nd, ok := s.FindNestingByNestKey(kinds.Server, "networks")
	if !ok {
		t.Fatal("networks nesting not found for server")
	}
	if nd.ChildKind != kinds.Network {
		t.Errorf("expected child kind network, got %s", nd.ChildKind)
	}
}

func TestInterfaceNestingDefinitions(t *testing.T) {
	s := schema.CoreSchema()

	// Interface: 6 global + 2 per-kind (vlans, cables) = 8
	ifaceNesting := s.GetNestingDefs(kinds.Interface)
	if len(ifaceNesting) != 8 {
		t.Fatalf("expected 8 nesting defs for interface, got %d", len(ifaceNesting))
	}

	nd, ok := s.FindNestingByNestKey(kinds.Interface, "interfaces")
	if !ok {
		t.Fatal("interfaces nesting not found for interface")
	}
	if nd.ChildKind != kinds.Interface {
		t.Errorf("expected child kind interface, got %s", nd.ChildKind)
	}
}

func TestParseNestedInterfaceVRRP(t *testing.T) {
	yaml := `
objects:
  - id: router-01
    kind: router
    name: Router 01
    spec:
      interfaces:
        - id: eth0-vrrp
          kind: interface
          name: VRRP Virtual Interface
          attributes:
            status: active
          spec:
            type: virtual
            ip_address:
              - 10.0.0.1
            interfaces:
              - id: eth0
                kind: interface
                name: eth0 - Primary
                attributes:
                  status: active
                spec:
                  type: ethernet
                  ip_address:
                    - 10.0.0.2
              - id: eth1
                kind: interface
                name: eth1 - Secondary
                attributes:
                  status: standby
                spec:
                  type: ethernet
                  ip_address:
                    - 10.0.0.3
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 router + 1 virtual interface + 2 physical interfaces = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check parent router
	router, ok := g.GetEntity("router-01")
	if !ok {
		t.Fatal("entity router-01 not found")
	}
	if router.Owner != "" {
		t.Errorf("router should be root, got owner %s", router.Owner)
	}

	// Check virtual interface
	viface, ok := g.GetEntity("eth0-vrrp")
	if !ok {
		t.Fatal("entity eth0-vrrp not found")
	}
	if viface.Owner != "router-01" {
		t.Errorf("expected owner router-01, got %s", viface.Owner)
	}
	if viface.Kind != kinds.Interface {
		t.Errorf("expected kind interface, got %s", viface.Kind)
	}
	if viface.Status != core.StatusActive {
		t.Errorf("expected status active, got %s", viface.Status)
	}
	if vtype, ok := viface.GetProperty("type"); !ok || vtype != "virtual" {
		t.Errorf("expected type virtual, got %v", vtype)
	}

	// Check primary physical interface
	eth0, ok := g.GetEntity("eth0")
	if !ok {
		t.Fatal("entity eth0 not found")
	}
	if eth0.Owner != "eth0-vrrp" {
		t.Errorf("expected owner eth0-vrrp, got %s", eth0.Owner)
	}
	if eth0.Kind != kinds.Interface {
		t.Errorf("expected kind interface, got %s", eth0.Kind)
	}
	if eth0.Status != core.StatusActive {
		t.Errorf("expected status active, got %s", eth0.Status)
	}

	// Check secondary physical interface
	eth1, ok := g.GetEntity("eth1")
	if !ok {
		t.Fatal("entity eth1 not found")
	}
	if eth1.Owner != "eth0-vrrp" {
		t.Errorf("expected owner eth0-vrrp, got %s", eth1.Owner)
	}
	if eth1.Status != core.StatusStandby {
		t.Errorf("expected status standby, got %s", eth1.Status)
	}
}

func TestParseNestedInterfaceLACP(t *testing.T) {
	yaml := `
objects:
  - id: router-01
    kind: router
    name: Router 01
    spec:
      interfaces:
        - id: bond0
          kind: interface
          name: LAG Bundle
          spec:
            type: bond
            interfaces:
              - id: eth0
                kind: interface
                attributes:
                  status: active
                spec:
                  type: ethernet
                  ip_address:
                    - 192.168.1.1
              - id: eth1
                kind: interface
                attributes:
                  status: standby
                spec:
                  type: ethernet
                  ip_address:
                    - 192.168.1.2
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// 1 router + 1 bond interface + 2 physical interfaces = 4
	if g.EntityCount() != 4 {
		t.Fatalf("expected 4 entities, got %d", g.EntityCount())
	}

	// Check bond interface
	bond, ok := g.GetEntity("bond0")
	if !ok {
		t.Fatal("entity bond0 not found")
	}
	if bond.Owner != "router-01" {
		t.Errorf("expected owner router-01, got %s", bond.Owner)
	}
	if btype, ok := bond.GetProperty("type"); !ok || btype != "bond" {
		t.Errorf("expected type bond, got %v", btype)
	}

	// Check that physical interfaces have no virtual IP
	eth0, ok := g.GetEntity("eth0")
	if !ok {
		t.Fatal("entity eth0 not found")
	}
	if eth0.Owner != "bond0" {
		t.Errorf("expected owner bond0, got %s", eth0.Owner)
	}
}

func TestRoundTripNestedEntities(t *testing.T) {
	yaml := `
objects:
  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter 1
    spec:
      racks:
        - id: rack-a01
          name: Rack A01
          spec:
            servers:
              - id: srv-01
                name: Server 01
                spec:
                  cpu_cores: 16
`

	// Parse original
	parser1 := NewParser()
	g1, err := parser1.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	if g1.EntityCount() != 3 {
		t.Fatalf("expected 3 entities, got %d", g1.EntityCount())
	}

	// Serialize
	serializer := NewSerializer()
	data, err := serializer.Serialize(g1)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Parse serialized
	parser2 := NewParser()
	g2, err := parser2.Parse(data)
	if err != nil {
		t.Fatalf("failed to parse serialized: %v", err)
	}

	if g2.EntityCount() != g1.EntityCount() {
		t.Errorf("entity count mismatch: %d vs %d", g1.EntityCount(), g2.EntityCount())
	}

	// Verify round-trip integrity
	for _, e1 := range g1.Entities() {
		e2, ok := g2.GetEntity(e1.ID)
		if !ok {
			t.Errorf("entity %s not found in round-trip", e1.ID)
			continue
		}
		if e1.Kind != e2.Kind {
			t.Errorf("kind mismatch for %s: %s vs %s", e1.ID, e1.Kind, e2.Kind)
		}
		if e1.Owner != e2.Owner {
			t.Errorf("owner mismatch for %s: %s vs %s", e1.ID, e1.Owner, e2.Owner)
		}
	}
}

func TestValidationNoSlashInID(t *testing.T) {
	yaml := `
objects:
  - id: bad/entity
    kind: server
    name: Bad Entity
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	s := schema.CoreSchema()
	engine := validation.NewEngine(s)
	validation.RegisterCoreRules(engine)
	result := engine.Validate(g, nil)

	found := false
	for _, f := range result.Findings {
		if f.RuleID == "no-slash-in-id" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected no-slash-in-id finding for entity with slash in ID")
	}
}

func TestValidationNestingParent(t *testing.T) {
	yaml := `
objects:
  - id: site-01
    kind: site
    name: Site 01

  - id: srv-01
    kind: server
    name: Server 01
    attributes:
      owner: site-01
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	s := schema.CoreSchema()
	engine := validation.NewEngine(s)
	validation.RegisterCoreRules(engine)
	result := engine.Validate(g, nil)

	for _, f := range result.Findings {
		if f.RuleID == "valid-nesting-parent" {
			t.Error("valid-nesting-parent warning should not be emitted for server owned by site (server is a valid child of site)")
		}
	}
}

func TestPathReferenceResolution(t *testing.T) {
	yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      networks:
        - id: net-mgmt
          name: Management
          interfaces:
            - id: eth0
              name: eth0
              spec:
                ip_address: 10.0.0.1

  - id: sw-01
    kind: switch
    name: Switch 01
    spec:
      interfaces:
        - id: port1
          name: port1

  - id: rel-connects
    type: connects
    participants:
      - srv-01/eth0
      - sw-01/port1
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	errs := ResolveReferences(g)
	if len(errs) != 0 {
		t.Errorf("expected no reference errors, got %v", errs)
	}
}

func TestParseEntityPropertyReference(t *testing.T) {
	yaml := `
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
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	vlan, ok := g.GetEntity("vlan-100")
	if !ok {
		t.Fatal("expected entity vlan-100")
	}

	v, ok := vlan.GetProperty("associated_network")
	if !ok {
		t.Fatal("expected property associated_network")
	}
	ref, ok := v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue, got %T", v)
	}
	if ref.RefTargetID() != "net-mgmt" {
		t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
	}

	// Verify reference resolution
	errs := ResolveReferences(g)
	if len(errs) != 0 {
		t.Errorf("expected no reference errors, got %v", errs)
	}
}

func TestParseEntityPropertyReferenceNotFound(t *testing.T) {
	yaml := `
objects:
  - id: vlan-100
    kind: vlan
    name: VLAN 100
    spec:
      vlan_id: 100
      associated_network: "@nonexistent"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	errs := ResolveReferences(g)
	if len(errs) == 0 {
		t.Error("expected reference error for nonexistent entity")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "nonexistent") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about nonexistent reference, got %v", errs)
	}
}

func TestParseRelationPropertyReference(t *testing.T) {
	yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01

  - id: net-mgmt
    kind: network
    name: Management Network

  - id: rel-uses
    type: depends_on
    participants:
      source: srv-01
      target: net-mgmt
    spec:
      dependency_type: "@net-mgmt"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rel, ok := g.GetRelation("rel-uses")
	if !ok {
		t.Fatal("expected relation rel-uses")
	}

	v, ok := rel.GetProperty("dependency_type")
	if !ok {
		t.Fatal("expected property dependency_type")
	}
	_, ok = v.(core.ReferenceValue)
	if !ok {
		t.Fatalf("expected ReferenceValue, got %T", v)
	}

	errs := ResolveReferences(g)
	if len(errs) != 0 {
		t.Errorf("expected no reference errors, got %v", errs)
	}
}

func TestParsePropertyPlainTextNotAffected(t *testing.T) {
	yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      platform: proxmox
      description: "A server without @ prefix"
`

	parser := NewParser()
	g, err := parser.Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	srv, ok := g.GetEntity("srv-01")
	if !ok {
		t.Fatal("expected entity srv-01")
	}

	// Plain string should NOT be converted to ReferenceValue
	v, ok := srv.GetProperty("platform")
	if !ok {
		t.Fatal("expected property platform")
	}
	if _, ok := v.(core.ReferenceValue); ok {
		t.Error("plain string should not be converted to ReferenceValue")
	}
	str, ok := v.(string)
	if !ok || str != "proxmox" {
		t.Errorf("expected plain string proxmox, got %v", v)
	}
}

func TestConvertPropertyValueRecursive(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantRef  bool
		wantType string
	}{
		{
			name:    "simple @ reference",
			input:   "@net-mgmt",
			wantRef: true,
		},
		{
			name:     "plain string",
			input:    "hello",
			wantType: "string",
		},
		{
			name: "list with @ reference",
			input: []interface{}{
				"@net-mgmt",
				"plain",
			},
			wantType: "list-of-strings",
		},
		{
			name: "nested map with @ reference",
			input: map[string]interface{}{
				"network": "@net-mgmt",
				"name":    "mgmt",
			},
			wantType: "map",
		},
		{
			name: "list of maps with @ reference",
			input: []interface{}{
				map[string]interface{}{
					"network": "@net-mgmt",
					"vlan":    float64(100),
				},
			},
			wantType: "list-of-maps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertPropertyValue(tt.input)

			if tt.wantRef {
				ref, ok := result.(core.ReferenceValue)
				if !ok {
					t.Errorf("expected ReferenceValue, got %T", result)
				}
				if ref.RefTargetID() != "net-mgmt" {
					t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
				}
				return
			}

			switch tt.wantType {
			case "list-of-strings":
				list, ok := result.([]interface{})
				if !ok {
					t.Fatalf("expected []interface{}, got %T", result)
				}
				if len(list) > 0 {
					if ref, ok := list[0].(core.ReferenceValue); !ok {
						t.Errorf("expected list[0] to be ReferenceValue, got %T", list[0])
					} else if ref.RefTargetID() != "net-mgmt" {
						t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
					}
				}
			case "list-of-maps":
				list, ok := result.([]interface{})
				if !ok {
					t.Fatalf("expected []interface{}, got %T", result)
				}
				if len(list) > 0 {
					m, ok := list[0].(map[string]interface{})
					if !ok {
						t.Fatalf("expected list[0] to be map, got %T", list[0])
					}
					if ref, ok := m["network"].(core.ReferenceValue); !ok {
						t.Errorf("expected network to be ReferenceValue, got %T", m["network"])
					} else if ref.RefTargetID() != "net-mgmt" {
						t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
					}
				}
			case "map":
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map[string]interface{}, got %T", result)
				}
				if ref, ok := m["network"].(core.ReferenceValue); !ok {
					t.Errorf("expected network to be ReferenceValue, got %T", m["network"])
				} else if ref.RefTargetID() != "net-mgmt" {
					t.Errorf("expected reference target net-mgmt, got %s", ref.RefTargetID())
				}
			case "string":
				str, ok := result.(string)
				if !ok {
					t.Fatalf("expected string, got %T", result)
				}
				if str != "hello" {
					t.Errorf("expected hello, got %s", str)
				}
			}
	})

	t.Run("server nesting container generates hosts relation", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      containers:
        - id: ctr-01
          name: Container 01
          spec:
            image: nginx:latest
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		srv, ok := g.GetEntity("srv-01")
		if !ok {
			t.Fatal("server not found")
		}
		if srv.Owner != "" {
			t.Errorf("expected server to be root, got owner %s", srv.Owner)
		}

		ctr, ok := g.GetEntity("ctr-01")
		if !ok {
			t.Fatal("container not found")
		}
		if ctr.Owner != "srv-01" {
			t.Errorf("expected container owner to be srv-01, got %s", ctr.Owner)
		}

		rel, ok := g.GetRelation("rel-auto-hosts-srv-01-ctr-01")
		if !ok || rel.Type != types.Hosts {
			t.Error("missing hosts relation from container to server")
		}
	})

	t.Run("vm nesting container generates hosts relation", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      vms:
        - id: vm-01
          name: VM 01
          spec:
            containers:
              - id: ctr-01
                name: Container 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		ctr, ok := g.GetEntity("ctr-01")
		if !ok {
			t.Fatal("container not found")
		}
		if ctr.Owner != "vm-01" {
			t.Errorf("expected container owner to be vm-01, got %s", ctr.Owner)
		}

		rel, ok := g.GetRelation("rel-auto-hosts-vm-01-ctr-01")
		if !ok || rel.Type != types.Hosts {
			t.Error("missing hosts relation from container to vm")
		}
	})

	t.Run("application nesting container generates hosts relation", func(t *testing.T) {
		yaml := `
objects:
  - id: app-01
    kind: application
    name: App 01
    spec:
      containers:
        - id: ctr-01
          name: Container 01
          spec:
            image: nginx:latest
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		ctr, ok := g.GetEntity("ctr-01")
		if !ok {
			t.Fatal("container not found")
		}
		if ctr.Owner != "app-01" {
			t.Errorf("expected container owner to be app-01, got %s", ctr.Owner)
		}

		rel, ok := g.GetRelation("rel-auto-hosts-app-01-ctr-01")
		if !ok || rel.Type != types.Hosts {
			t.Error("missing hosts relation from container to application")
		}
	})

	t.Run("container nesting application generates hosts relation", func(t *testing.T) {
		yaml := `
objects:
  - id: ctr-01
    kind: container
    name: Container 01
    spec:
      image: node:18
      applications:
        - id: app-01
          name: App 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		app, ok := g.GetEntity("app-01")
		if !ok {
			t.Fatal("application not found")
		}
		if app.Owner != "ctr-01" {
			t.Errorf("expected application owner to be ctr-01, got %s", app.Owner)
		}

		rel, ok := g.GetRelation("rel-auto-hosts-ctr-01-app-01")
		if !ok || rel.Type != types.Hosts {
			t.Error("missing hosts relation from application to container")
		}
	})

	t.Run("kubernetes multi-node scenario", func(t *testing.T) {
		yaml := `
objects:
  - id: cluster-prod-k8s
    kind: cluster
    name: Production K8s Cluster
  - id: srv-k8s-01
    kind: server
    name: K8s Node 01
    attributes:
      owner: cluster-prod-k8s
  - id: srv-k8s-02
    kind: server
    name: K8s Node 02
    attributes:
      owner: cluster-prod-k8s
  - id: app-web
    kind: application
    name: Web Application
    attributes:
      owner: cluster-prod-k8s
    spec:
      containers:
        - id: ctr-nginx-01
          name: nginx-01
          spec:
            image: nginx:latest
  - id: rel-hosts-01
    type: hosts
    participants:
      source: srv-k8s-01
      target: app-web
  - id: rel-hosts-02
    type: hosts
    participants:
      source: srv-k8s-02
      target: app-web
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// Verify entities
		srv1, ok := g.GetEntity("srv-k8s-01")
		if !ok || srv1.Owner != "cluster-prod-k8s" {
			t.Error("server 01 owner mismatch")
		}
		srv2, ok := g.GetEntity("srv-k8s-02")
		if !ok || srv2.Owner != "cluster-prod-k8s" {
			t.Error("server 02 owner mismatch")
		}
		app, ok := g.GetEntity("app-web")
		if !ok || app.Owner != "cluster-prod-k8s" {
			t.Error("application owner mismatch")
		}
		ctr, ok := g.GetEntity("ctr-nginx-01")
		if !ok || ctr.Owner != "app-web" {
			t.Error("container owner mismatch")
		}

		// Verify explicit hosts relations (multi-node)
		rel1, ok := g.GetRelation("rel-hosts-01")
		if !ok || rel1.Type != types.Hosts {
			t.Error("missing explicit hosts relation 01")
		}
		rel2, ok := g.GetRelation("rel-hosts-02")
		if !ok || rel2.Type != types.Hosts {
			t.Error("missing explicit hosts relation 02")
		}

		// Verify auto-relation from nesting
		autoRel, ok := g.GetRelation("rel-auto-hosts-app-web-ctr-nginx-01")
		if !ok || autoRel.Type != types.Hosts {
			t.Error("missing auto hosts relation from container to application")
		}
	})

	t.Run("container on multiple servers via hosts relations", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
  - id: srv-02
    kind: server
    name: Server 02
  - id: app-web
    kind: application
    name: Web Application
    spec:
      containers:
        - id: ctr-nginx-01
          name: nginx-01
          spec:
            image: nginx:latest
  - id: rel-hosts-01
    type: hosts
    participants:
      source: srv-01
      target: ctr-nginx-01
  - id: rel-hosts-02
    type: hosts
    participants:
      source: srv-02
      target: ctr-nginx-01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		ctr, ok := g.GetEntity("ctr-nginx-01")
		if !ok {
			t.Fatal("container not found")
		}
		if ctr.Owner != "app-web" {
			t.Errorf("expected container owner to be app-web, got %s", ctr.Owner)
		}

		rel1, ok := g.GetRelation("rel-hosts-01")
		if !ok || rel1.Type != types.Hosts {
			t.Error("missing hosts relation from srv-01 to container")
		}
		if rel1.Participants.Source != "srv-01" || rel1.Participants.Target != "ctr-nginx-01" {
			t.Errorf("wrong participants: %s -> %s", rel1.Participants.Source, rel1.Participants.Target)
		}

		rel2, ok := g.GetRelation("rel-hosts-02")
		if !ok || rel2.Type != types.Hosts {
			t.Error("missing hosts relation from srv-02 to container")
		}
		if rel2.Participants.Source != "srv-02" || rel2.Participants.Target != "ctr-nginx-01" {
			t.Errorf("wrong participants: %s -> %s", rel2.Participants.Source, rel2.Participants.Target)
		}

		autoRel, ok := g.GetRelation("rel-auto-hosts-app-web-ctr-nginx-01")
		if !ok || autoRel.Type != types.Hosts {
			t.Error("missing auto hosts relation from application to container")
		}
	})
}
}

func TestAutoRelationGeneration(t *testing.T) {
	t.Run("server nesting vm generates hosts relation", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      vms:
        - id: vm-01
          name: VM 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		rel, ok := g.GetRelation("rel-auto-hosts-srv-01-vm-01")
		if !ok {
			t.Fatal("auto-relation not generated")
		}
		if rel.Type != types.Hosts {
			t.Errorf("expected hosts, got %s", rel.Type)
		}
		if rel.Participants.Source != "srv-01" {
			t.Errorf("expected source srv-01, got %s", rel.Participants.Source)
		}
		if rel.Participants.Target != "vm-01" {
			t.Errorf("expected target vm-01, got %s", rel.Participants.Target)
		}
		if val, ok := rel.GetLabel("auto_generated"); !ok || val != "true" {
			t.Error("expected auto_generated label")
		}
	})

	t.Run("server nesting network generates belongs_to relation", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      networks:
        - id: net-01
          name: Network 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		rel, ok := g.GetRelation("rel-auto-belongs_to-net-01-srv-01")
		if !ok {
			t.Fatal("auto-relation not generated")
		}
		if rel.Type != types.BelongsTo {
			t.Errorf("expected belongs_to, got %s", rel.Type)
		}
		if rel.Participants.Source != "net-01" {
			t.Errorf("expected source net-01, got %s", rel.Participants.Source)
		}
		if rel.Participants.Target != "srv-01" {
			t.Errorf("expected target srv-01, got %s", rel.Participants.Target)
		}
	})

	t.Run("rack nesting server generates belongs_to relation", func(t *testing.T) {
		yaml := `
objects:
  - id: rack-01
    kind: rack
    name: Rack 01
    spec:
      servers:
        - id: srv-01
          name: Server 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		rel, ok := g.GetRelation("rel-auto-belongs_to-srv-01-rack-01")
		if !ok {
			t.Fatal("auto-relation not generated")
		}
		if rel.Type != types.BelongsTo {
			t.Errorf("expected belongs_to, got %s", rel.Type)
		}
	})

	t.Run("explicit relation skips auto-relation", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      vms:
        - id: vm-01
          name: VM 01
  - id: rel-hosts-1
    type: hosts
    participants:
      source: srv-01
      target: vm-01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// Explicit relation should exist
		_, ok := g.GetRelation("rel-hosts-1")
		if !ok {
			t.Fatal("explicit relation not found")
		}

		// Auto-relation should not be generated (duplicate skipped)
		autoRel, ok := g.GetRelation("rel-auto-hosts-srv-01-vm-01")
		if ok {
			t.Errorf("auto-relation should have been skipped, but found: %v", autoRel)
		}
	})

	t.Run("multi-level nesting", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      vms:
        - id: vm-01
          name: VM 01
          spec:
            applications:
              - id: app-01
                name: App 01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// server -> vm (hosts)
		rel1, ok := g.GetRelation("rel-auto-hosts-srv-01-vm-01")
		if !ok {
			t.Fatal("auto-relation server->vm not generated")
		}
		if rel1.Type != types.Hosts {
			t.Errorf("expected hosts, got %s", rel1.Type)
		}

		// vm -> application (hosts)
		rel2, ok := g.GetRelation("rel-auto-hosts-vm-01-app-01")
		if !ok {
			t.Fatal("auto-relation vm->app not generated")
		}
		if rel2.Type != types.Hosts {
			t.Errorf("expected hosts, got %s", rel2.Type)
		}
	})
}

func TestAutoRelationConfig(t *testing.T) {
	t.Run("disabled auto-relation", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      vms:
        - id: vm-01
          name: VM 01
`
		config := &AutoRelationConfig{
			Disabled: true,
		}
		parser := NewParserWithAutoRelationConfig(schema.CoreSchema(), config)
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		_, ok := g.GetRelation("rel-auto-hosts-srv-01-vm-01")
		if ok {
			t.Error("auto-relation should not be generated when disabled")
		}
	})

	t.Run("custom override", func(t *testing.T) {
		yaml := `
objects:
  - id: srv-01
    kind: server
    name: Server 01
    spec:
      vms:
        - id: vm-01
          name: VM 01
`
		config := &AutoRelationConfig{
			Overrides: map[string]AutoRelationMapping{
				"server.vms": {
					RelationType: types.DependsOn,
					Source:       "parent",
				},
			},
		}
		parser := NewParserWithAutoRelationConfig(schema.CoreSchema(), config)
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		rel, ok := g.GetRelation("rel-auto-depends_on-srv-01-vm-01")
		if !ok {
			t.Fatal("auto-relation not generated")
		}
		if rel.Type != types.DependsOn {
			t.Errorf("expected depends_on, got %s", rel.Type)
		}
	})
}

func TestAutoRelationIntegration(t *testing.T) {
	t.Run("nesting to flat round-trip preserves auto-relations", func(t *testing.T) {
		nestedYAML := `
objects:
  - id: site-01
    kind: site
    name: Site 01
    spec:
      racks:
        - id: rack-01
          name: Rack 01
          spec:
            servers:
              - id: srv-01
                name: Server 01
                spec:
                  vms:
                    - id: vm-01
                      name: VM 01
`
		parser := NewParser()
		g1, err := parser.Parse([]byte(nestedYAML))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		if g1.EntityCount() != 4 {
			t.Errorf("expected 4 entities, got %d", g1.EntityCount())
		}
		if g1.RelationCount() != 3 {
			t.Errorf("expected 3 auto-relations, got %d", g1.RelationCount())
		}

		// Verify specific relations
		rel1, ok := g1.GetRelation("rel-auto-belongs_to-rack-01-site-01")
		if !ok || rel1.Type != types.BelongsTo {
			t.Error("missing belongs_to rack->site")
		}

		rel2, ok := g1.GetRelation("rel-auto-belongs_to-srv-01-rack-01")
		if !ok || rel2.Type != types.BelongsTo {
			t.Error("missing belongs_to server->rack")
		}

		rel3, ok := g1.GetRelation("rel-auto-hosts-srv-01-vm-01")
		if !ok || rel3.Type != types.Hosts {
			t.Error("missing hosts server->vm")
		}

		// Serialize to flat YAML
		serializer := NewSerializer()
		data, err := serializer.Serialize(g1)
		if err != nil {
			t.Fatalf("failed to serialize: %v", err)
		}

		// Re-parse should produce equivalent graph
		parser2 := NewParser()
		g2, err := parser2.Parse(data)
		if err != nil {
			t.Fatalf("failed to re-parse: %v", err)
		}

		// All 4 entities should exist
		if g2.EntityCount() != 4 {
			t.Errorf("expected 4 entities after round-trip, got %d", g2.EntityCount())
		}

		// Verify ownership preserved
		rack, _ := g2.GetEntity("rack-01")
		if rack.Owner != "site-01" {
			t.Errorf("rack owner should be site-01, got %s", rack.Owner)
		}
		srv, _ := g2.GetEntity("srv-01")
		if srv.Owner != "rack-01" {
			t.Errorf("server owner should be rack-01, got %s", srv.Owner)
		}
		vm, _ := g2.GetEntity("vm-01")
		if vm.Owner != "srv-01" {
			t.Errorf("vm owner should be srv-01, got %s", vm.Owner)
		}
	})

	t.Run("all nesting types generate correct relations", func(t *testing.T) {
		yaml := `
objects:
  - id: site-01
    kind: site
    name: Site 01
    spec:
      clusters:
        - id: cluster-01
          name: Cluster 01
      firewalls:
        - id: fw-01
          name: Firewall 01
          spec:
            acls:
              - id: acl-01
                name: ACL 01
                spec:
                  acl_rules:
                    - id: rule-01
                      name: Rule 01
                      spec:
                        action: allow
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// site -> cluster (belongs_to, child source)
		r1, ok := g.GetRelation("rel-auto-belongs_to-cluster-01-site-01")
		if !ok || r1.Type != types.BelongsTo {
			t.Error("missing belongs_to cluster->site")
		}
		if r1.Participants.Source != "cluster-01" || r1.Participants.Target != "site-01" {
			t.Errorf("wrong participants: %s -> %s", r1.Participants.Source, r1.Participants.Target)
		}

		// site -> firewall (belongs_to, child source)
		r2, ok := g.GetRelation("rel-auto-belongs_to-fw-01-site-01")
		if !ok || r2.Type != types.BelongsTo {
			t.Error("missing belongs_to firewall->site")
		}

		// firewall -> acl (belongs_to, child source)
		r3, ok := g.GetRelation("rel-auto-belongs_to-acl-01-fw-01")
		if !ok || r3.Type != types.BelongsTo {
			t.Error("missing belongs_to acl->firewall")
		}

		// acl -> acl_rule (belongs_to, child source)
		r4, ok := g.GetRelation("rel-auto-belongs_to-rule-01-acl-01")
		if !ok || r4.Type != types.BelongsTo {
			t.Error("missing belongs_to acl_rule->acl")
		}
	})

	t.Run("explicit relation prevents auto-generation", func(t *testing.T) {
		yaml := `
objects:
  - id: rack-01
    kind: rack
    name: Rack 01
    spec:
      servers:
        - id: srv-01
          name: Server 01
  - id: rel-custom
    type: belongs_to
    participants:
      source: srv-01
      target: rack-01
`
		parser := NewParser()
		g, err := parser.Parse([]byte(yaml))
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		// Only the explicit relation should exist
		if g.RelationCount() != 1 {
			t.Errorf("expected 1 relation, got %d", g.RelationCount())
		}
		_, ok := g.GetRelation("rel-custom")
		if !ok {
			t.Error("explicit relation not found")
		}
		_, ok = g.GetRelation("rel-auto-belongs_to-srv-01-rack-01")
		if ok {
			t.Error("auto-relation should not exist when explicit one is present")
		}
	})
}
