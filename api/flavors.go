package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Flavor represents a single flavor object returned by the API.
type Flavor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Links       []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
}

// FlavorDetailResp represents the response for a single flavor detail request.
type FlavorDetailResp struct {
	Flavor struct {
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		RAM         int               `json:"ram"`
		Disk        int               `json:"disk"`
		VCPUs       int               `json:"vcpus"`
		Swap        string            `json:"swap"` // sometimes a string (""), sometimes an int
		RxTxFactor  float64           `json:"rxtx_factor"`
		IsPublic    bool              `json:"os-flavor-access:is_public"`
		Ephemeral   int               `json:"OS-FLV-EXT-DATA:ephemeral"`
		IsDisabled  bool              `json:"OS-FLV-DISABLED:disabled"`
		Description string            `json:"description,omitempty"` // microversion>=2.55
		ExtraSpecs  map[string]string `json:"extra_specs,omitempty"` // microversion>=2.61
	} `json:"flavor"`
}

// FlavorListResponse represents the response for listing flavors.
type FlavorListResponse struct {
	Flavors []Flavor `json:"flavors"`
}

// ListFlavors fetches the list of flavors from the stored compute URL
func ListFlavors(computeURL, token string, queryParams map[string]string) (FlavorListResponse, error) {
	var result FlavorListResponse

	url := fmt.Sprintf("%s/flavors", computeURL)

	if len(queryParams) > 0 {
		url += "?"
		for key, value := range queryParams {
			url += fmt.Sprintf("%s=%s&", key, value)
		}
		url = url[:len(url)-1]
	}

	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch flavors: %v", err)
	}
	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("flavors request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	if err := json.Unmarshal([]byte(apiResp.Response), &result); err != nil {
		return result, fmt.Errorf("failed to parse flavors response: %v", err)
	}
	return result, nil
}

// GetFlavorDetails fetches the details of a single flavor from the stored compute URL
func GetFlavorDetails(computeURL, token, flavorID string) (FlavorDetailResp, error) {
	var result FlavorDetailResp

	url := fmt.Sprintf("%s/flavors/%s", computeURL, flavorID)
	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to GET flavor: %v", err)
	}
	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("flavor details request failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}
	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse flavor details: %v", err)
	}
	return result, nil
}

// The flavor names are not unique, so this function a single flavor if only one is found,
// if multiple flavors or none are found, it returns an error.
func GetFlavorIDByName(computeURL, token, flavorName string) (string, error) {
	if isUuid(flavorName) {
		return flavorName, nil
	}

	flavors, err := ListFlavors(computeURL, token, nil)

	if err != nil {
		return "", err
	}

	var foundFlavors []Flavor

	for _, flavor := range flavors.Flavors {
		if strings.Contains(flavor.Name, flavorName) {
			foundFlavors = append(foundFlavors, flavor)
		}
	}

	if len(foundFlavors) == 0 {
		return "", fmt.Errorf("no flavors found for name %s", flavorName)
	}

	if len(foundFlavors) > 1 {
		return "", fmt.Errorf("multiple flavors found for name %s", flavorName)
	}

	return foundFlavors[0].ID, nil
}
