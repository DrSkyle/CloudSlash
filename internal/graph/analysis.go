package graph

// ImpactReport details what will be affected if a node is removed.
type ImpactReport struct {
	TargetNode      *Node
	DirectImpact    []*Node // Nodes directly depending on this
	CascadingImpact []*Node // Nodes strictly reachable only through this (if we did proper dominator analysis, for now just downstream)
	TotalRiskScore  int
}

// AnalyzeImpact performs a traversal to find everything that depends on the target node.
// It uses ReverseEdges (Who points to me?) to find dependencies.
func (g *Graph) AnalyzeImpact(nodeID string) *ImpactReport {
	g.Mu.RLock()
	defer g.Mu.RUnlock()

	targetNode, ok := g.Nodes[nodeID]
	if !ok {
		return nil
	}

	report := &ImpactReport{
		TargetNode: targetNode,
	}

	// 1. Direct Dependencies (Reverse Edges: Who needs me?)
	// e.g. EC2 needs EBS. If we delete EBS, EC2 breaks (or crashes).
	// Actually, usually it's "AttachedTo".
	// If EBS is attached to EC2:
	// EBS --AttachedTo--> EC2.
	// If we delete EBS, EC2 is affected (Disk check fails).
	//
	// Edge direction: Source -> Target.
	// "EBS AttachedTo EC2": Source=EBS, Target=EC2.
	// g.Edges[EBS] = [EC2].
	//
	// So "Downstream" (Forward Edges) are the things affected by this node's removal.
	// Wait, let's verify direction semantics.
	// Dependencies usually mean "A depends on B".
	// If A depends on B, then if B is removed, A is broken.
	// In DAGs, usually A -> B means A dependency B.
	// BUT in my `dag.go`, `EdgeTypeAttachedTo` (EBS -> EC2).
	// Does EBS depend on EC2? Or EC2 depend on EBS?
	// Physically, EBS volume is attached to Instance.
	// If Instance is deleted, Volume might be deleted (DeleteOnTermination).
	// If Volume is deleted, Instance might crash.
	//
	// Let's assume "Impact" = "What points to me?" vs "What do I point to?"
	// In the graph:
	// EBS -> EC2 (AttachedTo)
	// SecurityGroup -> EC2 (SecuredBy? No, EC2 -> SG usually).
	// Subnet -> VPC (Contains? No, VPC -> Subnet).
	//
	// Let's standardise on "Flow of Data/Control".
	// If A -> B, then A affects B.
	// So `g.Edges[nodeID]` provides the direct impact.

	directEdges := g.Edges[nodeID] // Targets
	for _, edge := range directEdges {
		if node, ok := g.Nodes[edge.TargetID]; ok {
			report.DirectImpact = append(report.DirectImpact, node)
			report.TotalRiskScore += node.RiskScore
		}
	}

	// 2. Cascading Impact (Recursive downstream)
	// BFS on forward edges
	visited := make(map[string]bool)
	queue := []string{}

	// Seed queue with direct children
	for _, child := range report.DirectImpact {
		visited[child.ID] = true
		queue = append(queue, child.ID)
	}

	// Mark target as visited so we don't loop back if cycle exists (shouldn't in DAG but safety first)
	visited[nodeID] = true

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		// Add to cascading (skipping direct ones which are already added? No, keep them separate or disjoint?)
		// Let's make CascadingImpact include indirect only?
		// Or all downstream?
		// Let's make it ALL downstream for simplicity of "Total Impact".
		if _, ok := g.Nodes[currentID]; ok {
			// Don't duplicate if it was already in direct impact?
			// Actually report.DirectImpact is a subset.
			// Let's just create a unique set of all impacted nodes.
			// Re-logic:
		}

		children := g.Edges[currentID]
		for _, childEdge := range children {
			if !visited[childEdge.TargetID] {
				visited[childEdge.TargetID] = true
				queue = append(queue, childEdge.TargetID)
				if node, ok := g.Nodes[childEdge.TargetID]; ok {
					report.CascadingImpact = append(report.CascadingImpact, node)
				}
			}
		}
	}

	return report
}
