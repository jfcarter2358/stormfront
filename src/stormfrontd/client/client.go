package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stormfrontd/client/auth"
	"stormfrontd/client/communication"
	"stormfrontd/config"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jfcarter2358/ceresdb-go/connection"
	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
)

var Client StormfrontClient
var Running = false
var JoinTokens []string
var AuthClient auth.ClientInformation

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
	TotalDiskSpace  int     `json:"total_disk"`
	FreeDiskSpace   int     `json:"free_disk"`
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

	diskInfo := DiskUsage("/")
	systemInfo.FreeDiskSpace = int(diskInfo.Free)
	systemInfo.TotalDiskSpace = int(diskInfo.All)

	memoryUsed := (float64(systemInfo.TotalMemory) * config.Config.ReservedMemoryPercentage) + float64(memoryReserved)
	systemInfo.MemoryAvailable = systemInfo.TotalMemory - int(memoryUsed)

	cpuUsed := (float64(systemInfo.Cores) * config.Config.ReservedCPUPercentage) + cpuReserved
	systemInfo.CPUAvailable = float64(systemInfo.Cores) - cpuUsed

	Client.System = systemInfo

	return nil
}

func updateSuccession() error {
	AuthClient = auth.ReadClientInformation()

	nodeData, err := connection.Query("get record stormfront.node")
	if len(nodeData) == 0 {
		log.Println("No followers to update")
		return nil
	}
	if err != nil {
		return err
	}
	var succession []StormfrontNode
	successionBytes, _ := json.Marshal(nodeData[0]["succession"])
	json.Unmarshal(successionBytes, &succession)

	newSuccession := []StormfrontNode{}
	newUnhealthy := []StormfrontNode{}
	newUnknown := []StormfrontNode{}

	for _, successor := range succession {
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

	reconcileApplications()

	newSuccession = dedupeNodes(newSuccession)
	newUnhealthy = dedupeNodes(newUnhealthy)
	newUnknown = dedupeNodes(newUnknown)

	nodeData, err = connection.Query("get record stormfront.node")
	if err != nil {
		log.Printf("Unable to contact database during node get, changes to node status not recorded: %v\n", err)
		return err
	}
	nodeData[0]["succession"] = newSuccession
	nodeData[0]["unhealthy"] = newUnhealthy
	nodeData[0]["unknown"] = newUnknown

	payload, _ := json.Marshal(nodeData)

	_, err = connection.Query(fmt.Sprintf("put record stormfront.node %s", payload))
	if err != nil {
		log.Printf("Unable to contact database during node put, changes to node status not recorded: %v\n", err)
		return err
	}

	return nil
}

func HealthCheckFollower() {
	for {
		err := updateSuccession()
		if err != nil {
			fmt.Printf("Encountered error updating succession: %v", err)
		}
		time.Sleep(HEALTH_CHECK_DELAY * time.Second)
	}
}

func HealthCheckLeader() {
	for {
		err := updateSuccession()
		if err != nil {
			fmt.Printf("Encountered error updating succession: %v", err)
		}
		time.Sleep(HEALTH_CHECK_DELAY * time.Second)
	}
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

		// Check follower healths
		go HealthCheckLeader()
	}

	err := updateSystemInfo()
	if err != nil {
		panic(err)
	}

	return nil
}
