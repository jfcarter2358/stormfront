package access

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/auth"
	"stormfront-cli/config"
	"stormfront-cli/logging"
)

var DebugRefreshHelpText = fmt.Sprintf(`usage: stormfront token access refresh [-l|--log-level <log level>] [-h|--help]
arguments:
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseRefreshArgs(args []string) error {
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
					return err
				}
				args = args[2:]
			} else {
				return errors.New("no value passed after log-level flag")
			}
		default:
			fmt.Printf("Invalid argument: %s\n", args[0])
			fmt.Println(DebugRefreshHelpText)
			os.Exit(1)
		}
	}

	return nil
}

func ExecuteRefresh() error {
	host, err := config.GetHost()
	if err != nil {
		return err
	}

	port, err := config.GetPort()
	if err != nil {
		return err
	}

	clientInfo := auth.ReadClientInformation()

	logging.Info("Refreshing stormfront client token...")

	requestURL := fmt.Sprintf("http://%s:%s/auth/refresh", host, port)

	logging.Debug("Sending GET request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

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
		json.Unmarshal(body, &clientInfo)

		err = auth.WriteClientInformation(clientInfo)

		return err
	} else {
		var data map[string]string
		if err := json.Unmarshal([]byte(responseBody), &data); err == nil {
			if errMessage, ok := data["error"]; ok {
				logging.Error(errMessage)
			}
		}
		logging.Fatal(fmt.Sprintf("Client has returned error with status code %v", resp.StatusCode))
	}

	return fmt.Errorf("client has returned error with status code %v", resp.StatusCode)
}
