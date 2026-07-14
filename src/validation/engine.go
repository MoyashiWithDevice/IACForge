package validation

import (
	"fmt"
	"sort"

	"IACForge/src/core"
	"IACForge/src/core/kinds"
	"IACForge/src/schema"
)

// Engine is the validation engine that evaluates rules against a graph.
type Engine struct {
	schema    *schema.Schema
	rules     map[string]RuleFunc
	ruleDefs  map[string]*Rule
}

// NewEngine creates a new validation engine with the given schema.
func NewEngine(s *schema.Schema) *Engine {
	e := &Engine{
		schema:   s,
		rules:    make(map[string]RuleFunc),
		ruleDefs: make(map[string]*Rule),
	}
	return e
}

// RegisterRule registers a validation rule with the engine.
func (e *Engine) RegisterRule(rule *Rule, fn RuleFunc) {
	e.rules[rule.ID] = fn
	e.ruleDefs[rule.ID] = rule
}

// Validate validates the given graph against all registered rules.
func (e *Engine) Validate(g *core.Graph, profile *schema.Profile) *Result {
	result := &Result{}

	for ruleID, fn := range e.rules {
		ruleDef := e.ruleDefs[ruleID]

		if profile != nil && len(profile.Rules) > 0 {
			if !profile.HasRule(ruleID) {
				continue
			}
		}

		ctx := &Context{
			Graph:   g,
			Schema:  e.schema,
			Profile: profile,
		}

		findings := fn(ctx)
		for i := range findings {
			if findings[i].RuleID == "" {
				findings[i].RuleID = ruleID
			}
			if findings[i].Severity == "" && ruleDef != nil {
				findings[i].Severity = ruleDef.Severity
			}
		}
		result.Findings = append(result.Findings, findings...)
	}

	if profile != nil && len(profile.RequiredKinds) > 0 {
		result.Findings = append(result.Findings, e.checkRequiredKinds(g, profile)...)
	}
	if profile != nil && len(profile.RequiredRelations) > 0 {
		result.Findings = append(result.Findings, e.checkRequiredRelations(g, profile)...)
	}

	result.Summary = e.computeSummary(result.Findings)
	result.Passed = result.Summary.Errors == 0

	return result
}

func (e *Engine) checkRequiredKinds(g *core.Graph, profile *schema.Profile) []Finding {
	var findings []Finding
	for _, kind := range profile.RequiredKinds {
		entities := g.EntitiesByKind(core.EntityKind(kind))
		if len(entities) == 0 {
			findings = append(findings, Finding{
				RuleID:   "profile-required-kind",
				Severity: SeverityError,
				Message:  fmt.Sprintf("profile requires at least one entity of kind %q", kind),
			})
		}
	}
	return findings
}

func (e *Engine) checkRequiredRelations(g *core.Graph, profile *schema.Profile) []Finding {
	var findings []Finding
	for _, relType := range profile.RequiredRelations {
		relations := g.RelationsByType(core.RelationType(relType))
		if len(relations) == 0 {
			findings = append(findings, Finding{
				RuleID:   "profile-required-relation",
				Severity: SeverityError,
				Message:  fmt.Sprintf("profile requires at least one relation of type %q", relType),
			})
		}
	}
	return findings
}

func (e *Engine) computeSummary(findings []Finding) Summary {
	s := Summary{
		TotalFindings: len(findings),
	}
	s.TotalRules = len(e.rules)
	for _, f := range findings {
		switch f.Severity {
		case SeverityError:
			s.Errors++
		case SeverityWarning:
			s.Warnings++
		case SeverityInfo:
			s.Infos++
		}
	}
	return s
}

// RegisterCoreRules registers all 14 core validation rules.
func RegisterCoreRules(engine *Engine) {
	registerGraphIntegrityRules(engine)
	registerEntityRules(engine)
	registerRelationRules(engine)
	registerOwnershipRules(engine)
	registerReferenceRules(engine)
}

func registerGraphIntegrityRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "unique-id",
		Name:     "Unique Identifier",
		Severity: SeverityError,
		Scope:    ScopeGraph,
	}, ruleUniqueID)

	e.RegisterRule(&Rule{
		ID:       "valid-reference",
		Name:     "Valid Reference",
		Severity: SeverityError,
		Scope:    ScopeGraph,
	}, ruleValidReference)

	e.RegisterRule(&Rule{
		ID:       "valid-owner",
		Name:     "Valid Owner",
		Severity: SeverityError,
		Scope:    ScopeGraph,
	}, ruleValidOwner)

	e.RegisterRule(&Rule{
		ID:       "single-owner",
		Name:     "Single Owner",
		Severity: SeverityError,
		Scope:    ScopeOwnership,
	}, ruleSingleOwner)
}

func registerEntityRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "required-kind",
		Name:     "Required Kind",
		Severity: SeverityError,
		Scope:    ScopeEntity,
	}, ruleRequiredKind)

	e.RegisterRule(&Rule{
		ID:       "required-name",
		Name:     "Required Name",
		Severity: SeverityError,
		Scope:    ScopeEntity,
	}, ruleRequiredName)

	e.RegisterRule(&Rule{
		ID:       "valid-kind",
		Name:     "Valid Kind",
		Severity: SeverityError,
		Scope:    ScopeEntity,
	}, ruleValidKind)

	e.RegisterRule(&Rule{
		ID:       "valid-status",
		Name:     "Valid Status",
		Severity: SeverityWarning,
		Scope:    ScopeEntity,
	}, ruleValidStatus)

	e.RegisterRule(&Rule{
		ID:       "valid-port-range",
		Name:     "Valid Port Range",
		Severity: SeverityError,
		Scope:    ScopeEntity,
	}, ruleValidPortRange)

	e.RegisterRule(&Rule{
		ID:       "valid-acl-rule-parent",
		Name:     "Valid ACL Rule Parent",
		Severity: SeverityError,
		Scope:    ScopeEntity,
	}, ruleValidACLRULEParent)
}

func registerRelationRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "required-type",
		Name:     "Required Type",
		Severity: SeverityError,
		Scope:    ScopeRelation,
	}, ruleRequiredType)

	e.RegisterRule(&Rule{
		ID:       "required-participants",
		Name:     "Required Participants",
		Severity: SeverityError,
		Scope:    ScopeRelation,
	}, ruleRequiredParticipants)

	e.RegisterRule(&Rule{
		ID:       "valid-type",
		Name:     "Valid Type",
		Severity: SeverityError,
		Scope:    ScopeRelation,
	}, ruleValidType)

	e.RegisterRule(&Rule{
		ID:       "valid-direction",
		Name:     "Valid Direction",
		Severity: SeverityError,
		Scope:    ScopeRelation,
	}, ruleValidDirection)

	e.RegisterRule(&Rule{
		ID:       "valid-cardinality",
		Name:     "Valid Cardinality",
		Severity: SeverityError,
		Scope:    ScopeRelation,
	}, ruleValidCardinality)

	e.RegisterRule(&Rule{
		ID:       "valid-participant-kind",
		Name:     "Valid Participant Kind",
		Severity: SeverityWarning,
		Scope:    ScopeRelation,
	}, ruleValidParticipantKind)
}

func registerOwnershipRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "ownership-tree",
		Name:     "Ownership Tree",
		Severity: SeverityError,
		Scope:    ScopeOwnership,
	}, ruleOwnershipTree)

	e.RegisterRule(&Rule{
		ID:       "no-ownership-cycle",
		Name:     "No Ownership Cycle",
		Severity: SeverityError,
		Scope:    ScopeOwnership,
	}, ruleNoOwnershipCycle)

	e.RegisterRule(&Rule{
		ID:       "root-entity",
		Name:     "Root Entity",
		Severity: SeverityError,
		Scope:    ScopeOwnership,
	}, ruleRootEntity)
}

func registerReferenceRules(e *Engine) {
	e.RegisterRule(&Rule{
		ID:       "dangling-reference",
		Name:     "Dangling Reference",
		Severity: SeverityError,
		Scope:    ScopeGraph,
	}, ruleDanglingReference)
}

// --- Rule Implementations ---

