package renderer

import (
	"fmt"
	"strings"

	"IACForge/src/view"
)

// MarkdownRenderer renders views as Markdown documents.
type MarkdownRenderer struct {
	id     string
	name   string
	format string
}

// NewMarkdownRenderer creates a new Markdown renderer.
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		id:     "markdown",
		name:   "Markdown Renderer",
		format: "md",
	}
}

// ID returns the renderer identifier.
func (r *MarkdownRenderer) ID() string {
	return r.id
}

// Name returns the renderer name.
func (r *MarkdownRenderer) Name() string {
	return r.name
}

// Format returns the output format.
func (r *MarkdownRenderer) Format() string {
	return r.format
}

// Render renders a view as Markdown.
func (r *MarkdownRenderer) Render(v *view.ViewResult, opts *RenderOptions) (*Artifact, error) {
	if opts == nil {
		opts = NewRenderOptions()
	}

	var md strings.Builder

	includeTOC := true
	if opts.Options != nil {
		if toc, ok := opts.Options["toc"].(bool); ok {
			includeTOC = toc
		}
	}

	md.WriteString("# ")
	md.WriteString(v.Title)
	md.WriteString("\n\n")

	if v.Description != "" {
		md.WriteString(v.Description)
		md.WriteString("\n\n")
	}

	if includeTOC && len(v.VisibleEntities) > 0 {
		md.WriteString("## Table of Contents\n\n")
		kindGroups := make(map[string][]string)
		for _, entity := range v.VisibleEntities {
			kindGroups[string(entity.Kind)] = append(kindGroups[string(entity.Kind)], entity.ID)
		}
		for kind, ids := range kindGroups {
			md.WriteString(fmt.Sprintf("- [%s](#%s) (%d)\n", kind, strings.ToLower(kind), len(ids)))
		}
		md.WriteString("\n")
	}

	for _, group := range v.Groups {
		md.WriteString(fmt.Sprintf("## %s\n\n", group.Name))
		md.WriteString(fmt.Sprintf("**Kind:** %s\n\n", group.Kind))
		if count, ok := group.Properties["count"]; ok {
			md.WriteString(fmt.Sprintf("**Members:** %v\n\n", count))
		}
	}

	if len(v.VisibleEntities) > 0 {
		md.WriteString("## Entities\n\n")

		kindGroups := make(map[string][]*view.Group)
		for _, entity := range v.VisibleEntities {
			key := string(entity.Kind)
			if _, ok := kindGroups[key]; !ok {
				kindGroups[key] = make([]*view.Group, 0)
			}
			kindGroups[key] = append(kindGroups[key], &view.Group{
				ID:   entity.ID,
				Kind: key,
				Name: entity.Name,
			})
		}

		for kind, entities := range kindGroups {
			md.WriteString(fmt.Sprintf("### %s\n\n", kind))
			md.WriteString("| ID | Name | Status |\n")
			md.WriteString("|----|------|--------|\n")
			for _, e := range entities {
				md.WriteString(fmt.Sprintf("| %s | %s | %s |\n", e.ID, e.Name, "-"))
			}
			md.WriteString("\n")
		}
	}

	if len(v.VisibleRelations) > 0 {
		md.WriteString("## Relations\n\n")
		md.WriteString("| ID | Type | Source | Target |\n")
		md.WriteString("|----|------|--------|--------|\n")
		for _, rel := range v.VisibleRelations {
			md.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", rel.ID, rel.Type, rel.Source(), rel.Target()))
		}
		md.WriteString("\n")
	}

	if len(v.Annotations) > 0 {
		md.WriteString("## Annotations\n\n")
		md.WriteString("| Entity | Property | Value |\n")
		md.WriteString("|--------|----------|-------|\n")
		for entityID, annotations := range v.Annotations {
			for prop, value := range annotations {
				md.WriteString(fmt.Sprintf("| %s | %s | %v |\n", entityID, prop, value))
			}
		}
		md.WriteString("\n")
	}

	artifact := NewArtifact(
		fmt.Sprintf("artifact-%s-%s", r.id, v.ViewID),
		r.id,
		v.ViewID,
		r.format,
		md.String(),
	)
	artifact.Metadata["title"] = v.Title
	artifact.Metadata["description"] = v.Description

	return artifact, nil
}
