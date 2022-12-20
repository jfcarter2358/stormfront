package client

import (
	"encoding/json"
	"fmt"
	"strings"

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

func getNodes() ([]StormfrontNodeType, error) {
	nodeData, err := connection.Query("get record stormfront.node")
	if err != nil {
		return nil, err
	}
	var succession []StormfrontNode
	var unhealthy []StormfrontNode
	var unknown []StormfrontNode

	successionBytes, _ := json.Marshal(nodeData[0]["succession"])
	unhealthyBytes, _ := json.Marshal(nodeData[0]["unhealthy"])
	unknownBytes, _ := json.Marshal(nodeData[0]["unknown"])

	json.Unmarshal(successionBytes, &succession)
	json.Unmarshal(unhealthyBytes, &unhealthy)
	json.Unmarshal(unknownBytes, &unknown)

	nodes := []StormfrontNodeType{}

	nodes = append(nodes, StormfrontNodeType{
		ID:     Client.Leader.ID,
		Host:   Client.Leader.Host,
		Port:   Client.Leader.Port,
		Type:   "Leader",
		Health: "Healthy",
	})
	for _, node := range succession {
		nodes = append(nodes, StormfrontNodeType{
			ID:     node.ID,
			Host:   node.Host,
			Port:   node.Port,
			Type:   "Follower",
			Health: "Healthy",
		})
	}
	for _, node := range unhealthy {
		nodes = append(nodes, StormfrontNodeType{
			ID:     node.ID,
			Host:   node.Host,
			Port:   node.Port,
			Type:   "Follower",
			Health: "Unhealthy",
		})
	}
	for _, node := range unknown {
		nodes = append(nodes, StormfrontNodeType{
			ID:     node.ID,
			Host:   node.Host,
			Port:   node.Port,
			Type:   "Follower",
			Health: "Unknown",
		})
	}

	return nodes, nil
}

func getApplications() ([]StormfrontApplication, error) {
	applicationData, err := connection.Query("get record stormfront.application")
	if err != nil {
		return nil, err
	}
	var applications []StormfrontApplication
	applicationBytes, _ := json.Marshal(applicationData)
	json.Unmarshal(applicationBytes, &applications)

	return applications, nil
}

func lookupFunc(domain string) (string, error) {
	parts := strings.Split(domain, ".")
	fmt.Printf("Split records: %v", parts)

	length := len(parts)
	hostname := parts[length-3]

	nodes, err := getNodes()
	if err != nil {
		return "", err
	}
	applications, err := getApplications()
	if err != nil {
		return "", err
	}

	for _, app := range applications {
		if app.Hostname == hostname {
			scheduledNode := app.Node
			for _, node := range nodes {
				if node.ID == scheduledNode {
					if node.Health != "Healthy" {
						return "", fmt.Errorf("node %s is unhealthy", node.ID)
					}
					return node.Host, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no node is currently serving hostname '%s'", hostname)
}