func ruleUniqueID(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	seen := make(map[string]bool)
	for _, e := range g.Entities() {
		if seen[e.ID] {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("duplicate entity ID: %q", e.ID),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
		seen[e.ID] = true
	}
	for _, r := range g.Relations() {
		if seen[r.ID] {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("duplicate relation ID: %q", r.ID),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
		seen[r.ID] = true
	}

	return findings
}

func ruleValidReference(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, r := range g.Relations() {
		for _, participantID := range r.ParticipantIDs() {
			_, found := g.ResolveReference(participantID)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("relation %q references non-existent object %q", r.ID, participantID),
					ObjectID:   r.ID,
					ObjectType: ObjectTypeRelation,
				})
			}
		}
	}

	return findings
}

func ruleValidOwner(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Owner != "" {
			_, found := g.GetEntity(e.Owner)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("entity %q has owner %q which does not exist", e.ID, e.Owner),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleSingleOwner(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Owner == "" {
			continue
		}
		_, found := g.GetEntity(e.Owner)
		if !found {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q references non-existent owner %q", e.ID, e.Owner),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleRequiredKind(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind == "" {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q is missing required 'kind' property", e.ID),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleRequiredName(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Name == "" {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q is missing required 'name' property", e.ID),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleValidKind(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, e := range g.Entities() {
		if !s.HasEntityKind(e.Kind) {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("entity %q has undefined kind %q", e.ID, e.Kind),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func ruleValidStatus(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Status != "" && !kinds.IsValidStatus(e.Status) {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("entity %q has invalid status %q", e.ID, e.Status),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	for _, r := range g.Relations() {
		if r.Status != "" && !kinds.IsValidStatus(r.Status) {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("relation %q has invalid status %q", r.ID, r.Status),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleValidPortRange(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind == kinds.OpenPort {
			portVal, ok := e.GetProperty("port")
			if !ok {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("open_port entity %q is missing required 'port' property", e.ID),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
				continue
			}

			var portNum int
			switch v := portVal.(type) {
			case int:
				portNum = v
			case int64:
				portNum = int(v)
			case float64:
				portNum = int(v)
			default:
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("open_port entity %q has non-numeric port value", e.ID),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
				continue
			}

			if portNum < 1 || portNum > 65535 {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("open_port entity %q has port %d outside valid range 1-65535", e.ID, portNum),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleValidACLRULEParent(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, e := range g.Entities() {
		if e.Kind == kinds.ACLRule {
			if e.Owner == "" {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("acl_rule entity %q has no owner", e.ID),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
				continue
			}

			ownerEntity, found := g.GetEntity(e.Owner)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("acl_rule entity %q has non-existent owner %q", e.ID, e.Owner),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
				continue
			}

			if ownerEntity.Kind != kinds.ACL {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("acl_rule entity %q must be owned by an acl, but owner %q is of kind %q", e.ID, e.Owner, ownerEntity.Kind),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}

func ruleRequiredType(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, r := range g.Relations() {
		if r.Type == "" {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q is missing required 'type' property", r.ID),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleRequiredParticipants(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, r := range g.Relations() {
		if r.Participants.Count() < 2 {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q must have at least 2 participants", r.ID),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleValidType(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, r := range g.Relations() {
		if r.Type != "" && !s.HasRelationType(r.Type) {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q has undefined type %q", r.ID, r.Type),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleValidDirection(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, r := range g.Relations() {
		typeDef, ok := s.GetRelationTypeDef(r.Type)
		if !ok {
			continue
		}

		if typeDef.Direction == schema.DirectionDirected {
			if r.Participants.Source == "" || r.Participants.Target == "" {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("directed relation %q must have both source and target", r.ID),
					ObjectID:   r.ID,
					ObjectType: ObjectTypeRelation,
				})
			}
		}
	}

	return findings
}

func ruleValidCardinality(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, r := range g.Relations() {
		typeDef, ok := s.GetRelationTypeDef(r.Type)
		if !ok || typeDef.Participants == nil {
			continue
		}

		count := r.Participants.Count()
		if typeDef.Participants.MinParticipants > 0 && count < typeDef.Participants.MinParticipants {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q has %d participants, minimum required is %d", r.ID, count, typeDef.Participants.MinParticipants),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
		if typeDef.Participants.MaxParticipants > 0 && count > typeDef.Participants.MaxParticipants {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("relation %q has %d participants, maximum allowed is %d", r.ID, count, typeDef.Participants.MaxParticipants),
				ObjectID:   r.ID,
				ObjectType: ObjectTypeRelation,
			})
		}
	}

	return findings
}

func ruleValidParticipantKind(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	s := ctx.Schema.(*schema.Schema)
	var findings []Finding

	for _, r := range g.Relations() {
		typeDef, ok := s.GetRelationTypeDef(r.Type)
		if !ok || typeDef.Participants == nil {
			continue
		}

		for _, participantID := range r.ParticipantIDs() {
			entity, found := g.GetEntity(participantID)
			if !found {
				continue
			}

			kindValid := false
			for _, allowedKind := range typeDef.Participants.SourceKinds {
				if entity.Kind == allowedKind {
					kindValid = true
					break
				}
			}
			if !kindValid {
				for _, allowedKind := range typeDef.Participants.TargetKinds {
					if entity.Kind == allowedKind {
						kindValid = true
						break
					}
				}
			}

			if !kindValid {
				findings = append(findings, Finding{
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("relation %q has participant %q of kind %q which is not typically allowed for type %q", r.ID, participantID, entity.Kind, r.Type),
					ObjectID:   r.ID,
					ObjectType: ObjectTypeRelation,
				})
			}
		}
	}

	return findings
}

func ruleOwnershipTree(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	if err := g.BuildOwnershipPaths(); err != nil {
		findings = append(findings, Finding{
			Severity: SeverityError,
			Message:  fmt.Sprintf("ownership tree is broken: %v", err),
		})
	}

	return findings
}

func ruleNoOwnershipCycle(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	visited := make(map[string]bool)
	for _, e := range g.Entities() {
		if e.Owner == "" {
			continue
		}
		if err := detectCycle(g, e.ID, visited); err != nil {
			findings = append(findings, Finding{
				Severity:   SeverityError,
				Message:    err.Error(),
				ObjectID:   e.ID,
				ObjectType: ObjectTypeEntity,
			})
		}
	}

	return findings
}

func detectCycle(g *core.Graph, entityID string, visited map[string]bool) error {
	path := make(map[string]bool)
	current := entityID

	for current != "" {
		if path[current] {
			return fmt.Errorf("ownership cycle detected involving entity %q", current)
		}
		path[current] = true

		e, found := g.GetEntity(current)
		if !found {
			break
		}
		current = e.Owner
	}

	for id := range path {
		visited[id] = true
	}

	return nil
}

func ruleRootEntity(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	var roots []string
	for _, e := range g.Entities() {
		if e.Owner == "" {
			roots = append(roots, e.ID)
		}
	}

	if len(roots) == 0 {
		findings = append(findings, Finding{
			Severity: SeverityError,
			Message:  "no root entity found (graph must have exactly one root)",
		})
	} else if len(roots) > 1 {
		sort.Strings(roots)
		findings = append(findings, Finding{
			Severity: SeverityError,
			Message:  fmt.Sprintf("multiple root entities found: %v", roots),
		})
	}

	return findings
}

func ruleDanglingReference(ctx *Context) []Finding {
	g := ctx.Graph.(*core.Graph)
	var findings []Finding

	for _, r := range g.Relations() {
		for _, participantID := range r.ParticipantIDs() {
			_, found := g.GetEntity(participantID)
			if !found {
				_, foundRel := g.GetRelation(participantID)
				if !foundRel {
					findings = append(findings, Finding{
						Severity:   SeverityError,
						Message:    fmt.Sprintf("relation %q references non-existent object %q", r.ID, participantID),
						ObjectID:   r.ID,
						ObjectType: ObjectTypeRelation,
					})
				}
			}
		}
	}

	for _, e := range g.Entities() {
		if e.Owner != "" {
			_, found := g.GetEntity(e.Owner)
			if !found {
				findings = append(findings, Finding{
					Severity:   SeverityError,
					Message:    fmt.Sprintf("entity %q references non-existent owner %q", e.ID, e.Owner),
					ObjectID:   e.ID,
					ObjectType: ObjectTypeEntity,
				})
			}
		}
	}

	return findings
}
