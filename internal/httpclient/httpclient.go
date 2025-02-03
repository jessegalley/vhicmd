package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
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

	if viper.GetBool("debug") {
		printDebugDivider("request")
		fmt.Printf("\033[1;32mURL:\033[0m %s\n", url)
		fmt.Printf("\033[1;32mMethod:\033[0m %s\n", req.Method)
		printDebugDivider("request headers")
		printHeaders(req.Header)
		if len(jsonData) > 0 {
			printDebugDivider("request body")
			fmt.Println(prettyPrintJSON(jsonData))
		}
	}

	client := &http.Client{
		Timeout: requestTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	if viper.GetBool("debug") {
		printDebugDivider("response")
		fmt.Printf("\033[1;32mStatus:\033[0m %s\n", resp.Status)
		printDebugDivider("response headers")
		printHeaders(resp.Header)

		if resp.Body != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			if len(bodyBytes) > 0 {
				printDebugDivider("response body")
				fmt.Println(prettyPrintJSON(bodyBytes))
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	return resp, nil
}

// SendRequestWithToken can handle both GET and POST requests with a timeout and a custom User-Agent.
func SendRequestWithToken(method, url, token string, body io.Reader) (*http.Response, error) {
	// Read and reassign the body for logging if it's not nil
	var bodyBytes []byte
	if body != nil && viper.GetBool("debug") {
		bodyBytes, _ = io.ReadAll(body)
		body = bytes.NewReader(bodyBytes) // Rebuild the body for reuse
	}

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
	req.Header.Set("X-OpenStack-Nova-API-Version", "2.72")

	if viper.GetBool("debug") {
		printDebugDivider("request")
		fmt.Printf("\033[1;32mURL:\033[0m %s\n", url)
		fmt.Printf("\033[1;32mMethod:\033[0m %s\n", req.Method)
		printDebugDivider("request headers")
		printHeaders(req.Header)
		if len(bodyBytes) > 0 {
			printDebugDivider("request body")
			fmt.Println(prettyPrintJSON(bodyBytes))
		}
	}

	// Send the request
	client := &http.Client{Timeout: requestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	if viper.GetBool("debug") {
		printDebugDivider("response")
		fmt.Printf("\033[1;32mStatus:\033[0m %s\n", resp.Status)
		printDebugDivider("response headers")
		printHeaders(resp.Header)

		if resp.Body != nil {
			respBytes, _ := io.ReadAll(resp.Body)
			if len(respBytes) > 0 {
				printDebugDivider("response body")
				fmt.Println(prettyPrintJSON(respBytes))
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(respBytes))
		}
	}

	return resp, nil
}

// SendLargePutRequest sends a PUT req but uses io.Reader for large uploads.
func SendLargePutRequest(url, token string, data io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Auth-Token", token)

	// No timeout for large uploads
	client := &http.Client{Timeout: 0}

	if viper.GetBool("debug") {
		printDebugDivider("request")
		fmt.Printf("\033[1;32mURL:\033[0m %s\n", url)
		fmt.Printf("\033[1;32mMethod:\033[0m %s\n", req.Method)
		printDebugDivider("request headers")
		printHeaders(req.Header)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	if viper.GetBool("debug") {
		printDebugDivider("response")
		fmt.Printf("\033[1;32mStatus:\033[0m %s\n", resp.Status)
		printDebugDivider("response headers")
		printHeaders(resp.Header)
	}

	return resp, nil
}

// -- DEBUGGING --

// printDebugDivider prints a section divider
func printDebugDivider(title string) {
	fmt.Printf("\n\033[1;36m=== %s ===\033[0m\n", strings.ToUpper(title))
}

// prettyPrintJSON formats JSON with indentation
func prettyPrintJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "    ")
	if err != nil {
		return string(data) // Fallback to raw if not valid JSON
	}
	return prettyJSON.String()
}

// printHeaders pretty prints HTTP headers
func printHeaders(headers http.Header) {
	for key, values := range headers {
		fmt.Printf("    \033[1;33m%s\033[0m: %s\n", key, strings.Join(values, ", "))
	}
}
