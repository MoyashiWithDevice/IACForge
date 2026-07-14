package extension

import (
	"fmt"

	"IACForge/src/core"
	"IACForge/src/schema"
)

// RelationTypesExtensionPoint manages relation type extensions.
type RelationTypesExtensionPoint struct {
	schema *schema.Schema
	types  map[core.RelationType]string // type -> extension ID that registered it
}

// NewRelationTypesExtensionPoint creates a new relation types extension point.
func NewRelationTypesExtensionPoint(s *schema.Schema) *RelationTypesExtensionPoint {
	return &RelationTypesExtensionPoint{
		schema: s,
		types:  make(map[core.RelationType]string),
	}
}

// Type returns the extension point type.
func (ep *RelationTypesExtensionPoint) Type() ExtensionPointType {
	return ExtensionPointRelationTypes
}

// Register registers all relation type contributions from the given extension.
func (ep *RelationTypesExtensionPoint) Register(ext *Extension) error {
	for _, contrib := range ext.RelationTypes {
		if ep.schema.HasRelationType(contrib.Type) {
			return fmt.Errorf("%w: relation type %q already defined in schema", ErrCoreConflict, contrib.Type)
		}
		ep.schema.AddRelationType(contrib.Type, contrib.Definition)
		ep.types[contrib.Type] = ext.Manifest.ID
	}
	return nil
}

// GetRelationTypesByExtension returns all relation types registered by a specific extension.
func (ep *RelationTypesExtensionPoint) GetRelationTypesByExtension(extensionID string) []core.RelationType {
	var result []core.RelationType
	for relType, extID := range ep.types {
		if extID == extensionID {
			result = append(result, relType)
		}
	}
	return result
}

// GetExtensionForType returns the extension ID that registered the given relation type.
func (ep *RelationTypesExtensionPoint) GetExtensionForType(relType core.RelationType) (string, bool) {
	extID, ok := ep.types[relType]
	return extID, ok
}

// AllExtendedTypes returns all relation types added by extensions.
func (ep *RelationTypesExtensionPoint) AllExtendedTypes() map[core.RelationType]string {
	result := make(map[core.RelationType]string, len(ep.types))
	for k, v := range ep.types {
		result[k] = v
	}
	return result
}
