package renderer

import (
	"fmt"

	"IACForge/src/view"
)

// RenderFormat renders a view result to a string in the requested format.
// Supported formats are "mermaid", "markdown", and "md".
func RenderFormat(v *view.ViewResult, format string) (string, error) {
	switch format {
	case "mermaid":
		artifact, err := NewMermaidRenderer().Render(v, NewRenderOptions())
		if err != nil {
			return "", err
		}
		return artifact.Content, nil
	case "markdown", "md":
		artifact, err := NewMarkdownRenderer().Render(v, NewRenderOptions())
		if err != nil {
			return "", err
		}
		return artifact.Content, nil
	default:
		return "", fmt.Errorf("unknown render format: %s (supported: mermaid, markdown)", format)
	}
}
