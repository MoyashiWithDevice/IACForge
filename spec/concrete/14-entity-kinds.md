# Entity Kinds

## Overview

Entity Kinds define the categories of objects that may exist in an infrastructure model.

Every Entity MUST define a kind.

The core specification defines the following Entity Kinds.

Implementations MAY introduce additional kinds through extensions.

---

## Common Properties

Every Entity shares the following common properties regardless of kind.

### Required

- id
- kind
- name

### Optional

- description
- status
- tags
- labels
- extensions

Individual Entity Kinds MAY define additional properties.

---

## Physical Infrastructure

### site

A physical location where infrastructure is deployed.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| address | string | no | - | Physical address |
| latitude | number | no | - | Geographic latitude |
| longitude | number | no | - | Geographic longitude |
| timezone | string | no | - | Timezone identifier |

#### Typical Ownership

- owned by: root (no owner specified)

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| racks | rack |
| clusters | cluster |

#### Typical Relations

- (none)

#### Example

```yaml
- id: site-tokyo-01
  kind: site
  name: Tokyo Datacenter 1
  status: active
  tags:
    - production
    - ap-northeast-1
  labels:
    region: asia-pacific
    tier: primary
```

---

### rack

A physical rack enclosure within a site.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| height_units | integer | no | 42 | Rack height in rack units (U) |
| power_capacity_watts | integer | no | - | Total power capacity in watts |
| max_load_kg | number | no | - | Maximum weight capacity in kg |

#### Typical Ownership

- owned by: site

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| servers | server |
| switches | switch |
| routers | router |
| firewalls | firewall |

#### Typical Relations

- belongs_to → site

#### Example

```yaml
- id: rack-a01
  kind: rack
  name: Rack A01
  status: active
  labels:
    row: A
    zone: dc-1
```

---

### server

A physical or virtual compute host.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| cpu_cores | integer | no | - | Total CPU cores |
| memory_gb | number | no | - | Total memory in GB |
| storage_gb | number | no | - | Total local storage in GB |
| platform | string | no | - | Virtualization platform (e.g., proxmox, vmware, kubernetes) |
| bios_version | string | no | - | BIOS/UEFI version |

#### Typical Ownership

- owned by: rack, site

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| networks | network |
| vms | vm |

#### Typical Relations

- belongs_to → rack
- belongs_to → site
- hosts → vm
- hosts → container

#### Example

```yaml
- id: srv-proxmox-01
  kind: server
  name: Proxmox Node 01
  status: active
  platform: proxmox
  cpu_cores: 32
  memory_gb: 128
  storage_gb: 2000
  ip_address: 10.0.1.10
```

---

### cable

A physical cable connecting two or more interfaces.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cable_type | string | no | copper | Cable type (copper, fiber, dac) |
| length_meters | number | no | - | Cable length in meters |
| connector_a | string | no | - | Connector type at end A |
| connector_b | string | no | - | Connector type at end B |

#### Typical Relations

- connects → interface (symmetric)

#### Example

```yaml
- id: cable-001
  kind: cable
  name: Patch Cable A01-01 to Switch-01-Port24
  cable_type: cat6a
  length_meters: 3.0
```

---

### power_distribution

A power distribution unit (PDU) or power feed.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| capacity_amps | integer | no | - | Total amperage capacity |
| voltage | number | no | 240 | Operating voltage |
| phases | integer | no | 1 | Number of phases |

#### Typical Relations

- belongs_to → rack
- connects → server
- connects → storage
- connects → switch

---

## Network

### network

A logical network or broadcast domain.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cidr | string | no | - | Network CIDR notation |
| gateway | string | no | - | Default gateway address |
| dns_servers | list[string] | no | - | DNS server addresses |
| vlan_id | integer | no | - | Associated VLAN ID |
| network_type | string | no | - | Network type (management, storage, vm, public) |

#### Typical Relations

- belongs_to → site
- belongs_to → cluster

#### Nestable Children

| Nest Key | Child Kind | Description |
|----------|------------|-------------|
| interfaces | interface | Network interfaces belonging to this network |

#### Interface Properties

When defining interfaces as nested children of a network, the following properties are available:

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| type | string | no | ethernet | Interface type (ethernet, fiber, wireless) |
| speed_mbps | integer | no | - | Interface speed in Mbps |
| mac_address | string | no | - | MAC address |
| ip_address | string | no | - | IP address if configured |
| mtu | integer | no | 1500 | Maximum transmission unit |

#### Interface Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| vlans | vlan |
| cables | cable |

#### Example

