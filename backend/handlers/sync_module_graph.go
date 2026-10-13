package handlers

import (
	"database/sql"
	"fmt"

	"backend/models"
)

func syncModuleGraph(tx *sql.Tx, moduleID string, nodesValue any, edgesValue any, deletedNodes, deletedEdges []models.GraphDelete) error {
	nodes, err := normalizeJSONArray(nodesValue)
	if err != nil {
		return fmt.Errorf("invalid nodes payload: %w", err)
	}
	edges, err := normalizeJSONArray(edgesValue)
	if err != nil {
		return fmt.Errorf("invalid edges payload: %w", err)
	}

	nodeIDs, err := activeNodeIDs(tx, moduleID)
	if err != nil {
		return err
	}
	for idx, node := range nodes {
		nodeID := stringField(node, "id")
		if nodeID == "" {
			nodeID = "node_" + GenerateUUID()
			node["id"] = nodeID
		}
		nodeIDs[nodeID] = true

		doc := mapField(node, "doc")
		if err := validateSpecificationNode(node); err != nil {
			return err
		}
		slaValue, slaUnit := parseSLA(stringField(doc, "sla"))
		expectedVersion, hasVersion := rowVersion(node)
		if err := upsertNode(tx, moduleID, nodeID, idx, node, doc, slaValue, slaUnit, expectedVersion, hasVersion); err != nil {
			return err
		}
		roles := mapField(node, "roles")
		if err := syncRoleFacet(tx, nodeID, "uiux", mapField(roles, "uiux")); err != nil {
			return err
		}
		if err := syncRoleFacet(tx, nodeID, "frontend", mapField(roles, "frontend")); err != nil {
			return err
		}
		if err := syncRoleFacet(tx, nodeID, "backend", mapField(roles, "backend")); err != nil {
			return err
		}
	}

	for idx, edge := range edges {
		fromID := stringField(edge, "from")
		toID := stringField(edge, "to")
		if fromID == "" || toID == "" || fromID == toID || !nodeIDs[fromID] || !nodeIDs[toID] {
			return &graphSyncError{Code: graphInvalidCode, Message: "edge source and target must be active nodes in the same module"}
		}

		edgeID := stringField(edge, "id")
		if edgeID == "" {
			edgeID = "edge_" + GenerateUUID()
		}
		expectedVersion, hasVersion := rowVersion(edge)
		if err := upsertEdge(tx, moduleID, edgeID, idx, edge, expectedVersion, hasVersion); err != nil {
			return err
		}
	}
	for _, node := range nodes {
		if err := syncBusinessDetails(tx, moduleID, stringField(node, "id"), mapField(node, "doc")); err != nil {
			return err
		}
	}

	if err := tombstoneEdges(tx, moduleID, deletedEdges); err != nil {
		return err
	}
	if err := tombstoneNodes(tx, moduleID, deletedNodes); err != nil {
		return err
	}

	return nil
}
