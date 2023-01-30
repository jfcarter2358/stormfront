package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/config"
	"stormfront-cli/logging"
)

var ApplicationUpdateHelpText = fmt.Sprintf(`usage: stormfront application update <application id> [-d|--definition <application definition file path>] [-H|--host <stormfront host>] [-p|--port <stormfront port>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-i|--id           The ID of the application to update
	-d|--definition   The path to the JSON file defining the application
	-H|--host         The host of the stormfront client to connect to, defaults to "localhost"
	-p|--port         The port of the stormfront client to connect to, defaults to "6626"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseUpdateArgs(args []string) (string, string, string, string, error) {
	host := "localhost"
	port := "6626"
	definition := ""
	id := ""
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping", envLogLevel)
		}
	}

	for len(args) > 0 {
		switch args[0] {
		case "-i", "--id":
			if len(args) > 1 {
				id = args[1]
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after id flag")
			}
		case "-d", "--definition":
			if len(args) > 1 {
				definition = args[1]
				args = args[2:]
			} else {
				return "", "", "", "", errors.New("no value passed after definition flag")
			}
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
			fmt.Println(ApplicationUpdateHelpText)
			os.Exit(1)
		}
	}

	return host, port, definition, id, nil
}

func ExecuteUpdate(host, port, definition, id string) error {
	logging.Info("Creating application...")

	requestURL := fmt.Sprintf("http://%s:%s/api/application/%s", host, port, id)

	logging.Debug("Sending PATCH request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	apiToken, err := config.GetAPIToken()
	if err != nil {
		return err
	}

	file, _ := ioutil.ReadFile(definition)
	data := map[string]interface{}{}

	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		panic(err)
	}
	postBody, _ := json.Marshal(data)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	req, _ := http.NewRequest("PATCH", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("X-Stormfront-API %s", apiToken))
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
		fmt.Println(responseBody)
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
