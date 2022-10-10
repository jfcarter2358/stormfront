package api_token

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"stormfront-cli/auth"
	"stormfront-cli/logging"
)

var APITokenRevokeHelpText = fmt.Sprintf(`usage: stormfront api-token revoke -t <token> [-H|--host <stormfront host>] [-p|--port <stormfront port>] [-l|--log-level <log level>] [-h|--help]
arguments:
	-t|--token        The API token to revoke
	-H|--host         The host of the stormfront client to connect to, defaults to "localhost"
	-p|--port         The port of the stormfront client to connect to, defaults to "6626"
	-l|--log-level    Sets the log level of the CLI. valid levels are: %s, defaults to %s
	-h|--help         Show this help message and exit`, logging.GetDefaults(), logging.INFO_NAME)

func ParseRevokeArgs(args []string) (string, string, string, error) {
	host := "localhost"
	port := "6626"
	token := ""
	envLogLevel, present := os.LookupEnv("STORMFRONT_LOG_LEVEL")
	if present {
		if err := logging.SetLevel(envLogLevel); err != nil {
			fmt.Printf("Env logging level %s (from STORMFRONT_LOG_LEVEL) is invalid, skipping", envLogLevel)
		}
	}

	for len(args) > 0 {
		switch args[0] {
		case "-t", "--token":
			if len(args) > 1 {
				token = args[1]
				args = args[2:]
			} else {
				return "", "", "", errors.New("no value passed after token flag")
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
			fmt.Println(APITokenRevokeHelpText)
			os.Exit(1)
		}
	}

	if token == "" {
		return "", "", "", errors.New("no token passed to revoke")
	}

	return token, host, port, nil
}

func ExecuteRevoke(token, host, port string) error {
	logging.Info("Getting stormfront client health...")

	requestURL := fmt.Sprintf("http://%s:%s/auth/api", host, port)

	logging.Debug("Sending DELETE request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	clientInfo := auth.ReadClientInformation()

	httpClient := &http.Client{}
	req, _ := http.NewRequest("DELETE", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", clientInfo.AccessToken))
	req.Header.Set("X-Stormfront-API", token)
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
		logging.Success("API token revoked")
	} else {
		logging.Fatal(fmt.Sprintf("Client has returned error with status code %v", resp.StatusCode))
	}

	return nil
}
