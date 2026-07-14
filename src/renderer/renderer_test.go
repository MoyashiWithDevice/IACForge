package renderer

import (
	"testing"

	"IACForge/src/core"
	"IACForge/src/view"
)

func TestNewSVGRenderer(t *testing.T) {
	r := NewSVGRenderer()
	if r.ID() != "svg" {
		t.Errorf("expected ID 'svg', got '%s'", r.ID())
	}
	if r.Name() != "SVG Renderer" {
		t.Errorf("expected Name 'SVG Renderer', got '%s'", r.Name())
	}
	if r.Format() != "svg" {
		t.Errorf("expected Format 'svg', got '%s'", r.Format())
	}
}

func TestNewMermaidRenderer(t *testing.T) {
	r := NewMermaidRenderer()
	if r.ID() != "mermaid" {
		t.Errorf("expected ID 'mermaid', got '%s'", r.ID())
	}
	if r.Format() != "mmd" {
		t.Errorf("expected Format 'mmd', got '%s'", r.Format())
	}
}

func TestNewMarkdownRenderer(t *testing.T) {
	r := NewMarkdownRenderer()
	if r.ID() != "markdown" {
		t.Errorf("expected ID 'markdown', got '%s'", r.ID())
	}
	if r.Format() != "md" {
		t.Errorf("expected Format 'md', got '%s'", r.Format())
	}
}

func TestNewJSONRenderer(t *testing.T) {
	r := NewJSONRenderer()
	if r.ID() != "json" {
		t.Errorf("expected ID 'json', got '%s'", r.ID())
	}
	if r.Format() != "json" {
		t.Errorf("expected Format 'json', got '%s'", r.Format())
	}
}

func TestNewArtifact(t *testing.T) {
	artifact := NewArtifact("art-1", "svg", "view-1", "svg", "<svg></svg>")
	if artifact.ID != "art-1" {
		t.Errorf("expected ID 'art-1', got '%s'", artifact.ID)
	}
	if artifact.RendererID != "svg" {
		t.Errorf("expected RendererID 'svg', got '%s'", artifact.RendererID)
	}
	if artifact.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestNewTheme(t *testing.T) {
	theme := NewTheme("dark", "Dark Theme")
	if theme.ID != "dark" {
		t.Errorf("expected ID 'dark', got '%s'", theme.ID)
	}
	if theme.Colors == nil {
		t.Error("expected non-nil colors")
	}
	if theme.Typography == nil {
		t.Error("expected non-nil typography")
	}
}

func TestNewRenderOptions(t *testing.T) {
	opts := NewRenderOptions()
	if opts.Width != 800 {
		t.Errorf("expected Width 800, got %f", opts.Width)
	}
	if opts.Height != 600 {
		t.Errorf("expected Height 600, got %f", opts.Height)
	}
}

func TestSVGRendererRender(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	g.AddEntity(e1)
	e2 := core.NewEntity("srv-2", "server", "Server 2")
	g.AddEntity(e2)

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewSVGRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "svg" {
		t.Errorf("expected format 'svg', got '%s'", artifact.Format)
	}
	if len(artifact.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestMermaidRendererRender(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	g.AddEntity(e1)
	e2 := core.NewEntity("srv-2", "server", "Server 2")
	g.AddEntity(e2)
	r1 := core.NewDirectedRelation("rel-1", "connects", "srv-1", "srv-2")
	g.AddRelation(r1)

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewMermaidRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "mmd" {
		t.Errorf("expected format 'mmd', got '%s'", artifact.Format)
	}
	if len(artifact.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestMarkdownRendererRender(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	g.AddEntity(e1)

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewMarkdownRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "md" {
		t.Errorf("expected format 'md', got '%s'", artifact.Format)
	}
	if len(artifact.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestJSONRendererRender(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	g.AddEntity(e1)

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	renderer := NewJSONRenderer()
	artifact, err := renderer.Render(result, nil)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "json" {
		t.Errorf("expected format 'json', got '%s'", artifact.Format)
	}
	if len(artifact.Content) == 0 {
		t.Error("expected non-empty content")
	}
}

func TestLayoutEngineHierarchical(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("site-1", "site", "Site 1")
	g.AddEntity(e1)
	e2 := core.NewEntity("rack-1", "rack", "Rack 1")
	e2.SetOwner("site-1")
	g.AddEntity(e2)

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	layoutEngine := NewLayoutEngine(&LayoutConfig{
		Type:      "hierarchical",
		Direction: "top-down",
		Spacing:   50,
		Padding:   20,
	})

	layout := layoutEngine.ComputeLayout(result)
	if len(layout.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(layout.Nodes))
	}
	if layout.Width <= 0 {
		t.Errorf("expected positive width, got %f", layout.Width)
	}
	if layout.Height <= 0 {
		t.Errorf("expected positive height, got %f", layout.Height)
	}
}

func TestLayoutEngineForceDirected(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	g.AddEntity(e1)
	e2 := core.NewEntity("srv-2", "server", "Server 2")
	g.AddEntity(e2)

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	layoutEngine := NewLayoutEngine(&LayoutConfig{
		Type:    "force-directed",
		Spacing: 100,
		Padding: 20,
	})

	layout := layoutEngine.ComputeLayout(result)
	if len(layout.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(layout.Nodes))
	}
}

func TestSVGRendererRenderWithTheme(t *testing.T) {
	g := core.NewGraph()
	e1 := core.NewEntity("srv-1", "server", "Server 1")
	g.AddEntity(e1)

	v := view.NewView("test-view", "Test View")
	engine := view.NewEngine(g)
	result, err := engine.Apply(v)
	if err != nil {
		t.Fatalf("failed to apply view: %v", err)
	}

	theme := NewTheme("dark", "Dark Theme")
	opts := &RenderOptions{
		Width:  1024,
		Height: 768,
		Theme:  theme,
	}

	renderer := NewSVGRenderer()
	artifact, err := renderer.Render(result, opts)
	if err != nil {
		t.Fatalf("failed to render: %v", err)
	}

	if artifact.Format != "svg" {
		t.Errorf("expected format 'svg', got '%s'", artifact.Format)
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"srv-1", "srv_1"},
		{"my.server", "my_server"},
		{"path/to/file", "path_to_file"},
	}

	for _, test := range tests {
		result := sanitizeID(test.input)
		if result != test.expected {
			t.Errorf("sanitizeID(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"<script>", "&lt;script&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
	}

	for _, test := range tests {
		result := escapeXML(test.input)
		if result != test.expected {
			t.Errorf("escapeXML(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestEscapeMermaid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{`"quoted"`, "'quoted'"},
		{"line\nbreak", "line break"},
	}

	for _, test := range tests {
		result := escapeMermaid(test.input)
		if result != test.expected {
			t.Errorf("escapeMermaid(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}
