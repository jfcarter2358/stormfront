package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/logging"
	"strconv"
	"strings"
)

var JoinHelpText = fmt.Sprintf(`usage: stormfront join -l <leader URL> -j <join-token> [-H <stormfront host>] [-p <stormfront port>] [-l <log level>] [-h|--help]
arguments:
	-H|--host          The host of the stormfront client to connect to, defaults to "localhost"
	-p|--port          The port of the stormfront client to connect to, defaults to "6674"
	-L|--leader        URL of the leader to join in the form <host>:<port>
	-j|--join-token    Join token to use to connect to the leader
	-l|--log-level     Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help          Show this help message and exit`, logging.GetDefaults(), logging.INFO_NAME)

func ParseJoinArgs(args []string) (string, string, string, string, error) {
	host := "localhost"
	port := "6674"
	joinToken := ""
	leader := ""
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping", envLogLevel)
		}
	}

	for len(args) > 0 {
		switch args[0] {
		case "-H", "--host":
			if len(args) > 1 {
				host = args[1]
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after host flag")
			}
		case "-j", "--join-token":
			if len(args) > 1 {
				joinToken = args[1]
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after log-level flag")
			}
		case "-L", "--leader":
			if len(args) > 1 {
				leader = args[1]
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after log-level flag")
			}
		case "-p", "--port":
			if len(args) > 1 {
				port = args[1]
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after port flag")
			}
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", "", "", "", err
				}
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after log-level flag")
			}
		default:
			fmt.Printf("Invalid argument: %s\n", args[0])
			fmt.Println(JoinHelpText)
			os.Exit(1)
		}
	}

	return host, port, leader, joinToken, nil
}

func ExecuteJoin(host, port, leader, joinToken string) error {

	logging.Info("Deploying stormfront node...")

	requestURL := fmt.Sprintf("http://%s:%s/api/join", host, port)

	logging.Debug("Sending POST request to daemon...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	parts := strings.Split(leader, ":")

	intPort, err := strconv.Atoi(parts[1])

	if err != nil {
		logging.Fatal(fmt.Sprintf("Invalid port number: %s", parts[1]))
	}

	data := map[string]interface{}{"leader_host": parts[0], "leader_port": intPort, "join_token": joinToken}

	postBody, _ := json.Marshal(data)
	postBodyBuffer := bytes.NewBuffer(postBody)

	resp, err := http.Post(requestURL, "application/json", postBodyBuffer)
	if err != nil {
		return err
	}

	logging.Debug("Done!")

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	responseBody := string(body)

	logging.Debug(fmt.Sprintf("Status code: %v", resp.StatusCode))
	logging.Debug(fmt.Sprintf("Response body: %s", responseBody))

	if resp.StatusCode == http.StatusOK {
		logging.Success("Done!")
	} else {
		var data map[string]string
		if err := json.Unmarshal([]byte(responseBody), &data); err == nil {
			if errMessage, ok := data["error"]; ok {
				logging.Error(errMessage)
			}
		}
		logging.Fatal(fmt.Sprintf("Client has returned error with status code %v", resp.StatusCode))
	}

	return nil
}
