package kinds

import "IACForge/src/core"

const (
	Site             core.EntityKind = "site"
	Rack             core.EntityKind = "rack"
	Server           core.EntityKind = "server"
	Interface        core.EntityKind = "interface"
	Cable            core.EntityKind = "cable"
	PowerDistribution core.EntityKind = "power_distribution"
	Network          core.EntityKind = "network"
	VLAN             core.EntityKind = "vlan"
	Switch           core.EntityKind = "switch"
	Router           core.EntityKind = "router"
	Firewall         core.EntityKind = "firewall"
	ACL              core.EntityKind = "acl"
	ACLRule          core.EntityKind = "acl_rule"
	VM               core.EntityKind = "vm"
	Container        core.EntityKind = "container"
	Application      core.EntityKind = "application"
	OpenPort         core.EntityKind = "open_port"
	Storage          core.EntityKind = "storage"
	Volume           core.EntityKind = "volume"
	Cluster          core.EntityKind = "cluster"
	AvailabilityZone core.EntityKind = "availability_zone"
)

var AllKinds = []core.EntityKind{
	Site, Rack, Server, Interface, Cable, PowerDistribution,
	Network, VLAN, Switch, Router, Firewall, ACL, ACLRule,
	VM, Container, Application, OpenPort,
	Storage, Volume,
	Cluster, AvailabilityZone,
}

var ValidStatuses = []core.Status{
	core.StatusPlanned,
	core.StatusActive,
	core.StatusMaintenance,
	core.StatusDeprecated,
	core.StatusOffline,
	core.StatusStandby,
}

func IsValidKind(k core.EntityKind) bool {
	for _, kind := range AllKinds {
		if kind == k {
			return true
		}
	}
	return false
}

func IsValidStatus(s core.Status) bool {
	for _, status := range ValidStatuses {
		if status == s {
			return true
		}
	}
	return false
}
