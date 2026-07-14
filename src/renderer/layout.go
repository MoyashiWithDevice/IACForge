package renderer

import (
	"math"

	"IACForge/src/view"
)

// LayoutEngine computes spatial arrangement of elements.
type LayoutEngine struct {
	config *LayoutConfig
}

// NewLayoutEngine creates a new LayoutEngine.
func NewLayoutEngine(config *LayoutConfig) *LayoutEngine {
	if config == nil {
		config = &LayoutConfig{
			Type:      "hierarchical",
			Direction: "top-down",
			Spacing:   50,
			Padding:   20,
		}
	}
	return &LayoutEngine{config: config}
}

// ComputeLayout computes positions for all nodes and edges.
func (le *LayoutEngine) ComputeLayout(result *view.ViewResult) *LayoutResult {
	switch le.config.Type {
	case "force-directed":
		return le.computeForceDirectedLayout(result)
	default:
		return le.computeHierarchicalLayout(result)
	}
}

// computeHierarchicalLayout arranges nodes in a tree-like structure.
func (le *LayoutEngine) computeHierarchicalLayout(result *view.ViewResult) *LayoutResult {
	levels := le.buildLevels(result)
	spacing := le.config.Spacing
	padding := le.config.Padding

	if spacing == 0 {
		spacing = 50
	}
	if padding == 0 {
		padding = 20
	}

	nodeWidth := 120.0
	nodeHeight := 40.0

	var layoutResult LayoutResult
	layoutResult.Nodes = make([]NodePosition, 0)
	layoutResult.Edges = make([]EdgePosition, 0)

	maxWidth := 0.0
	for levelIdx, level := range levels {
		y := padding + float64(levelIdx)*(nodeHeight+spacing)
		for nodeIdx, entity := range level {
			x := padding + float64(nodeIdx)*(nodeWidth+spacing)

			layoutResult.Nodes = append(layoutResult.Nodes, NodePosition{
				ID: entity.ID,
				Position: Position{
					X: x,
					Y: y,
				},
				Width:  nodeWidth,
				Height: nodeHeight,
			})

			right := x + nodeWidth
			if right > maxWidth {
				maxWidth = right
			}
		}
	}

	for _, rel := range result.VisibleRelations {
		sourcePos := le.findNodePosition(layoutResult.Nodes, rel.Source())
		targetPos := le.findNodePosition(layoutResult.Nodes, rel.Target())
		if sourcePos != nil && targetPos != nil {
			edge := EdgePosition{
				ID:     rel.ID,
				Source: rel.Source(),
				Target: rel.Target(),
				Points: []Position{
					{X: sourcePos.Position.X + sourcePos.Width/2, Y: sourcePos.Position.Y + sourcePos.Height},
					{X: targetPos.Position.X + targetPos.Width/2, Y: targetPos.Position.Y},
				},
			}
			layoutResult.Edges = append(layoutResult.Edges, edge)
		}
	}

	layoutResult.Width = maxWidth + padding
	layoutResult.Height = padding + float64(len(levels))*(nodeHeight+spacing)

	return &layoutResult
}

// buildLevels builds hierarchical levels from entities.
func (le *LayoutEngine) buildLevels(result *view.ViewResult) [][]*view.Group {
	groups := make(map[string][]string)
	entityMap := make(map[string]interface{})
	levelMap := make(map[string]int)

	for _, entity := range result.VisibleEntities {
		entityMap[entity.ID] = entity
		levelMap[entity.ID] = 0
	}

	for _, group := range result.Groups {
		for _, memberID := range group.Members {
			groups[memberID] = append(groups[memberID], group.ID)
		}
	}

	maxLevel := 0
	for _, entity := range result.VisibleEntities {
		level := 0
		if groupIDs, ok := groups[entity.ID]; ok && len(groupIDs) > 0 {
			level = 1
		}
		levelMap[entity.ID] = level
		if level > maxLevel {
			maxLevel = level
		}
	}

	levels := make([][]*view.Group, maxLevel+1)
	for i := range levels {
		levels[i] = make([]*view.Group, 0)
	}

	for _, entity := range result.VisibleEntities {
		level := levelMap[entity.ID]
		levels[level] = append(levels[level], &view.Group{
			ID:      entity.ID,
			Kind:    string(entity.Kind),
			Name:    entity.Name,
			Members: []string{entity.ID},
		})
	}

	return levels
}

