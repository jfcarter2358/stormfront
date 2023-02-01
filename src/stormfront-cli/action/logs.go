package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"stormfront-cli/config"
	"stormfront-cli/logging"
)

func GetLogsById(id string) (string, error) {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return "", nil
	}
	logging.Info("Getting logs...")

	requestURL := fmt.Sprintf("http://%s:%s/api/application/%s/logs", host, port, id)

	logging.Debug("Sending GET request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	apiToken, err := config.GetAPIToken()
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Stormfront-API %s", apiToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	logging.Debug("Done!")

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	responseBody := string(body)

	logging.Debug(fmt.Sprintf("Status code: %v", resp.StatusCode))
	logging.Debug(fmt.Sprintf("Response body: %s", responseBody))

	if resp.StatusCode == http.StatusOK {
		responseJSON := map[string]string{}
		json.Unmarshal(body, &responseJSON)
		return responseJSON["logs"], nil
	} else {
		var data map[string]string
		if err := json.Unmarshal([]byte(responseBody), &data); err == nil {
			if errMessage, ok := data["error"]; ok {
				logging.Error(errMessage)
				return "", errors.New(errMessage)
			}
		}
		return "", fmt.Errorf("client has returned error with status code %v", resp.StatusCode)
	}
}

func GetLogsByNameNamespace(name, namespace string) (string, error) {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return "", nil
	}

	logging.Info("Getting applications...")

	requestURL := fmt.Sprintf("http://%s:%s/api/application", host, port)

	logging.Debug("Sending GET request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	apiToken, err := config.GetAPIToken()
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Stormfront-API %s", apiToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	logging.Debug("Done!")

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	responseBody := string(body)

	logging.Debug(fmt.Sprintf("Status code: %v", resp.StatusCode))
	logging.Debug(fmt.Sprintf("Response body: %s", responseBody))

	if resp.StatusCode == http.StatusOK {
		data, err := ParseJSON(responseBody)
		if err != nil {
			return "", err
		}
		data, err = FilterNamespace(data, namespace)
		if err != nil {
			return "", err
		}
		for _, application := range data {
			if application["name"].(string) == name {
				id := application["id"].(string)
				output, err := GetLogsById(id)
				return output, err
			}
		}
		return "", fmt.Errorf("no application with name %s in namespace %s exists", name, namespace)
	}
	return "", fmt.Errorf("request failed with status code %d", resp.StatusCode)
}
