package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

var AuthClients []ClientInformation

type ClientInformation struct {
	ID              string `json:"id"`
	Secret          string `json:"secret"`
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dataDir := fmt.Sprintf("%s/.stormfront", homeDir)
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

func VerifyToken(token string) int {
	for _, authClient := range AuthClients {
		if token == authClient.AccessToken {
			start, _ := time.Parse(time.RFC3339, authClient.TokenIssued)
			end, _ := time.Parse(time.RFC3339, authClient.TokenExpiration)
			if !inTimeSpan(start, end, time.Now()) {
				return http.StatusNotAcceptable
			}
			return http.StatusOK
		}
	}
	return http.StatusUnauthorized
}

func RefreshClient(token string) (ClientInformation, error) {
	for idx, authClient := range AuthClients {
		if token == authClient.RefreshToken {
			rand.Seed(time.Now().UnixNano())
			currentTime := time.Now()
			expiration := currentTime.Add(time.Hour * 6)
			authClient.AccessToken = GenToken(128)
			authClient.RefreshToken = GenToken(128)
			authClient.TokenIssued = currentTime.Format(time.RFC3339)
			authClient.TokenExpiration = expiration.Format(time.RFC3339)
			AuthClients[idx] = authClient
			return authClient, nil
		}
	}
	return ClientInformation{}, fmt.Errorf("no client with refresh token exists")
}
