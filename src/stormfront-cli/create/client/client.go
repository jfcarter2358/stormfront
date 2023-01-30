package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/auth"
	"stormfront-cli/config"
	"stormfront-cli/logging"

	"github.com/google/uuid"
)

var ClientHelpText = fmt.Sprintf(`usage: stormfront create client [-H|--host <stormfront host>] [-p|--port <stormfront port>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-H|--host         The host of the Stormfront daemon to connect to. Defaults to "localhost"
	-p|--port         The port of the Stormfront daemon to connect to. Defaults to "6674"
	-c|--client-port  The port that the Stormfront client will be deployed to. Defaults to "6626"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.ERROR_NAME)

func ParseClientArgs(args []string) (string, string, string, error) {
	host := "localhost"
	port := "6674"
	clientPort := "6626"
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
				return "", "", "", errors.New("no value passed after host flag")
			}
		case "-p", "--port":
			if len(args) > 1 {
				port = args[1]
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after port flag")
			}
		case "-c", "--client-port":
			if len(args) > 1 {
				clientPort = args[1]
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after client port flag")
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
			fmt.Println(ClientHelpText)
			os.Exit(1)
		}
	}

	return host, port, clientPort, nil
}

func ExecuteClient(host, port, clientPort string) error {
	logging.Info("Deploying stormfront client...")

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

	if resp.StatusCode == http.StatusOK {
		apiToken, err := auth.GetAPIToken(host, clientPort)
		if err != nil {
			return err
		}

		clusterName := uuid.New().String()

		clusterData := config.ClusterConfig{
			Name:             clusterName,
			Token:            apiToken,
			CurrentNamespace: "default",
			Namespaces:       []string{"default"},
			Host:             host,
			Port:             port,
		}

		err = config.AddCluster(clusterData)
		if err != nil {
			return err
		}

		err = config.ChangeCluster(clusterName)
		if err != nil {
			return err
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
