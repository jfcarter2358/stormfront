package client

import (
	"encoding/json"
	"fmt"

	"github.com/jfcarter2358/ceresdb-go/connection"
)

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

func getHostFromNode(nodeId string) (string, error) {
	nodeData, err := connection.Query("get record stormfront.node")
	if err != nil {
		return "", err
	}

	var succession []StormfrontNode
	successionBytes, _ := json.Marshal(nodeData[0]["succession"])
	json.Unmarshal(successionBytes, &succession)

	if nodeId == Client.Leader.ID {
		return fmt.Sprintf("%s:%d", Client.Leader.Host, Client.Leader.Port), nil
	}

	for _, node := range succession {
		if nodeId == node.ID {
			return fmt.Sprintf("%s:%d", node.Host, node.Port), nil
		}
	}

	return "", fmt.Errorf("no healthy node with id %s exists", nodeId)

}
