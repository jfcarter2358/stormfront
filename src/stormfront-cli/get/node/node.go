package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/config"
	"stormfront-cli/logging"
	"stormfront-cli/utils"
	"strings"

	"gopkg.in/yaml.v2"
)

var NodeHelpText = fmt.Sprintf(`usage: stormfront get node [<node id>] [-o|--output <output>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseNodeArgs(args []string) (string, string, error) {
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
		case "-o", "--output":
			if len(args) > 1 {
				switch args[1] {
				case "table", "yaml", "json":
					output = args[1]
				default:
					return "", "", fmt.Errorf("invalid output value %s, allowed values are 'table', 'yaml', and 'json")
				}
				args = args[2:]
			} else {
				return "", "", errors.New("no value passed after output flag")
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
			if strings.HasPrefix(args[0], "-") {
				fmt.Printf("Invalid argument: %s\n", args[0])
				fmt.Println(NodeHelpText)
				os.Exit(1)
			} else {
				id = args[0]
				args = args[1:]
			}
		}
	}

	return id, output, nil
}

func ExecuteNode(id, output string) error {
	host, err := config.GetHost()
	if err != nil {
		return err
	}
	port, err := config.GetPort()
	if err != nil {
		return err
	}

	logging.Info("Getting node...")

	requestURL := ""
	if id == "" {
		requestURL = fmt.Sprintf("http://%s:%s/api/node", host, port)
	} else {
		requestURL = fmt.Sprintf("http://%s:%s/api/node/%s", host, port, id)
	}

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
		headers := []string{
			"id",
			"host",
			"port",
			"health",
			"type",
		}
		types := []string{
			"string",
			"string",
			"int",
			"string",
			"string",
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
				fmt.Println(string(contents))
			} else {
				var data map[string]interface{}
				json.Unmarshal([]byte(responseBody), &data)
				contents, _ := yaml.Marshal(&data)
				fmt.Println(string(contents))
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
