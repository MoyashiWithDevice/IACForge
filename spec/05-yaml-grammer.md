# YAML Grammar

## Overview

YAML is the first standardized serialization format for the Object Model.

This specification defines the canonical YAML representation.

Other authoring formats MAY exist.

---

## Document Structure

A YAML document consists of a single Graph.

The document root MUST contain:

objects:

Additional sections MAY be defined by future specifications.

---

## Objects

Every Object MUST appear exactly once.

Objects are distinguished by their object type.

Entity

Relation

---

## Entity

An Entity object MUST define:

id

kind

name

Optional properties follow the Entity specification.

---

## Relation

A Relation object MUST define:

id

type

participants

Optional properties follow the Relation specification.

---

## References

References use canonical identifiers.

A parser MUST resolve every reference.

Unknown references are validation errors.

---

## Ownership

Ownership MUST be represented using the standard contains Relation.

Hierarchical nesting MUST NOT change semantics.

Nested syntax MAY exist in authoring formats.

The canonical format always represents ownership explicitly.

---

## Unknown Properties

Unknown properties MUST be preserved.

The core specification ignores unknown properties.

Extensions MAY interpret them.

---

## Ordering

Object order has no semantic meaning.

Implementations SHOULD preserve ordering when possible.

---

## Comments

Comments are not part of the Object Model.

Mappings SHOULD preserve comments whenever possible.

---

## Round-trip

Parsing and serializing a canonical YAML document SHOULD preserve semantic equivalence.

Formatting differences are acceptable.

Semantic differences are not.
