package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
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

	if token := resp.Header.Get("X-Subject-Token"); token != "" {
		apiResp.TokenHeader = token
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResp, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode >= 400 {
		apiResp.Response = FormatErrorResponse(string(bodyBytes))
	} else {
		apiResp.Response = string(bodyBytes)
	}

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

	if resp.StatusCode >= 400 {
		// Format error responses
		apiResp.Response = FormatErrorResponse(string(body))
	} else {
		apiResp.Response = string(body)
	}

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

	if resp.StatusCode >= 400 {
		apiResp.Response = FormatErrorResponse(string(body))
	} else {
		apiResp.Response = string(body)
	}

	return apiResp, nil
}

// callPATCH helper for PATCH requests
func callPATCH(url, token string, body interface{}) (ApiResponse, error) {
	apiResp := ApiResponse{}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return apiResp, fmt.Errorf("error marshaling JSON payload: %v", err)
	}

	resp, err := httpclient.SendRequestWithToken("PATCH", url, token, bytes.NewBuffer(jsonData))
	if err != nil {
		return apiResp, fmt.Errorf("error making HTTP PATCH request: %v", err)
	}
	defer resp.Body.Close()

	apiResp.ResponseCode = resp.StatusCode

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResp, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode >= 400 {
		apiResp.Response = FormatErrorResponse(string(bodyBytes))
	} else {
		apiResp.Response = string(bodyBytes)
	}

	return apiResp, nil
}

// callPUT helper for PUT requests
func callPUT(url, token string, body interface{}) (ApiResponse, error) {
	apiResp := ApiResponse{}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return apiResp, fmt.Errorf("error marshaling JSON payload: %v", err)
	}

	resp, err := httpclient.SendRequestWithToken("PUT", url, token, bytes.NewBuffer(jsonData))
	if err != nil {
		return apiResp, fmt.Errorf("error making HTTP PUT request: %v", err)
	}
	defer resp.Body.Close()

	apiResp.ResponseCode = resp.StatusCode

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResp, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode >= 400 {
		apiResp.Response = FormatErrorResponse(string(bodyBytes))
	} else {
		apiResp.Response = string(bodyBytes)
	}

	return apiResp, nil
}

func isUuid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
