package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Image represents a virtual machine image.
type Image struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	Visibility    string   `json:"visibility"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	Size          int64    `json:"size"`
	MinDisk       int      `json:"min_disk"`
	MinRAM        int      `json:"min_ram"`
	Owner         string   `json:"owner"`
	ContainerFmt  string   `json:"container_format"`
	DiskFmt       string   `json:"disk_format"`
	Tags          []string `json:"tags"`
	Protected     bool     `json:"protected"`
	DirectURL     string   `json:"direct_url,omitempty"`
	TraitRequired string   `json:"trait:CUSTOM_HCI_122E856B9E9C4D80A0F8C21591B5AFCB,omitempty"`
}

// ImageListResponse is the structure for the API response.
type ImageListResponse struct {
	Images []Image `json:"images"`
	Schema string  `json:"schema"`
	First  string  `json:"first"`
	Next   string  `json:"next,omitempty"`
}

// ListImages fetches the list of images with optional filters and sorting.
func ListImages(computeURL, token string, queryParams map[string]string) (ImageListResponse, error) {
	var result ImageListResponse

	baseURL, err := url.Parse(fmt.Sprintf("%s/v2/images", computeURL))
	if err != nil {
		return result, fmt.Errorf("failed to parse URL: %v", err)
	}

	query := baseURL.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	baseURL.RawQuery = query.Encode()

	apiResp, err := callGET(baseURL.String(), token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch images: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("image list request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse image list response: %v", err)
	}

	return result, nil
}
