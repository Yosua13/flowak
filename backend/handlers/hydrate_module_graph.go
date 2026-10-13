package handlers

import (
	"database/sql"
	"encoding/json"

	"backend/db"
	"backend/models"
)

func hydrateModuleGraph(module *models.Module) error {
	rows, err := db.DB.Query(`
		SELECT id, type, label, x, y, actor, trigger, input_desc, process_desc,
			output_desc, business_rules, exception_path, system_context, sla_value,
			sla_unit, priority, risk_level, acceptance_criteria, outcome, trigger_type, preconditions, reference_links, metadata, row_version
		FROM workflow_nodes
		WHERE module_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`, module.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	nodes := []map[string]any{}
	nodeIndex := map[string]map[string]any{}
	for rows.Next() {
		var id, nodeType, label string
		var x, y float64
		var actor, trigger, input, process, output, rules, exceptionPath, systemContext sql.NullString
		var slaValue sql.NullFloat64
		var slaUnit, priority, riskLevel, acceptanceCriteria, outcome, triggerType, preconditions, referenceLinks sql.NullString
		var rowVersion int
		var metadata []byte
		if err := rows.Scan(&id, &nodeType, &label, &x, &y, &actor, &trigger, &input, &process, &output, &rules, &exceptionPath, &systemContext, &slaValue, &slaUnit, &priority, &riskLevel, &acceptanceCriteria, &outcome, &triggerType, &preconditions, &referenceLinks, &metadata, &rowVersion); err != nil {
			return err
		}

		doc := map[string]any{
			"actor":   nullStringValue(actor),
			"input":   nullStringValue(input),
			"process": nullStringValue(process),
			"output":  nullStringValue(output),
			"rules":   nullStringValue(rules),
			"system":  nullStringValue(systemContext),
			"sla":     formatSLA(slaValue, slaUnit),
		}
		putIfString(doc, "trigger", trigger)
		putIfString(doc, "exceptionPath", exceptionPath)
		putIfString(doc, "priority", priority)
		putIfString(doc, "riskLevel", riskLevel)
		putIfString(doc, "acceptanceCriteria", acceptanceCriteria)
		putIfString(doc, "outcome", outcome)
		putIfString(doc, "triggerType", triggerType)
		putIfString(doc, "preconditions", preconditions)
		putIfString(doc, "referenceLinks", referenceLinks)

		node := map[string]any{
			"id":         id,
			"type":       nodeType,
			"label":      label,
			"x":          x,
			"y":          y,
			"rowVersion": rowVersion,
			"doc":        doc,
			"roles":      map[string]any{},
		}
		var meta map[string]any
		if json.Unmarshal(metadata, &meta) == nil {
			if notes := mapField(meta, "legacy_notes"); len(notes) > 0 {
				node["legacyNotes"] = notes
			}
		}
		nodes = append(nodes, node)
		nodeIndex[id] = node
	}

	if len(nodes) == 0 {
		return nil
	}
	if err := hydrateBusinessDetails(module.ID, nodeIndex); err != nil {
		return err
	}

	if err := hydrateRoleTasks(module.ID, nodeIndex); err != nil {
		return err
	}
	if err := hydrateUiuxSpecs(module.ID, nodeIndex); err != nil {
		return err
	}
	if err := hydrateFrontendSpecs(module.ID, nodeIndex); err != nil {
		return err
	}
	if err := hydrateBackendContracts(module.ID, nodeIndex); err != nil {
		return err
	}
	for _, node := range nodes {
		node["completeness"] = specificationCompleteness(node)
	}

	edgeRows, err := db.DB.Query(`
		SELECT id, from_node_id, to_node_id, label, condition_text, row_version
		FROM workflow_edges
		WHERE module_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`, module.ID)
	if err != nil {
		return err
	}
	defer edgeRows.Close()

	edges := []map[string]any{}
	for edgeRows.Next() {
		var id, fromID, toID string
		var label, condition sql.NullString
		var rowVersion int
		if err := edgeRows.Scan(&id, &fromID, &toID, &label, &condition, &rowVersion); err != nil {
			return err
		}
		edge := map[string]any{
			"id":         id,
			"from":       fromID,
			"to":         toID,
			"rowVersion": rowVersion,
		}
		putIfString(edge, "label", label)
		putIfString(edge, "condition", condition)
		edges = append(edges, edge)
	}

	nodesJSON, _ := json.Marshal(nodes)
	edgesJSON, _ := json.Marshal(edges)
	module.Nodes = string(nodesJSON)
	module.Edges = string(edgesJSON)
	return nil
}

func hydrateBusinessDetails(moduleID string, nodeIndex map[string]map[string]any) error {
	rules, err := db.DB.Query(`SELECT r.node_id, COALESCE(r.rule_code, ''), r.severity, r.description FROM node_business_rules r JOIN workflow_nodes n ON n.id = r.node_id WHERE n.module_id = $1 ORDER BY r.sort_order`, moduleID)
	if err != nil {
		return err
	}
	defer rules.Close()
	for rules.Next() {
		var nodeID, code, severity, description string
		if err := rules.Scan(&nodeID, &code, &severity, &description); err != nil {
			return err
		}
		node := nodeIndex[nodeID]
		doc, _ := node["doc"].(map[string]any)
		current, _ := doc["rules"].([]map[string]any)
		doc["rules"] = append(current, map[string]any{"code": code, "severity": severity, "description": description})
	}
	if err := rules.Err(); err != nil {
		return err
	}
	outcomes, err := db.DB.Query(`SELECT o.node_id, o.outcome, COALESCE(o.edge_id, '') FROM node_decision_outcomes o JOIN workflow_nodes n ON n.id = o.node_id WHERE n.module_id = $1 ORDER BY o.sort_order`, moduleID)
	if err != nil {
		return err
	}
	defer outcomes.Close()
	for outcomes.Next() {
		var nodeID, outcome, edgeID string
		if err := outcomes.Scan(&nodeID, &outcome, &edgeID); err != nil {
			return err
		}
		node := nodeIndex[nodeID]
		doc, _ := node["doc"].(map[string]any)
		current, _ := doc["decisionOutcomes"].([]map[string]any)
		item := map[string]any{"outcome": outcome}
		if edgeID != "" {
			item["edgeId"] = edgeID
		}
		doc["decisionOutcomes"] = append(current, item)
	}
	return outcomes.Err()
}

func hydrateRoleTasks(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT t.node_id, t.role_key, COALESCE(u.name, t.assignee_label, '') AS assignee,
			t.status, COALESCE(t.due_date::text, '') AS due_date, COALESCE(t.notes, '') AS notes
		FROM node_role_tasks t
		JOIN workflow_nodes n ON n.id = t.node_id
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, roleKey, assignee, status, dueDate, notes string
		if err := rows.Scan(&nodeID, &roleKey, &assignee, &status, &dueDate, &notes); err != nil {
			return err
		}
		role := roleMap(nodeIndex, nodeID, roleKey)
		role["assignee"] = assignee
		role["status"] = status
		if dueDate != "" {
			role["dueDate"] = dueDate
		}
		if notes != "" {
			role["notes"] = notes
		}
	}
	return nil
}

