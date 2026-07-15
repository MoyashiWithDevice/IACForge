# Extension Development Reference

This guide explains how to create plugins (extensions) for IACForge.

## Overview

IACForge supports runtime plugin loading via Go's `plugin` package.
Place compiled `.so` files in the `extensions/` directory, and they will be automatically loaded at startup.

## Quick Start

### 1. Create a plugin project

```bash
mkdir my-plugin
cd my-plugin
go mod init my-plugin
```

### 2. Create `main.go`

```go
//go:build plugin

package main

import (
	"IACForge/src/core"
	"IACForge/src/extension"
	"IACForge/src/schema"
)

func Extension() *extension.Extension {
	return &extension.Extension{
		Manifest: &extension.Manifest{
			ID:          "my-organization.my-plugin",
			Name:        "My Custom Plugin",
			Version:     "1.0.0",
			Author:      "Your Name",
			Description: "Adds custom entity kinds for our infrastructure",
			Namespace:   "myorg",
			ExtensionPoints: []string{
				string(extension.ExtensionPointEntityKinds),
			},
		},
		EntityKinds: []extension.EntityKindContribution{
			{
				Kind: core.EntityKind("custom_database"),
				Definition: &schema.EntityKindDefinition{
					Description: "Custom database server",
					Properties: []schema.PropertyDefinition{
						{
							Name:     "engine",
							Type:     schema.PropertyTypeString,
							Required: true,
							Description: "Database engine (e.g., postgres, mysql)",
							Constraints: &schema.Constraint{
								Enum: []string{"postgres", "mysql", "mariadb", "sqlite"},
							},
						},
						{
							Name:     "version",
							Type:     schema.PropertyTypeString,
							Required: false,
							Description: "Database version",
						},
					},
				},
			},
		},
	}
}

func main() {}
```

### 3. Build the plugin

```bash
go build -buildmode=plugin -o my-plugin.so .
```

### 4. Deploy

```bash
cp my-plugin.so /path/to/iacforge/extensions/
```

## Manifest

The `Manifest` struct defines plugin metadata:

