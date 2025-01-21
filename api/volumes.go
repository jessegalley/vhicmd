package api

import (
	"encoding/json"
	"fmt"
)

// VolumeStruct represents a block storage volume.
type VolumeStruct struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int    `json:"size"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// VolumeListResponse represents the response for listing volumes.
type VolumeListResponse struct {
	Volumes []VolumeStruct `json:"volumes"`
}

// CreateVolumeRequest represents the payload for creating a volume.
type CreateVolumeRequest struct {
	Volume struct {
		Name        string `json:"name"`
		Size        int    `json:"size"`
		Description string `json:"description,omitempty"`
		ImageRef    string `json:"imageRef,omitempty"`
		VolumeType  string `json:"volume_type,omitempty"`
	} `json:"volume"`
}

// MarshalJSON ensures CreateVolumeRequest implements the Payload interface.
func (r CreateVolumeRequest) MarshalJSON() ([]byte, error) {
	type Alias CreateVolumeRequest
	return json.Marshal(Alias(r))
}

// CreateVolumeResponse represents the response after creating a volume.
type CreateVolumeResponse struct {
	Volume VolumeStruct `json:"volume"`
}

// ListVolumes fetches the list of volumes.
func ListVolumes(storageURL, token string, queryParams map[string]string) (VolumeListResponse, error) {
	var result VolumeListResponse

	url := fmt.Sprintf("%s/volumes/detail", storageURL)
	if len(queryParams) > 0 {
		url += "?"
		for key, value := range queryParams {
			url += fmt.Sprintf("%s=%s&", key, value)
		}
		url = url[:len(url)-1]
	}

	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch volumes: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("list volumes request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse volumes response: %v", err)
	}

	return result, nil
}

// CreateVolume sends a request to create a new volume.
func CreateVolume(storageURL, token string, request CreateVolumeRequest) (CreateVolumeResponse, error) {
	var result CreateVolumeResponse

	url := fmt.Sprintf("%s/volumes", storageURL)

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return result, fmt.Errorf("failed to create volume: %v", err)
	}

	if apiResp.ResponseCode != 202 { // 202 Accepted
		return result, fmt.Errorf("volume creation request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse volume creation response: %v", err)
	}

	return result, nil
}
