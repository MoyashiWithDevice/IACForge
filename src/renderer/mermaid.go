package renderer

import (
	"fmt"
	"strings"

	"IACForge/src/view"
)

// MermaidRenderer renders views as Mermaid diagrams.
type MermaidRenderer struct {
	id     string
	name   string
	format string
}

// NewMermaidRenderer creates a new Mermaid renderer.
func NewMermaidRenderer() *MermaidRenderer {
	return &MermaidRenderer{
		id:     "mermaid",
		name:   "Mermaid Renderer",
		format: "mmd",
	}
}

// ID returns the renderer identifier.
func (r *MermaidRenderer) ID() string {
	return r.id
}

// Name returns the renderer name.
func (r *MermaidRenderer) Name() string {
	return r.name
}

// Format returns the output format.
func (r *MermaidRenderer) Format() string {
	return r.format
}

// Render renders a view as Mermaid diagram.
func (r *MermaidRenderer) Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error) {
	if opts == nil {
		opts = NewRenderOptions()
	}

	direction := "TB"
	if opts.Options != nil {
		if d, ok := opts.Options["direction"].(string); ok {
			direction = d
		}
	}

	var mermaid strings.Builder

	mermaid.WriteString("graph ")
	mermaid.WriteString(direction)
	mermaid.WriteString("\n")

	for _, entity := range v.VisibleEntities {
		mermaid.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", sanitizeID(entity.ID), escapeMermaid(entity.Name)))
	}

	for _, group := range v.Groups {
		mermaid.WriteString(fmt.Sprintf("    subgraph %s[\"%s\"]\n", sanitizeID(group.ID), escapeMermaid(group.Name)))
		for _, memberID := range group.Members {
			mermaid.WriteString(fmt.Sprintf("        %s\n", sanitizeID(memberID)))
		}
		mermaid.WriteString("    end\n")
	}

	for _, rel := range v.VisibleRelations {
		source := sanitizeID(rel.Source())
		target := sanitizeID(rel.Target())
		edgeLabel := string(rel.Type)
		if rel.Description != "" {
			edgeLabel = rel.Description
		}
		mermaid.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", source, escapeMermaid(edgeLabel), target))
	}

	artifact := NewArtifact(
		fmt.Sprintf("artifact-%s-%s", r.id, v.ViewID),
		r.id,
		v.ViewID,
		r.format,
		mermaid.String(),
	)
	artifact.Metadata["title"] = v.Title
	artifact.Metadata["description"] = v.Description

	return artifact, nil
}

// sanitizeID replaces special characters for Mermaid IDs.
func sanitizeID(id string) string {
	result := strings.ReplaceAll(id, "-", "_")
	result = strings.ReplaceAll(result, ".", "_")
	result = strings.ReplaceAll(result, "/", "_")
	return result
}

// escapeMermaid escapes special Mermaid characters.
func escapeMermaid(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
