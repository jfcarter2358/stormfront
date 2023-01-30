package namespace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/config"
	"stormfront-cli/logging"
	"strings"
)

var NamespaceHelpText = fmt.Sprintf(`usage: stormfront delete namespace <namespace name> [-l|--log-level <log level>] [-h|--help]
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseNamespaceArgs(args []string) (string, error) {
	name := ""
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping", envLogLevel)
		}
	}

	for len(args) > 0 {
		switch args[0] {
		case "-l", "--log-level":
			if len(args) > 1 {
				err := logging.SetLevel(args[1])
				if err != nil {
					return "", err
				}
				args = args[2:]
			} else {
				return "", errors.New("no value passed after log-level flag")
			}
		default:
			if strings.HasPrefix(args[0], "-") || name != "" {
				fmt.Printf("Invalid argument: %s\n", args[0])
				fmt.Println(NamespaceHelpText)
				os.Exit(1)
			} else {
				name = args[0]
				args = args[1:]
			}
		}
	}

	if name == "" {
		return "", errors.New("name argument is required")
	}

	return name, nil
}

func ExecuteNamespace(name string) error {
	host, err := config.GetHost()
	if err != nil {
		return err
	}
	port, err := config.GetPort()
	if err != nil {
		return err
	}

	logging.Info("Deleting namespace...")

	requestURL := fmt.Sprintf("http://%s:%s/api/application", host, port)

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
		var data []map[string]interface{}
		if strings.HasPrefix(responseBody, "[") {
			json.Unmarshal([]byte(responseBody), &data)
		} else {
			json.Unmarshal([]byte(fmt.Sprintf("[%s]", responseBody)), &data)
		}

		for _, app := range data {
			if app["namespace"].(string) == name {

				logging.Info("Deleting application...")

				requestURL := fmt.Sprintf("http://%s:%s/api/application/%s", host, port, app["id"].(string))

				logging.Debug("Sending DELETE request to client...")
				logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

				apiToken, err := config.GetAPIToken()
				if err != nil {
					return err
				}

				httpClient := &http.Client{}
				req, _ := http.NewRequest("DELETE", requestURL, nil)
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

				if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
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
			}
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

	err = config.RemoveNamespace(name)
	return err
}
