package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stormfrontd/client/auth"
	"stormfrontd/client/communication"
	"stormfrontd/client/lightning"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var Client StormfrontClient
var Running = false
var JoinTokens []string
var AuthClient auth.ClientInformation
var SuccessionLock = false

const HEALTH_CHECK_DELAY = 30
const UPDATE_RETRY_DELAY = 1
const UPDATE_MAX_TRIES = 3

type StormfrontClient struct {
	Type       string           `json:"type"`
	Leader     StormfrontNode   `json:"leader"`
	Succession []StormfrontNode `json:"succession"`
	Updated    string           `json:"updated"`
	Host       string           `json:"host"`
	Port       int              `json:"port"`
	Healthy    bool             `json:"healthy"`
	Router     *gin.Engine      `json:"-"`
	Server     *http.Server     `json:"-"`
}

type StormfrontNode struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetHealth(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	if Client.Healthy {
		c.Status(http.StatusOK)
		return
	} else {
		c.Status(http.StatusServiceUnavailable)
	}
}

func GetState(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	c.JSON(http.StatusOK, Client)
}

func UpdateFollowerSuccession(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	var succession []StormfrontNode
	c.BindJSON(&succession)

	currentTime := time.Now()
	Client.Updated = currentTime.Format(time.RFC3339)

	Client.Succession = succession
}

func updateSuccession(successors []StormfrontNode) []StormfrontNode {
	AuthClient = auth.ReadClientInformation()

	postBody, _ := json.Marshal(Client.Succession)

	newSuccession := []StormfrontNode{}

	for _, successor := range successors {
		foundSuccessor := false
		for counter := 0; counter < UPDATE_MAX_TRIES; counter++ {
			fmt.Printf("Trying to reach follower at %s:%v/api/health, %v of %v\n", successor.Host, successor.Port, counter+1, UPDATE_MAX_TRIES)
			status, _, err := communication.Get(successor.Host, successor.Port, "api/health", AuthClient)
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
			foundSuccessor = true
			break
		}
		if foundSuccessor {
			newSuccession = append(newSuccession, successor)
		}
	}

	for _, successor := range newSuccession {
		status, _, err := communication.Post(successor.Host, successor.Port, "api/update/succession", AuthClient, postBody)
		if err != nil {
			fmt.Printf("Encountered error: %v\n", err.Error())
			continue
		}
		if status != http.StatusOK {
			fmt.Printf("Encountered non-200 status: %v\n", status)
			continue
		}
	}

	return newSuccession
}

func RegisterFollower(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	var follower StormfrontNode
	c.BindJSON(&follower)

	currentTime := time.Now()
	Client.Updated = currentTime.Format(time.RFC3339)

	for SuccessionLock {
		time.Sleep(10 * time.Millisecond)
	}
	SuccessionLock = true
	Client.Succession = append(Client.Succession, follower)
	Client.Succession = updateSuccession(Client.Succession)
	SuccessionLock = false

	c.Status(http.StatusOK)
}

func DeregisterFollower(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

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
	Client.Succession = updateSuccession(Client.Succession)
	SuccessionLock = false

	c.Status(http.StatusOK)
}

func GetJoinCommand(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	joinToken := auth.GenToken(128)
	JoinTokens = append(JoinTokens, joinToken)

	joinCommand := fmt.Sprintf("./stormfront daemon join -L %s:%v -j %s", Client.Leader.Host, Client.Leader.Port, joinToken)

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

func verifyAccessToken(token string) int {
	if Client.Type == "Follower" {
		status, _, err := communication.Get(Client.Leader.Host, Client.Leader.Port, "auth/check", AuthClient)
		if err != nil {
			fmt.Printf("Encountered error: %v\n", err.Error())
			return http.StatusInternalServerError
		}
		return status
	}
	tokenStatus := auth.VerifyToken(token)
	return tokenStatus
}

func CheckAccessToken(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
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
	fmt.Println("Health check loop started")
	for {
		for SuccessionLock {
			fmt.Println("Session is locked, waiting...")
			time.Sleep(10 * time.Millisecond)
		}
		fmt.Println("Session is unlocked!")
		SuccessionLock = true
		fmt.Println("Updating succession...")
		Client.Succession = updateSuccession(Client.Succession)
		SuccessionLock = false
		fmt.Println("Waiting until next health check")
		time.Sleep(HEALTH_CHECK_DELAY * time.Second)
	}
}

func GetBolt(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

	boltId := c.Param("id")

	bolt, err := lightning.GetBolt(boltId)

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, bolt)
}

func PostBolt(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		c.Status(http.StatusUnauthorized)
		return
	}
	token = splitToken[1]

	status := verifyAccessToken(token)
	if status != http.StatusOK {
		c.Status(status)
		return
	}

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

		node := StormfrontNode{ID: uuid.New().String(), Host: Client.Host, Port: Client.Port}

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

	return nil
}
