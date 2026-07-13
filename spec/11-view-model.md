# View Model

## Overview

A View defines how a projected Graph is presented to a consumer.

A View does not modify the Graph.

A View does not define rendering.

A View defines perspective.

---

## Purpose

A View provides a meaningful interpretation of infrastructure knowledge.

Different Views expose different aspects of the same Graph.

Examples include:

- Physical Infrastructure
- Logical Infrastructure
- Network Topology
- Storage Topology
- Virtualization
- Security Zones
- Service Dependencies
- Inventory

---

## Relationship to Projection

A View consumes the output of a Projection.

The Projection determines the available Graph.

The View determines how that Graph should be interpreted.

The View never changes the Graph.

---

## View Definition

A View MAY define:

- title
- description
- intended audience
- grouping rules
- visibility rules
- annotation rules
- layout hints

View definitions contain no rendering information.

---

## Visibility

A View MAY hide Objects.

Hidden Objects remain part of the underlying Graph.

Visibility affects presentation only.

---

## Grouping

A View MAY organize Objects into logical groups.

Examples include:

- Rack
- VLAN
- Cluster
- Availability Zone
- Application Stack

Grouping does not alter ownership.

Grouping does not alter Relations.

---

## Annotation

Views MAY attach annotations.

Examples include:

- Calculated utilization
- Interface speed
- Host counts
- Warning indicators

Annotations are ephemeral.

Annotations never modify the Graph.

---

## Audience

Views MAY specify their intended audience.

Examples include:

- Network Engineers
- Infrastructure Engineers
- Security Engineers
- Developers
- Operators
- Management

Audience information is descriptive only.

---

## Composition

Multiple Views MAY consume the same Projection.

A View MAY also consume different Projection outputs.

Views are reusable.

---

## Persistence

Views are not canonical data.

Views SHOULD be regenerated whenever possible.

Persisting Views is implementation-specific.
