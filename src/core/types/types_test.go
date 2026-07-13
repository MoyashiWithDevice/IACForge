package types

import (
	"testing"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
)

func TestAllRelationTypesCount(t *testing.T) {
	if len(AllRelationTypes) != 11 {
		t.Errorf("expected 11 relation types, got %d", len(AllRelationTypes))
	}
}

func TestIsValidRelationType(t *testing.T) {
	tests := []struct {
		typ   core.RelationType
		valid bool
	}{
		{Connects, true},
		{Hosts, true},
		{DependsOn, true},
		{BelongsTo, true},
		{ReplicatesTo, true},
		{BacksUp, true},
		{Monitors, true},
		{ManagedBy, true},
		{MountedOn, true},
		{AppliesTo, true},
		{ListensOn, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.typ), func(t *testing.T) {
			if got := IsValidRelationType(tt.typ); got != tt.valid {
				t.Errorf("IsValidRelationType(%s) = %v, want %v", tt.typ, got, tt.valid)
			}
		})
	}
}

func TestGetDefaultDirection(t *testing.T) {
	symmetricTypes := []core.RelationType{Connects}
	for _, rt := range symmetricTypes {
		if d := GetDefaultDirection(rt); d != core.DirectionSymmetric {
			t.Errorf("GetDefaultDirection(%s) = %s, want symmetric", rt, d)
		}
	}

	directedTypes := []core.RelationType{
		Hosts, DependsOn, BelongsTo, ReplicatesTo,
		BacksUp, Monitors, ManagedBy, MountedOn, AppliesTo, ListensOn,
	}
	for _, rt := range directedTypes {
		if d := GetDefaultDirection(rt); d != core.DirectionDirected {
			t.Errorf("GetDefaultDirection(%s) = %s, want directed", rt, d)
		}
	}
}

func TestGetDefaultDirectionUnknown(t *testing.T) {
	d := GetDefaultDirection("unknown")
	if d != core.DirectionDirected {
		t.Errorf("GetDefaultDirection(unknown) = %s, want directed", d)
	}
}

func TestRelationTypeValues(t *testing.T) {
	expected := map[core.RelationType]string{
		Connects:     "connects",
		Hosts:        "hosts",
		DependsOn:    "depends_on",
		BelongsTo:    "belongs_to",
		ReplicatesTo: "replicates_to",
		BacksUp:      "backs_up",
		Monitors:     "monitors",
		ManagedBy:    "managed_by",
		MountedOn:    "mounted_on",
		AppliesTo:    "applies_to",
		ListensOn:    "listens_on",
	}

	for typ, value := range expected {
		if string(typ) != value {
			t.Errorf("type %v has wrong string value: got %s, want %s", typ, string(typ), value)
		}
	}
}

func TestSourceKindsPopulated(t *testing.T) {
	for _, rt := range AllRelationTypes {
		if _, ok := RelationTypeSourceKinds[rt]; !ok {
			t.Errorf("missing source kinds for relation type %s", rt)
		}
	}
}

func TestTargetKindsPopulated(t *testing.T) {
	for _, rt := range AllRelationTypes {
		if _, ok := RelationTypeTargetKinds[rt]; !ok {
			t.Errorf("missing target kinds for relation type %s", rt)
		}
	}
}

func TestIsSourceKindValid(t *testing.T) {
	if !IsSourceKindValid(Connects, kinds.Interface) {
		t.Error("expected interface to be valid source for connects")
	}
	if IsSourceKindValid(Connects, kinds.Server) {
		t.Error("expected server to be invalid source for connects")
	}
	if !IsSourceKindValid(Hosts, kinds.Server) {
		t.Error("expected server to be valid source for hosts")
	}
	if !IsSourceKindValid(Hosts, kinds.VM) {
		t.Error("expected vm to be valid source for hosts")
	}
	if !IsSourceKindValid(MountedOn, kinds.Volume) {
		t.Error("expected volume to be valid source for mounted_on")
	}
}

func TestIsTargetKindValid(t *testing.T) {
	if !IsTargetKindValid(Connects, kinds.Interface) {
		t.Error("expected interface to be valid target for connects")
	}
	if IsTargetKindValid(Connects, kinds.Server) {
		t.Error("expected server to be invalid target for connects")
	}
	if !IsTargetKindValid(Hosts, kinds.VM) {
		t.Error("expected vm to be valid target for hosts")
	}
	if !IsTargetKindValid(Hosts, kinds.Application) {
		t.Error("expected application to be valid target for hosts")
	}
	if !IsTargetKindValid(DependsOn, kinds.Storage) {
		t.Error("expected storage to be valid target for depends_on")
	}
	if !IsTargetKindValid(DependsOn, kinds.Network) {
		t.Error("expected network to be valid target for depends_on")
	}
}
