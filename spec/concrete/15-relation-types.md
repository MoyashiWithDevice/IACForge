# Relation Types

## Overview

Relation Types define the semantics of connections between Entities.

Every Relation MUST define a type.

The core specification defines the following Relation Types.

Implementations MAY introduce additional types through extensions.

---

## Common Properties

Every Relation shares the following properties regardless of type.

### Required

- id
- type
- participants

### Optional

- description
- status
- tags
- labels
- extensions

Individual Relation Types MAY define additional properties.

---

## Directionality

Relations are classified by directionality.

| Type | Description |
|------|-------------|
| directed | Has a source and target participant |
| symmetric | All participants are equal |
| undirected | Connection with no inherent direction |

---

## Core Relation Types

### connects

Represents a physical or logical connection between Entities.

| Property | Value |
|----------|-------|
| Direction | symmetric |
| Participants | Two or more Entities |
| Cardinality | N:N |

#### Constraints

- Connects is symmetric; order of participants does not matter.
- Typically connects interfaces via cables.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| connection_type | string | no | - | Type of connection (physical, logical, virtual) |
| bandwidth_mbps | integer | no | - | Connection bandwidth in Mbps |

#### Examples

```yaml
- id: rel-connects-srv-sw
  type: connects
  participants:
    - srv-proxmox-01/eno1
    - sw-core-01/port1
  connection_type: physical
  bandwidth_mbps: 10000
  status: active

- id: rel-connects-sw-sw
  type: connects
  participants:
    - sw-core-01/port24
    - sw-access-01/port24
  connection_type: physical
  bandwidth_mbps: 10000
```

---

### hosts

Represents an execution or hosting relationship.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Hosting Entity |
| Target | Hosted Entity |
| Cardinality | 1:N |

#### Constraints

- Source provides resources to target.
- Target executes or runs on source.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| (none) | - | - | - | Uses only common relation properties |

#### Examples

```yaml
- id: rel-hosts-server-vm
  type: hosts
  participants:
    source: srv-proxmox-01
    target: vm-web-01
  status: active

- id: rel-hosts-vm-app
  type: hosts
  participants:
    source: vm-web-01
    target: app-web-server
  status: active
```

---

### depends_on

Represents a directional dependency between Entities.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Dependent Entity |
| Target | Dependency Entity |
| Cardinality | N:N |

#### Constraints

- If A depends_on B, A requires B but B does not require A.
- Dependencies may be cyclic (A → B → A is valid).

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| dependency_type | string | no | - | Type of dependency (runtime, build, network, storage) |
| critical | boolean | no | false | Whether failure causes cascading failure |

#### Examples

```yaml
- id: rel-depends-app-db
  type: depends_on
  participants:
    source: app-web-server
    target: app-database
  dependency_type: runtime
  critical: true
  status: active

- id: rel-depends-vm-storage
  type: depends_on
  participants:
    source: vm-web-01
    target: vol-web-data
  dependency_type: storage
  critical: true
```

---

### belongs_to

Represents logical membership or association.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Member Entity |
| Target | Group Entity |
| Cardinality | N:N |

#### Constraints

- Membership does not imply ownership.
- An Entity may belong to multiple groups.

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| (none) | - | - | - | Uses only common relation properties |

#### Examples

```yaml
- id: rel-belongsto-vm-cluster
  type: belongs_to
  participants:
    source: vm-web-01
    target: cluster-prod-01
  status: active

- id: rel-belongsto-intf-network
  type: belongs_to
  participants:
    source: vm-web-01/eth0
    target: mgmt-network-01
  status: active
```

---

## Extended Relation Types

The following relation types are defined by the core specification for common use cases.

### replicates_to

Represents data replication between Entities.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Primary Entity |
| Target | Replica Entity |
| Cardinality | 1:N |

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| replication_type | string | no | synchronous | Replication type (synchronous, asynchronous) |
| lag_seconds | number | no | - | Replication lag in seconds |

#### Examples

```yaml
- id: rel-replicates-db
  type: replicates_to
  participants:
    source: app-database-primary
    target: app-database-replica
  replication_type: asynchronous
  lag_seconds: 0.5
```

---

### backs_up

Represents a backup relationship.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Source Entity |
| Target | Backup Entity |
| Cardinality | 1:N |

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| backup_type | string | no | - | Backup type (full, incremental, differential) |
| schedule | string | no | - | Backup schedule (cron expression) |
| retention_days | integer | no | - | Backup retention in days |

