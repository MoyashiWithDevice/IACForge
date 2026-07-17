# Entity Kinds

[← README](README.md)

---

## Physical Infrastructure

### site

A physical location where infrastructure is deployed.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| address | string | no | - | Physical address |
| latitude | number | no | - | Geographic latitude |
| longitude | number | no | - | Geographic longitude |
| timezone | string | no | - | Timezone identifier |

**Ownership:** Root (no owner specified)

```yaml
- id: site-tokyo-01
  kind: site
  name: Tokyo Datacenter 1
  attributes:
    status: active
    tags:
      - production
    labels:
      region: asia-pacific
      tier: primary
```

---

### rack

A physical rack enclosure within a site.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| height_units | integer | no | 42 | Rack height in rack units (U) |
| power_capacity_watts | integer | no | - | Total power capacity in watts |
| max_load_kg | number | no | - | Maximum weight capacity in kg |

**Ownership:** site

```yaml
- id: rack-a01
  kind: rack
  name: Rack A01
  attributes:
    owner: site-tokyo-01
    status: active
    labels:
      row: A
      zone: dc-1
  spec:
    height_units: 42
```

---

### server

A physical or virtual compute host.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| cpu | list[object] | no | - | CPU configurations |
| memory | list[object] | no | - | Memory modules |
| storage | list[object] | no | - | Local storage devices |
| platform | string | no | - | Virtualization platform (proxmox, vmware, kubernetes) |
| bios_version | string | no | - | BIOS/UEFI version |

##### cpu Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cores | integer | no | - | Number of CPU cores |
| architecture | string | no | - | CPU architecture (x86_64, arm64) |

##### memory Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| size_gb | number | yes | - | Memory module size in GB |
| speed | integer | no | - | Memory speed in MHz |
| type | string | no | - | Memory type (ddr4, ddr5, lpddr4, lpddr5) |

##### storage Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| size_gb | number | no | - | Storage size in GB |
| type | string | no | - | Storage type (ssd, hdd, nvme) |

**Ownership:** rack

**Note:** IP addresses are not direct properties of server. Use interface entities to assign IP addresses.

```yaml
- id: srv-proxmox-01
  kind: server
  name: Proxmox Node 01
  attributes:
    owner: rack-a01
    status: active
  spec:
    platform: proxmox
    cpu:
      - cores: 16
        architecture: x86_64
      - cores: 16
        architecture: x86_64
    memory:
      - size_gb: 64
        speed: 3200
        type: ddr4
      - size_gb: 64
        speed: 3200
        type: ddr4
    storage:
      - size_gb: 500
        type: ssd
      - size_gb: 500
        type: ssd
```

---

### interface

A network interface on a device.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| type | string | no | ethernet | Interface type (ethernet, fiber, wireless) |
| speed_mbps | integer | no | - | Interface speed in Mbps |
| mac_address | string | no | - | MAC address |
| ip_address | string | no | - | IP address if configured |
| mtu | integer | no | 1500 | Maximum transmission unit |

**Ownership:** server, switch, router, vm

**Nestable Children:** vlans (vlan), cables (cable)

```yaml
- id: eno1
  kind: interface
  name: eno1
  attributes:
    owner: srv-proxmox-01
  spec:
    type: ethernet
    speed_mbps: 10000
    mac_address: "aa:bb:cc:dd:ee:f0"
    ip_address: 10.0.1.10
    vlans:
      - id: eno1-vlan100
        spec:
          vlan_id: 100
    cables:
      - id: cable-eno1-sw01
        spec:
          cable_type: cat6a
          length_meters: 2.0
```

---

### cable

A physical cable connecting two or more interfaces.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cable_type | string | no | copper | Cable type (copper, fiber, dac, cat6a) |
| length_meters | number | no | - | Cable length in meters |
| connector_a | string | no | - | Connector type at end A |
| connector_b | string | no | - | Connector type at end B |

**Ownership:** Optional (typically no owner)

```yaml
- id: cable-001
  kind: cable
  name: Patch Cable SRV01-SW01
  spec:
    cable_type: cat6a
    length_meters: 3.0
```

---

### power_distribution

A power distribution unit (PDU) or power feed.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| capacity_amps | integer | no | - | Total amperage capacity |
| voltage | number | no | 240 | Operating voltage |
| phases | integer | no | 1 | Number of phases |

**Ownership:** rack

```yaml
- id: pdu-rack-01
  kind: power_distribution
  name: PDU Rack A01
  attributes:
    owner: rack-a01
  spec:
    capacity_amps: 30
    voltage: 240
    phases: 3
```

---

## Network

### network

A logical network or broadcast domain.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cidr | string | no | - | Network CIDR notation |
| gateway | string | no | - | Default gateway address |
| dns_servers | list[string] | no | - | DNS server addresses |
| vlan_id | integer | no | - | Associated VLAN ID |
| network_type | string | no | - | Network type (management, storage, vm, public) |

