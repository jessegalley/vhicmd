package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jessegalley/vhicmd/internal/httpclient"
)

// ApiResponse represents the response from an API call
type ApiResponse struct {
	TokenHeader  string
	ResponseCode int
	Response     string
}

// callPOST is a helper for POST requests. If you need to pass a token, supply it via the `token` parameter.
func callPOST(url, token string, body interface{}) (ApiResponse, error) {
	apiResp := ApiResponse{}

	// Marshal the request struct (whatever type it is) into JSON.
	jsonData, err := json.Marshal(body)
	if err != nil {
		return apiResp, fmt.Errorf("error marshaling JSON payload: %v", err)
	}

	resp, err := httpclient.SendRequestWithToken("POST", url, token, bytes.NewBuffer(jsonData))
	if err != nil {
		return apiResp, fmt.Errorf("error making HTTP POST request: %v", err)
	}
	defer resp.Body.Close()

	apiResp.ResponseCode = resp.StatusCode

	// Check if there's a token in the header
	if token := resp.Header.Get("X-Subject-Token"); token != "" {
		apiResp.TokenHeader = token
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResp, fmt.Errorf("error reading response body: %v", err)
	}
	apiResp.Response = string(bodyBytes)

	return apiResp, nil
}

// callGET is a helper for GET requests that requires a token in the X-Auth-Token header.
func callGET(url, token string) (ApiResponse, error) {
	apiResp := ApiResponse{}

	resp, err := httpclient.SendRequestWithToken("GET", url, token, nil)
	if err != nil {
		return apiResp, fmt.Errorf("error making HTTP GET request: %v", err)
	}
	defer resp.Body.Close()

	apiResp.ResponseCode = resp.StatusCode

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResp, fmt.Errorf("error reading response body: %v", err)
	}
	apiResp.Response = string(body)

	return apiResp, nil
}

// callDELETE is a helper for DELETE requests
func callDELETE(url, token string) (ApiResponse, error) {
	apiResp := ApiResponse{}

	resp, err := httpclient.SendRequestWithToken("DELETE", url, token, nil)
	if err != nil {
		return apiResp, fmt.Errorf("error making HTTP DELETE request: %v", err)
	}
	defer resp.Body.Close()

	apiResp.ResponseCode = resp.StatusCode

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResp, fmt.Errorf("error reading response body: %v", err)
	}
	apiResp.Response = string(body)

	return apiResp, nil
}

// callBigPUT is a helper for large binary PUT requests
func callBigPUT(url, token string, data io.Reader) (ApiResponse, error) {
	apiResp := ApiResponse{}

	resp, err := httpclient.SendLargePutRequest(url, token, data)
	if err != nil {
		return apiResp, fmt.Errorf("error making HTTP PUT request: %v", err)
	}
	defer resp.Body.Close()

	apiResp.ResponseCode = resp.StatusCode
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResp, fmt.Errorf("error reading response body: %v", err)
	}
	apiResp.Response = string(bodyBytes)

	return apiResp, nil
}
