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
	"application": `{"id":"STRING","node":"STRING","name":"STRING","image":"STRING","hostname":"STRING","env":"DICT","ports":"DICT","mounts":"DICT","memory":"INT","cpu":"FLOAT","status":"DICT"}`,
	"leader":      `{"id":"STRING","succession":"LIST","unhealthy":"LIST","unknown":"LIST"}`,
	"node":        `{"id":"STRING","host":"STRING","port":"INT","system":"DICT","health":"STRING"}`,
	"client":      `{"id":"STRING","type":"STRING","leader":"DICT","succession":"LIST","unhealthy":"LIST","unknown":"LIST","updated":"STRING","host":"STRING","port":"INT","healthy":"BOOL","applications":"LIST","system":"DICT"}`,
}

// ID           string                  `json:"id" yaml:"id"`
// 	Type         string                  `json:"type" yaml:"type"`
// 	Leader       StormfrontNode          `json:"leader" yaml:"leader"`
// 	Succession   []StormfrontNode        `json:"succession" yaml:"succession"`
// 	Unhealthy    []StormfrontNode        `json:"unhealthy" yaml:"unhealthy"`
// 	Unknown      []StormfrontNode        `json:"unknown" yaml:"unknown"`
// 	Updated      string                  `json:"updated" yaml:"updated"`
// 	Host         string                  `json:"host" yaml:"host"`
// 	Port         int                     `json:"port" yaml:"port"`
// 	Healthy      bool                    `json:"healthy" yaml:"healthy"`
// 	Router       *gin.Engine             `json:"-" yaml:"-"`
// 	Server       *http.Server            `json:"-" yaml:"-"`
// 	Applications []StormfrontApplication `json:"applications" yaml:"applications"`
// 	System       StormfrontSystemInfo    `json:"system" yaml:"system"`

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
