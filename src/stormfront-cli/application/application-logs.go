package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/auth"
	"stormfront-cli/logging"
)

var ApplicationLogsHelpText = fmt.Sprintf(`usage: stormfront application logs [-i|--id <application id>] [-H|--host <stormfront host>] [-p|--port <stormfront port>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-i|--id           The ID of the application to get
	-H|--host         The host of the stormfront client to connect to, defaults to "localhost"
	-p|--port         The port of the stormfront client to connect to, defaults to "6626"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.INFO_NAME)

func ParseLogsArgs(args []string) (string, string, string, error) {
	host := "localhost"
	port := "6626"
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
				return "", "", "", errors.New("no value passed after id flag")
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
			fmt.Println(ApplicationLogsHelpText)
			os.Exit(1)
		}
	}

	return host, port, id, nil
}

func ExecuteLogs(host, port, id string) error {
	logging.Info("Getting application...")

	requestURL := fmt.Sprintf("http://%s:%s/api/application/%s/logs", host, port, id)

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

	responseData := map[string]string{}

	json.Unmarshal(body, &responseData)

	logging.Debug(fmt.Sprintf("Status code: %v", resp.StatusCode))
	logging.Debug(responseBody)

	fmt.Println(responseData["logs"])

	return nil
}