#### Examples

```yaml
- id: rel-backup-vm
  type: backs_up
  participants:
    source: vm-web-01
    target: vol-web-backup
  backup_type: incremental
  schedule: "0 2 * * *"
  retention_days: 30
```

---

### monitors

Represents a monitoring relationship.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Monitoring Entity |
| Target | Monitored Entity |
| Cardinality | 1:N |

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| monitor_type | string | no | - | Monitor type (agent, agentless, snmp, api) |
| interval_seconds | integer | no | 60 | Monitoring interval in seconds |

#### Examples

```yaml
- id: rel-monitor-prometheus
  type: monitors
  participants:
    source: app-prometheus
    target: srv-proxmox-01
  monitor_type: snmp
  interval_seconds: 30
```

---

### managed_by

Represents a management relationship.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Managed Entity |
| Target | Management Entity |
| Cardinality | N:1 |

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| management_type | string | no | - | Management type (configuration, orchestration, monitoring) |

#### Examples

```yaml
- id: rel-managed-vm
  type: managed_by
  participants:
    source: vm-web-01
    target: app-ansible
  management_type: configuration
```

---

### mounted_on

Represents a mounting relationship (storage).

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Volume Entity |
| Target | Compute Entity |
| Cardinality | N:1 |

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| mount_point | string | no | - | Mount point path |
| filesystem | string | no | - | Filesystem type |
| options | string | no | - | Mount options |

#### Examples

```yaml
- id: rel-mounted-data
  type: mounted_on
  participants:
    source: vol-web-data
    target: vm-web-01
  mount_point: /data
  filesystem: ext4
  options: "rw,noatime"
```

---

### applies_to

Represents that an ACL is applied to a network target (interface, firewall, host).

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | ACL Entity |
| Target | Network Target Entity |
| Cardinality | N:N |

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| (none) | - | - | - | Uses only common relation properties |

#### Examples

```yaml
- id: rel-applies-web-acl
  type: applies_to
  participants:
    source: acl-web-ingress
    target: srv-web-01/eth0
  status: active

- id: rel-applies-fw-acl
  type: applies_to
  participants:
    source: acl-web-ingress
    target: fw-core-01
  status: active
```

---

### listens_on

Represents that an open port is listening on a network interface or address.

| Property | Value |
|----------|-------|
| Direction | directed |
| Source | Open Port Entity |
| Target | Interface or Host Entity |
| Cardinality | N:1 |

#### Properties

| Property | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| (none) | - | - | - | Uses only common relation properties |

#### Examples

```yaml
- id: rel-listens-nginx
  type: listens_on
  participants:
    source: port-443-nginx
    target: vm-web-01/eth0
  status: active

- id: rel-listens-postgres
  type: listens_on
  participants:
    source: port-5432-postgres
    target: vm-web-01/eth0
  status: active
```

---

## Participant Constraints

### Allowed Participant Kinds

Each Relation Type MAY define which Entity Kinds may participate.

#### Core Type Constraints

| Relation Type | Source Kinds | Target Kinds |
|---------------|--------------|--------------|
| connects | interface | interface |
| hosts | server, vm, container | vm, container, application |
| depends_on | vm, container, application | vm, container, application, storage, network |
| belongs_to | vm, container, interface, server, switch, router, firewall, storage, acl, acl_rule, open_port | cluster, network, site, firewall, interface, server, vm, container, application |
| applies_to | acl | interface, firewall, server, vm, container |
| listens_on | open_port | interface, server, vm, container |

### Cardinality

Cardinality defines how many instances of a relation may exist between participants.

| Type | Cardinality | Description |
|------|-------------|-------------|
| connects | N:N | Many-to-many connections |
| hosts | 1:N | One host, many guests |
| depends_on | N:N | Many-to-many dependencies |
| belongs_to | N:N | Many members, many groups |
| applies_to | N:N | One ACL, many targets |
| listens_on | N:1 | Many ports, one interface |

---

## Status Values

Every Relation MAY have a status.

The core specification defines the following statuses:

| Status | Description |
|--------|-------------|
| planned | Relation is planned but not yet active |
| active | Relation is operational |
| maintenance | Relation is under maintenance |
| deprecated | Relation is scheduled for removal |
| offline | Relation is not operational |

---

## Equality

Relations are uniquely identified by their identifier.

Changing participants modifies an existing Relation.

Changing the identifier creates a different Relation.
