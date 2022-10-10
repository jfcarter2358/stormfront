package application

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/auth"
	"stormfront-cli/logging"
)

var ApplicationCreateHelpText = fmt.Sprintf(`usage: stormfront application create [-d|--definition <application definition file path>] [-H|--host <stormfront host>] [-p|--port <stormfront port>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-d|--definition   The path to the JSON file defining the application
	-H|--host         The host of the stormfront client to connect to, defaults to "localhost"
	-p|--port         The port of the stormfront client to connect to, defaults to "6626"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.INFO_NAME)

func ParseCreateArgs(args []string) (string, string, string, error) {
	host := "localhost"
	port := "6626"
	definition := ""
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping", envLogLevel)
		}
	}

	for len(args) > 0 {
		switch args[0] {
		case "-d", "--definition":
			if len(args) > 1 {
				definition = args[1]
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after definition flag")
			}
		case "-H", "--host":
			if len(args) > 1 {
				host = args[1]
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after host flag")
			}
		case "-p", "--port":
			if len(args) > 1 {
				port = args[1]
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after port flag")
			}
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", "", "", err
				}
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after log-level flag")
			}
		default:
			fmt.Printf("Invalid argument: %s\n", args[0])
			fmt.Println(ApplicationCreateHelpText)
			os.Exit(1)
		}
	}

	return host, port, definition, nil
}

func ExecuteCreate(host, port, definition string) error {
	logging.Info("Creating application...")

	requestURL := fmt.Sprintf("http://%s:%s/api/application", host, port)

	logging.Debug("Sending POST request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	clientInfo := auth.ReadClientInformation()

	file, _ := ioutil.ReadFile(definition)
	data := map[string]interface{}{}

	err := json.Unmarshal([]byte(file), &data)
	if err != nil {
		panic(err)
	}
	postBody, _ := json.Marshal(data)
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
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

	fmt.Println(responseBody)

	return nil
}