func hydrateUiuxSpecs(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT s.node_id, COALESCE(s.screen_name, ''), COALESCE(s.prototype_url, ''),
			COALESCE(s.wireframe_url, ''), COALESCE(s.state_notes, ''), COALESCE(s.accessibility_notes, ''), COALESCE(s.user_goal, ''), COALESCE(s.surface, ''), COALESCE(s.figma_frame_url, ''), COALESCE(s.design_version, ''), COALESCE(s.screen_states, ''), COALESCE(s.interactions, ''), COALESCE(s.content_messages, ''), COALESCE(s.responsive_intent, '')
		FROM node_uiux_specs s
		JOIN workflow_nodes n ON n.id = s.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, screen, prototype, wireframe, stateNotes, accessibility, userGoal, surface, figmaFrame, designVersion, screenStates, interactions, contentMessages, responsiveIntent string
		if err := rows.Scan(&nodeID, &screen, &prototype, &wireframe, &stateNotes, &accessibility, &userGoal, &surface, &figmaFrame, &designVersion, &screenStates, &interactions, &contentMessages, &responsiveIntent); err != nil {
			return err
		}
		role := roleMap(nodeIndex, nodeID, "uiux")
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
	return nil
}

func hydrateFrontendSpecs(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT s.node_id, COALESCE(s.page_name, ''), COALESCE(s.route_path, ''),
			COALESCE(s.interaction_notes, ''), COALESCE(s.validation_notes, ''),
			COALESCE(s.state_handling, ''), COALESCE(s.handoff_url, ''), COALESCE(s.experience_name, ''), COALESCE(s.entry_exit_behavior, ''), COALESCE(s.input_requirements, ''), COALESCE(s.api_references, ''), COALESCE(s.analytics_intent, ''), COALESCE(s.feature_availability, '')
		FROM node_frontend_specs s
		JOIN workflow_nodes n ON n.id = s.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, page, route, interaction, validation, stateHandling, handoff, experience, entryExit, inputs, apiRefs, analytics, availability string
		if err := rows.Scan(&nodeID, &page, &route, &interaction, &validation, &stateHandling, &handoff, &experience, &entryExit, &inputs, &apiRefs, &analytics, &availability); err != nil {
			return err
		}
		role := roleMap(nodeIndex, nodeID, "frontend")
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
	return nil
}

