package client

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func dedupeNodes(nodes []StormfrontNode) []StormfrontNode {
	out := []StormfrontNode{}

	for _, node := range nodes {
		shouldAdd := true
		for _, uniqueNode := range out {
			if node.ID == uniqueNode.ID {
				shouldAdd = false
				break
			}
		}
		if shouldAdd {
			out = append(out, node)
		}
	}

	return out
}
