# YAML Syntax

## Overview

YAML is the canonical serialization format for the Object Model.

This specification defines the concrete YAML syntax.

All compliant implementations MUST support this syntax.

---

## Document Structure

A YAML document represents a single Graph.

The document root contains:

```yaml
objects:
  # Entities and Relations go here
```

---

## Entity Syntax

An Entity is defined with the following structure.

### Required Top-Level Properties

| Property | Type | Description |
|----------|------|-------------|
| id | string | Unique identifier |
| kind | string | Entity kind |
| name | string | Human-readable name |

### Attributes Section

The `attributes` sub-key contains optional properties common to all entities.

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| owner | string | - | Parent Entity identifier for ownership |
| description | string | - | Documentation |
| status | enum | - | Lifecycle state |
| tags | list[string] | - | Labels |
| labels | map[string] | - | Key-value metadata |
| extensions | map[string] | - | Extension data |

### Spec Section

The `spec` sub-key contains kind-specific properties (platform, cpu_cores, memory_gb, etc.).

### Basic Entity

```yaml
objects:
  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter 1
```

### Entity with All Properties

```yaml
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
```

---

## Nested Entity Definitions

Child entities can be defined inline within their parent entity's definition,
instead of as separate top-level objects. This provides a more concise and
hierarchical representation of the infrastructure.

### Nesting Rules

Each entity kind defines which child kinds can be nested and under which key.

| Parent Kind | Nest Key | Child Kind |
|-------------|----------|------------|
| site | racks | rack |
| site | clusters | cluster |
| rack | servers | server |
| rack | switches | switch |
| rack | routers | router |
| rack | firewalls | firewall |
| server | networks | network |
| server | vms | vm |
| switch | interfaces | interface |
| router | interfaces | interface |
| firewall | interfaces | interface |
| firewall | acls | acl |
| vm | networks | network |
| vm | applications | application |
| network | interfaces | interface |
| application | open_ports | open_port |
| acl | acl_rules | acl_rule |

### Syntax

Nested children are defined as lists under the appropriate nest key.
The nest key can appear either inside the `spec` section or at the entity
definition level.

```yaml
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
                spec:
                  ip_address: 172.31.0.15
                  type: ethernet
      vms:
        - id: vm-web-01
          name: Web Server 01
          spec:
            cpu_cores: 4
            memory_gb: 8
```

### Optional Fields in Nested Definitions

| Field | Required | Notes |
|-------|----------|-------|
| id | optional | Auto-generated from parent ID if omitted |
| kind | optional | Inferred from the nest key |
| name | optional | Defaults to ID if omitted |
| spec | optional | Kind-specific properties |

### ID Auto-Generation

When a nested entity omits its `id`, one is generated using the pattern:
`{parent-id}-{child-kind}`

If that ID is already in use (e.g., multiple children of the same kind),
a numeric suffix is appended: `{parent-id}-{child-kind}-{n}`

### Ownership

Nested entities automatically receive their parent's ID as their `owner`.
The `owner` field should NOT be specified in nested definitions.

### Reference Syntax for Nested Entities

Nested entities can be referenced by their ID:

```yaml
participants:
  source: eth1
  target: sw-core-01/port1
```

Or by path notation:

```yaml
participants:
  source: srv-proxmox-01/net-private/eth1
  target: sw-core-01/port1
```

### Mixed Definitions

Flat and nested definitions can be mixed in the same file:

```yaml
objects:
  # Flat definition
  - id: rack-a01
    kind: rack
    name: Rack A01
    attributes:
      owner: site-tokyo-01

  # Nested definition
  - id: srv-proxmox-01
    kind: server
    name: Proxmox Node 01
    spec:
      networks:
        - id: net-private
          spec:
            interfaces:
              - id: eth1
```

---

## Relation Syntax

A Relation is defined with the following structure.

### Required Top-Level Properties

| Property | Type | Description |
|----------|------|-------------|
| id | string | Unique identifier |
| type | string | Relation type |
| participants | list[string] or map | Entity references |

### Participant Formats

#### List Format (Symmetric Relations)

For symmetric relations like `connects`:

```yaml
participants:
  - srv-01/eno1
  - sw-01/port1
```

#### Map Format (Directed Relations)

For directed relations like `hosts`, `depends_on`:

```yaml
participants:
  source: site-tokyo-01
  target: rack-a01
```

### Attributes Section

The `attributes` sub-key contains optional properties common to all relations.

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| description | string | - | Documentation |
| status | enum | - | Lifecycle state |
| tags | list[string] | - | Labels |
| labels | map[string] | - | Key-value metadata |
| extensions | map[string] | - | Extension data |

### Spec Section

The `spec` sub-key contains relation-type-specific properties (connection_type, bandwidth_mbps, etc.).

### Relation with All Properties

```yaml
objects:
  - id: rel-connects-srv-sw
    type: connects
    attributes:
      description: "Physical connection between server and switch"
      status: active
      tags:
        - networking
      labels:
        speed: 10g
    spec:
      connection_type: physical
      bandwidth_mbps: 10000
    participants:
      - srv-proxmox-01/eno1
      - sw-core-01/port1
```

