package module

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"backend/internal/domain/module"
)

// FindGraphByModuleID loads normalized workflow nodes, edges, doc, and facets for a module.
func FindGraphByModuleID(ctx context.Context, db *sql.DB, moduleID string) (*module.ModuleGraph, error) {
	var g module.ModuleGraph
	err := db.QueryRowContext(ctx, "SELECT id, project_id, name, description, schema_version, version FROM modules WHERE id = $1 AND status = 'active'", moduleID).
		Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.SchemaVersion, &g.Version)
	if err == sql.ErrNoRows {
		return nil, module.ErrModuleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query module: %w", err)
	}

	// 1. Query nodes
	rows, err := db.QueryContext(ctx, `
		SELECT id, type, label, x, y, actor, trigger, input_desc, process_desc,
			output_desc, business_rules, exception_path, system_context, sla_value,
			sla_unit, priority, risk_level, acceptance_criteria, outcome, trigger_type, preconditions, reference_links, metadata, row_version
		FROM workflow_nodes
		WHERE module_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`, moduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow nodes: %w", err)
	}
	defer rows.Close()

	nodes := make([]module.Node, 0)
	nodeIndex := make(map[string]*module.Node)

	for rows.Next() {
		var n module.Node
		var actor, trigger, input, process, output, rules, exceptionPath, systemContext sql.NullString
		var slaValue sql.NullFloat64
		var slaUnit, priority, riskLevel, acceptanceCriteria, outcome, triggerType, preconditions, referenceLinks sql.NullString
		var metadata []byte

		if err := rows.Scan(&n.ID, &n.Type, &n.Label, &n.X, &n.Y, &actor, &trigger, &input, &process, &output, &rules, &exceptionPath, &systemContext, &slaValue, &slaUnit, &priority, &riskLevel, &acceptanceCriteria, &outcome, &triggerType, &preconditions, &referenceLinks, &metadata, &n.RowVersion); err != nil {
			return nil, fmt.Errorf("failed to scan workflow node: %w", err)
		}

		doc := map[string]any{
			"actor":   actor.String,
			"input":   input.String,
			"process": process.String,
			"output":  output.String,
			"rules":   rules.String,
			"system":  systemContext.String,
			"sla":     formatSLA(slaValue, slaUnit),
		}
		if trigger.Valid && trigger.String != "" {
			doc["trigger"] = trigger.String
		}
		if exceptionPath.Valid && exceptionPath.String != "" {
			doc["exceptionPath"] = exceptionPath.String
		}
		if priority.Valid && priority.String != "" {
			doc["priority"] = priority.String
		}
		if riskLevel.Valid && riskLevel.String != "" {
			doc["riskLevel"] = riskLevel.String
		}
		if acceptanceCriteria.Valid && acceptanceCriteria.String != "" {
			doc["acceptanceCriteria"] = acceptanceCriteria.String
		}
		if outcome.Valid && outcome.String != "" {
			doc["outcome"] = outcome.String
		}
		if triggerType.Valid && triggerType.String != "" {
			doc["triggerType"] = triggerType.String
		}
		if preconditions.Valid && preconditions.String != "" {
			doc["preconditions"] = preconditions.String
		}
		if referenceLinks.Valid && referenceLinks.String != "" {
			doc["referenceLinks"] = referenceLinks.String
		}

		n.Doc = doc
		n.Roles = make(map[string]any)

		var meta map[string]any
		if json.Unmarshal(metadata, &meta) == nil {
			if notes := mapField(meta, "legacy_notes"); len(notes) > 0 {
				n.LegacyNotes = notes
			}
		}

		nodes = append(nodes, n)
		nodeIndex[n.ID] = &nodes[len(nodes)-1]
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Hydrate business details and role specs if nodes exist
	if len(nodes) > 0 {
		if err := hydrateDetails(ctx, db, moduleID, nodeIndex); err != nil {
			return nil, err
		}
	}

	// 2. Query edges
	edgeRows, err := db.QueryContext(ctx, `
		SELECT id, from_node_id, to_node_id, label, condition_text, row_version
		FROM workflow_edges
		WHERE module_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`, moduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow edges: %w", err)
	}
	defer edgeRows.Close()

	edges := make([]module.Edge, 0)
	for edgeRows.Next() {
		var e module.Edge
		var label, condition sql.NullString
		if err := edgeRows.Scan(&e.ID, &e.From, &e.To, &label, &condition, &e.RowVersion); err != nil {
			return nil, fmt.Errorf("failed to scan workflow edge: %w", err)
		}
		if label.Valid && label.String != "" {
			e.Label = label.String
		}
		if condition.Valid && condition.String != "" {
			e.Condition = condition.String
		}
		edges = append(edges, e)
	}
	if err := edgeRows.Err(); err != nil {
		return nil, err
	}

	g.Nodes = nodes
	g.Edges = edges
	return &g, nil
}

