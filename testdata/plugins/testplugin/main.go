//go:build plugin

package main

import (
	"IACForge/src/core"
	"IACForge/src/extension"
	"IACForge/src/schema"
	"IACForge/src/validation"
)

func Extension() *extension.Extension {
	return &extension.Extension{
		Manifest: &extension.Manifest{
			ID:            "test-plugin",
			Name:          "Test Plugin",
			Version:       "1.0.0",
			Author:        "IACForge Team",
			Namespace:     "test",
			Description:   "Example plugin demonstrating all extension point types",
			ExtensionPoints: []string{
				string(extension.ExtensionPointEntityKinds),
				string(extension.ExtensionPointRelationTypes),
				string(extension.ExtensionPointValidationRules),
			},
		},
		EntityKinds: []extension.EntityKindContribution{
			{
				Kind: core.EntityKind("test_plugin_vm"),
				Definition: &schema.EntityKindDefinition{
					Description: "Custom VM kind from test plugin",
					Properties: []schema.PropertyDefinition{
						{
							Name:     "hypervisor",
							Type:     schema.PropertyTypeString,
							Required: true,
							Description: "Hypervisor type",
							Constraints: &schema.Constraint{
								Enum: []string{"kvm", "xen", "vmware"},
							},
						},
					},
				},
			},
		},
		RelationTypes: []extension.RelationTypeContribution{
			{
				Type: core.RelationType("test_manages"),
				Definition: &schema.RelationTypeDefinition{
					Direction:   schema.DirectionDirected,
					Description: "Test management relation",
					Participants: &schema.ParticipantConstraints{
						SourceKinds: []core.EntityKind{"server"},
						TargetKinds: []core.EntityKind{"test_plugin_vm"},
					},
				},
			},
		},
		ValidationRules: []extension.ValidationRuleContribution{
			{
				Rule: &validation.Rule{
					ID:          "test-rule-vm-count",
					Name:        "VM Count Limit",
					Description: "Ensures no more than 10 VMs per server",
					Severity:    validation.SeverityWarning,
					Scope:       validation.ScopeGraph,
				},
				Fn: func(ctx *validation.Context) []validation.Finding {
					return nil
				},
			},
		},
	}
}

func main() {}
