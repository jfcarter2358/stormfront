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

func GetAllRoutes(namespace string) ([]map[string]interface{}, error) {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return []map[string]interface{}{}, nil
	}

	logging.Info("Getting routes...")

	requestURL := fmt.Sprintf("http://%s:%s/api/route", host, port)

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

func GetRouteById(id string) ([]map[string]interface{}, error) {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return []map[string]interface{}{}, nil
	}

	logging.Info(fmt.Sprintf("Getting route %s...", id))

	requestURL := fmt.Sprintf("http://%s:%s/api/route/%s", host, port, id)

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

func GetRouteByNameNamespace(name, namespace string) ([]map[string]interface{}, error) {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return []map[string]interface{}{}, nil
	}

	logging.Info("Getting routes...")

	requestURL := fmt.Sprintf("http://%s:%s/api/route", host, port)

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
		for _, route := range data {
			if route["name"].(string) == name {
				return []map[string]interface{}{route}, nil
			}
		}
		return []map[string]interface{}{}, fmt.Errorf("no route with name %s in namespace %s exists", name, namespace)
	}
	return []map[string]interface{}{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
}

func DeleteRouteById(id string) error {
	host, port, err := GetConnectionDetails()
	if err != nil {
		return nil
	}

	logging.Info(fmt.Sprintf("Deleting route %s...", id))

	requestURL := fmt.Sprintf("http://%s:%s/api/route/%s", host, port, id)

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

func DeleteRouteByNameNamespace(name, namespace string) error {
	routes, err := GetRouteByNameNamespace(name, namespace)
	if err != nil {
		return err
	}

	id := routes[0]["id"].(string)

	err = DeleteRouteById(id)
	return err
}
