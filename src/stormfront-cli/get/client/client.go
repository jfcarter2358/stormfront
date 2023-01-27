package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/auth"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
	"strings"

	"gopkg.in/yaml.v2"
)

var ClientHelpText = fmt.Sprintf(`usage: stormfront get client [<client id>] [-H|--host <stormfront host>] [-p|--port <stormfront port>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-H|--host         The host of the stormfront client to connect to, defaults to "localhost"
	-p|--port         The port of the stormfront client to connect to, defaults to "6626"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseClientArgs(args []string) (string, string, string, string, error) {
	host := "localhost"
	port := "6626"
	id := ""
	output := "table"
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
		case "-p", "--port":
			if len(args) > 1 {
				port = args[1]
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after port flag")
			}
		case "-o", "--output":
			if len(args) > 1 {
				output = args[1]
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after output flag")
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
			if strings.HasPrefix(args[0], "-") {
				fmt.Printf("Invalid argument: %s\n", args[0])
				fmt.Println(ClientHelpText)
				os.Exit(1)
			} else {
				id = args[0]
				args = args[1:]
			}
		}
	}

	if output != "table" && output != "yaml" && output != "json" {
		return "", "", "", "", fmt.Errorf(`invalid output option "%s", valid arguments are "table", "yaml", and "json"`, output)
	}

	return host, port, id, output, nil
}

func ExecuteClient(host, port, id, output string) error {
	logging.Info("Getting client...")

	requestURL := ""
	if id == "" {
		requestURL = fmt.Sprintf("http://%s:%s/api/client", host, port)
	} else {
		requestURL = fmt.Sprintf("http://%s:%s/api/client/%s", host, port, id)
	}

	logging.Debug("Sending GET request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	clientInfo := auth.ReadClientInformation()

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", clientInfo.AccessToken))
	resp, err := httpClient.Do(req)
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
		headers := []string{
			"id",
			"type",
			"updated",
			"host",
			"port",
			"healthy",
		}
		types := []string{
			"string",
			"string",
			"string",
			"string",
			"int",
			"bool",
		}

		switch output {
		case "table":
			var data []map[string]interface{}
			if strings.HasPrefix(responseBody, "[") {
				json.Unmarshal([]byte(responseBody), &data)
			} else {
				json.Unmarshal([]byte(fmt.Sprintf("[%s]", responseBody)), &data)
			}
			utils.PrintTable(data, headers, types)
		case "yaml":
			if strings.HasPrefix(responseBody, "[") {
				var data []map[string]interface{}
				json.Unmarshal([]byte(responseBody), &data)
				contents, _ := yaml.Marshal(&data)
				fmt.Println(contents)
			} else {
				var data map[string]interface{}
				json.Unmarshal([]byte(responseBody), &data)
				contents, _ := yaml.Marshal(&data)
				fmt.Println(contents)
			}
		case "json":
			fmt.Println(responseBody)
		}
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
