package api

import (
	"encoding/json"
	"fmt"
)

// CatalogResponse represents the response from the /v3/auth/catalog endpoint.
type CatalogResponse struct {
	Catalog []struct {
		Type      string `json:"type"`
		Name      string `json:"name"`
		Endpoints []struct {
			Region    string `json:"region"`
			Interface string `json:"interface"`
			URL       string `json:"url"`
		} `json:"endpoints"`
	} `json:"catalog"`
}

// GetCatalog fetches the service catalog from the Identity API.
func GetCatalog(host, token string) (CatalogResponse, error) {
	var result CatalogResponse

	// will use this to get all the other API endpoints
	url := fmt.Sprintf("https://%s:5000/v3/auth/catalog", host)

	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch service catalog: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("service catalog request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse service catalog response: %v", err)
	}

	return result, nil
}
