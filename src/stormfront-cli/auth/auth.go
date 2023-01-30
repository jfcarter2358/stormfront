package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type ClientInformation struct {
	ID              string `json:"id"`
	Secret          string `json:"secret"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	TokenExpiration string `json:"token_expiration"`
	TokenIssued     string `json:"token_issued"`
}

func getDataDirectory() string {
	dataDir := "/var/stormfront"
	return dataDir
}

func ReadClientInformation() ClientInformation {
	var clientInfo ClientInformation

	jsonFile, _ := os.Open(fmt.Sprintf("%s/auth.json", getDataDirectory()))

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &clientInfo)

	return clientInfo
}

func WriteClientInformation(clientInfo ClientInformation) error {
	clientData, _ := json.MarshalIndent(clientInfo, "", "    ")

	err := ioutil.WriteFile(fmt.Sprintf("%s/auth.json", getDataDirectory()), clientData, 0644)

	return err
}

func GetAPIToken(host, port string) (string, error) {
	requestURL := fmt.Sprintf("http://%s:%s/auth/api", host, port)

	clientInfo := ReadClientInformation()

	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", clientInfo.AccessToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	responseBody := string(body)

	if resp.StatusCode == http.StatusOK {
		responseJSON := map[string]string{}
		json.Unmarshal(body, &responseJSON)
		return responseJSON["token"], nil
	} else {
		var data map[string]string
		if err := json.Unmarshal([]byte(responseBody), &data); err == nil {
			if errMessage, ok := data["error"]; ok {
				return "", errors.New(errMessage)
			}
		}
		return "", fmt.Errorf("client has returned error with status code %v", resp.StatusCode)
	}
}
