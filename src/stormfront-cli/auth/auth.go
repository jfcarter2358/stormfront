package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
