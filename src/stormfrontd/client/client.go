package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stormfrontd/client/auth"
	"stormfrontd/client/communication"
	"stormfrontd/client/dns"
	"stormfrontd/config"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jfcarter2358/ceresdb-go/connection"
)

var Client StormfrontClient
var Running = false
var JoinTokens []string
var AuthClient auth.ClientInformation

const HEALTH_CHECK_DELAY = 10
const UPDATE_RETRY_DELAY = 1
const UPDATE_MAX_TRIES = 3

type StormfrontClient struct {
	ID           string                  `json:"id" yaml:"id"`
	Type         string                  `json:"type" yaml:"type"`
	Leader       StormfrontNode          `json:"leader" yaml:"leader"`
	Succession   []StormfrontNode        `json:"succession" yaml:"succession"`
	Unhealthy    []StormfrontNode        `json:"unhealthy" yaml:"unhealthy"`
	Unknown      []StormfrontNode        `json:"unknown" yaml:"unknown"`
	Updated      string                  `json:"updated" yaml:"updated"`
	Host         string                  `json:"host" yaml:"host"`
	Port         int                     `json:"port" yaml:"port"`
	Healthy      bool                    `json:"healthy" yaml:"healthy"`
	Router       *gin.Engine             `json:"-" yaml:"-"`
	Server       *http.Server            `json:"-" yaml:"-"`
	Applications []StormfrontApplication `json:"applications" yaml:"applications"`
	System       StormfrontSystemInfo    `json:"system" yaml:"system"`
}

func Initialize(joinToken string) error {
	Client.Router = gin.Default()

	InitializeRoutes(Client.Type)

	// Initialize DNS server
	server := dns.NewDNSServer(53, Client.Host)
	server.AddZoneData("stormfront", nil, lookupFunc, dns.DNSForwardLookupZone)
	go server.StartAndServe()

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

	if Client.Type == "Leader" {
		err := InitializeLeader()
		if err != nil {
			panic(err)
		}
	} else {
		err := InitializeFollower(joinToken)
		if err != nil {
			panic(err)
		}
	}

	err := updateSystemInfo()
	if err != nil {
		panic(err)
	}

	return nil
}

func InitializeFollower(joinToken string) error {
	if config.Config.CeresDBHost == "" {
		config.Config.CeresDBHost = Client.Host
	}
	connection.Initialize(CERESDB_USERNAME, config.Config.CeresDBPassword, config.Config.CeresDBHost, config.Config.CeresDBPort)

	Client.ID = uuid.New().String()

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

	node := StormfrontNode{ID: Client.ID, Host: Client.Host, Port: Client.Port, System: StormfrontSystemInfo{}, Health: "Healthy", Type: "Follower"}

	postBody, _ := json.Marshal(node)

	status, _, err = communication.Post(Client.Leader.Host, Client.Leader.Port, "api/register", AuthClient, postBody)
	if err != nil {
		fmt.Printf("Encountered error: %v\n", err.Error())
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("unable to contact client at %s:%v, received status code %v", Client.Leader.Host, Client.Leader.Port, status)
	}

	clientData, _ := json.Marshal(Client)
	_, err = connection.Query(fmt.Sprintf(`post record stormfront.client %s`, clientData))
	if err != nil {
		fmt.Printf("database error: %v", err)
		return err
	}

	// Check follower healths
	go HealthCheckFollower()

	return nil
}

func InitializeLeader() error {
	time.Sleep(5 * time.Second)
	err := CreateDatabases()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		panic(err)
	}

	Client.ID = Client.Leader.ID
	AuthClient = auth.CreateClientInformation()
	auth.WriteClientInformation(AuthClient)

	authData, _ := json.Marshal(AuthClient)
	_, err = connection.Query(fmt.Sprintf("post record stormfront.auth %s", authData))

	if err != nil {
		panic(err)
	}

	leaderDataRaw := StormfrontLeader{
		ID:         Client.ID,
		Succession: make([]string, 0),
		Healthy:    make([]string, 0),
		Unhealthy:  make([]string, 0),
		Unknown:    make([]string, 0),
	}
	leaderData, _ := json.Marshal(leaderDataRaw)

	_, err = connection.Query(fmt.Sprintf("post record stormfront.leader %s", leaderData))

	if err != nil {
		panic(err)
	}

	nodeDataRaw := StormfrontNode{
		ID:     Client.ID,
		Host:   Client.Host,
		Port:   Client.Port,
		System: StormfrontSystemInfo{},
		Health: "Healthy",
		Type:   "Leader",
	}
	nodeData, _ := json.Marshal(nodeDataRaw)

	_, err = connection.Query(fmt.Sprintf("post record stormfront.node %s", nodeData))

	if err != nil {
		panic(err)
	}

	clientData, _ := json.Marshal(Client)
	_, err = connection.Query(fmt.Sprintf(`post record stormfront.client %s`, clientData))
	if err != nil {
		fmt.Printf("database error: %v", err)
		return err
	}

	// Check follower healths
	go HealthCheckLeader()

	return nil
}
