package kinds

import (
	"testing"

	"IACForge/src/core"
)

func TestAllKindsCount(t *testing.T) {
	if len(AllKinds) != 21 {
		t.Errorf("expected 21 entity kinds, got %d", len(AllKinds))
	}
}

func TestIsValidKind(t *testing.T) {
	tests := []struct {
		kind  core.EntityKind
		valid bool
	}{
		{Site, true},
		{Rack, true},
		{Server, true},
		{Interface, true},
		{Cable, true},
		{PowerDistribution, true},
		{Network, true},
		{VLAN, true},
		{Switch, true},
		{Router, true},
		{Firewall, true},
		{ACL, true},
		{ACLRule, true},
		{VM, true},
		{Container, true},
		{Application, true},
		{OpenPort, true},
		{Storage, true},
		{Volume, true},
		{Cluster, true},
		{AvailabilityZone, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			if got := IsValidKind(tt.kind); got != tt.valid {
				t.Errorf("IsValidKind(%s) = %v, want %v", tt.kind, got, tt.valid)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status core.Status
		valid  bool
	}{
		{core.StatusPlanned, true},
		{core.StatusActive, true},
		{core.StatusMaintenance, true},
		{core.StatusDeprecated, true},
		{core.StatusOffline, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := IsValidStatus(tt.status); got != tt.valid {
				t.Errorf("IsValidStatus(%s) = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestKindValues(t *testing.T) {
	expected := map[core.EntityKind]string{
		Site:              "site",
		Rack:              "rack",
		Server:            "server",
		Interface:         "interface",
		Cable:             "cable",
		PowerDistribution: "power_distribution",
		Network:           "network",
		VLAN:              "vlan",
		Switch:            "switch",
		Router:            "router",
		Firewall:          "firewall",
		ACL:               "acl",
		ACLRule:           "acl_rule",
		VM:                "vm",
		Container:         "container",
		Application:       "application",
		OpenPort:          "open_port",
		Storage:           "storage",
		Volume:            "volume",
		Cluster:           "cluster",
		AvailabilityZone:  "availability_zone",
	}

	for kind, value := range expected {
		if string(kind) != value {
			t.Errorf("kind %v has wrong string value: got %s, want %s", kind, string(kind), value)
		}
	}
}
