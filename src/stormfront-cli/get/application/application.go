package application

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

var ApplicationHelpText = fmt.Sprintf(`usage: stormfront application get [<application id>] [-o|--output <output>] [-n|--namespace] [-a|--all-namespaces] [-l|--log-level <log level>] [-h|--help]
arguments:
	-o|--output            Output format to print to console, valid options are "table", "yaml", and "json"
	-n|--namespace         Namespace to grab applications from
	-a|--all-namespaces    Show applications from all namespaces. Supersedes 'namespace' flag
	-l|--log-level         Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help              Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseApplicationArgs(args []string) (string, string, string, bool, error) {
	id := ""
	output := "table"
	namespace := ""
	allNamespaces := false
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
					return "", "", "", false, fmt.Errorf("invalid output value %s, allowed values are 'table', 'yaml', and 'json", args[1])
				}
				args = args[2:]
			} else {
				return "", "", "", false, errors.New("no value passed after output flag")
			}
		case "-a", "--all-namespaces":
			allNamespaces = true
			args = args[1:]
		case "-n", "--namespace":
			if len(args) > 1 {
				output = args[1]
				args = args[2:]
			} else {
				return "", "", "", false, errors.New("no value passed after namespace flag")
			}
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", "", "", false, err
				}
				args = args[2:]
			} else {
				return "", "", "", false, errors.New("no value passed after log-level flag")
			}
		default:
			if strings.HasPrefix(args[0], "-") || id != "" {
				fmt.Printf("Invalid argument: %s\n", args[0])
				fmt.Println(ApplicationHelpText)
				os.Exit(1)
			} else {
				id = args[0]
				args = args[1:]
			}
		}
	}

	return id, output, namespace, allNamespaces, nil
}

func ExecuteApplication(id, output, namespace string, allNamespaces bool) error {
	host, err := config.GetHost()
	if err != nil {
		return err
	}
	port, err := config.GetPort()
	if err != nil {
		return err
	}

	logging.Info("Getting application...")

	requestURL := ""
	if id == "" {
		requestURL = fmt.Sprintf("http://%s:%s/api/application", host, port)
	} else {
		requestURL = fmt.Sprintf("http://%s:%s/api/application/%s", host, port, id)
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
			"name",
			"image",
			"node",
			"hostname",
			"namespace",
		}
		types := []string{
			"string",
			"string",
			"string",
			"string",
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
			if !allNamespaces {
				if namespace == "" {
					namespace, err = config.GetNamespace()
					if err != nil {
						return err
					}
				}
				data = utils.Filter(data, "namespace", namespace)
			}
			utils.PrintTable(data, headers, types)
		case "yaml":
			var data []map[string]interface{}
			if strings.HasPrefix(responseBody, "[") {
				var tempData []map[string]interface{}
				json.Unmarshal([]byte(responseBody), &tempData)
				data = tempData
			} else {
				var tempData map[string]interface{}
				json.Unmarshal([]byte(responseBody), &tempData)
				data = []map[string]interface{}{tempData}
			}
			if !allNamespaces {
				if namespace == "" {
					namespace, err = config.GetNamespace()
					if err != nil {
						return err
					}
				}
				data = utils.Filter(data, "namespace", namespace)
			}
			contents, _ := yaml.Marshal(&data)
			fmt.Println(string(contents))
		case "json":
			var data []map[string]interface{}
			if strings.HasPrefix(responseBody, "[") {
				var tempData []map[string]interface{}
				json.Unmarshal([]byte(responseBody), &tempData)
				data = tempData
			} else {
				var tempData map[string]interface{}
				json.Unmarshal([]byte(responseBody), &tempData)
				data = []map[string]interface{}{tempData}
			}
			if !allNamespaces {
				if namespace == "" {
					namespace, err = config.GetNamespace()
					if err != nil {
						return err
					}
				}
				data = utils.Filter(data, "namespace", namespace)
			}
			contents, _ := json.Marshal(&data)
			fmt.Println(string(contents))
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
