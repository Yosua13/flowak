package module

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/domain/module"
)

// UpsertNodes synchronizes workflow nodes, facets, business rules, and decision outcomes.
func UpsertNodes(ctx context.Context, tx *sql.Tx, moduleID string, nodes []module.Node) error {
	for idx, n := range nodes {
		nodeID := n.ID
		if nodeID == "" {
			nodeID = "node_" + generateUUID()
			nodes[idx].ID = nodeID
		}

		doc := n.Doc
		if doc == nil {
			doc = make(map[string]any)
		}
		slaValue, slaUnit := module.ParseSLAAny(stringField(doc, "sla"))

		expectedVersion := n.RowVersion
		hasVersion := expectedVersion > 0

		metadata := jsonText(map[string]any{"source": "frontend_graph", "legacy_notes": n.LegacyNotes})

		var currentVersion int
		err := tx.QueryRowContext(ctx, "SELECT row_version FROM workflow_nodes WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL", nodeID, moduleID).Scan(&currentVersion)
		if err == sql.ErrNoRows {
			var foreignModule string
			if err = tx.QueryRowContext(ctx, "SELECT module_id FROM workflow_nodes WHERE id = $1", nodeID).Scan(&foreignModule); err == nil {
				return module.ErrInvalidGraphReference
			}

			_, err = tx.ExecContext(ctx, `
				INSERT INTO workflow_nodes (
					id, module_id, type, label, x, y, actor, trigger, input_desc, process_desc,
					output_desc, business_rules, exception_path, system_context, sla_value, sla_unit,
					priority, risk_level, acceptance_criteria, outcome, trigger_type, preconditions, reference_links, metadata, sort_order, row_version, updated_at
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
					$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24::jsonb, $25, 1, CURRENT_TIMESTAMP
				)
			`,
				nodeID,
				moduleID,
				normalizeNodeType(n.Type),
				defaultString(n.Label, "Langkah Alur"),
				n.X,
				n.Y,
				nullableString(stringField(doc, "actor")),
				nullableString(firstString(doc, "trigger", "event")),
				nullableString(stringField(doc, "input")),
				nullableString(stringField(doc, "process")),
				nullableString(stringField(doc, "output")),
				nullableString(rulesText(doc)),
				nullableString(firstString(doc, "exceptionPath", "exception_path")),
				nullableString(stringField(doc, "system")),
				slaValue,
				slaUnit,
				defaultString(firstString(doc, "priority"), "medium"),
				defaultString(firstString(doc, "riskLevel", "risk_level"), "medium"),
				nullableString(firstString(doc, "acceptanceCriteria", "acceptance_criteria")),
				nullableString(stringField(doc, "outcome")),
				nullableString(firstString(doc, "triggerType", "trigger_type")),
				nullableString(stringField(doc, "preconditions")),
				nullableString(firstString(doc, "referenceLinks", "reference_links")),
				metadata,
				idx,
			)
			if err != nil {
				return fmt.Errorf("failed to insert workflow node: %w", err)
			}
		} else if err != nil {
			return err
		} else {
			// Row exists -> check version
			if !hasVersion || expectedVersion != currentVersion {
				return module.ErrGraphConflict
			}

			result, err := tx.ExecContext(ctx, `UPDATE workflow_nodes SET
				type = $3, label = $4, x = $5, y = $6, actor = $7, trigger = $8, input_desc = $9, process_desc = $10,
				output_desc = $11, business_rules = $12, exception_path = $13, system_context = $14, sla_value = $15, sla_unit = $16,
				priority = $17, risk_level = $18, acceptance_criteria = $19, outcome = $20, trigger_type = $21, preconditions = $22, reference_links = $23, metadata = $24::jsonb, sort_order = $25,
				row_version = row_version + 1, updated_at = CURRENT_TIMESTAMP
				WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL AND row_version = $26
				AND (type, label, x, y, actor, trigger, input_desc, process_desc, output_desc, business_rules, exception_path, system_context, sla_value, sla_unit, priority, risk_level, acceptance_criteria, outcome, trigger_type, preconditions, reference_links, metadata, sort_order)
				IS DISTINCT FROM ($3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24::jsonb, $25)`,
				nodeID, moduleID, normalizeNodeType(n.Type), defaultString(n.Label, "Langkah Alur"), n.X, n.Y,
				nullableString(stringField(doc, "actor")), nullableString(firstString(doc, "trigger", "event")), nullableString(stringField(doc, "input")), nullableString(stringField(doc, "process")),
				nullableString(stringField(doc, "output")), nullableString(rulesText(doc)), nullableString(firstString(doc, "exceptionPaths", "exceptionPath", "exception_path")), nullableString(stringField(doc, "system")),
				slaValue, slaUnit, defaultString(firstString(doc, "priority"), "medium"), defaultString(firstString(doc, "riskLevel", "risk_level"), "medium"), nullableString(firstString(doc, "acceptanceCriteria", "acceptance_criteria")),
				nullableString(stringField(doc, "outcome")), nullableString(firstString(doc, "triggerType", "trigger_type")), nullableString(stringField(doc, "preconditions")), nullableString(firstString(doc, "referenceLinks", "reference_links")), metadata, idx, expectedVersion)
			if err != nil {
				return err
			}
			changed, err := result.RowsAffected()
			if err != nil || changed > 0 {
				if err != nil {
					return err
				}
			} else {
				if err := tx.QueryRowContext(ctx, "SELECT row_version FROM workflow_nodes WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL", nodeID, moduleID).Scan(&currentVersion); err != nil {
					return err
				}
				if currentVersion != expectedVersion {
					return module.ErrGraphConflict
				}
			}
		}

		// Sync roles
		if len(n.Roles) > 0 {
			if err := syncRoleFacet(ctx, tx, nodeID, "uiux", mapField(n.Roles, "uiux")); err != nil {
				return err
			}
			if err := syncRoleFacet(ctx, tx, nodeID, "frontend", mapField(n.Roles, "frontend")); err != nil {
				return err
			}
			if err := syncRoleFacet(ctx, tx, nodeID, "backend", mapField(n.Roles, "backend")); err != nil {
				return err
			}
		}

		// Sync business details (rules and outcomes)
		if err := syncBusinessDetails(ctx, tx, moduleID, nodeID, doc); err != nil {
			return err
		}
	}

	return nil
}