func hydrateBackendContracts(moduleID string, nodeIndex map[string]map[string]any) error {
	rows, err := db.DB.Query(`
		SELECT c.node_id, COALESCE(c.method, ''), COALESCE(c.endpoint_path, ''),
			COALESCE(c.auth_policy, ''), COALESCE(c.request_example::text, ''),
			COALESCE(c.response_example::text, ''), COALESCE(c.status_code, ''),
			COALESCE(c.error_codes::text, ''), COALESCE(c.service_capability, ''), COALESCE(c.api_references, ''), COALESCE(c.business_validation, ''), COALESCE(c.dependency_references, ''), COALESCE(c.idempotency_notes, ''), COALESCE(c.caching_notes, ''), COALESCE(c.security_notes, ''), COALESCE(c.observability_intent, ''), c.sla_value, COALESCE(c.sla_unit, '')
		FROM node_api_contracts c
		JOIN workflow_nodes n ON n.id = c.node_id
		WHERE n.module_id = $1
	`, moduleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID, method, endpoint, auth, requestBody, responseBody, statusCode, errorCodes, capability, apiRefs, validation, dependencies, idempotency, caching, security, observability, slaUnit string
		var slaValue sql.NullFloat64
		if err := rows.Scan(&nodeID, &method, &endpoint, &auth, &requestBody, &responseBody, &statusCode, &errorCodes, &capability, &apiRefs, &validation, &dependencies, &idempotency, &caching, &security, &observability, &slaValue, &slaUnit); err != nil {
			return err
		}
		role := roleMap(nodeIndex, nodeID, "backend")
		role["method"] = method
		role["endpoint"] = endpoint
		if auth != "" && !validReference(auth) && !oneOf(auth, "none", "inherit") {
			role["auth"] = "{{API_TOKEN}}"
			node := nodeIndex[nodeID]
			notes := mapField(node, "legacyNotes")
			notes["backend.auth"] = "Legacy authorization removed; choose a variable reference."
			node["legacyNotes"] = notes
		} else {
			role["auth"] = auth
		}
		role["request"] = scrubLegacyDisplay(jsonbDisplay(requestBody))
		role["response"] = scrubLegacyDisplay(jsonbDisplay(responseBody))
		role["statusCode"] = statusCode
		if errorCodes != "" {
			role["errorCodes"] = scrubLegacyDisplay(jsonbDisplay(errorCodes))
		}
		role["serviceCapability"], role["apiReferences"], role["businessValidation"], role["dependencyReferences"] = capability, apiRefs, validation, dependencies
		role["idempotencyNotes"], role["cachingNotes"], role["securityNotes"], role["observabilityIntent"] = idempotency, caching, security, observability
		role["sla"] = formatSLA(slaValue, sql.NullString{String: slaUnit, Valid: slaUnit != ""})
	}
	return nil
}
