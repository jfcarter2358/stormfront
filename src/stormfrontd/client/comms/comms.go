package comms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"stormfrontd/client/auth"
)

func Get(host string, port int, path string, AuthClient auth.ClientInformation) (int, string, error) {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://%s:%v/%s", host, port, path)
	req, _ := http.NewRequest("GET", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.AccessToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return -1, "", err
	}
	if resp.StatusCode == http.StatusNotAcceptable {
		refreshURL := fmt.Sprintf("http://%s:%v/auth/refresh", host, port)
		refreshClient := &http.Client{}
		refreshReq, _ := http.NewRequest("GET", refreshURL, nil)
		refreshReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.RefreshToken))
		refreshResp, err := refreshClient.Do(refreshReq)
		if err != nil {
			fmt.Printf("Unable to contact client at %s:%v/api/health\n", host, port)
			return -1, "", err
		}
		defer refreshResp.Body.Close()
		//Read the response body
		body, err := ioutil.ReadAll(refreshResp.Body)
		if err != nil {
			return -1, "", err
		}
		refreshBody := string(body)
		json.Unmarshal([]byte(refreshBody), &AuthClient)
		auth.WriteClientInformation(AuthClient)
		// resent the request with the new access token
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.AccessToken))
		resp, err = httpClient.Do(req)
		if err != nil {
			return -1, "", err
		}
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, "", err
	}
	responseBody := string(body)

	return resp.StatusCode, responseBody, nil
}

func Delete(host string, port int, path string, AuthClient auth.ClientInformation) (int, string, error) {
	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://%s:%v/%s", host, port, path)
	req, _ := http.NewRequest("DELETE", requestURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.AccessToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return -1, "", err
	}
	if resp.StatusCode == http.StatusNotAcceptable {
		refreshURL := fmt.Sprintf("http://%s:%v/auth/refresh", host, port)
		refreshClient := &http.Client{}
		refreshReq, _ := http.NewRequest("GET", refreshURL, nil)
		refreshReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.RefreshToken))
		refreshResp, err := refreshClient.Do(refreshReq)
		if err != nil {
			fmt.Printf("Unable to contact client at %s:%v/api/health\n", host, port)
			return -1, "", err
		}
		defer refreshResp.Body.Close()
		//Read the response body
		body, err := ioutil.ReadAll(refreshResp.Body)
		if err != nil {
			return -1, "", err
		}
		refreshBody := string(body)
		json.Unmarshal([]byte(refreshBody), &AuthClient)
		auth.WriteClientInformation(AuthClient)
		// resent the request with the new access token
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.AccessToken))
		resp, err = httpClient.Do(req)
		if err != nil {
			return -1, "", err
		}
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, "", err
	}
	responseBody := string(body)

	return resp.StatusCode, responseBody, nil
}

func Post(host string, port int, path string, AuthClient auth.ClientInformation, postBody []byte) (int, string, error) {
	postBodyBuffer := bytes.NewBuffer(postBody)

	httpClient := &http.Client{}
	requestURL := fmt.Sprintf("http://%s:%v/%s", host, port, path)
	req, _ := http.NewRequest("POST", requestURL, postBodyBuffer)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.AccessToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		return -1, "", err
	}
	if resp.StatusCode == http.StatusNotAcceptable {
		refreshURL := fmt.Sprintf("http://%s:%v/auth/refresh", host, port)
		refreshClient := &http.Client{}
		refreshReq, _ := http.NewRequest("GET", refreshURL, nil)
		refreshReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.RefreshToken))
		refreshResp, err := refreshClient.Do(refreshReq)
		if err != nil {
			fmt.Printf("Unable to contact client at %s:%v/api/health\n", host, port)
			return -1, "", err
		}
		defer refreshResp.Body.Close()
		//Read the response body
		body, err := ioutil.ReadAll(refreshResp.Body)
		if err != nil {
			return -1, "", err
		}
		refreshBody := string(body)
		json.Unmarshal([]byte(refreshBody), &AuthClient)
		auth.WriteClientInformation(AuthClient)
		// resent the request with the new access token
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AuthClient.AccessToken))
		resp, err = httpClient.Do(req)
		if err != nil {
			return -1, "", err
		}
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, "", err
	}
	responseBody := string(body)

	return resp.StatusCode, responseBody, nil
}
