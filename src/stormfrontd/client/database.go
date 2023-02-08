package client

import (
	"fmt"
	"stormfrontd/config"

	"github.com/jfcarter2358/ceresdb-go/connection"
)

const CERESDB_USERNAME = "ceresdb"

var Collections = map[string]string{
	"auth":        `{"id":"STRING","access_token":"STRING","refresh_token":"STRING","token_expiration":"STRING","token_issued":"STRING"}`,
	"api":         `{"token":"STRING"}`,
	"application": `{"id":"STRING","node":"STRING","name":"STRING","image":"STRING","hostname":"STRING","env":"DICT","ports":"DICT","mounts":"DICT","memory":"INT","cpu":"FLOAT","status":"DICT","namespace":"STRING"}`,
	"leader":      `{"id":"STRING","succession":"LIST","unhealthy":"LIST","unknown":"LIST","healthy":"LIST"}`,
	"node":        `{"id":"STRING","host":"STRING","port":"INT","system":"DICT","health":"STRING","type":"STRING"}`,
	"client":      `{"id":"STRING","type":"STRING","leader":"DICT","succession":"LIST","unhealthy":"LIST","unknown":"LIST","updated":"STRING","host":"STRING","port":"INT","healthy":"BOOL","applications":"LIST","system":"DICT"}`,
	"route":       `{"id":"STRING","hostname":"STRING","port":"INT","namespace":"STRING","alias":"STRING","name":"STRING"}`,
}

func CreateDatabases() error {
	fmt.Println("Initializing CeresDB connection")
	if config.Config.CeresDBHost == "" {
		config.Config.CeresDBHost = Client.Host
	}
	connection.Initialize(CERESDB_USERNAME, config.Config.CeresDBPassword, config.Config.CeresDBHost, config.Config.CeresDBPort)
	fmt.Println("Done!")

	// fmt.Printf("Username: %s\n", connection.Username)
	// fmt.Printf("Password: %s\n", connection.Password)
	// fmt.Printf("Host: %s\n", connection.Host)
	// fmt.Printf("Port: %d\n", connection.Port)

	fmt.Println("Creating stormfront database")
	data, err := connection.Query("post database stormfront")
	if err != nil {
		if data != nil {
			fmt.Printf("Data: %v\n", data)
		}
		return err
	}
	fmt.Println("Done!")
	for name, schema := range Collections {
		fmt.Printf("Creating %s collection\n", name)
		fmt.Printf("post collection stormfront.%s %s\n", name, schema)
		_, err := connection.Query(fmt.Sprintf("post collection stormfront.%s %s", name, schema))
		if err != nil {
			fmt.Printf("Data: %v\n", data)
			return err
		}
		fmt.Println("Done!")
	}
	return nil
}
