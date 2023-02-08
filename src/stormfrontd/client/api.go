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

func GetJoinToken(c *gin.Context) {
	joinToken := auth.GenToken(128)
	JoinTokens = append(JoinTokens, joinToken)

	c.JSON(http.StatusOK, gin.H{"token": joinToken})
}

func RevokeJoinToken(c *gin.Context) {
	token := c.Param("token")

	found := false
	for idx, joinToken := range JoinTokens {
		if joinToken == token {
			JoinTokens = append(JoinTokens[:idx], JoinTokens[idx+1:]...)
			found = true
		}
	}

	if found {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusNotFound)
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

func GetAllNodes(c *gin.Context) {

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/node", Client.Leader.Host, Client.Leader.Port))
		return
	}

	nodes, err := getNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, nodes)
}

func GetNode(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/node/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	data, err := connection.Query(fmt.Sprintf(`get record stormfront.node | filter id = '%s'`, id))
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var nodes []StormfrontNode
	nodeBytes, _ := json.Marshal(nodeData)
	json.Unmarshal(nodeBytes, &nodes)

	applicationData, err := connection.Query("get record stormfront.application")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
			fmt.Printf("Available CPU: %v, requested CPU: %v, available memory: %v, requested memory: %v\n", node.System.CPUAvailable, app.CPU, node.System.MemoryAvailable, app.Memory)
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
			fmt.Printf("Insufficient resources to schedule on node %s\n", node.ID)
		}
	} else {
		for _, node := range nodes {
			fmt.Printf("Available CPU: %v, requested CPU: %v, available memory: %v, requested memory: %v\n", node.System.CPUAvailable, app.CPU, node.System.MemoryAvailable, app.Memory)
			if node.System.CPUAvailable >= app.CPU && node.System.MemoryAvailable >= app.Memory {
				app.Node = node.ID
				appBytes, _ := json.Marshal(app)
				_, err := connection.Query(fmt.Sprintf("post record stormfront.application %s", string(appBytes)))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create application: %v", err.Error())})
					return
				}
				c.JSON(http.StatusCreated, gin.H{"id": app.ID})
				return
			}
			fmt.Printf("Insufficient resources to schedule on node %s: Available CPU: %v, requested CPU: %v, available memory: %v, requested memory: %v\n", node.ID, node.System.CPUAvailable, app.CPU, node.System.MemoryAvailable, app.Memory)
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient resources to schedule"})
}

func GetAllApplications(c *gin.Context) {

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application", Client.Leader.Host, Client.Leader.Port))
		return
	}

	applications, err := getApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

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

func RestartApplication(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application/%s/restart", Client.Leader.Host, Client.Leader.Port, id))
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

	var app StormfrontApplication

	appBytes, _ := json.Marshal(data[0])

	json.Unmarshal(appBytes, &app)

	destroyApplication(app.Name, false)
	deployApplication(app, false, false)
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

	data, err := connection.Query(fmt.Sprintf(`get record stormfront.application | filter id = '%s'`, id))
	if err != nil {
		fmt.Println("Error 1")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if len(data) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	app := data[0]
	nodeId := app["node"].(string)

	if nodeId != Client.ID {
		if Client.Type != "Leader" {
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/application/%s/logs", Client.Leader.Host, Client.Leader.Port, id))
			return
		}
		nodeToQuery, err := getHostFromNode(nodeId)
		if err != nil {
			fmt.Println("Error 2")
			fmt.Printf("Encountered error getting container logs: %v\n", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s/api/application/%s/logs", nodeToQuery, id))
		return
	}

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s logs %s", config.Config.ContainerEngine, app["name"].(string)))
	var outb bytes.Buffer
	cmd.Stdout = &outb
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error 3")
		fmt.Println(outb.String())
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

func GetClient(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/client/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	data, err := connection.Query(fmt.Sprintf(`get record stormfront.client | filter id = '%s'`, id))
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

func GetAllClients(c *gin.Context) {
	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/client", Client.Leader.Host, Client.Leader.Port))
		return
	}

	clients, err := getClients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, clients)
}

func CreateRoute(c *gin.Context) {

	if Client.Type != "Leader" {
		fmt.Printf("Node is not leader, redirecting to http://%s:%v/api/route\n", Client.Leader.Host, Client.Leader.Port)
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/route", Client.Leader.Host, Client.Leader.Port))
		return
	}

	var route StormfrontRoute
	c.BindJSON(&route)
	route.ID = uuid.NewString()

	routeBytes, _ := json.Marshal(route)
	_, err := connection.Query(fmt.Sprintf("post record stormfront.route %s", string(routeBytes)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to create route: %v", err.Error())})
		return
	}
	c.Status(http.StatusCreated)
}

func GetRoute(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/route/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	data, err := connection.Query(fmt.Sprintf(`get record stormfront.route | filter id = '%s'`, id))
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

func GetAllRoutes(c *gin.Context) {
	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/route", Client.Leader.Host, Client.Leader.Port))
		return
	}

	data, err := connection.Query(`get record stormfront.route`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, data)
}

func DeleteRoute(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s:%v/api/route/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	_, err := connection.Query(fmt.Sprintf(`get record stormfront.route | filter id = '%s' | delete record stormfront.route -`, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.Status(http.StatusNoContent)

}
