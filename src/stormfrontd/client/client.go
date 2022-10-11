package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"stormfrontd/client/auth"
	"stormfrontd/client/communication"
	"stormfrontd/client/lightning"
	"stormfrontd/config"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
)

var Client StormfrontClient
var Running = false
var JoinTokens []string
var AuthClient auth.ClientInformation
var SuccessionLock = false
var Applications []StormfrontApplication = []StormfrontApplication{}

const HEALTH_CHECK_DELAY = 10
const UPDATE_RETRY_DELAY = 1
const UPDATE_MAX_TRIES = 3

type StormfrontClient struct {
	ID           string                  `json:"id"`
	Type         string                  `json:"type"`
	Leader       StormfrontNode          `json:"leader"`
	Succession   []StormfrontNode        `json:"succession"`
	Unhealthy    []StormfrontNode        `json:"unhealthy"`
	Unknown      []StormfrontNode        `json:"unknown"`
	Updated      string                  `json:"updated"`
	Host         string                  `json:"host"`
	Port         int                     `json:"port"`
	Healthy      bool                    `json:"healthy"`
	Router       *gin.Engine             `json:"-"`
	Server       *http.Server            `json:"-"`
	Applications []StormfrontApplication `json:"applications"`
	System       StormfrontSystemInfo    `json:"system"`
}

type StormfrontSystemInfo struct {
	MemoryUsage     float64 `json:"memory_usage"`
	MemoryAvailable int     `json:"memory_available"`
	CPUUsage        float64 `json:"cpu_usage"`
	CPUAvailable    float64 `json:"cpu_available"`
	Cores           int     `json:"cores"`
	TotalMemory     int     `json:"total_memory"`
	FreeMemory      int     `json:"free_memory"`
}

type StormfrontNode struct {
	ID     string               `json:"id"`
	Host   string               `json:"host"`
	Port   int                  `json:"port"`
	System StormfrontSystemInfo `json:"system"`
}

