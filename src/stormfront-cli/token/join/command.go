package join

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/config"
	"stormfront-cli/logging"
)

var GetJoinCommandHelpText = fmt.Sprintf(`usage: stormfront token join command [-H|--host <stormfront host>] [-p|--port <stormfront port>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-H|--host         The host of the stormfront client to connect to, defaults to "localhost"
	-p|--port         The port of the stormfront client to connect to, defaults to "6626"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseGetJoinCommandArgs(args []string) (string, string, error) {
	host := "localhost"
	port := "6626"
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
				return "", "", errors.New("no value passed after host flag")
			}
		case "-p", "--port":
			if len(args) > 1 {
				port = args[1]
				args = args[2:]
			} else {
				return "", "", errors.New("no value passed after port flag")
			}
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", "", err
				}
				args = args[2:]
			} else {
				return "", "", errors.New("no value passed after log-level flag")
			}
		default:
			fmt.Printf("Invalid argument: %s\n", args[0])
			fmt.Println(GetJoinCommandHelpText)
			os.Exit(1)
		}
	}

	return host, port, nil
}

func ExecuteGetJoinCommand(host, port string) error {
	logging.Info("Getting join command from leader...")

	requestURL := fmt.Sprintf("http://%s:%v/auth/join/command", host, port)

	logging.Debug("Sending GET request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	apiToken, err := config.GetAPIToken()
	if err != nil {
		return err
	}

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
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
		var data map[string]string
		json.Unmarshal([]byte(responseBody), &data)
		fmt.Println(data["join_command"])
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
