# View Model

## Overview

A View is a projection of a Graph.

Views never modify the underlying Graph.

Multiple Views may exist for the same Graph.

Each View represents a different perspective.

---

## Purpose

Views provide human-readable representations of infrastructure.

Examples include:

- Physical topology
- Logical topology
- Network topology
- Rack layout
- Application dependency
- Documentation
- Inventory

The Graph remains unchanged.

---

## Relationship to Query

A View is defined by a Query.

The Query selects Objects and Relations.

The View determines how the result is presented.

---

## View Pipeline

Every View consists of three conceptual stages.

Graph

↓

Query

↓

Projection

↓

Rendering

The Query determines what is included.

The Projection determines how Objects are interpreted.

Rendering determines presentation.

---

## Projection

Projection transforms a Graph subset into a View Model.

Projection MAY:

- group Objects
- hide Objects
- aggregate Objects
- annotate Objects

Projection MUST NOT modify the underlying Graph.

---

## Rendering Independence

Views do not define rendering.

The same View may be rendered as:

- SVG
- PNG
- Mermaid
- D2
- Graphviz
- Markdown
- HTML
- JSON

---

## Stability

Views are ephemeral.

They are generated from the Graph.

Views are never stored as the canonical model.

---

## View Metadata

Views MAY define metadata.

Examples include:

- title
- description
- theme
- layout

View metadata never affects the Graph.

---

## Extensibility

Implementations MAY define custom View types.

Custom Views MUST preserve the semantics of the underlying Graph.
