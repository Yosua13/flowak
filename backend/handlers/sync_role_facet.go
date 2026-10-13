package handlers

import (
	"database/sql"
)

func syncRoleFacet(tx *sql.Tx, nodeID, roleKey string, facet map[string]any) error {
	if len(facet) == 0 {
		return nil
	}

	assigneeLabel := stringField(facet, "assignee")
	assigneeID, err := resolveAssigneeID(tx, assigneeLabel)
	if err != nil {
		return err
	}
	reviewerID, err := resolveAssigneeID(tx, stringField(facet, "reviewer"))
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO node_role_tasks (
			id, node_id, role_key, assignee_id, assignee_label, reviewer_id, status, readiness, due_date, notes, metadata, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, CURRENT_TIMESTAMP)
		ON CONFLICT (node_id, role_key) DO UPDATE SET
			assignee_id = EXCLUDED.assignee_id,
			assignee_label = EXCLUDED.assignee_label,
			reviewer_id = EXCLUDED.reviewer_id,
			status = EXCLUDED.status,
			readiness = EXCLUDED.readiness,
			due_date = EXCLUDED.due_date,
			notes = EXCLUDED.notes,
			metadata = EXCLUDED.metadata,
			updated_at = CURRENT_TIMESTAMP
	`,
		roleTaskID(nodeID, roleKey),
		nodeID,
		roleKey,
		assigneeID,
		nullableString(assigneeLabel),
		reviewerID,
		normalizeStatus(stringField(facet, "status")),
		normalizeStatus(firstString(facet, "readiness", "status")),
		nullableDate(firstString(facet, "dueDate", "due_date")),
		nullableString(firstString(facet, "notes", "note")),
		jsonText(map[string]any{"source": "frontend_graph"}),
	)
	if err != nil {
		return err
	}

	switch roleKey {
	case "uiux":
		_, err = tx.Exec(`
			INSERT INTO node_uiux_specs (
				node_id, screen_name, persona, prototype_url, wireframe_url, state_notes,
				accessibility_notes, user_goal, surface, figma_frame_url, design_version, screen_states, interactions, content_messages, responsive_intent, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				screen_name = EXCLUDED.screen_name,
				persona = EXCLUDED.persona,
				prototype_url = EXCLUDED.prototype_url,
				wireframe_url = EXCLUDED.wireframe_url,
				state_notes = EXCLUDED.state_notes,
				accessibility_notes = EXCLUDED.accessibility_notes,
				user_goal = EXCLUDED.user_goal, surface = EXCLUDED.surface, figma_frame_url = EXCLUDED.figma_frame_url, design_version = EXCLUDED.design_version,
				screen_states = EXCLUDED.screen_states, interactions = EXCLUDED.interactions, content_messages = EXCLUDED.content_messages, responsive_intent = EXCLUDED.responsive_intent,
				metadata = EXCLUDED.metadata,
				updated_at = CURRENT_TIMESTAMP
		`,
			nodeID,
			nullableString(firstString(facet, "screen", "screenName", "screen_name")),
			nullableString(firstString(facet, "persona")),
			nullableString(firstString(facet, "prototypeUrl", "prototype_url", "link")),
			nullableString(firstString(facet, "wireframeUrl", "wireframe_url")),
			nullableString(firstString(facet, "stateNotes", "state_notes")),
			nullableString(firstString(facet, "accessibilityNotes", "accessibility_notes")),
			nullableString(firstString(facet, "userGoal", "user_goal")), nullableString(stringField(facet, "surface")), nullableString(firstString(facet, "figmaFrameUrl", "figma_frame_url", "link")), nullableString(firstString(facet, "designVersion", "design_version")),
			nullableString(firstString(facet, "screenStates", "screen_states", "stateNotes")), nullableString(stringField(facet, "interactions")), nullableString(firstString(facet, "contentMessages", "content_messages")), nullableString(firstString(facet, "responsiveIntent", "responsive_intent")),
			jsonText(map[string]any{"source": "frontend_graph"}),
		)
	case "frontend":
		_, err = tx.Exec(`
			INSERT INTO node_frontend_specs (
				node_id, page_name, route_path, interaction_notes, validation_notes,
				state_handling, handoff_url, experience_name, entry_exit_behavior, input_requirements, api_references, analytics_intent, feature_availability, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				page_name = EXCLUDED.page_name,
				route_path = EXCLUDED.route_path,
				interaction_notes = EXCLUDED.interaction_notes,
				validation_notes = EXCLUDED.validation_notes,
				state_handling = EXCLUDED.state_handling,
				handoff_url = EXCLUDED.handoff_url,
				experience_name = EXCLUDED.experience_name, entry_exit_behavior = EXCLUDED.entry_exit_behavior, input_requirements = EXCLUDED.input_requirements,
				api_references = EXCLUDED.api_references, analytics_intent = EXCLUDED.analytics_intent, feature_availability = EXCLUDED.feature_availability,
				metadata = EXCLUDED.metadata,
				updated_at = CURRENT_TIMESTAMP
		`,
			nodeID,
			nullableString(firstString(facet, "page", "pageName", "page_name", "component")),
			nullableString(firstString(facet, "route", "routePath", "route_path")),
			nullableString(firstString(facet, "interaction", "interactionNotes", "interaction_notes")),
			nullableString(firstString(facet, "validation", "validationNotes", "validation_notes")),
			nullableString(firstString(facet, "state", "stateHandling", "state_handling")),
			nullableString(firstString(facet, "handoffLink", "handoffUrl", "handoff_url", "link")),
			nullableString(firstString(facet, "experienceName", "experience_name", "page")), nullableString(firstString(facet, "entryExitBehavior", "entry_exit_behavior")), nullableString(firstString(facet, "inputRequirements", "input_requirements")),
			nullableString(firstString(facet, "apiReferences", "api_references")), nullableString(firstString(facet, "analyticsIntent", "analytics_intent")), nullableString(firstString(facet, "featureAvailability", "feature_availability")),
			jsonText(map[string]any{"source": "frontend_graph"}),
		)
	case "backend":
		requestJSON := jsonbFromText(stringField(facet, "request"))
		responseJSON := jsonbFromText(stringField(facet, "response"))
		errorJSON := jsonbFromText(firstString(facet, "errorCodes", "error_codes"))
		_, err = tx.Exec(`
			INSERT INTO node_api_contracts (
				node_id, method, endpoint_path, auth_policy, request_example, response_example,
				status_code, error_codes, curl_example, service_capability, api_references, business_validation, dependency_references, idempotency_notes, caching_notes, security_notes, observability_intent, sla_value, sla_unit, metadata, updated_at
			) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8::jsonb, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21::jsonb, CURRENT_TIMESTAMP)
			ON CONFLICT (node_id) DO UPDATE SET
				method = EXCLUDED.method,
				endpoint_path = EXCLUDED.endpoint_path,
				auth_policy = EXCLUDED.auth_policy,
				request_example = EXCLUDED.request_example,
				response_example = EXCLUDED.response_example,
				status_code = EXCLUDED.status_code,
				error_codes = EXCLUDED.error_codes,
				curl_example = EXCLUDED.curl_example,
				service_capability = EXCLUDED.service_capability, api_references = EXCLUDED.api_references, business_validation = EXCLUDED.business_validation, dependency_references = EXCLUDED.dependency_references,
				idempotency_notes = EXCLUDED.idempotency_notes, caching_notes = EXCLUDED.caching_notes, security_notes = EXCLUDED.security_notes, observability_intent = EXCLUDED.observability_intent,
				sla_value = EXCLUDED.sla_value, sla_unit = EXCLUDED.sla_unit,
				metadata = EXCLUDED.metadata,
				updated_at = CURRENT_TIMESTAMP
		`,
			nodeID,
			nullableString(normalizeMethod(stringField(facet, "method"))),
			nullableString(stringField(facet, "endpoint")),
			nullableString(firstString(facet, "auth", "authPolicy", "auth_policy")),
			requestJSON,
			responseJSON,
			nullableString(firstString(facet, "statusCode", "status_code")),
			errorJSON,
			nullableString(stringField(facet, "curl")),
			nullableString(firstString(facet, "serviceCapability", "service_capability")), nullableString(firstString(facet, "apiReferences", "api_references")), nullableString(firstString(facet, "businessValidation", "business_validation")), nullableString(firstString(facet, "dependencyReferences", "dependency_references")),
			nullableString(firstString(facet, "idempotencyNotes", "idempotency_notes")), nullableString(firstString(facet, "cachingNotes", "caching_notes")), nullableString(firstString(facet, "securityNotes", "security_notes")), nullableString(firstString(facet, "observabilityIntent", "observability_intent")),
			func() any { value, _ := parseSLA(stringField(facet, "sla")); return value }(), func() any { _, unit := parseSLA(stringField(facet, "sla")); return unit }(),
			jsonText(map[string]any{"source": "frontend_graph"}),
		)
	}

	return err
}