```yaml
- id: mgmt-network-01
  kind: network
  name: Management Network
  spec:
    cidr: 10.0.0.0/24
    gateway: 10.0.0.1
    network_type: management
    interfaces:
      - id: eth0
        spec:
          ip_address: 10.0.0.10
          type: ethernet
          speed_mbps: 10000
          mac_address: "aa:bb:cc:dd:ee:f0"
```

---

### vlan

A virtual LAN configuration.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| vlan_id | integer | yes | - | VLAN identifier (1-4094) |
| name | string | no | - | VLAN name |
| associated_network | string | no | - | Reference to parent network |

#### Typical Relations

- belongs_to → network
- belongs_to → site

---

### switch

A network switch.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| ports | integer | no | - | Total port count |
| managed | boolean | no | true | Whether switch is managed |
| stackable | boolean | no | false | Whether switch supports stacking |

#### Typical Ownership

- owned by: rack

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| interfaces | interface |

#### Typical Relations

- belongs_to → rack
- connects → server (via cable)
- connects → switch (via cable)
- connects → router (via cable)

---

### router

A network router.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| interfaces | integer | no | - | Number of interfaces |

#### Typical Ownership

- owned by: rack

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| interfaces | interface |

#### Typical Relations

- belongs_to → rack
- connects → switch (via cable)
- connects → firewall (via cable)

---

### firewall

A network firewall.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| throughput_gbps | number | no | - | Maximum throughput in Gbps |

#### Typical Ownership

- owned by: rack

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| interfaces | interface |
| acls | acl |

#### Typical Relations

- belongs_to → rack
- connects → router (via cable)

---

### acl

An Access Control List containing ordered rules for filtering network traffic.

An ACL is a container entity that holds `acl_rule` children in evaluation order.

Rules are evaluated top-to-bottom; the first matching rule determines the action.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| default_action | string | no | deny | Default action when no rule matches (allow, deny) |
| direction | string | no | - | Traffic direction this ACL applies to (inbound, outbound, both) |
| protocol | string | no | any | Protocol filter (tcp, udp, icmp, any) |

#### Typical Ownership

- owned by: firewall, interface, server, vm, container

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| acl_rules | acl_rule |

#### Typical Relations

- belongs_to → firewall
- belongs_to → interface
- belongs_to → server
- belongs_to → vm
- belongs_to → container
- applies_to → interface (via applies_to)
- applies_to → firewall (via applies_to)

#### Example

```yaml
- id: acl-web-ingress
  kind: acl
  name: Web Server Ingress ACL
  status: active
  direction: inbound
  default_action: deny
  labels:
    environment: production
    tier: web
```

---

### acl_rule

A single rule within an Access Control List.

ACL rules are evaluated in order within their parent ACL.

The first matching rule determines whether traffic is allowed or denied.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| action | string | yes | - | Rule action (allow, deny) |
| protocol | string | no | any | Protocol (tcp, udp, icmp, any) |
| source_address | string | no | any | Source IP address or CIDR |
| source_port | string | no | any | Source port or range (e.g., "80", "1024-65535") |
| destination_address | string | no | any | Destination IP address or CIDR |
| destination_port | string | no | any | Destination port or range (e.g., "443", "8080-8090") |
| enabled | boolean | no | true | Whether this rule is active |

#### Typical Ownership

- owned by: acl

#### Typical Relations

- (none)

#### Example

```yaml
- id: acl-rule-allow-https
  kind: acl_rule
  name: Allow HTTPS
  action: allow
  protocol: tcp
  source_address: 0.0.0.0/0
  destination_port: "443"
  enabled: true

- id: acl-rule-allow-ssh
  kind: acl_rule
  name: Allow SSH from Management
  action: allow
  protocol: tcp
  source_address: 10.0.0.0/24
  destination_port: "22"
  enabled: true
```

---

## Compute

### vm

A virtual machine.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cpu_cores | integer | no | - | Number of virtual CPUs |
| memory_gb | number | no | - | Memory in GB |
| storage_gb | number | no | - | Virtual disk size in GB |
| os | string | no | - | Operating system |
| os_version | string | no | - | Operating system version |

#### Typical Ownership

- owned by: server

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| networks | network |
| applications | application |

#### Typical Relations

- belongs_to → server
- belongs_to → cluster
- hosts → application

#### Example

```yaml
- id: vm-web-01
  kind: vm
  name: Web Server 01
  cpu_cores: 4
  memory_gb: 8
  storage_gb: 100
  os: ubuntu
  os_version: "22.04"
  ip_address: 10.0.2.10
```

---

### container

