package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"stormfrontd/client/auth"
	"stormfrontd/config"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jfcarter2358/ceresdb-go/connection"
)

func GetJoinCommand(c *gin.Context) {
	joinToken := auth.GenToken(128)
	JoinTokens = append(JoinTokens, joinToken)

	joinCommand := fmt.Sprintf("stormfront join -L %s:%v -j %s", Client.Leader.Host, Client.Leader.Port, joinToken)

	c.JSON(http.StatusOK, gin.H{"join_command": joinCommand})
}

func GetAccessToken(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	if contains(JoinTokens, token) {

		removeIdx := -1
		for idx, joinToken := range JoinTokens {
			if joinToken == token {
				removeIdx = idx
			}
		}
		JoinTokens = append(JoinTokens[:removeIdx], JoinTokens[removeIdx+1:]...)
		clientInfo := auth.CreateClientInformation()

		authData, _ := json.Marshal(clientInfo)
		_, err := connection.Query(fmt.Sprintf("post record stormfront.auth %s", authData))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, clientInfo)
		return
	}

	c.Status(http.StatusUnauthorized)
}

func CheckAccessToken(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := auth.VerifyAccessToken(token)
	c.Status(status)
}

func RefreshAccessToken(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	newClient, err := auth.RefreshClient(token)

	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, newClient)
}

func GetHealth(c *gin.Context) {
	if Client.Healthy {
		err := updateSystemInfo()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, Client.System)
		return
	} else {
		c.Status(http.StatusServiceUnavailable)
	}
}

func GetState(c *gin.Context) {
	c.JSON(http.StatusOK, Client)
}

func GetNodes(c *gin.Context) {
	token := c.Request.Header.Get("X-Stormfront-API")

	status := auth.VerifyAPIToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	nodeData, err := connection.Query("get record stormfront.node")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
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

	c.JSON(http.StatusOK, nodes)
}

func CreateApplication(c *gin.Context) {

	if Client.Type != "Leader" {
		fmt.Printf("Node is not leader, redirecting to http://%s:%v/api/application\n", Client.Leader.Host, Client.Leader.Port)
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application", Client.Leader.Host, Client.Leader.Port))
		return
	}

	var app StormfrontApplication
	c.BindJSON(&app)
	app.ID = uuid.NewString()

	nodeData, err := connection.Query("get record stormfront.node")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	var succession []StormfrontNode
	successionBytes, _ := json.Marshal(nodeData[0]["succession"])
	json.Unmarshal(successionBytes, &succession)

	nodes := append(succession, StormfrontNode{
		ID:     Client.ID,
		Host:   Client.Host,
		Port:   Client.Port,
		System: Client.System,
	})

	applicationData, err := connection.Query("get record stormfront.application")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	var applications []StormfrontApplication
	applicationBytes, _ := json.Marshal(applicationData)
	json.Unmarshal(applicationBytes, &applications)

	for _, runningApp := range applications {
		for exposedPort := range runningApp.Ports {
			for desiredPort := range app.Ports {
				if exposedPort == desiredPort {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Port already allocated"})
					return
				}
			}
		}
	}
	if app.Node != "" {
		for _, node := range nodes {
			if node.ID != app.Node {
				continue
			}
			if node.System.CPUAvailable >= app.CPU && node.System.MemoryAvailable >= app.Memory {
				app.Node = node.ID
				appBytes, _ := json.Marshal(app)
				_, err := connection.Query(fmt.Sprintf("post record stormfront.application %s", string(appBytes)))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create application: %v", err.Error())})
					return
				}
				c.Status(http.StatusCreated)
				return
			}
		}
	} else {
		for _, node := range nodes {
			fmt.Printf("Available CPU: %v, requested CPU: %v, available memory: %v, requested memory: %v", node.System.CPUAvailable, app.CPU, node.System.MemoryAvailable, app.Memory)
			if node.System.CPUAvailable >= app.CPU && node.System.MemoryAvailable >= app.Memory {
				app.Node = node.ID
				appBytes, _ := json.Marshal(app)
				_, err := connection.Query(fmt.Sprintf("post record stormfront.application %s", string(appBytes)))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create application: %v", err.Error())})
					return
				}
				c.Status(http.StatusCreated)
				return
			}
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient resources to schedule"})
}

