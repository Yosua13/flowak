/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { Module, Node, Edge, ID } from './types';

/**
 * Generates a short random unique ID
 */
export function uid(): string {
  return Math.random().toString(36).substring(2, 9);
}

/**
 * Checks if a node ID is unique inside the module
 */
export function assertUniqueNodeId(module: Module, id: ID): boolean {
  return !module.nodes.some((n) => n.id === id);
}

/**
 * Validates if an edge can be added based on invariants:
 * INV-1: from/to must refer to valid node IDs in the module
 * INV-2: no duplicate edges between same from/to
 * INV-3: no self-loops
 */
export function canAddEdge(module: Module, from: ID, to: ID): boolean {
  if (from === to) return false; // INV-3: No self loop

  const fromExists = module.nodes.some((n) => n.id === from);
  const toExists = module.nodes.some((n) => n.id === to);
  if (!fromExists || !toExists) return false; // INV-1: Node must exist

  const edgeExists = module.edges.some((e) => e.from === from && e.to === to);
  if (edgeExists) return false; // INV-2: No duplicates

  return true;
}

/**
 * Immutably adds a new edge to the module if valid
 */
export function addEdge(module: Module, from: ID, to: ID, label?: string): Module {
  if (!canAddEdge(module, from, to)) {
    return module;
  }

  const newEdge: Edge = {
    id: `edge_${uid()}`,
    from,
    to,
    label,
  };

  return {
    ...module,
    edges: [...module.edges, newEdge],
  };
}

/**
 * Immutably deletes a node and cascades to delete touching edges
 * INV-4: delete node => cascade delete connected edges
 */
export function deleteNode(module: Module, nodeId: ID): Module {
  const filteredNodes = module.nodes.filter((n) => n.id !== nodeId);
  const filteredEdges = module.edges.filter((e) => e.from !== nodeId && e.to !== nodeId);

  return {
    ...module,
    nodes: filteredNodes,
    edges: filteredEdges,
  };
}
