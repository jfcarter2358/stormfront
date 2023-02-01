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

	var nodes []StormfrontNode
	nodeBytes, _ := json.Marshal(nodeData)
	json.Unmarshal(nodeBytes, &nodes)

	for _, node := range nodes {
		if nodeId == node.ID {
			return fmt.Sprintf("%s:%d", node.Host, node.Port), nil
		}
	}

	return "", fmt.Errorf("no healthy node with id %s exists", nodeId)

}

func getNodes() ([]StormfrontNode, error) {
	nodeData, err := connection.Query("get record stormfront.node")
	if err != nil {
		return nil, err
	}
	nodes := []StormfrontNode{}
	nodeBytes, _ := json.Marshal(nodeData)
	json.Unmarshal(nodeBytes, &nodes)

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

func getClients() ([]StormfrontClient, error) {
	clientData, err := connection.Query("get record stormfront.client")
	if err != nil {
		return nil, err
	}
	var clients []StormfrontClient
	clientBytes, _ := json.Marshal(clientData)
	json.Unmarshal(clientBytes, &clients)

	return clients, nil
}

func lookupFunc(domain string) (string, error) {
	fmt.Printf("DNS request received for %s", domain)
	parts := strings.Split(domain, ".")

	length := len(parts)
	namespace := parts[length-2]
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
		if app.Hostname == hostname && app.Namespace == namespace {
			scheduledNode := app.Node
			for _, node := range nodes {
				if node.ID == scheduledNode {
					if node.Health != "Healthy" {
						return "", fmt.Errorf("node %s is unhealthy", node.ID)
					}
					fmt.Printf("Routing traffic from %s to %s", domain, node.Host)
					return node.Host, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no node is currently serving hostname '%s' with namespace '%s'", hostname, namespace)
}
