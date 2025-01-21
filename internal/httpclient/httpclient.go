package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

const requestTimeout = 5 * time.Second

const userAgent = "vhicmd v0.1"

// SendRequest sends a POST request with JSON data, a timeout, and a custom User-Agent.
// Use for authenticating.
func SendRequest(url string, jsonData []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{
		Timeout: requestTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	return resp, nil
}

// SendRequestWithToken can handle both GET and POST requests with a timeout and a custom User-Agent.
func SendRequestWithToken(method, url, token string, body io.Reader) (*http.Response, error) {
	// Read and reassign the body for logging if it's not nil
	//var bodyBytes []byte
	//if body != nil {
	//	bodyBytes, _ = io.ReadAll(body)
	//	body = bytes.NewReader(bodyBytes) // Rebuild the body for reuse
	//}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	if method == "POST" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if token != "" {
		req.Header.Set("X-Auth-Token", token)
	}

	// We need this header set for BlockDeviceMappingV2.VolumeType
	req.Header.Set("X-OpenStack-Nova-API-Version", "2.67")

	// Log the complete request
	//fmt.Printf("\nRequest:\n")
	//fmt.Printf("Method: %s\n", req.Method)
	//fmt.Printf("URL: %s\n", req.URL.String())
	//fmt.Println("Headers:")
	//for key, values := range req.Header {
	//	for _, value := range values {
	//		fmt.Printf("  %s: %s\n", key, value)
	//	}
	//}
	//if bodyBytes != nil {
	//	fmt.Println("Body:")
	//	var prettyJSON bytes.Buffer
	//	if json.Indent(&prettyJSON, bodyBytes, "", "  ") == nil {
	//		fmt.Println(prettyJSON.String())
	//	} else {
	//		fmt.Println(string(bodyBytes))
	//	}
	//}

	// Send the request
	client := &http.Client{Timeout: requestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	// Log the response body if available
	//if resp.Body != nil {
	//	defer resp.Body.Close()
	//	respBody, _ := io.ReadAll(resp.Body)
	//	fmt.Printf("\nResponse:\n")
	//	fmt.Printf("Status: %s\n", resp.Status)
	//	fmt.Println("Headers:")
	//	for key, values := range resp.Header {
	//		for _, value := range values {
	//			fmt.Printf("  %s: %s\n", key, value)
	//		}
	//	}
	//	fmt.Println("Body:")
	//	var prettyJSON bytes.Buffer
	//	if json.Indent(&prettyJSON, respBody, "", "  ") == nil {
	//		fmt.Println(prettyJSON.String())
	//	} else {
	//		fmt.Println(string(respBody))
	//	}
	//	// Restore the response body for further use
	//	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	//}

	return resp, nil
}
