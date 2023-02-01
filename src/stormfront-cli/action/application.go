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

func GetAllApplications(namespace string) ([]map[string]interface{}, error) {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return []map[string]interface{}{}, nil
	}

	logging.Info("Getting applications...")

	requestURL := fmt.Sprintf("http://%s:%s/api/application", host, port)

	logging.Debug("Sending GET request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	apiToken, err := config.GetAPIToken()
	if err != nil {
		return []map[string]interface{}{}, err
	}

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Stormfront-API %s", apiToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return []map[string]interface{}{}, err
	}

	logging.Debug("Done!")

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []map[string]interface{}{}, err
	}
	responseBody := string(body)

	logging.Debug(fmt.Sprintf("Status code: %v", resp.StatusCode))
	logging.Debug(fmt.Sprintf("Response body: %s", responseBody))

	if resp.StatusCode == http.StatusOK {
		data, err := ParseJSON(responseBody)
		if err != nil {
			return []map[string]interface{}{}, err
		}
		data, err = FilterNamespace(data, namespace)
		if err != nil {
			return []map[string]interface{}{}, err
		}
		return data, nil
	}
	return []map[string]interface{}{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
}

func GetApplicationById(id string) ([]map[string]interface{}, error) {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return []map[string]interface{}{}, nil
	}

	logging.Info(fmt.Sprintf("Getting application %s...", id))

	requestURL := fmt.Sprintf("http://%s:%s/api/application/%s", host, port, id)

	logging.Debug("Sending GET request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	apiToken, err := config.GetAPIToken()
	if err != nil {
		return []map[string]interface{}{}, err
	}

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Stormfront-API %s", apiToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return []map[string]interface{}{}, err
	}

	logging.Debug("Done!")

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []map[string]interface{}{}, err
	}
	responseBody := string(body)

	logging.Debug(fmt.Sprintf("Status code: %v", resp.StatusCode))
	logging.Debug(fmt.Sprintf("Response body: %s", responseBody))

	if resp.StatusCode == http.StatusOK {
		data, err := ParseJSON(responseBody)
		if err != nil {
			return []map[string]interface{}{}, err
		}
		return data, nil
	}
	return []map[string]interface{}{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
}

func GetApplicationByNameNamespace(name, namespace string) ([]map[string]interface{}, error) {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return []map[string]interface{}{}, nil
	}

	logging.Info("Getting applications...")

	requestURL := fmt.Sprintf("http://%s:%s/api/application", host, port)

	logging.Debug("Sending GET request to client...")
	logging.Trace(fmt.Sprintf("Sending request to %s", requestURL))

	apiToken, err := config.GetAPIToken()
	if err != nil {
		return []map[string]interface{}{}, err
	}

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("X-Stormfront-API %s", apiToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return []map[string]interface{}{}, err
	}

	logging.Debug("Done!")

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []map[string]interface{}{}, err
	}
	responseBody := string(body)

	logging.Debug(fmt.Sprintf("Status code: %v", resp.StatusCode))
	logging.Debug(fmt.Sprintf("Response body: %s", responseBody))

	if resp.StatusCode == http.StatusOK {
		data, err := ParseJSON(responseBody)
		if err != nil {
			return []map[string]interface{}{}, err
		}
		data, err = FilterNamespace(data, namespace)
		if err != nil {
			return []map[string]interface{}{}, err
		}
		for _, application := range data {
			if application["name"].(string) == name {
				return []map[string]interface{}{application}, nil
			}
		}
		return []map[string]interface{}{}, fmt.Errorf("no application with name %s in namespace %s exists", name, namespace)
	}
	return []map[string]interface{}{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
}

func DeleteApplicationById(id string) error {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return nil
	}

	logging.Info(fmt.Sprintf("Deleting application %s...", id))

	requestURL := fmt.Sprintf("http://%s:%s/api/application/%s", host, port, id)

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
		logging.Success("Done!")
	} else {
		var data map[string]string
		if err := json.Unmarshal([]byte(responseBody), &data); err == nil {
			if errMessage, ok := data["error"]; ok {
				return errors.New(errMessage)
			}
			return nil
		}
		return fmt.Errorf("client has returned error with status code %v", resp.StatusCode)
	}
	return nil
}

func DeleteApplicationByNameNamespace(name, namespace string) error {
	applications, err := GetApplicationByNameNamespace(name, namespace)
	if err != nil {
		return err
	}

	id := applications[0]["id"].(string)

	err = DeleteApplicationById(id)
	return err
}
