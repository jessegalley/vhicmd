package api

import (
	"encoding/json"
	"fmt"
	"time"
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

// CreateVolumeResponse represents the response after creating a volume.
type CreateVolumeResponse struct {
	Volume VolumeStruct `json:"volume"`
}

type SetBootableRequest struct {
	OsSetBootable struct {
		Bootable bool `json:"bootable"`
	} `json:"os-set_bootable"`
}

// SetVolumeBootable sets a volume’s bootable flag
func SetVolumeBootable(storageURL, token, volumeID string, bootable bool) error {
	url := fmt.Sprintf("%s/volumes/%s/action", storageURL, volumeID)

	request := SetBootableRequest{}
	request.OsSetBootable.Bootable = bootable

	resp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to set bootable flag: %v", err)
	}
	if resp.ResponseCode != 200 {
		return fmt.Errorf("failed to set bootable flag [%d]: %s", resp.ResponseCode, resp.Response)
	}
	return nil
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
	if apiResp.ResponseCode != 202 {
		return result, fmt.Errorf("volume creation request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse volume creation response: %v", err)
	}
	return result, nil
}

// DeleteVolume sends a request to delete a volume.
func DeleteVolume(storageURL, token, volumeID string) error {
	url := fmt.Sprintf("%s/volumes/%s", storageURL, volumeID)

	resp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %v", err)
	}
	if resp.ResponseCode != 204 {
		return fmt.Errorf("failed to delete volume [%d]: %s", resp.ResponseCode, resp.Response)
	}
	return nil
}

// WaitForVolumeStatus polls volume status until it matches target or times out
func WaitForVolumeStatus(storageURL, token, volumeID, targetStatus string) error {
	maxAttempts := 30 // ~5 minutes with 10s intervals
	for i := 0; i < maxAttempts; i++ {
		resp, err := ListVolumes(storageURL, token, map[string]string{"id": volumeID})
		if err != nil {
			return fmt.Errorf("failed to get volume status: %v", err)
		}
		if len(resp.Volumes) == 0 {
			return fmt.Errorf("volume %s not found", volumeID)
		}
		status := resp.Volumes[0].Status
		if status == targetStatus {
			return nil
		}
		if status == "error" {
			return fmt.Errorf("volume entered error state while waiting for %s", targetStatus)
		}
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("timeout waiting for volume to become %s", targetStatus)
}
