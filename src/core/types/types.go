package types

import (
	"IACForge/src/core"
	"IACForge/src/core/kinds"
)

const (
	Connects     core.RelationType = "connects"
	Hosts        core.RelationType = "hosts"
	DependsOn    core.RelationType = "depends_on"
	BelongsTo    core.RelationType = "belongs_to"
	ReplicatesTo core.RelationType = "replicates_to"
	BacksUp      core.RelationType = "backs_up"
	Monitors     core.RelationType = "monitors"
	ManagedBy    core.RelationType = "managed_by"
	MountedOn    core.RelationType = "mounted_on"
	AppliesTo    core.RelationType = "applies_to"
	ListensOn    core.RelationType = "listens_on"
)

var AllRelationTypes = []core.RelationType{
	Connects, Hosts, DependsOn, BelongsTo,
	ReplicatesTo, BacksUp, Monitors, ManagedBy,
	MountedOn, AppliesTo, ListensOn,
}

var RelationTypeDirections = map[core.RelationType]core.Direction{
	Connects:     core.DirectionSymmetric,
	Hosts:        core.DirectionDirected,
	DependsOn:    core.DirectionDirected,
	BelongsTo:    core.DirectionDirected,
	ReplicatesTo: core.DirectionDirected,
	BacksUp:      core.DirectionDirected,
	Monitors:     core.DirectionDirected,
	ManagedBy:    core.DirectionDirected,
	MountedOn:    core.DirectionDirected,
	AppliesTo:    core.DirectionDirected,
	ListensOn:    core.DirectionDirected,
}

var RelationTypeSourceKinds = map[core.RelationType][]core.EntityKind{
	Connects:     {kinds.Interface},
	Hosts:        {kinds.Server, kinds.VM, kinds.Container},
	DependsOn:    {kinds.VM, kinds.Container, kinds.Application},
	BelongsTo:    {kinds.VM, kinds.Container, kinds.Interface, kinds.Server, kinds.Switch, kinds.Router, kinds.Firewall, kinds.Storage, kinds.ACL, kinds.ACLRule, kinds.OpenPort},
	ReplicatesTo: {kinds.VM, kinds.Container, kinds.Application, kinds.Storage},
	BacksUp:      {kinds.VM, kinds.Container, kinds.Application, kinds.Storage, kinds.Volume},
	Monitors:     {kinds.Server, kinds.VM, kinds.Container, kinds.Application},
	ManagedBy:    {kinds.Server, kinds.VM, kinds.Container, kinds.Application, kinds.Switch, kinds.Router, kinds.Firewall, kinds.Storage},
	MountedOn:    {kinds.Volume},
	AppliesTo:    {kinds.ACL},
	ListensOn:    {kinds.OpenPort},
}

var RelationTypeTargetKinds = map[core.RelationType][]core.EntityKind{
	Connects:     {kinds.Interface},
	Hosts:        {kinds.VM, kinds.Container, kinds.Application},
	DependsOn:    {kinds.VM, kinds.Container, kinds.Application, kinds.Storage, kinds.Network},
	BelongsTo:    {kinds.Cluster, kinds.Network, kinds.Site, kinds.Firewall, kinds.Interface, kinds.Server, kinds.VM, kinds.Container, kinds.Application},
	ReplicatesTo: {kinds.VM, kinds.Container, kinds.Application, kinds.Storage},
	BacksUp:      {kinds.Volume, kinds.Storage},
	Monitors:     {kinds.Server, kinds.VM, kinds.Container, kinds.Application, kinds.Storage},
	ManagedBy:    {kinds.Server, kinds.VM, kinds.Container, kinds.Application},
	MountedOn:    {kinds.Server, kinds.VM, kinds.Container, kinds.Storage},
	AppliesTo:    {kinds.Interface, kinds.Firewall, kinds.Server, kinds.VM, kinds.Container},
	ListensOn:    {kinds.Interface, kinds.Server, kinds.VM, kinds.Container},
}

func IsValidRelationType(t core.RelationType) bool {
	for _, rt := range AllRelationTypes {
		if rt == t {
			return true
		}
	}
	return false
}

func GetDefaultDirection(t core.RelationType) core.Direction {
	if d, ok := RelationTypeDirections[t]; ok {
		return d
	}
	return core.DirectionDirected
}

func IsSourceKindValid(relType core.RelationType, kind core.EntityKind) bool {
	allowed, ok := RelationTypeSourceKinds[relType]
	if !ok {
		return true
	}
	for _, k := range allowed {
		if k == kind {
			return true
		}
	}
	return false
}

func IsTargetKindValid(relType core.RelationType, kind core.EntityKind) bool {
	allowed, ok := RelationTypeTargetKinds[relType]
	if !ok {
		return true
	}
	for _, k := range allowed {
		if k == kind {
			return true
		}
	}
	return false
}
