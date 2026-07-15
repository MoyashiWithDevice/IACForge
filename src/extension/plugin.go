package extension

// Plugin is the interface that all IACForge plugins must implement.
// Each plugin must export a function that returns an Extension.
//
// Usage:
//
//	//go:build plugin
//	package main
//
//	import "IACForge/src/extension"
//
//	func Extension() *extension.Extension {
//		return &extension.Extension{
//			Manifest: &extension.Manifest{
//				ID:        "my-plugin",
//				Namespace: "myorg",
//				Version:   "1.0.0",
//			},
//			EntityKinds: []extension.EntityKindContribution{...},
//		}
//	}
type Plugin interface {
	Extension() *Extension
}

// PluginFunc is a function that returns an Extension.
// This is the type that plugins must export.
type PluginFunc func() *Extension
