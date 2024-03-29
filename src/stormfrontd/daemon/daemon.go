package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"stormfrontd/client"
	"stormfrontd/client/auth"
	"stormfrontd/config"
	"stormfrontd/database"
	"stormfrontd/utils"
	"time"

	"github.com/google/uuid"
	"github.com/jfcarter2358/ceresdb-go/connection"
)

func Deploy() error {

	if client.Running {
		return errors.New("client is already running")
	}

	currentTime := time.Now()
	hostname, err := utils.GetIP(config.Config.InterfaceName)
	if err != nil {
		return err
	}

	client.Client = client.StormfrontClient{
		Type: "Leader",
		Leader: client.StormfrontNode{
			ID:   uuid.NewString(),
			Host: hostname,
			Port: config.Config.ClientPort,
		},
		Succession: []client.StormfrontNode{},
		Updated:    currentTime.Format(time.RFC3339),
		Host:       hostname,
		Port:       config.Config.ClientPort,
		Healthy:    true,
	}

	err = database.Deploy("")

	if err != nil {
		database.Destroy()
		return err
	}

	err = client.Initialize("")

	if err != nil {
		Destroy()
	}

	return err
}

func Destroy() error {
	if !client.Running {
		return errors.New("no running client to destroy")
	}

	client.Running = false

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Client.Server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	database.Destroy()

	if client.Client.Type == "Follower" {
		node := client.StormfrontNode{Host: client.Client.Host, Port: client.Client.Port}

		postBody, _ := json.Marshal(node)
		postBodyBuffer := bytes.NewBuffer(postBody)

		clientInfo := auth.ReadClientInformation()

		requestURL := fmt.Sprintf("http://%s:%v/api/register", client.Client.Leader.Host, client.Client.Leader.Port)
		httpClient := &http.Client{}
		req, _ := http.NewRequest("DELETE", requestURL, postBodyBuffer)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", clientInfo.AccessToken))
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("unable to contact client at %s:%v, received status code %v", client.Client.Leader.Host, client.Client.Leader.Port, resp.StatusCode)
		}
	}

	clientIDs, err := connection.Query(fmt.Sprintf(`get record stormfront.client .id | filter id = "%s"`, client.Client.ID))
	if err != nil {
		fmt.Printf("database error: %v", err)
		return err
	}
	_, err = connection.Query(fmt.Sprintf(`delete record stormfront.client %s`, clientIDs[0][".id"].(string)))
	if err != nil {
		fmt.Printf("database error: %v", err)
		return err
	}

	return nil
}

func Join(leaderHost string, leaderPort int, joinToken string) error {
	if client.Running {
		return errors.New("client is already running")
	}

	currentTime := time.Now()
	hostname, err := utils.GetIP(config.Config.InterfaceName)
	if err != nil {
		return err
	}

	client.Client = client.StormfrontClient{
		Type: "Follower",
		Leader: client.StormfrontNode{
			Host: leaderHost,
			Port: leaderPort,
		},
		Succession: []client.StormfrontNode{},
		Updated:    currentTime.Format(time.RFC3339),
		Host:       hostname,
		Port:       config.Config.ClientPort,
		Healthy:    true,
	}

	err = database.Deploy(fmt.Sprintf("%s:7437", leaderHost))

	if err != nil {
		database.Destroy()
		return err
	}

	err = client.Initialize(joinToken)

	if err != nil {
		Destroy()
	}

	return err
}

func Restart() error {
	err := Destroy()
	if err != nil {
		return err
	}
	Deploy()
	return nil
}
