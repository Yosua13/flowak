package module

// DetectCycles detects whether a directed cycle exists in the given nodes and edges graph.
// It returns ErrCycleDetected if a loop is detected, and nil if the graph is acyclic.
func DetectCycles(nodes []Node, edges []Edge) error {
	if len(edges) == 0 {
		return nil
	}

	// Adjacency list
	adj := make(map[string][]string)
	nodeSet := make(map[string]struct{})

	for _, n := range nodes {
		nodeSet[n.ID] = struct{}{}
	}

	for _, e := range edges {
		if e.From == "" || e.To == "" {
			continue
		}
		// Direct self-loop
		if e.From == e.To {
			return ErrCycleDetected
		}
		adj[e.From] = append(adj[e.From], e.To)
		nodeSet[e.From] = struct{}{}
		nodeSet[e.To] = struct{}{}
	}

	// 0: unvisited, 1: visiting (in stack), 2: visited
	state := make(map[string]int)

	var hasCycle func(curr string) bool
	hasCycle = func(curr string) bool {
		state[curr] = 1 // Mark visiting

		for _, neighbor := range adj[curr] {
			neighborState := state[neighbor]
			if neighborState == 1 {
				// Back-edge found -> cycle!
				return true
			}
			if neighborState == 0 {
				if hasCycle(neighbor) {
					return true
				}
			}
		}

		state[curr] = 2 // Mark visited
		return false
	}

	for nodeID := range nodeSet {
		if state[nodeID] == 0 {
			if hasCycle(nodeID) {
				return ErrCycleDetected
			}
		}
	}

	return nil
}
