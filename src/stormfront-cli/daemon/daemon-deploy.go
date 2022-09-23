package daemon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/logging"
)

var DaemonDeployHelpText = fmt.Sprintf(`usage: stormfront daemon deploy [-H|--host <stormfront host>] [-p|--port <stormfront port>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-H|--host         The host of the stormfront daemon to connect to, defaults to "localhost"
	-p|--port         The port of the stormfront daemon to connect to, defaults to "6674"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.INFO_NAME)

func ParseDeployArgs(args []string) (string, string, error) {
	host := "localhost"
	port := "6674"

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
			fmt.Println(DaemonDeployHelpText)
			os.Exit(1)
		}
	}

	return host, port, nil
}

func ExecuteDeploy(host, port string) error {
	logging.Info("Deploying stormfront node...")

	requestURL := fmt.Sprintf("http://%s:%s/api/deploy", host, port)

	logging.Debug("Sending POST request to daemon...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	postBody, _ := json.Marshal(map[string]string{})
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

	logging.Success("Done!")

	return nil
}
