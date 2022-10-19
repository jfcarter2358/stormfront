package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jfcarter2358/ceresdb-go/connection"
)

var AuthClients []ClientInformation
var APITokens []string

type ClientInformation struct {
	ID              string `json:"id"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	TokenExpiration string `json:"token_expiration"`
	TokenIssued     string `json:"token_issued"`
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenToken(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getDataDirectory() string {
	dataDir := "/var/stormfront"
	return dataDir
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func CreateClientInformation() ClientInformation {
	rand.Seed(time.Now().UnixNano())
	currentTime := time.Now()
	expiration := currentTime.Add(time.Hour * 6)

	clientInfo := ClientInformation{
		ID:              uuid.New().String(),
		AccessToken:     GenToken(128),
		RefreshToken:    GenToken(128),
		TokenExpiration: expiration.Format(time.RFC3339),
		TokenIssued:     currentTime.Format(time.RFC3339),
	}

	AuthClients = append(AuthClients, clientInfo)

	return clientInfo
}

func WriteClientInformation(clientInfo ClientInformation) error {
	clientData, _ := json.MarshalIndent(clientInfo, "", "    ")

	err := ioutil.WriteFile(fmt.Sprintf("%s/auth.json", getDataDirectory()), clientData, 0644)

	return err
}

func ReadClientInformation() ClientInformation {
	var clientInfo ClientInformation

	jsonFile, _ := os.Open(fmt.Sprintf("%s/auth.json", getDataDirectory()))

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &clientInfo)

	return clientInfo
}

func VerifyAccessToken(token string) int {
	data, err := connection.Query(fmt.Sprintf(`get record stormfront.auth | filter access_token = "%s"`, token))
	if err != nil {
		return http.StatusInternalServerError
	}
	if len(data) > 0 {
		issued := data[0]["token_issued"].(string)
		expiration := data[0]["token_expiration"].(string)
		start, _ := time.Parse(time.RFC3339, issued)
		end, _ := time.Parse(time.RFC3339, expiration)
		if !inTimeSpan(start, end, time.Now()) {
			return http.StatusNotAcceptable
		}
		return http.StatusOK
	}
	return http.StatusUnauthorized
}

func VerifyAPIToken(token string) int {
	for _, apiToken := range APITokens {
		if token == apiToken {
			return http.StatusOK
		}
	}
	return http.StatusUnauthorized
}

func RefreshClient(token string) (ClientInformation, error) {
	data, err := connection.Query(fmt.Sprintf(`get record stormfront.auth | filter refresh_token = "%s"`, token))
	if err != nil {
		return ClientInformation{}, fmt.Errorf("database error: %v", err)
	}
	if len(data) == 0 {
		return ClientInformation{}, fmt.Errorf("no auth information with refresh token exists")
	}
	rand.Seed(time.Now().UnixNano())
	currentTime := time.Now()
	expiration := currentTime.Add(time.Hour * 6)
	authClient := ClientInformation{}

	data[0]["access_token"] = GenToken(128)
	data[0]["refresh_token"] = GenToken(128)
	data[0]["token_issued"] = currentTime.Format(time.RFC3339)
	data[0]["token_expiration"] = expiration.Format(time.RFC3339)

	authData, _ := json.Marshal(data[0])

	json.Unmarshal(authData, &authClient)

	_, err = connection.Query(fmt.Sprintf(`put record stormfront.auth %s`, string(authData)))
	if err != nil {
		return ClientInformation{}, fmt.Errorf("database error: %v", err)
	}

	return authClient, nil
}
