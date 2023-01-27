package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stormfrontd/client/auth"
	"stormfrontd/client/communication"
	"time"

	"github.com/jfcarter2358/ceresdb-go/connection"
)

func updateSuccession() error {
	AuthClient = auth.ReadClientInformation()

	nodeData, err := connection.Query("get record stormfront.leader")
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

	nodeData, err = connection.Query("get record stormfront.leader")
	if err != nil {
		log.Printf("Unable to contact database during node get, changes to node status not recorded: %v\n", err)
		return err
	}
	nodeData[0]["succession"] = newSuccession
	nodeData[0]["unhealthy"] = newUnhealthy
	nodeData[0]["unknown"] = newUnknown

	payload, _ := json.Marshal(nodeData)

	_, err = connection.Query(fmt.Sprintf("put record stormfront.leader %s", payload))
	if err != nil {
		log.Printf("Unable to contact database during node put, changes to node status not recorded: %v\n", err)
		return err
	}

	return nil
}
