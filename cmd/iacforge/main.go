package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"IACForge/src/core"
	iacmcp "IACForge/src/mcp"
	"IACForge/src/parser"
	"IACForge/src/query"
	"IACForge/src/renderer"
	"IACForge/src/schema"
	"IACForge/src/validation"
	"IACForge/src/view"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "validate":
		cmdValidate(args)
	case "info":
		cmdInfo(args)
	case "render":
		cmdRender(args)
	case "query":
		cmdQuery(args)
	case "mcp":
		cmdMCP(args)
	case "version":
		fmt.Println("iacforge 0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: iacforge <command> [path] [options]

Commands:
  validate   Validate a YAML infrastructure model
  info       Show summary of a YAML infrastructure model
  render     Render a view to an artifact (svg, markdown, mermaid, json)
  query      Execute a query against a model
  mcp        Start MCP server (SSE transport)
  version    Print version

Path argument is optional. When omitted, the current directory
is scanned recursively for .yaml/.yml files and merged into a single graph.
IDs must be globally unique across all files.

Examples:
  iacforge validate
  iacforge validate models/
  iacforge validate model.yaml
  iacforge info
  iacforge render --format markdown
  iacforge query --kind server
  iacforge mcp --port 8080`)
}

func loadGraph(path string) (*core.Graph, error) {
	if path == "" {
		path = "."
	}
	p := parser.NewParser()
	return p.Load(path)
}

// extractPath extracts the path argument from args (non-flag argument).
// Returns the path and remaining args.
func extractPath(args []string) (string, []string) {
	if len(args) == 0 {
		return "", args
	}
	if !strings.HasPrefix(args[0], "-") {
		return args[0], args[1:]
	}
	return "", args
}

func cmdValidate(args []string) {
	path, args := extractPath(args)

	g, err := loadGraph(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	s := schema.CoreSchema()
	v := validation.NewEngine(s)
	validation.RegisterCoreRules(v)

	var profile *schema.Profile
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--profile" && i+1 < len(args) {
			profilePath := args[i+1]
			profile, err = loadProfile(profilePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Profile error: %v\n", err)
				os.Exit(1)
			}
			break
		}
	}

	result := v.Validate(g, profile)

	if result.Passed {
		fmt.Println("Validation PASSED")
	} else {
		fmt.Println("Validation FAILED")
	}

	fmt.Printf("Rules: %d, Findings: %d (errors: %d, warnings: %d, info: %d)\n",
		result.Summary.TotalRules, result.Summary.TotalFindings,
		result.Summary.Errors, result.Summary.Warnings, result.Summary.Infos)

	if len(result.Findings) > 0 {
		fmt.Println("\nFindings:")
		for _, f := range result.Findings {
			objInfo := ""
			if f.ObjectID != "" {
				objInfo = fmt.Sprintf(" [%s]", f.ObjectID)
			}
			fmt.Printf("  %-7s %s%s: %s\n", strings.ToUpper(string(f.Severity)), f.RuleID, objInfo, f.Message)
		}
	}

	if !result.Passed {
		os.Exit(1)
	}
}

func cmdInfo(args []string) {
	path, _ := extractPath(args)

	g, err := loadGraph(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	entities := g.Entities()
	relations := g.Relations()

	kindCounts := make(map[core.EntityKind]int)
	for _, e := range entities {
		kindCounts[e.Kind]++
	}

	relTypeCounts := make(map[core.RelationType]int)
	for _, r := range relations {
		relTypeCounts[r.Type]++
	}

	fmt.Printf("Graph Summary\n")
	fmt.Printf("  Entities:  %d\n", len(entities))
	fmt.Printf("  Relations: %d\n", len(relations))

	if len(kindCounts) > 0 {
		fmt.Println("\n  Entity Kinds:")
		for _, k := range sortedKinds(kindCounts) {
			fmt.Printf("    %-20s %d\n", k, kindCounts[k])
		}
	}

	if len(relTypeCounts) > 0 {
		fmt.Println("\n  Relation Types:")
		for _, rt := range sortedRelTypes(relTypeCounts) {
			fmt.Printf("    %-20s %d\n", rt, relTypeCounts[rt])
		}
	}
}

func cmdRender(args []string) {
	path, args := extractPath(args)

	format := "markdown"
	output := ""

	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--format":
			format = args[i+1]
			i++
		case "--output":
			output = args[i+1]
			i++
		}
	}

	g, err := loadGraph(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	v := view.NewView("default", "Infrastructure View")
	v.AddVisibility(view.NewVisibilityRule(view.VisibilityTargetEntities, view.VisibilityActionShow))
	v.AddVisibility(view.NewVisibilityRule(view.VisibilityTargetRelations, view.VisibilityActionShow))

	ve := view.NewEngine(g)
	vr, err := ve.Apply(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "View error: %v\n", err)
		os.Exit(1)
	}

	var r renderer.Renderer
	switch format {
	case "svg":
		r = renderer.NewSVGRenderer()
	case "markdown", "md":
		r = renderer.NewMarkdownRenderer()
		format = "markdown"
	case "mermaid":
		r = renderer.NewMermaidRenderer()
	case "json":
		r = renderer.NewJSONRenderer()
	default:
		fmt.Fprintf(os.Stderr, "unknown format: %s (supported: svg, markdown, mermaid, json)\n", format)
		os.Exit(1)
	}

	artifact, err := r.Render(vr, renderer.NewRenderOptions())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Render error: %v\n", err)
		os.Exit(1)
	}

	if output != "" {
		if err := os.WriteFile(output, []byte(artifact.Content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Output written to %s\n", output)
	} else {
		fmt.Print(artifact.Content)
	}
}

func cmdQuery(args []string) {
	path, args := extractPath(args)

	kind := ""
	relType := ""

	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--kind":
			kind = args[i+1]
			i++
		case "--type":
			relType = args[i+1]
			i++
		}
	}

	g, err := loadGraph(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	qe := query.NewEngine(g)
	q := query.NewQuery()
	q.Select = query.NewSelectClause()

	if kind != "" {
		sel := q.Select.AddEntity(core.EntityKind(kind))
		_ = sel
	} else if relType != "" {
		q.Select.AddRelation(core.RelationType(relType))
	} else {
		sel := q.Select.AddEntity("")
		_ = sel
	}

	result, err := qe.Execute(q)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Results: %d items\n\n", result.Count)
	for _, item := range result.Results {
		fmt.Printf("  %-12s %s (%s)\n", item.Type, item.ID, item.Path)
	}
}

func loadProfile(path string) (*schema.Profile, error) {
	p := parser.NewParser()
	g, err := p.ParseFile(path)
	if err != nil {
		return nil, err
	}

	profile := schema.NewProfile("custom")
	for _, e := range g.Entities() {
		profile.AddRequiredKind(string(e.Kind))
	}
	for _, r := range g.Relations() {
		profile.AddRequiredRelation(string(r.Type))
	}

	return profile, nil
}

func sortedKinds(counts map[core.EntityKind]int) []core.EntityKind {
	kinds := make([]core.EntityKind, 0, len(counts))
	for k := range counts {
		kinds = append(kinds, k)
	}
	for i := 0; i < len(kinds); i++ {
		for j := i + 1; j < len(kinds); j++ {
			if string(kinds[i]) > string(kinds[j]) {
				kinds[i], kinds[j] = kinds[j], kinds[i]
			}
		}
	}
	return kinds
}

func sortedRelTypes(counts map[core.RelationType]int) []core.RelationType {
	types := make([]core.RelationType, 0, len(counts))
	for t := range counts {
		types = append(types, t)
	}
	for i := 0; i < len(types); i++ {
		for j := i + 1; j < len(types); j++ {
			if string(types[i]) > string(types[j]) {
				types[i], types[j] = types[j], types[i]
			}
		}
	}
	return types
}

func cmdMCP(args []string) {
	port := 8080
	stdio := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			if i+1 < len(args) {
				p, err := strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "invalid port: %s\n", args[i+1])
					os.Exit(1)
				}
				port = p
				i++
			}
		case "--stdio":
			stdio = true
		}
	}

	sm := iacmcp.NewSessionManager()
	s := iacmcp.NewMCPServer(sm)

	if stdio {
		if err := iacmcp.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		addr := fmt.Sprintf(":%d", port)
		fmt.Fprintf(os.Stderr, "Starting IACForge MCP server on %s\n", addr)

		if err := iacmcp.NewSSEServer(s, addr); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	}
}