**Ownership:** Optional (typically site or cluster)

```yaml
- id: mgmt-network-01
  kind: network
  name: Management Network
  spec:
    cidr: 10.0.0.0/24
    gateway: 10.0.0.1
    network_type: management
    dns_servers:
      - 8.8.8.8
      - 8.8.4.4
```

---

### vlan

A virtual LAN configuration.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| vlan_id | integer | yes | - | VLAN identifier (1-4094) |
| name | string | no | - | VLAN name |
| associated_network | string | no | - | Reference to parent network |

**Ownership:** network, site

```yaml
- id: vlan-100
  kind: vlan
  name: Production VLAN
  attributes:
    owner: mgmt-network-01
  spec:
    vlan_id: 100
    associated_network: mgmt-network-01
```

---

### switch

A network switch.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| ports | integer | no | - | Total port count |
| managed | boolean | no | true | Whether switch is managed |
| stackable | boolean | no | false | Whether switch supports stacking |

**Ownership:** rack

```yaml
- id: sw-core-01
  kind: switch
  name: Core Switch 01
  attributes:
    owner: rack-a01
    status: active
  spec:
    manufacturer: cisco
    model: Catalyst 9300
    ports: 48
    managed: true
```

---

### router

A network router.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| interfaces | integer | no | - | Number of interfaces |

**Ownership:** rack

```yaml
- id: rt-core-01
  kind: router
  name: Core Router 01
  attributes:
    owner: rack-a01
  spec:
    manufacturer: mikrotik
    model: CCR1036
    interfaces: 36
```

---

### firewall

A network firewall.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| serial_number | string | no | - | Serial number |
| throughput_gbps | number | no | - | Maximum throughput in Gbps |

**Ownership:** rack

```yaml
- id: fw-core-01
  kind: firewall
  name: Core Firewall 01
  attributes:
    owner: rack-a01
  spec:
    manufacturer: paloalto
    model: PA-3260
    throughput_gbps: 10.0
```

---

### acl

An Access Control List containing ordered rules for filtering network traffic.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| default_action | string | no | deny | Default action when no rule matches (allow, deny) |
| direction | string | no | - | Traffic direction (inbound, outbound, both) |
| protocol | string | no | any | Protocol filter (tcp, udp, icmp, any) |

**Ownership:** firewall, interface, server, vm, container

```yaml
- id: acl-web-ingress
  kind: acl
  name: Web Server Ingress ACL
  attributes:
    owner: vm-web-01
    status: active
    labels:
      environment: production
  spec:
    direction: inbound
    default_action: deny
```

---

### acl_rule

A single rule within an Access Control List.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| action | string | yes | - | Rule action (allow, deny) |
| protocol | string | no | any | Protocol (tcp, udp, icmp, any) |
| source_address | string | no | any | Source IP address or CIDR |
| source_port | string | no | any | Source port or range (e.g., "80", "1024-65535") |
| destination_address | string | no | any | Destination IP address or CIDR |
| destination_port | string | no | any | Destination port or range (e.g., "443", "8080-8090") |
| enabled | boolean | no | true | Whether this rule is active |

**Ownership:** acl

```yaml
- id: acl-rule-allow-https
  kind: acl_rule
  name: Allow HTTPS
  attributes:
    owner: acl-web-ingress
  spec:
    action: allow
    protocol: tcp
    source_address: 0.0.0.0/0
    destination_port: "443"
    enabled: true

- id: acl-rule-allow-ssh
  kind: acl_rule
  name: Allow SSH from Management
  attributes:
    owner: acl-web-ingress
  spec:
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

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cpu | list[object] | no | - | Virtual CPU configurations |
| memory | list[object] | no | - | Memory modules |
| storage | list[object] | no | - | Virtual disk configurations |
| os | string | no | - | Operating system |
| os_version | string | no | - | Operating system version |

##### cpu Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cores | integer | no | - | Number of virtual CPU cores |
| architecture | string | no | - | CPU architecture (x86_64, arm64) |

##### memory Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| size_gb | number | yes | - | Memory module size in GB |
| speed | integer | no | - | Memory speed in MHz |
| type | string | no | - | Memory type (ddr4, ddr5, lpddr4, lpddr5) |

##### storage Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| size_gb | number | no | - | Disk size in GB |
| type | string | no | - | Disk type (ssd, hdd, nvme) |

**Ownership:** server

**Note:** IP addresses are not direct properties of vm. Use interface entities to assign IP addresses.

```yaml
- id: vm-web-01
  kind: vm
  name: Web Server 01
  attributes:
    owner: srv-proxmox-01
    status: active
  spec:
    cpu:
      - cores: 4
        architecture: x86_64
    memory:
      - size_gb: 8
        speed: 3200
        type: ddr4
    storage:
      - size_gb: 100
        type: ssd
    os: ubuntu
    os_version: "22.04"
