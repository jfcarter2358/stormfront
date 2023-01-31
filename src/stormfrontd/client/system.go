package client

import (
	"encoding/json"
	"fmt"
	"stormfrontd/config"
	"time"

	"github.com/jfcarter2358/ceresdb-go/connection"
	"github.com/pbnjay/memory"
	"github.com/shirou/gopsutil/cpu"
)

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

	// update client information
	clientIDs, err := connection.Query(fmt.Sprintf(`get record stormfront.client .id | filter id = "%s"`, Client.ID))
	if err != nil {
		fmt.Printf("database error: %v", err)
		return err
	}
	clientData, _ := json.Marshal(Client)
	var clientMap map[string]interface{}
	json.Unmarshal(clientData, &clientMap)
	clientMap[".id"] = clientIDs[0][".id"]
	clientData, _ = json.Marshal(clientMap)
	_, err = connection.Query(fmt.Sprintf(`put record stormfront.client %s`, clientData))
	if err != nil {
		fmt.Printf("database error: %v", err)
		return err
	}

	// update node information
	nodeData, err := connection.Query(fmt.Sprintf(`get record stormfront.node | filter id = "%s"`, Client.ID))
	if err != nil {
		return err
	}

	nodeBytes, err := json.Marshal(nodeData[0])
	if err != nil {
		return err
	}

	var node map[string]interface{}
	err = json.Unmarshal(nodeBytes, &node)
	if err != nil {
		return err
	}

	node["system"] = systemInfo

	nodeMarshalled, err := json.Marshal(node)
	if err != nil {
		return err
	}
	fmt.Println(string(nodeMarshalled))

	_, err = connection.Query(fmt.Sprintf(`put record stormfront.node %s`, nodeMarshalled))
	if err != nil {
		fmt.Printf("database error: %v", err)
		return err
	}

	return nil
}
