package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Network represents a single network object in the response.
type Network struct {
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	Status                  string   `json:"status"`
	ProjectID               string   `json:"project_id"`
	AdminStateUp            bool     `json:"admin_state_up"`
	MTU                     int      `json:"mtu"`
	Shared                  bool     `json:"shared"`
	RouterExternal          bool     `json:"router:external"`
	AvailabilityZones       []string `json:"availability_zones"`
	ProviderNetworkType     string   `json:"provider:network_type"`
	ProviderPhysicalNetwork string   `json:"provider:physical_network"`
}

// NetworkListResponse represents the response for listing networks.
type NetworkListResponse struct {
	Networks []Network `json:"networks"`
}

// ListNetworks fetches the list of networks available to the project.
func ListNetworks(baseURL, token string, queryParams map[string]string) (NetworkListResponse, error) {
	var result NetworkListResponse

	// Construct the request URL with query parameters.
	baseURL += "/v2.0/networks"
	if len(queryParams) > 0 {
		params := url.Values{}
		for key, value := range queryParams {
			params.Add(key, value)
		}
		baseURL += "?" + params.Encode()
	}

	// Send a GET request to fetch the networks.
	apiResp, err := callGET(baseURL, token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch networks: %v", err)
	}

	// Check for a successful response code.
	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("list networks request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	// Parse the JSON response.
	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse networks response: %v", err)
	}

	return result, nil
}