```go
type Manifest struct {
	ID             string   `yaml:"id"`
	Name           string   `yaml:"name"`
	Version        string   `yaml:"version"`
	Author         string   `yaml:"author,omitempty"`
	Description    string   `yaml:"description,omitempty"`
	SpecVersion    string   `yaml:"spec_version,omitempty"`
	SchemaVersion  string   `yaml:"schema_version,omitempty"`
	Namespace      string   `yaml:"namespace"`
	Dependencies   []string `yaml:"dependencies,omitempty"`
	ExtensionPoints []string `yaml:"extension_points"`
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `ID` | Yes | Unique identifier (use namespace prefix) |
| `Name` | Yes | Human-readable name |
| `Version` | Yes | Semantic version (e.g., `1.0.0`) |
| `Namespace` | Yes | Prevents naming conflicts |
| `Author` | No | Plugin author |
| `Description` | No | What the plugin does |
| `SpecVersion` | No | Compatible IACForge spec version |
| `SchemaVersion` | No | Compatible schema version |
| `Dependencies` | No | List of required plugin IDs |
| `ExtensionPoints` | No | Types of extensions provided |

## Extension Points

IACForge supports 4 extension point types:

### entity_kinds

Add custom entity types to the schema.

```go
EntityKinds: []extension.EntityKindContribution{
	{
		Kind: core.EntityKind("my_custom_vm"),
		Definition: &schema.EntityKindDefinition{
			Description: "Description of this entity kind",
			Properties: []schema.PropertyDefinition{
				{
					Name:     "cpu_count",
					Type:     schema.PropertyTypeInteger,
					Required: true,
					Description: "Number of CPU cores",
					Constraints: &schema.Constraint{
						Min: float6Ptr(1),
						Max: float6Ptr(128),
					},
				},
			},
		},
	},
},
```

**Core entity kinds that CANNOT be redefined:**
`site`, `rack`, `server`, `interface`, `cable`, `power_distribution`, `network`, `vlan`, `switch`, `router`, `firewall`, `acl`, `acl_rule`, `vm`, `container`, `application`, `open_port`, `storage`, `volume`, `cluster`, `availability_zone`

### relation_types

Add custom relation types.

```go
RelationTypes: []extension.RelationTypeContribution{
	{
		Type: core.RelationType("custom_manages"),
		Definition: &schema.RelationTypeDefinition{
			Direction: schema.DirectionDirected,
			Description: "Custom management relation",
			Participants: &schema.ParticipantConstraints{
				SourceKinds: []core.EntityKind{"server", "vm"},
				TargetKinds: []core.EntityKind{"application"},
			},
		},
	},
},
```

**Direction types:**
- `directed` - Has source and target
- `symmetric` - Same meaning in both directions
- `undirected` - No directionality

**Core relation types that CANNOT be redefined:**
`connects`, `hosts`, `depends_on`, `belongs_to`, `replicates_to`, `backs_up`, `monitors`, `managed_by`, `mounted_on`, `applies_to`, `listens_on`

### validation_rules

Add custom validation rules.

```go
ValidationRules: []extension.ValidationRuleContribution{
	{
		Rule: &validation.Rule{
			ID:          "my-custom-rule",
			Name:        "Custom Validation Rule",
			Description: "Ensures custom business logic",
			Severity:    validation.SeverityWarning,
			Scope:       validation.ScopeEntity,
		},
		Fn: func(ctx *validation.Context) []validation.Finding {
			g := ctx.Graph.(*core.Graph)
			var findings []validation.Finding

			for _, e := range g.Entities() {
				if e.Kind == "my_custom_vm" {
					// Your validation logic here
				}
			}

			return findings
		},
	},
},
```

**Severity levels:** `info`, `warning`, `error`

**Scopes:** `graph`, `entity`, `relation`, `ownership`

### renderers

Add custom output renderers.

```go
Renderers: []extension.RendererContribution{
	{
		Renderer: &MyCustomRenderer{},
	},
},
```

Your renderer must implement:

```go
type Renderer interface {
	Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error)
	ID() string
	Name() string
	Format() string
}
```

## Dependencies

Declare dependencies on other plugins:

```go
Manifest: &extension.Manifest{
	ID:           "myorg.dependent-plugin",
	Dependencies: []string{"myorg.base-plugin"},
},
```

- Load order is automatically resolved (topological sort)
- Circular dependencies cause an error
- Missing dependencies cause an error

## Namespace Conventions

Use reverse-domain notation to prevent conflicts:

```
com.github.username.plugin-name
org.company.infrastructure-type
```

## Property Types

| Type | Go Type | Description |
|------|---------|-------------|
| `string` | `string` | Text values |
| `integer` | `int`, `int64` | Whole numbers |
| `number` | `float64` | Decimal numbers |
| `boolean` | `bool` | True/false |
| `list` | `[]interface{}` | Array of values |
| `map` | `map[string]interface{}` | Key-value pairs |
| `reference` | `string` | Reference to another entity |
| `enum` | `string` | One of predefined values |

## Constraints

```go
&schema.Constraint{
	Min:         float6Ptr(0),       // Minimum numeric value
	Max:         float6Ptr(100),     // Maximum numeric value
	MinLength:   intPtr(1),          // Minimum string length
	MaxLength:   intPtr(255),        // Maximum string length
	Pattern:     stringPtr("^[a-z]"),// Regex pattern
	Enum:        []string{"a", "b"}, // Allowed values
	UniqueItems: boolPtr(true),      // List items must be unique
}
```

## Complete Example

See `testdata/plugins/testplugin/main.go` for a working example.

## Troubleshooting

### Plugin fails to load

1. Ensure Go versions match between host and plugin
2. Verify the plugin exports `Extension() *extension.Extension`
3. Check that dependencies are built with `-buildmode=plugin`

### Namespace conflicts

Each namespace can only be used by one plugin. Use unique namespaces.

### Core conflicts

You cannot redefine core entity kinds or relation types. Use different names with your namespace prefix.