func hydrateDetails(ctx context.Context, db *sql.DB, moduleID string, nodeIndex map[string]*module.Node) error {
	// Business rules
	bRows, err := db.QueryContext(ctx, `SELECT r.node_id, COALESCE(r.rule_code, ''), r.severity, r.description FROM node_business_rules r JOIN workflow_nodes n ON n.id = r.node_id WHERE n.module_id = $1 ORDER BY r.sort_order`, moduleID)
	if err == nil {
		defer bRows.Close()
		for bRows.Next() {
			var nodeID, code, severity, description string
			if err := bRows.Scan(&nodeID, &code, &severity, &description); err == nil {
				if n, ok := nodeIndex[nodeID]; ok {
					curr, _ := n.Doc["rules"].([]map[string]any)
					n.Doc["rules"] = append(curr, map[string]any{"code": code, "severity": severity, "description": description})
				}
			}
		}
	}

	// Decision outcomes
	dRows, err := db.QueryContext(ctx, `SELECT o.node_id, o.outcome, COALESCE(o.edge_id, '') FROM node_decision_outcomes o JOIN workflow_nodes n ON n.id = o.node_id WHERE n.module_id = $1 ORDER BY o.sort_order`, moduleID)
	if err == nil {
		defer dRows.Close()
		for dRows.Next() {
			var nodeID, outcome, edgeID string
			if err := dRows.Scan(&nodeID, &outcome, &edgeID); err == nil {
				if n, ok := nodeIndex[nodeID]; ok {
					curr, _ := n.Doc["decisionOutcomes"].([]map[string]any)
					item := map[string]any{"outcome": outcome}
					if edgeID != "" {
						item["edgeId"] = edgeID
					}
					n.Doc["decisionOutcomes"] = append(curr, item)
				}
			}
		}
	}

	// Role tasks
	tRows, err := db.QueryContext(ctx, `
		SELECT t.node_id, t.role_key, COALESCE(u.name, t.assignee_label, '') AS assignee,
			t.status, COALESCE(t.due_date::text, '') AS due_date, COALESCE(t.notes, '') AS notes
		FROM node_role_tasks t
		JOIN workflow_nodes n ON n.id = t.node_id
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE n.module_id = $1
	`, moduleID)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var nodeID, roleKey, assignee, status, dueDate, notes string
			if err := tRows.Scan(&nodeID, &roleKey, &assignee, &status, &dueDate, &notes); err == nil {
				if n, ok := nodeIndex[nodeID]; ok {
					role := ensureRoleMap(n, roleKey)
					role["assignee"] = assignee
					role["status"] = status
					if dueDate != "" {
						role["dueDate"] = dueDate
					}
					if notes != "" {
						role["notes"] = notes
					}
				}
			}
		}
	}

	// UI/UX specs
	uRows, err := db.QueryContext(ctx, `
		SELECT s.node_id, COALESCE(s.screen_name, ''), COALESCE(s.prototype_url, ''),
			COALESCE(s.wireframe_url, ''), COALESCE(s.state_notes, ''), COALESCE(s.accessibility_notes, ''), COALESCE(s.user_goal, ''), COALESCE(s.surface, ''), COALESCE(s.figma_frame_url, ''), COALESCE(s.design_version, ''), COALESCE(s.screen_states, ''), COALESCE(s.interactions, ''), COALESCE(s.content_messages, ''), COALESCE(s.responsive_intent, '')
		FROM node_uiux_specs s
		JOIN workflow_nodes n ON n.id = s.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err == nil {
		defer uRows.Close()
		for uRows.Next() {
			var nodeID, screen, prototype, wireframe, stateNotes, accessibility, userGoal, surface, figmaFrame, designVersion, screenStates, interactions, contentMessages, responsiveIntent string
			if err := uRows.Scan(&nodeID, &screen, &prototype, &wireframe, &stateNotes, &accessibility, &userGoal, &surface, &figmaFrame, &designVersion, &screenStates, &interactions, &contentMessages, &responsiveIntent); err == nil {
				if n, ok := nodeIndex[nodeID]; ok {
					role := ensureRoleMap(n, "uiux")
					role["screen"] = screen
					role["link"] = prototype
					if wireframe != "" {
						role["wireframeUrl"] = wireframe
					}
					if stateNotes != "" {
						role["stateNotes"] = stateNotes
					}
					if accessibility != "" {
						role["accessibilityNotes"] = accessibility
					}
					role["userGoal"], role["surface"], role["figmaFrameUrl"], role["designVersion"] = userGoal, surface, figmaFrame, designVersion
					role["screenStates"], role["interactions"], role["contentMessages"], role["responsiveIntent"] = screenStates, interactions, contentMessages, responsiveIntent
				}
			}
		}
	}

	// Frontend specs
	fRows, err := db.QueryContext(ctx, `
		SELECT s.node_id, COALESCE(s.page_name, ''), COALESCE(s.route_path, ''),
			COALESCE(s.interaction_notes, ''), COALESCE(s.validation_notes, ''),
			COALESCE(s.state_handling, ''), COALESCE(s.handoff_url, ''), COALESCE(s.experience_name, ''), COALESCE(s.entry_exit_behavior, ''), COALESCE(s.input_requirements, ''), COALESCE(s.api_references, ''), COALESCE(s.analytics_intent, ''), COALESCE(s.feature_availability, '')
		FROM node_frontend_specs s
		JOIN workflow_nodes n ON n.id = s.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err == nil {
		defer fRows.Close()
		for fRows.Next() {
			var nodeID, page, route, interaction, validation, stateHandling, handoff, experience, entryExit, inputs, apiRefs, analytics, availability string
			if err := fRows.Scan(&nodeID, &page, &route, &interaction, &validation, &stateHandling, &handoff, &experience, &entryExit, &inputs, &apiRefs, &analytics, &availability); err == nil {
				if n, ok := nodeIndex[nodeID]; ok {
					role := ensureRoleMap(n, "frontend")
					role["page"] = page
					role["component"] = page
					role["route"] = route
					role["interaction"] = interaction
					role["validation"] = validation
					role["state"] = stateHandling
					role["handoffLink"] = handoff
					role["link"] = handoff
					role["experienceName"], role["entryExitBehavior"], role["inputRequirements"] = experience, entryExit, inputs
					role["apiReferences"], role["analyticsIntent"], role["featureAvailability"] = apiRefs, analytics, availability
				}
			}
		}
	}

	// Backend API contracts
	cRows, err := db.QueryContext(ctx, `
		SELECT c.node_id, COALESCE(c.method, ''), COALESCE(c.endpoint_path, ''),
			COALESCE(c.auth_policy, ''), COALESCE(c.request_example::text, ''),
			COALESCE(c.response_example::text, ''), COALESCE(c.status_code, ''),
			COALESCE(c.error_codes::text, ''), COALESCE(c.service_capability, ''), COALESCE(c.api_references, ''), COALESCE(c.business_validation, ''), COALESCE(c.dependency_references, ''), COALESCE(c.idempotency_notes, ''), COALESCE(c.caching_notes, ''), COALESCE(c.security_notes, ''), COALESCE(c.observability_intent, ''), c.sla_value, COALESCE(c.sla_unit, '')
		FROM node_api_contracts c
		JOIN workflow_nodes n ON n.id = c.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var nodeID, method, endpoint, auth, requestBody, responseBody, statusCode, errorCodes, capability, apiRefs, validation, dependencies, idempotency, caching, security, observability, slaUnit string
			var slaValue sql.NullFloat64
			if err := cRows.Scan(&nodeID, &method, &endpoint, &auth, &requestBody, &responseBody, &statusCode, &errorCodes, &capability, &apiRefs, &validation, &dependencies, &idempotency, &caching, &security, &observability, &slaValue, &slaUnit); err == nil {
				if n, ok := nodeIndex[nodeID]; ok {
					role := ensureRoleMap(n, "backend")
					role["method"] = method
					role["endpoint"] = endpoint
					role["auth"] = auth
					role["request"] = requestBody
					role["response"] = responseBody
					role["statusCode"] = statusCode
					if errorCodes != "" {
						role["errorCodes"] = errorCodes
					}
					role["serviceCapability"], role["apiReferences"], role["businessValidation"], role["dependencyReferences"] = capability, apiRefs, validation, dependencies
					role["idempotencyNotes"], role["cachingNotes"], role["securityNotes"], role["observabilityIntent"] = idempotency, caching, security, observability
					role["sla"] = formatSLA(slaValue, sql.NullString{String: slaUnit, Valid: slaUnit != ""})
				}
			}
		}
	}

	return nil
}

func ensureRoleMap(n *module.Node, roleKey string) map[string]any {
	if n.Roles == nil {
		n.Roles = make(map[string]any)
	}
	role, ok := n.Roles[roleKey].(map[string]any)
	if !ok || role == nil {
		role = make(map[string]any)
		n.Roles[roleKey] = role
	}
	return role
}