// computeForceDirectedLayout arranges nodes using a physics simulation.
func (le *LayoutEngine) computeForceDirectedLayout(result *view.ViewResult) *LayoutResult {
	nodeCount := len(result.VisibleEntities)
	if nodeCount == 0 {
		return &LayoutResult{
			Nodes:  []NodePosition{},
			Edges:  []EdgePosition{},
			Width:  0,
			Height: 0,
		}
	}

	nodeWidth := 120.0
	nodeHeight := 40.0
	spacing := le.config.Spacing
	if spacing == 0 {
		spacing = 100
	}

	nodes := make([]NodePosition, 0, nodeCount)
	for i, entity := range result.VisibleEntities {
		angle := 2 * math.Pi * float64(i) / float64(nodeCount)
		radius := spacing * math.Sqrt(float64(nodeCount))
		nodes = append(nodes, NodePosition{
			ID: entity.ID,
			Position: Position{
				X: radius * math.Cos(angle),
				Y: radius * math.Sin(angle),
			},
			Width:  nodeWidth,
			Height: nodeHeight,
		})
	}

	for iter := 0; iter < 50; iter++ {
		le.applyForces(nodes, result)
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, node := range nodes {
		if node.Position.X < minX {
			minX = node.Position.X
		}
		if node.Position.Y < minY {
			minY = node.Position.Y
		}
		if node.Position.X+node.Width > maxX {
			maxX = node.Position.X + node.Width
		}
		if node.Position.Y+node.Height > maxY {
			maxY = node.Position.Y + node.Height
		}
	}

	padding := le.config.Padding
	if padding == 0 {
		padding = 20
	}

	for i := range nodes {
		nodes[i].Position.X -= minX - padding
		nodes[i].Position.Y -= minY - padding
	}

	var layoutResult LayoutResult
	layoutResult.Nodes = nodes
	layoutResult.Edges = make([]EdgePosition, 0)
	layoutResult.Width = maxX - minX + 2*padding
	layoutResult.Height = maxY - minY + 2*padding

	for _, rel := range result.VisibleRelations {
		edge := EdgePosition{
			ID:     rel.ID,
			Source: rel.Source(),
			Target: rel.Target(),
		}
		layoutResult.Edges = append(layoutResult.Edges, edge)
	}

	return &layoutResult
}

// applyForces applies repulsive and attractive forces.
func (le *LayoutEngine) applyForces(nodes []NodePosition, result *view.ViewResult) {
	repulsion := 1000.0
 attraction := 0.01
 damping := 0.9

	velocities := make([]Position, len(nodes))

	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			dx := nodes[j].Position.X - nodes[i].Position.X
			dy := nodes[j].Position.Y - nodes[i].Position.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 1 {
				dist = 1
			}

			force := repulsion / (dist * dist)
			fx := force * dx / dist
			fy := force * dy / dist

			velocities[i].X -= fx
			velocities[i].Y -= fy
			velocities[j].X += fx
			velocities[j].Y += fy
		}
	}

	for _, rel := range result.VisibleRelations {
		sourceIdx := le.findNodeIndex(nodes, rel.Source())
		targetIdx := le.findNodeIndex(nodes, rel.Target())
		if sourceIdx >= 0 && targetIdx >= 0 {
			dx := nodes[targetIdx].Position.X - nodes[sourceIdx].Position.X
			dy := nodes[targetIdx].Position.Y - nodes[sourceIdx].Position.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 1 {
				dist = 1
			}

			force := attraction * dist
			fx := force * dx / dist
			fy := force * dy / dist

			velocities[sourceIdx].X += fx
			velocities[sourceIdx].Y += fy
			velocities[targetIdx].X -= fx
			velocities[targetIdx].Y -= fy
		}
	}

	for i := range nodes {
		nodes[i].Position.X += velocities[i].X * damping
		nodes[i].Position.Y += velocities[i].Y * damping
	}
}

// findNodePosition finds a node position by ID.
func (le *LayoutEngine) findNodePosition(nodes []NodePosition, id string) *NodePosition {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}

// findNodeIndex finds a node index by ID.
func (le *LayoutEngine) findNodeIndex(nodes []NodePosition, id string) int {
	for i := range nodes {
		if nodes[i].ID == id {
			return i
		}
	}
	return -1
}
