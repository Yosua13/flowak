package module_test

import (
	"errors"
	"testing"

	"backend/internal/domain/module"
)

func TestDetectCycles_AcyclicGraphs(t *testing.T) {
	tests := []struct {
		name  string
		nodes []module.Node
		edges []module.Edge
	}{
		{
			name:  "empty graph",
			nodes: []module.Node{},
			edges: []module.Edge{},
		},
		{
			name: "linear chain",
			nodes: []module.Node{
				{ID: "node_1"}, {ID: "node_2"}, {ID: "node_3"},
			},
			edges: []module.Edge{
				{ID: "e1", From: "node_1", To: "node_2"},
				{ID: "e2", From: "node_2", To: "node_3"},
			},
		},
		{
			name: "diamond DAG",
			nodes: []module.Node{
				{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"},
			},
			edges: []module.Edge{
				{ID: "e1", From: "A", To: "B"},
				{ID: "e2", From: "A", To: "C"},
				{ID: "e3", From: "B", To: "D"},
				{ID: "e4", From: "C", To: "D"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := module.DetectCycles(tc.nodes, tc.edges)
			if err != nil {
				t.Fatalf("expected acyclic graph to pass, got: %v", err)
			}
		})
	}
}

func TestDetectCycles_CyclicGraphs(t *testing.T) {
	tests := []struct {
		name  string
		nodes []module.Node
		edges []module.Edge
	}{
		{
			name: "self-loop",
			nodes: []module.Node{
				{ID: "node_1"},
			},
			edges: []module.Edge{
				{ID: "e1", From: "node_1", To: "node_1"},
			},
		},
		{
			name: "2-node cycle",
			nodes: []module.Node{
				{ID: "A"}, {ID: "B"},
			},
			edges: []module.Edge{
				{ID: "e1", From: "A", To: "B"},
				{ID: "e2", From: "B", To: "A"},
			},
		},
		{
			name: "multi-node cycle in branch",
			nodes: []module.Node{
				{ID: "start"}, {ID: "A"}, {ID: "B"}, {ID: "C"},
			},
			edges: []module.Edge{
				{ID: "e1", From: "start", To: "A"},
				{ID: "e2", From: "A", To: "B"},
				{ID: "e3", From: "B", To: "C"},
				{ID: "e4", From: "C", To: "A"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := module.DetectCycles(tc.nodes, tc.edges)
			if !errors.Is(err, module.ErrCycleDetected) {
				t.Fatalf("expected ErrCycleDetected, got: %v", err)
			}
		})
	}
}