A containerized workload.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| image | string | no | - | Container image |
| image_tag | string | no | latest | Image tag |
| cpu_limit | string | no | - | CPU limit (e.g., "2.0") |
| memory_limit | string | no | - | Memory limit (e.g., "512Mi") |
| ports | list[integer] | no | - | Exposed ports |

#### Typical Relations

- belongs_to → vm
- belongs_to → server (direct)
- hosts → application

---

### application

A software application or service.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| version | string | no | - | Application version |
| port | integer | no | - | Primary listening port |
| protocol | string | no | - | Network protocol (http, https, tcp, udp) |
| url | string | no | - | Application URL if applicable |

#### Typical Relations

- belongs_to → vm
- belongs_to → container
- depends_on → vm
- depends_on → application

#### Nestable Children

| Nest Key | Child Kind |
|----------|------------|
| open_ports | open_port |

#### Example

```yaml
- id: app-web-server
  kind: application
  name: Nginx Web Server
  version: "1.24.0"
  port: 443
  protocol: https
```

---

### open_port

A listening or open network port on a host, VM, container, or application.

Represents a discovered or declared port that is accepting connections.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| port | integer | yes | - | Port number (1-65535) |
| protocol | string | yes | - | Transport protocol (tcp, udp) |
| state | string | no | listening | Port state (listening, established, closed) |
| address | string | no | 0.0.0.0 | Listening IP address |
| process | string | no | - | Process or service name using this port |
| pid | integer | no | - | Process ID if known |

#### Typical Relations

- belongs_to → server
- belongs_to → vm
- belongs_to → container
- belongs_to → application
- listens_on → interface (via listens_on)

#### Example

```yaml
- id: port-443-nginx
  kind: open_port
  name: Nginx HTTPS
  port: 443
  protocol: tcp
  state: listening
  address: 0.0.0.0
  process: nginx

- id: port-5432-postgres
  kind: open_port
  name: PostgreSQL
  port: 5432
  protocol: tcp
  state: listening
  address: 10.0.2.10
  process: postgres
```

---

## Storage

### storage

A storage system or array.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| total_capacity_gb | number | no | - | Total raw capacity in GB |
| usable_capacity_gb | number | no | - | Usable capacity after redundancy |
| raid_level | string | no | - | RAID level if applicable |
| protocol | string | no | - | Storage protocol (nfs, iscsi, fc, local) |

#### Typical Ownership

- owned by: rack

#### Typical Relations

- belongs_to → rack
- hosts → vm (for boot storage)

---

### volume

A logical storage volume.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| capacity_gb | number | no | - | Volume capacity in GB |
| filesystem | string | no | - | Filesystem type if mounted |
| mount_point | string | no | - | Mount point if applicable |
| thin_provisioned | boolean | no | false | Whether volume is thin provisioned |

#### Typical Relations

- belongs_to → storage
- belongs_to → server (local disks)
- hosts → vm

---

## Logical

### cluster

A logical grouping of compute resources.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cluster_type | string | no | - | Cluster type (compute, storage, hyperconverged) |
| ha_enabled | boolean | no | false | Whether HA is enabled |
| drs_enabled | boolean | no | false | Whether DRS is enabled |

#### Typical Ownership

- owned by: site

#### Typical Relations

- belongs_to → site
- belongs_to → network

#### Example

```yaml
- id: cluster-prod-01
  kind: cluster
  name: Production Cluster 01
  cluster_type: hyperconverged
  ha_enabled: true
  drs_enabled: true
```

---

### availability_zone

A logical availability zone within a site.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| redundancy | string | no | - | Redundancy level (n+1, 2n, etc.) |

#### Typical Ownership

- owned by: site

#### Typical Relations

- belongs_to → site

---

## Vendor Kinds

Vendor-specific Entity Kinds MUST NOT replace core kinds.

Vendors MAY introduce additional kinds through extensions.

### Extension Naming Convention

Extension kinds MUST use namespace prefixes.

Examples:

- `proxmox.vm` - Proxmox-specific VM extension
- `kubernetes.pod` - Kubernetes pod
- `aws.vpc` - AWS Virtual Private Cloud
- `network.switch` - Extended switch properties

---

## Status Values

Every Entity MAY have a status.

The core specification defines the following statuses:

| Status | Description |
|--------|-------------|
| planned | Entity is planned but not yet deployed |
| active | Entity is operational |
| maintenance | Entity is under maintenance |
| deprecated | Entity is scheduled for removal |
| offline | Entity is not operational |

Implementations MAY introduce additional statuses.

---

## Equality

Two Entities are considered different if their identifiers differ.

Changing properties does not create a new Entity.

Changing an identifier creates a different Entity.