func GetAllApplications(c *gin.Context) {

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application", Client.Leader.Host, Client.Leader.Port))
		return
	}

	applicationData, err := connection.Query("get record stormfront.application")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	var applications []StormfrontApplication
	applicationBytes, _ := json.Marshal(applicationData)
	json.Unmarshal(applicationBytes, &applications)

	c.JSON(http.StatusOK, applications)
}

func GetApplication(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	data, err := connection.Query(fmt.Sprintf(`get record stormfront.application | filter id = '%s'`, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if len(data) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, data[0])
}

func DeleteApplication(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	_, err := connection.Query(fmt.Sprintf(`get record stormfront.application .id | filter id = '%s' | delete record stormfront.application -`, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.Status(http.StatusNoContent)
}

// TODO: Re-implement the UpdateApplication logic
// func UpdateApplication(c *gin.Context) {
// 	id := c.Param("id")

// 	if Client.Type != "Leader" {
// 		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application/%s", Client.Leader.Host, Client.Leader.Port, id))
// 		return
// 	}

// 	var applicationDefinition StormfrontApplication
// 	c.BindJSON(&applicationDefinition)

// 	for idx, app := range Applications {
// 		if app.ID == id {
// 			applicationDefinition.ID = app.ID
// 			if applicationDefinition.Node != app.Node {
// 				if applicationDefinition.Node == "" {
// 					applicationDefinition.Node = app.Node
// 				} else {
// 					c.JSON(http.StatusInternalServerError, gin.H{"error": "Node specification not allowed"})
// 				}
// 			}
// 			if applicationDefinition.Name != app.Name {
// 				if applicationDefinition.Name == "" {
// 					applicationDefinition.Name = app.Name
// 				} else {
// 					c.JSON(http.StatusInternalServerError, gin.H{"error": "Renaming not allowed in application update"})
// 				}
// 			}
// 			if applicationDefinition.Hostname != app.Hostname {
// 				if applicationDefinition.Hostname == "" {
// 					applicationDefinition.Hostname = app.Hostname
// 				} else {
// 					c.JSON(http.StatusInternalServerError, gin.H{"error": "Hostname change not allowed in application update"})
// 				}
// 			}
// 			if applicationDefinition.Memory != app.Memory {
// 				if applicationDefinition.Memory == 0 {
// 					applicationDefinition.Memory = app.Memory
// 				} else {
// 					c.JSON(http.StatusInternalServerError, gin.H{"error": "Memory spec change not allowed in application update"})
// 				}
// 			}
// 			if applicationDefinition.CPU != app.CPU {
// 				if applicationDefinition.CPU == 0 {
// 					applicationDefinition.CPU = app.CPU
// 				} else {
// 					c.JSON(http.StatusInternalServerError, gin.H{"error": "CPU spec change not allowed in application update"})
// 				}
// 			}
// 			if applicationDefinition.Node != app.Node {
// 				if applicationDefinition.Node == "" {
// 					applicationDefinition.Node = app.Node
// 				} else {
// 					c.JSON(http.StatusInternalServerError, gin.H{"error": "Node specification not allowed"})
// 				}
// 			}

// 			if applicationDefinition.Env == nil {
// 				applicationDefinition.Env = app.Env
// 			}
// 			if applicationDefinition.Ports == nil {
// 				applicationDefinition.Ports = app.Ports
// 			}
// 			if applicationDefinition.Image == "" {
// 				applicationDefinition.Image = app.Image
// 			}

// 			Applications[idx] = applicationDefinition

// 			c.Status(http.StatusOK)
// 			return
// 		}
// 	}

// 	c.Status(http.StatusNotFound)
// }

func GetApplicationLogs(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application/%s/logs", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	data, err := connection.Query(fmt.Sprintf(`get record stormfront.application | filter id = '%s'`, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if len(data) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	app := data[0]

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s logs %s", config.Config.ContainerEngine, app["name"].(string)))
	var outb bytes.Buffer
	cmd.Stdout = &outb
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Encountered error getting container logs: %v\n", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logs := outb.String()
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func GetAPIToken(c *gin.Context) {
	apiToken := auth.GenToken(128)

	auth.APITokens = append(auth.APITokens, apiToken)

	c.JSON(http.StatusOK, gin.H{"token": apiToken})
}

func RevokeAPIToken(c *gin.Context) {
	apiToken := c.Request.Header.Get("X-Stormfront-API")

	for idx, token := range auth.APITokens {
		if token == apiToken {
			auth.APITokens = append(auth.APITokens[:idx], auth.APITokens[idx+1:]...)
			break
		}
	}

	c.Status(http.StatusOK)
}