type StormfrontNodeType struct {
	ID     string `json:"id"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Type   string `json:"type"`
	Health string `json:"health"`
}

type StormfrontUpdatePackage struct {
	AuthClients  []auth.ClientInformation `json:"auth_clients"`
	APITokens    []string                 `json:"api_tokens"`
	Succession   []StormfrontNode         `json:"succession"`
	Applications []StormfrontApplication  `json:"applications"`
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetNodes(c *gin.Context) {
	token := c.Request.Header.Get("X-Stormfront-API")

	status := auth.VerifyAPIToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	nodes := []StormfrontNodeType{}

	nodes = append(nodes, StormfrontNodeType{
		ID:     Client.Leader.ID,
		Host:   Client.Leader.Host,
		Port:   Client.Leader.Port,
		Type:   "Leader",
		Health: "Healthy",
	})
	for _, node := range Client.Succession {
		nodes = append(nodes, StormfrontNodeType{
			ID:     node.ID,
			Host:   node.Host,
			Port:   node.Port,
			Type:   "Follower",
			Health: "Healthy",
		})
	}
	for _, node := range Client.Unhealthy {
		nodes = append(nodes, StormfrontNodeType{
			ID:     node.ID,
			Host:   node.Host,
			Port:   node.Port,
			Type:   "Follower",
			Health: "Unhealthy",
		})
	}
	for _, node := range Client.Unknown {
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
		c.Redirect(http.StatusFound, fmt.Sprintf("http://%s:%v/api/application", Client.Leader.Host, Client.Leader.Port))
		return
	}

	var app StormfrontApplication
	c.BindJSON(&app)
	app.ID = uuid.NewString()

	nodes := append(Client.Succession, StormfrontNode{
		ID:     Client.ID,
		Host:   Client.Host,
		Port:   Client.Port,
		System: Client.System,
	})

	for _, runningApp := range Applications {
		for exposedPort := range runningApp.Ports {
			for desiredPort := range app.Ports {
				if exposedPort == desiredPort {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Port already allocated"})
					return
				}
			}
		}
	}
	for _, node := range nodes {
		if node.System.CPUAvailable >= app.CPU && node.System.MemoryAvailable >= app.Memory {
			app.Node = node.ID
			Applications = append(Applications, app)
			c.Status(http.StatusCreated)
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Insufficient resources to schedule"})
}

func GetLocalApplications(c *gin.Context) {

	c.JSON(http.StatusOK, Client.Applications)
}

func GetAllApplications(c *gin.Context) {

	if Client.Type != "Leader" {
		c.Redirect(http.StatusFound, fmt.Sprintf("http://%s:%v/api/application", Client.Leader.Host, Client.Leader.Port))
		return
	}

	c.JSON(http.StatusOK, Applications)
}

func GetApplication(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusFound, fmt.Sprintf("http://%s:%v/api/application/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	for _, app := range Applications {
		if app.ID == id {
			c.JSON(http.StatusOK, app)
			return
		}
	}

	c.Status(http.StatusNotFound)
}

func DeleteApplication(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusFound, fmt.Sprintf("http://%s:%v/api/application/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	for idx, app := range Applications {
		if app.ID == id {

			Applications = append(Applications[:idx], Applications[idx+1:]...)

			c.Status(http.StatusOK)
			return
		}
	}

	c.Status(http.StatusNotFound)
}

func UpdateApplication(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusFound, fmt.Sprintf("http://%s:%v/api/application/%s", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	var applicationDefinition StormfrontApplication
	c.BindJSON(&applicationDefinition)

	for idx, app := range Applications {
		if app.ID == id {
			applicationDefinition.ID = app.ID
			if applicationDefinition.Node != app.Node {
				if applicationDefinition.Node == "" {
					applicationDefinition.Node = app.Node
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Node specification not allowed"})
				}
			}
			if applicationDefinition.Name != app.Name {
				if applicationDefinition.Name == "" {
					applicationDefinition.Name = app.Name
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Renaming not allowed in application update"})
				}
			}
			if applicationDefinition.Hostname != app.Hostname {
				if applicationDefinition.Hostname == "" {
					applicationDefinition.Hostname = app.Hostname
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Hostname change not allowed in application update"})
				}
			}
			if applicationDefinition.Memory != app.Memory {
				if applicationDefinition.Memory == 0 {
					applicationDefinition.Memory = app.Memory
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Memory spec change not allowed in application update"})
				}
			}
			if applicationDefinition.CPU != app.CPU {
				if applicationDefinition.CPU == 0 {
					applicationDefinition.CPU = app.CPU
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "CPU spec change not allowed in application update"})
				}
			}
			if applicationDefinition.Node != app.Node {
				if applicationDefinition.Node == "" {
					applicationDefinition.Node = app.Node
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Node specification not allowed"})
				}
			}

			if applicationDefinition.Env == nil {
				applicationDefinition.Env = app.Env
			}
			if applicationDefinition.Ports == nil {
				applicationDefinition.Ports = app.Ports
			}
			if applicationDefinition.Image == "" {
				applicationDefinition.Image = app.Image
			}

			Applications[idx] = applicationDefinition

			c.Status(http.StatusOK)
			return
		}
	}

	c.Status(http.StatusNotFound)
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

func updateSystemInfo() error {
	systemInfo := StormfrontSystemInfo{}

	cores, err := cpu.Counts(true)
	if err != nil {
		return err
	}
	usage, err := cpu.Percent(time.Second, false)
	if err != nil {
		return err
	}
	totalMemory := memory.TotalMemory()
	freeMemory := memory.FreeMemory()

	systemInfo.Cores = cores
	systemInfo.CPUUsage = usage[0]
	systemInfo.TotalMemory = int(totalMemory)
	systemInfo.FreeMemory = int(freeMemory)
	if systemInfo.TotalMemory != 0 {
		systemInfo.MemoryUsage = float64(freeMemory) / float64(totalMemory)
	} else {
		systemInfo.MemoryUsage = -1
	}

	memoryReserved := 0
	cpuReserved := 0.0
	for _, application := range Client.Applications {
		if application.Node == Client.ID {
			memoryReserved += application.Memory
			cpuReserved += application.CPU
		}
	}

	memoryUsed := (float64(systemInfo.TotalMemory) * config.Config.ReservedMemoryPercentage) + float64(memoryReserved)
	systemInfo.MemoryAvailable = systemInfo.TotalMemory - int(memoryUsed)

	cpuUsed := (float64(systemInfo.Cores) * config.Config.ReservedCPUPercentage) + cpuReserved
	systemInfo.CPUAvailable = float64(systemInfo.Cores) - cpuUsed

	Client.System = systemInfo

	return nil
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

func GetApplicationLogs(c *gin.Context) {
	id := c.Param("id")

	if Client.Type != "Leader" {
		c.Redirect(http.StatusFound, fmt.Sprintf("http://%s:%v/api/application/%s/logs", Client.Leader.Host, Client.Leader.Port, id))
		return
	}

	for _, app := range Applications {
		if app.ID == id {
			cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("docker logs %s", app.Name))
			var outb bytes.Buffer
			cmd.Stdout = &outb
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Encountered error getting container logs: %v\n", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			logs := outb.String()
			c.JSON(http.StatusOK, gin.H{"logs": logs})
			return
		}
	}

	c.Status(http.StatusNotFound)
}

func UpdateState(c *gin.Context) {
	var updatePackage StormfrontUpdatePackage

	c.BindJSON(&updatePackage)

	currentTime := time.Now()
	Client.Updated = currentTime.Format(time.RFC3339)

	reconcileApplications(updatePackage)

	Client.Succession = updatePackage.Succession
	auth.APITokens = updatePackage.APITokens
	auth.AuthClients = updatePackage.AuthClients
	Client.Applications = append([]StormfrontApplication(nil), updatePackage.Applications...)
	Client.Applications = getApplicationStatus(Client.Applications)
}

func updateFollowers(successors []StormfrontNode) ([]StormfrontNode, []StormfrontNode, []StormfrontNode) {
	AuthClient = auth.ReadClientInformation()

	localApplications := []StormfrontApplication{}

	updatePackage := StormfrontUpdatePackage{
		Succession:   Client.Succession,
		APITokens:    auth.APITokens,
		AuthClients:  auth.AuthClients,
		Applications: Applications,
	}

	postBody, _ := json.Marshal(updatePackage)

	newSuccession := []StormfrontNode{}
	newUnhealthy := []StormfrontNode{}
	newUnknown := []StormfrontNode{}

	for _, successor := range successors {
		foundSuccessor := false
		for counter := 0; counter < UPDATE_MAX_TRIES; counter++ {
			fmt.Printf("Trying to reach follower at %s:%v/api/health, %v of %v\n", successor.Host, successor.Port, counter+1, UPDATE_MAX_TRIES)
			status, body, err := communication.Get(successor.Host, successor.Port, "api/health", AuthClient)
			if err != nil {
				fmt.Printf("Encountered error: %v\n", err.Error())
				time.Sleep(UPDATE_RETRY_DELAY * time.Second)
				continue
			}
			if status != http.StatusOK {
				fmt.Printf("Encountered error: %v\n", err.Error())
				time.Sleep(UPDATE_RETRY_DELAY * time.Second)
				continue
			}
			json.Unmarshal([]byte(body), &successor.System)
			foundSuccessor = true
			break
		}
		if foundSuccessor {
			newSuccession = append(newSuccession, successor)
		} else {
			newUnknown = append(newUnknown, successor)
		}
	}

	for _, successor := range newSuccession {
		status, _, err := communication.Post(successor.Host, successor.Port, "api/state", AuthClient, postBody)
		if err != nil {
			fmt.Printf("Encountered error: %v\n", err.Error())
			newUnhealthy = append(newUnhealthy, successor)
			continue
		} else {
			if status != http.StatusOK {
				fmt.Printf("Encountered non-200 status: %v\n", status)
				newUnhealthy = append(newUnhealthy, successor)
				continue
			}
		}
		status, responseBody, err := communication.Get(successor.Host, successor.Port, "api/application/local", AuthClient)
		if err != nil {
			fmt.Printf("Encountered error: %v\n", err.Error())
			newUnhealthy = append(newUnhealthy, successor)
			continue
		} else {
			if status != http.StatusOK {
				fmt.Printf("Encountered non-200 status: %v\n", status)
				newUnhealthy = append(newUnhealthy, successor)
				continue
			}
		}
		successorApplications := []StormfrontApplication{}
		json.Unmarshal([]byte(responseBody), &successorApplications)
		localApplications = append(localApplications, successorApplications...)
	}

	reconcileApplications(updatePackage)

	for _, localApp := range localApplications {
		for idx, app := range Applications {
			fmt.Println("Updating status!")
			if app.ID == localApp.ID {
				app.Status = localApp.Status
				Applications[idx] = app
			}
		}
	}

	Applications = getApplicationStatus(Applications)
	Client.Applications = append([]StormfrontApplication(nil), Applications...)

	newSuccession = dedupeNodes(newSuccession)
	newUnhealthy = dedupeNodes(newUnhealthy)
	newUnknown = dedupeNodes(newUnknown)

	return newSuccession, newUnhealthy, newUnknown
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

func RegisterFollower(c *gin.Context) {
	var follower StormfrontNode
	c.BindJSON(&follower)

	currentTime := time.Now()
	Client.Updated = currentTime.Format(time.RFC3339)

	for SuccessionLock {
		time.Sleep(10 * time.Millisecond)
	}
	SuccessionLock = true
	Client.Succession = append(Client.Succession, follower)
	Client.Succession, Client.Unhealthy, Client.Unknown = updateFollowers(append(Client.Succession, append(Client.Unknown, Client.Unhealthy...)...))
	SuccessionLock = false

	c.Status(http.StatusOK)
}

func DeregisterFollower(c *gin.Context) {
	var follower StormfrontNode
	c.BindJSON(&follower)

	currentTime := time.Now()
	Client.Updated = currentTime.Format(time.RFC3339)

	removeIdx := -1

	for idx, node := range Client.Succession {
		if node.Host == follower.Host && node.Port == follower.Port {
			removeIdx = idx
			break
		}
	}

	for SuccessionLock {
		time.Sleep(10 * time.Millisecond)
	}
	SuccessionLock = true
	if removeIdx != -1 {
		Client.Succession = append(Client.Succession[:removeIdx], Client.Succession[removeIdx+1:]...)
	}
	Client.Succession, Client.Unhealthy, Client.Unknown = updateFollowers(append(Client.Succession, append(Client.Unknown, Client.Unhealthy...)...))
	SuccessionLock = false

	c.Status(http.StatusOK)
}

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

func HealthCheck() {
	for {
		for SuccessionLock {
			time.Sleep(10 * time.Millisecond)
		}
		SuccessionLock = true
		Client.Succession, Client.Unhealthy, Client.Unknown = updateFollowers(append(Client.Succession, append(Client.Unknown, Client.Unhealthy...)...))
		SuccessionLock = false
		time.Sleep(HEALTH_CHECK_DELAY * time.Second)
	}
}

func GetBolt(c *gin.Context) {
	boltId := c.Param("id")

	bolt, err := lightning.GetBolt(boltId)

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, bolt)
}

func PostBolt(c *gin.Context) {
	var boltConstructor lightning.BoltConstructor
	c.BindJSON(&boltConstructor)

	bolt, idx := lightning.CreateBolt(boltConstructor.Command)

	go lightning.RunBolt(&lightning.Bolts[idx])

	c.JSON(http.StatusOK, bolt)
}

func Initialize(joinToken string) error {
	Client.Router = gin.Default()

	InitializeRoutes(Client.Type)

	Client.Server = &http.Server{
		Addr:    ":" + strconv.Itoa(Client.Port),
		Handler: Client.Router,
	}

	Client.ID = uuid.New().String()

	Running = true

	// Start serving the application
	go func() {
		if err := Client.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	if Client.Type == "Follower" {
		AuthClient = auth.ClientInformation{}
		AuthClient.AccessToken = joinToken

		status, body, err := communication.Get(Client.Leader.Host, Client.Leader.Port, "auth/token", AuthClient)
		if err != nil {
			fmt.Printf("Encountered error: %v\n", err.Error())
			return err
		}
		if status != http.StatusOK {
			return fmt.Errorf("unable to contact client at %s:%v, received status code %v", Client.Leader.Host, Client.Leader.Port, status)
		}

		json.Unmarshal([]byte(body), &AuthClient)

		auth.WriteClientInformation(AuthClient)

		node := StormfrontNode{ID: Client.ID, Host: Client.Host, Port: Client.Port}

		postBody, _ := json.Marshal(node)

		status, _, err = communication.Post(Client.Leader.Host, Client.Leader.Port, "api/register", AuthClient, postBody)
		if err != nil {
			fmt.Printf("Encountered error: %v\n", err.Error())
			return err
		}
		if status != http.StatusOK {
			return fmt.Errorf("unable to contact client at %s:%v, received status code %v", Client.Leader.Host, Client.Leader.Port, status)
		}
	} else {
		AuthClient = auth.CreateClientInformation()
		auth.WriteClientInformation(AuthClient)

		// Check follower healths
		go HealthCheck()
	}

	err := updateSystemInfo()
	if err != nil {
		panic(err)
	}

	return nil
}
