package api

import (
	"encoding/json"
	"fmt"
	"time"
)

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

// GetVolumeDetails fetches detailed information about a specific volume
func GetVolumeDetails(storageURL, token, volumeID string) (VolumeDetail, error) {
	var wrapper struct {
		Volume VolumeDetail `json:"volume"`
	}

	url := fmt.Sprintf("%s/volumes/%s", storageURL, volumeID)

	apiResp, err := callGET(url, token)
	if err != nil {
		return wrapper.Volume, fmt.Errorf("failed to fetch volume details: %v", err)
	}
	if apiResp.ResponseCode != 200 {
		return wrapper.Volume, fmt.Errorf("get volume details request failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &wrapper)
	if err != nil {
		return wrapper.Volume, fmt.Errorf("failed to parse volume details: %v", err)
	}

	return wrapper.Volume, nil
}

func GetVolumeIDByName(storageURL, token, volumeName string) (string, error) {
	if isUuid(volumeName) {
		return volumeName, nil
	}
	qP := make(map[string]string)
	volumes, err := ListVolumes(storageURL, token, qP)
	if err != nil {
		return "", err
	}

	foundVolumes := []Volume{}
	for _, volume := range volumes.Volumes {
		if volume.Name == volumeName {
			foundVolumes = append(foundVolumes, volume)
		}
	}

	if len(foundVolumes) == 0 {
		return "", fmt.Errorf("no volumes found for name %s", volumeName)
	}
	if len(foundVolumes) > 1 {
		return "", fmt.Errorf("multiple volumes found for name %s", volumeName)
	}
	return foundVolumes[0].ID, nil
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

// DetachVolume sends a request to detach a volume from a VM.
func DetachVolume(computeURL, token, vmID, volumeID string) error {
	url := fmt.Sprintf("%s/servers/%s/os-volume_attachments/%s", computeURL, vmID, volumeID)

	apiResp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to detach volume: %v", err)
	}
	if apiResp.ResponseCode != 202 {
		return fmt.Errorf("volume detachment failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}
	return nil
}

// AttachVolumeRequest represents a request to attach a volume to a server
type AttachVolumeRequest struct {
	VolumeAttachment struct {
		VolumeID string `json:"volumeId"`
	} `json:"volumeAttachment"`
}

// AttachVolume attaches a volume to a VM. This is an asynchronous operation.
func AttachVolume(computeURL, token, vmID, volumeID string) error {
	url := fmt.Sprintf("%s/servers/%s/os-volume_attachments", computeURL, vmID)

	request := AttachVolumeRequest{}
	request.VolumeAttachment.VolumeID = volumeID

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to attach volume: %v", err)
	}
	if apiResp.ResponseCode != 200 {
		return fmt.Errorf("volume attachment failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}
	return nil
}

// CreateVolumeTransfer creates a new volume transfer
func CreateVolumeTransfer(storageURL, token, volumeID, name string) (VolumeTransfer, error) {
	var result CreateTransferResponse
	url := fmt.Sprintf("%s/volume-transfers", storageURL)

	request := CreateTransferRequest{}
	request.Transfer.VolumeID = volumeID
	request.Transfer.Name = name

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return result.Transfer, fmt.Errorf("failed to create volume transfer: %v", err)
	}

	if apiResp.ResponseCode != 202 {
		return result.Transfer, fmt.Errorf("create transfer failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result.Transfer, fmt.Errorf("failed to parse transfer response: %v", err)
	}

	return result.Transfer, nil
}

// AcceptVolumeTransfer accepts a volume transfer
func AcceptVolumeTransfer(storageURL, token, transferID, authKey string) error {
	url := fmt.Sprintf("%s/volume-transfers/%s/accept", storageURL, transferID)

	request := AcceptTransferRequest{}
	request.Accept.AuthKey = authKey

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to accept volume transfer: %v", err)
	}

	if apiResp.ResponseCode != 202 {
		return fmt.Errorf("accept transfer failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// DeleteVolumeTransfer deletes/cancels a volume transfer
func DeleteVolumeTransfer(storageURL, token, transferID string) error {
	url := fmt.Sprintf("%s/volume-transfers/%s", storageURL, transferID)

	apiResp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to delete volume transfer: %v", err)
	}

	if apiResp.ResponseCode != 202 {
		return fmt.Errorf("delete transfer failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// ListVolumeTypeAccess lists projects with access to a volume type
func ListVolumeTypeAccess(storageURL, token, volumeTypeID string) ([]VolumeTypeAccess, error) {
	var result VolumeTypeAccessList
	url := fmt.Sprintf("%s/types/%s/os-volume-type-access", storageURL, volumeTypeID)

	apiResp, err := callGET(url, token)
	if err != nil {
		return nil, fmt.Errorf("failed to list volume type access: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return nil, fmt.Errorf("list access failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access list: %v", err)
	}

	return result.VolumeTypeAccess, nil
}

// AddVolumeTypeAccess adds project access to a volume type
func AddVolumeTypeAccess(storageURL, token, volumeTypeID, projectID string) error {
	url := fmt.Sprintf("%s/types/%s/action", storageURL, volumeTypeID)

	request := AddProjectAccessRequest{}
	request.AddProjectAccess.Project = projectID

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to add volume type access: %v", err)
	}

	if apiResp.ResponseCode != 202 {
		return fmt.Errorf("add access failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// RemoveVolumeTypeAccess removes project access from a volume type
func RemoveVolumeTypeAccess(storageURL, token, volumeTypeID, projectID string) error {
	url := fmt.Sprintf("%s/types/%s/action", storageURL, volumeTypeID)

	request := struct {
		RemoveProjectAccess struct {
			Project string `json:"project"`
		} `json:"removeProjectAccess"`
	}{}
	request.RemoveProjectAccess.Project = projectID

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to remove volume type access: %v", err)
	}

	if apiResp.ResponseCode != 202 {
		return fmt.Errorf("remove access failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}