---

## Reference Syntax

References use Entity identifiers.

### Simple Reference

```yaml
source: srv-proxmox-01
target: vm-web-01
```

### Qualified Reference

For unambiguous references:

```yaml
source: /site-tokyo-01/rack-a01/srv-proxmox-01
target: vm-web-01
```

### Interface Reference

Interfaces are referenced with path notation:

```yaml
participants:
  - srv-proxmox-01/eno1
  - sw-core-01/port24
```

---

## Complete Example

### Infrastructure Model

```yaml
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

  - id: srv-proxmox-02
    kind: server
    name: Proxmox Node 02
    attributes:
      owner: rack-a01
      status: active
    spec:
      platform: proxmox
      cpu_cores: 32
      memory_gb: 128
      storage_gb: 2000
      ip_address: 10.0.1.11

  # Network with Interfaces
  - id: mgmt-network-01
    kind: network
    name: Management Network
    spec:
      cidr: 10.0.0.0/24
      gateway: 10.0.0.1
      network_type: management
      interfaces:
        - id: eno1
          spec:
            type: ethernet
            speed_mbps: 10000
            mac_address: "aa:bb:cc:dd:ee:f0"
            ip_address: 10.0.1.10
        - id: eno2
          spec:
            type: ethernet
            speed_mbps: 10000
            mac_address: "aa:bb:cc:dd:ee:f1"

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
      storage_gb: 100
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

  # Open Ports
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

  # ACLs
  - id: acl-web-ingress
    kind: acl
    name: Web Server Ingress ACL
    attributes:
      owner: vm-web-01
      status: active
    spec:
      direction: inbound
      default_action: deny

  # ACL Rules
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

  # Cluster
  - id: cluster-prod-01
    kind: cluster
    name: Production Cluster 01
    attributes:
      status: active
    spec:
      cluster_type: hyperconverged
      ha_enabled: true

  # Cables
  - id: cable-001
    kind: cable
    name: Patch Cable SRV01-SW01
    spec:
      cable_type: cat6a
      length_meters: 3.0

  # Connection Relations (connects)
  - id: rel-connects-srv-sw
    type: connects
    spec:
      connection_type: physical
      bandwidth_mbps: 10000
    participants:
      - mgmt-network-01/eno1
      - mgmt-network-01/eno2

  # Hosting Relations (hosts)
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

  # Membership Relations (belongs_to)
  - id: rel-belongsto-vm-cluster
    type: belongs_to
    participants:
      source: vm-web-01
      target: cluster-prod-01

  - id: rel-belongsto-intf-network
    type: belongs_to
    participants:
      source: mgmt-network-01/eno1
      target: mgmt-network-01

  # ACL Application Relations (applies_to)
  - id: rel-applies-web-acl
    type: applies_to
    participants:
      source: acl-web-ingress
      target: mgmt-network-01/eno1

  # Open Port Relations (belongs_to)
  - id: rel-belongsto-port-nginx
    type: belongs_to
    participants:
      source: port-443-nginx
      target: app-web-server

  - id: rel-belongsto-port-postgres
    type: belongs_to
    participants:
      source: port-5432-postgres
      target: vm-web-01

  # Port Listening Relations (listens_on)
  - id: rel-listens-nginx
    type: listens_on
    participants:
      source: port-443-nginx
      target: mgmt-network-01/eno1

  - id: rel-listens-postgres
    type: listens_on
    participants:
      source: port-5432-postgres
      target: mgmt-network-01/eno1
```

---

## Validation

### Required Fields

| Object | Required Fields |
|--------|-----------------|
| Entity | id, kind, name |
| Relation | id, type, participants |

### Ownership Validation

- Root Entity MUST NOT specify owner
- Non-root Entity MUST specify exactly one owner
- Owner identifier MUST reference an existing Entity
- Ownership MUST form exactly one tree

### Reference Validation

- References MUST point to existing Objects
- Interface references use path notation (entity/interface)
- Unknown references are validation errors

### Identifier Rules

- IDs MUST be unique within their scope
- IDs SHOULD be descriptive and stable
- IDs SHOULD follow naming conventions (kebab-case recommended)

---

## Naming Conventions

### Recommended Patterns

| Pattern | Example |
|---------|---------|
| kebab-case | `srv-proxmox-01` |
| snake_case | `mgmt_network_01` |
| camelCase | `mgmtNetwork01` |

### Kind Naming

- Use lowercase
- Use singular form
- Examples: `server`, `vm`, `interface`

### Relation Type Naming

- Use snake_case
- Use verbs or verb phrases
- Examples: `connects`, `hosts`, `depends_on`

---

## Comments

Comments are preserved during round-trip conversion.

```yaml
# Site information
objects:
  - id: site-tokyo-01
    kind: site
    name: Tokyo Datacenter 1
    # Primary location
    attributes:
      status: active
```
