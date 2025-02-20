package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// UpdateVM updates basic VM properties like name and description
func UpdateVM(computeURL, token, vmID string, request UpdateVMRequest) error {
	url := fmt.Sprintf("%s/servers/%s", computeURL, vmID)

	apiResp, err := callPUT(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to update VM: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return fmt.Errorf("update VM failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// UpdateVMName updates just the VM's name
func UpdateVMName(computeURL, token, vmID, newName string) error {
	request := UpdateVMRequest{}
	request.Server.Name = newName

	return UpdateVM(computeURL, token, vmID, request)
}

// UpdateVMMetadata updates all metadata for a VM
func UpdateVMMetadata(computeURL, token, vmID string, metadata map[string]string) error {
	url := fmt.Sprintf("%s/servers/%s/metadata", computeURL, vmID)

	request := UpdateMetadataRequest{
		Metadata: metadata,
	}

	apiResp, err := callPUT(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to update VM metadata: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return fmt.Errorf("update metadata failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// UpdateVMMetadataItem updates a single metadata item
func UpdateVMMetadataItem(computeURL, token, vmID, key, value string) error {
	url := fmt.Sprintf("%s/servers/%s/metadata/%s", computeURL, vmID, key)

	request := UpdateMetadataItemRequest{}
	request.Meta.Value = value

	apiResp, err := callPUT(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to update VM metadata item: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return fmt.Errorf("update metadata item failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// DeleteVMMetadataItem deletes a single metadata item
func DeleteVMMetadataItem(computeURL, token, vmID, key string) error {
	url := fmt.Sprintf("%s/servers/%s/metadata/%s", computeURL, vmID, key)

	apiResp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to delete VM metadata item: %v", err)
	}

	if apiResp.ResponseCode != 204 {
		return fmt.Errorf("delete metadata item failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// GetVMMetadata gets all metadata for a VM
func GetVMMetadata(computeURL, token, vmID string) (map[string]string, error) {
	var result struct {
		Metadata map[string]string `json:"metadata"`
	}

	url := fmt.Sprintf("%s/servers/%s/metadata", computeURL, vmID)

	apiResp, err := callGET(url, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM metadata: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return nil, fmt.Errorf("get metadata failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata response: %v", err)
	}

	return result.Metadata, nil
}

// ResizeVM changes the flavor of a VM (requires a subsequent confirm or revert)
func ResizeVM(computeURL, token, vmID, flavorID string) error {
	url := fmt.Sprintf("%s/servers/%s/action", computeURL, vmID)

	request := struct {
		Resize struct {
			FlavorRef string `json:"flavorRef"`
		} `json:"resize"`
	}{}
	request.Resize.FlavorRef = flavorID

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to resize VM: %v", err)
	}

	if apiResp.ResponseCode != 202 {
		return fmt.Errorf("resize failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// ConfirmResize confirms a VM resize operation
func ConfirmResize(computeURL, token, vmID string) error {
	url := fmt.Sprintf("%s/servers/%s/action", computeURL, vmID)

	request := struct {
		ConfirmResize *struct{} `json:"confirmResize"`
	}{
		ConfirmResize: &struct{}{},
	}

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to confirm resize: %v", err)
	}

	if apiResp.ResponseCode != 204 {
		return fmt.Errorf("confirm resize failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// RevertResize reverts a VM resize operation
func RevertResize(computeURL, token, vmID string) error {
	url := fmt.Sprintf("%s/servers/%s/action", computeURL, vmID)

	request := struct {
		RevertResize *struct{} `json:"revertResize"`
	}{
		RevertResize: &struct{}{},
	}

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to revert resize: %v", err)
	}

	if apiResp.ResponseCode != 202 {
		return fmt.Errorf("revert resize failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// ListVMs fetches the list of virtual machines.
func ListVMs(computeURL, token string, queryParams map[string]string) (VMListResponse, error) {
	var result VMListResponse

	baseURL := fmt.Sprintf("%s/servers", computeURL)
	if len(queryParams) > 0 {
		baseURL += "?"
		for key, value := range queryParams {
			baseURL += fmt.Sprintf("%s=%s&", key, value)
		}
		baseURL = strings.TrimSuffix(baseURL, "&")
	}

	apiResp, err := callGET(baseURL, token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch VMs: %v", err)
	}
	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("VM list request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse VM list response: %v", err)
	}
	return result, nil
}

// CreateVM sends a request to create a new VM using callPOST.
func CreateVM(computeURL, token string, request CreateVMRequest) (CreateVMResponse, error) {
	var result CreateVMResponse

	url := fmt.Sprintf("%s/servers", computeURL)

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return result, fmt.Errorf("failed to send VM create request: %v", err)
	}

	if apiResp.ResponseCode != 202 { // 202 Accepted
		return result, fmt.Errorf("VM create request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse VM create response: %v", err)
	}
	return result, nil
}

// GetVMNetworks fetches the list of networks attached to a VM.
func GetVMNetworks(computeURL, token, vmID string) (VMNetworkListResponse, error) {
	var result VMNetworkListResponse

	url := fmt.Sprintf("%s/servers/%s/os-interface", computeURL, vmID)

	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch VM networks: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("VM networks request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse VM networks response: %v", err)
	}
	return result, nil
}

// GetVMDetails fetches detailed information about a specific VM.
func GetVMDetails(computeURL, token, vmID string) (VMDetail, error) {
	var result VMDetail

	url := fmt.Sprintf("%s/servers/%s", computeURL, vmID)
	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch VM details: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("VM details request failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	var wrapper struct {
		Server VMDetail `json:"server"`
	}
	if err := json.Unmarshal([]byte(apiResp.Response), &wrapper); err != nil {
		return result, fmt.Errorf("failed to parse VM details: %v", err)
	}
	vm := wrapper.Server

	// If the flavor data is missing, do an extra GET /flavors/{id}.
	if vm.Flavor.RAM == 0 && vm.Flavor.VCPUs == 0 && vm.Flavor.Disk == 0 {
		flavorID := vm.Flavor.ID
		if flavorID != "" {
			flv, err := GetFlavorDetails(computeURL, token, flavorID)
			if err == nil {
				vm.Flavor.RAM = flv.Flavor.RAM
				vm.Flavor.VCPUs = flv.Flavor.VCPUs
				vm.Flavor.Disk = flv.Flavor.Disk
				vm.Flavor.Ephemeral = flv.Flavor.Ephemeral
				// If needed, handle flv.Flavor.Swap logic
				vm.Flavor.Swap = 0
				vm.Flavor.OriginalName = flv.Flavor.Name
				vm.Flavor.ExtraSpecs = flv.Flavor.ExtraSpecs
			}
		}
	}

	return vm, nil
}

// StopVM sends a request to stop a VM and waits for it to be fully stopped
func StopVM(computeURL, token, vmID string) error {
	url := fmt.Sprintf("%s/servers/%s/action", computeURL, vmID)

	// Send the stop request
	request := ActionRequest{OsStop: &struct{}{}}

	resp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to send stop request: %v", err)
	}
	if resp.ResponseCode != 202 {
		return fmt.Errorf("stop request failed: %s", resp.Response)
	}

	// Poll until shutdown complete or error
	maxAttempts := 30 // 5 minutes total (10 second intervals)
	attempts := 0
	for {
		if attempts >= maxAttempts {
			return fmt.Errorf("timeout waiting for VM to stop")
		}
		vmDetails, err := GetVMDetails(computeURL, token, vmID)
		if err != nil {
			return fmt.Errorf("failed to get VM details while stopping: %v", err)
		}
		if vmDetails.Status == "ERROR" {
			return fmt.Errorf("VM entered error state while stopping")
		}
		if vmDetails.Status == "SHUTOFF" {
			return nil
		}
		time.Sleep(10 * time.Second)
		attempts++
	}
}

// RebootVM sends a request to perform a reboot (HARD or SOFT) on a VM.
func RebootVM(computeURL, token, vmID string, rebootType string) error {
	// Default to SOFT if none specified
	if rebootType == "" {
		rebootType = "SOFT"
	}
	if rebootType != "HARD" && rebootType != "SOFT" {
		return fmt.Errorf("invalid reboot type: %s", rebootType)
	}

	url := fmt.Sprintf("%s/servers/%s/action", computeURL, vmID)
	var rebootRequest RebootRequestPayload
	rebootRequest.Reboot.Type = rebootType

	resp, err := callPOST(url, token, rebootRequest)
	if err != nil {
		return fmt.Errorf("failed to send reboot request: %v", err)
	}
	if resp.ResponseCode != 202 {
		return fmt.Errorf("reboot request failed: %s", resp.Response)
	}

	// Poll until the VM becomes ACTIVE or timeout
	maxAttempts := 30
	attempts := 0
	for {
		if attempts >= maxAttempts {
			return fmt.Errorf("timeout waiting for VM to reboot")
		}
		vmDetails, err := GetVMDetails(computeURL, token, vmID)
		if err != nil {
			return fmt.Errorf("failed to fetch VM details during reboot: %v", err)
		}
		if vmDetails.Status == "ERROR" {
			return fmt.Errorf("VM entered error state during reboot")
		}
		if vmDetails.Status == "ACTIVE" {
			return nil
		}
		time.Sleep(10 * time.Second)
		attempts++
	}
}

// WaitForStatus waits for a VM to reach a given status or returns error on timeout/error
func WaitForStatus(computeURL, token, vmID string, targetStatus string) (VMDetail, error) {
	maxAttempts := 30
	for attempts := 0; attempts < maxAttempts; attempts++ {
		vmDetails, err := GetVMDetails(computeURL, token, vmID)
		if err != nil {
			return VMDetail{}, fmt.Errorf("failed to get VM details: %v", err)
		}
		if strings.EqualFold(vmDetails.Status, "ERROR") {
			return VMDetail{}, fmt.Errorf("VM creation failed: status ERROR")
		}
		if strings.EqualFold(vmDetails.Status, targetStatus) {
			return vmDetails, nil
		}
		time.Sleep(10 * time.Second)
	}
	return VMDetail{}, fmt.Errorf("timeout waiting for VM to reach status %q", targetStatus)
}

// DeleteVM sends a request to delete a VM.
func DeleteVM(computeURL, token, vmID string) error {
	url := fmt.Sprintf("%s/servers/%s", computeURL, vmID)

	resp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %v", err)
	}
	if resp.ResponseCode != 204 {
		return fmt.Errorf("failed to delete VM [%d]: %s", resp.ResponseCode, resp.Response)
	}
	return nil
}

// GetVMIDByName fetches the ID of a VM by its name.
func GetVMIDByName(computeURL, token, vmName string) (string, error) {
	if isUuid(vmName) {
		return vmName, nil
	}

	vms, err := ListVMs(computeURL, token, nil)
	if err != nil {
		return "", err
	}

	var foundVMs []VM

	for _, vm := range vms.Servers {
		if strings.Contains(vm.Name, vmName) {
			foundVMs = append(foundVMs, vm)
		}
	}

	if len(foundVMs) == 0 {
		return "", fmt.Errorf("no VMs found for name %s", vmName)
	}

	if len(foundVMs) > 1 {
		return "", fmt.Errorf("multiple VMs found for name %s", vmName)
	}

	return foundVMs[0].ID, nil
}

// GetVMNameByID fetches the name of a VM by its ID.
func GetVMNameByID(computeURL, token, vmID string) (string, error) {
	vm, err := GetVMDetails(computeURL, token, vmID)
	if err != nil {
		return "", err
	}
	return vm.Name, nil
}