func syncRoleFacet(ctx context.Context, tx *sql.Tx, nodeID, roleKey string, facet map[string]any) error {
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

	_, err = tx.ExecContext(ctx, `
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
		_, err = tx.ExecContext(ctx, `
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
		_, err = tx.ExecContext(ctx, `
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
		slaVal, slaUnt := module.ParseSLAAny(stringField(facet, "sla"))
		_, err = tx.ExecContext(ctx, `
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
			slaVal, slaUnt,
			jsonText(map[string]any{"source": "frontend_graph"}),
		)
	}

	return err
}

func syncBusinessDetails(ctx context.Context, tx *sql.Tx, moduleID, nodeID string, doc map[string]any) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM node_business_rules WHERE node_id = $1", nodeID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM node_decision_outcomes WHERE node_id = $1", nodeID); err != nil {
		return err
	}
	for idx, rule := range arrayMaps(doc["rules"]) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO node_business_rules (id,node_id,rule_code,severity,description,sort_order) VALUES ($1,$2,$3,$4,$5,$6)`,
			fmt.Sprintf("rule_%s_%d", nodeID, idx), nodeID, nullableString(stringField(rule, "code")), normalizeSeverity(stringField(rule, "severity")), defaultString(stringField(rule, "description"), "Rule"), idx); err != nil {
			return err
		}
	}
	for idx, outcome := range arrayMaps(doc["decisionOutcomes"]) {
		edgeID := firstString(outcome, "edgeId", "edge_id")
		if edgeID != "" {
			var exists bool
			if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM workflow_edges WHERE id = $1 AND module_id = $2 AND deleted_at IS NULL)", edgeID, moduleID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return module.ErrInvalidGraphReference
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO node_decision_outcomes (id,node_id,edge_id,outcome,sort_order) VALUES ($1,$2,$3,$4,$5)`,
			fmt.Sprintf("outcome_%s_%d", nodeID, idx), nodeID, nullableString(edgeID), defaultString(stringField(outcome, "outcome"), "Outcome"), idx); err != nil {
			return err
		}
	}
	return nil
}
