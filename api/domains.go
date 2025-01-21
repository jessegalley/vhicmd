package api

import (
	"encoding/json"
	"fmt"
)

// DomainListResponse is the JSON structure returned by GET /v3/domains
type DomainListResponse struct {
	Domains []struct {
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		ID          string `json:"id"`
		Name        string `json:"name"`
	} `json:"domains"`
}

// ListDomains calls GET /v3/domains using the token for authentication.
func ListDomains(host, token string) (DomainListResponse, error) {
	var result DomainListResponse

	url := fmt.Sprintf("https://%s:5000/v3/domains", host)
	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to list domains: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("list domains failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("error unmarshalling domain list: %v", err)
	}

	return result, nil
}
