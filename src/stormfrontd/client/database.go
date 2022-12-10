package client

import (
	"fmt"
	"stormfrontd/config"

	"github.com/jfcarter2358/ceresdb-go/connection"
)

const CERESDB_USERNAME = "ceresdb"
const CERESDB_PORT = 7437

var Collections = map[string]string{
	"auth":        `{"id":"STRING","access_token":"STRING","refresh_token":"STRING","token_expiration":"STRING","token_issued":"STRING"}`,
	"api":         `{"token":"STRING"}`,
	"application": `{"id":"STRING","node":"STRING","name":"STRING","image":"STRING","hostname":"STRING","env":"DICT","ports":"DICT","memory":"INT","cpu":"FLOAT","status":"DICT"}`,
	"node":        `{"id":"STRING","succession":"LIST","unhealthy":"LIST","unknown":"LIST"}`,
}

func CreateDatabases() error {
	fmt.Println("Initializing CeresDB connection")
	connection.Initialize(CERESDB_USERNAME, config.Config.CeresDBPassword, Client.Host, CERESDB_PORT)
	fmt.Println("Done!")

	fmt.Printf("Username: %s\n", connection.Username)
	fmt.Printf("Password: %s\n", connection.Password)
	fmt.Printf("Host: %s\n", connection.Host)
	fmt.Printf("Port: %d\n", connection.Port)

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