```

---

### container

A containerized workload.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| image | string | no | - | Container image |
| image_tag | string | no | latest | Image tag |
| cpu_limit | string | no | - | CPU limit (e.g., "2.0") |
| memory_limit | string | no | - | Memory limit (e.g., "512Mi") |
| ports | list[integer] | no | - | Exposed ports |

**Ownership:** vm, server

```yaml
- id: ctr-nginx-01
  kind: container
  name: Nginx Container
  attributes:
    owner: vm-web-01
  spec:
    image: nginx
    image_tag: "1.24"
    cpu_limit: "1.0"
    memory_limit: "256Mi"
    ports:
      - 80
      - 443
```

---

### application

A software application or service.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| version | string | no | - | Application version |
| port | integer | no | - | Primary listening port |
| protocol | string | no | - | Network protocol (http, https, tcp, udp) |
| url | string | no | - | Application URL if applicable |

**Ownership:** vm, container

```yaml
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
```

---

### open_port

A listening or open network port on a host, VM, container, or application.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| port | integer | yes | - | Port number (1-65535) |
| protocol | string | yes | - | Transport protocol (tcp, udp) |
| state | string | no | listening | Port state (listening, established, closed) |
| address | string | no | 0.0.0.0 | Listening IP address |
| process | string | no | - | Process or service name using this port |
| pid | integer | no | - | Process ID if known |

**Ownership:** server, vm, container, application

```yaml
- id: port-443-nginx
  kind: open_port
  name: Nginx HTTPS
  attributes:
    owner: app-web-server
  spec:
    port: 443
    protocol: tcp
    state: listening
    address: 0.0.0.0
    process: nginx

- id: port-5432-postgres
  kind: open_port
  name: PostgreSQL
  attributes:
    owner: vm-web-01
  spec:
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

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| manufacturer | string | no | - | Hardware manufacturer |
| model | string | no | - | Hardware model |
| total_capacity_gb | number | no | - | Total raw capacity in GB |
| usable_capacity_gb | number | no | - | Usable capacity after redundancy |
| raid_level | string | no | - | RAID level if applicable |
| protocol | string | no | - | Storage protocol (nfs, iscsi, fc, local) |

**Ownership:** rack

```yaml
- id: storage-nas-01
  kind: storage
  name: NAS Storage 01
  attributes:
    owner: rack-a01
  spec:
    manufacturer: synology
    model: DS1621+
    total_capacity_gb: 48000
    usable_capacity_gb: 32000
    raid_level: raid6
    protocol: nfs
```

---

### volume

A logical storage volume.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| capacity_gb | number | no | - | Volume capacity in GB |
| filesystem | string | no | - | Filesystem type if mounted |
| mount_point | string | no | - | Mount point if applicable |
| thin_provisioned | boolean | no | false | Whether volume is thin provisioned |

**Ownership:** storage, server

```yaml
- id: vol-web-data
  kind: volume
  name: Web Data Volume
  attributes:
    owner: storage-nas-01
  spec:
    capacity_gb: 500
    filesystem: ext4
    mount_point: /data
    thin_provisioned: false
```

---

## Logical

### cluster

A logical grouping of compute resources.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| cluster_type | string | no | - | Cluster type (compute, storage, hyperconverged) |
| ha_enabled | boolean | no | false | Whether HA is enabled |
| drs_enabled | boolean | no | false | Whether DRS is enabled |

**Ownership:** site

```yaml
- id: cluster-prod-01
  kind: cluster
  name: Production Cluster 01
  attributes:
    status: active
  spec:
    cluster_type: hyperconverged
    ha_enabled: true
    drs_enabled: true
```

---

### availability_zone

A logical availability zone within a site.

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| redundancy | string | no | - | Redundancy level (n+1, 2n, etc.) |

**Ownership:** site

```yaml
- id: az-tokyo-a
  kind: availability_zone
  name: Tokyo Zone A
  attributes:
    owner: site-tokyo-01
  spec:
    redundancy: 2n
```

---

## Entity Kinds一覧

| Kind | Category | Description |
|------|----------|-------------|
| site | Physical | Physical location |
| rack | Physical | Physical rack enclosure |
| server | Physical | Physical or virtual compute host |
| interface | Network | Network interface |
| cable | Physical | Physical cable |
| power_distribution | Physical | PDU or power feed |
| network | Network | Logical network |
| vlan | Network | Virtual LAN |
| switch | Network | Network switch |
| router | Network | Network router |
| firewall | Network | Network firewall |
| acl | Network | Access Control List |
| acl_rule | Network | Individual ACL rule |
| vm | Compute | Virtual machine |
| container | Compute | Containerized workload |
| application | Compute | Software application |
| open_port | Compute | Listening network port |
| storage | Storage | Storage system |
| volume | Storage | Logical storage volume |
| cluster | Logical | Logical compute grouping |
| availability_zone | Logical | Logical availability zone |
