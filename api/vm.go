package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CreateVMRequest defines the payload structure for creating a VM.
type CreateVMRequest struct {
	Server struct {
		Name      string `json:"name"`
		FlavorRef string `json:"flavorRef"`
		ImageRef  string `json:"imageRef,omitempty"`
		//Networks             []map[string]string      `json:"networks"`
		Networks             string                   `json:"networks"` // for special "none" case
		BlockDeviceMappingV2 []map[string]interface{} `json:"block_device_mapping_v2,omitempty"`
		Metadata             map[string]string        `json:"metadata,omitempty"`
		UserData             string                   `json:"user_data,omitempty"`
	} `json:"server"`
}

// CreateVMResponse defines the structure of the response for creating a VM.
type CreateVMResponse struct {
	Server struct {
		ID    string `json:"id"`
		Links []struct {
			Href string `json:"href"`
			Rel  string `json:"rel"`
		} `json:"links"`
		AdminPass string `json:"adminPass,omitempty"`
	} `json:"server"`
}

// ServerImage for referencing an image ID
type ServerImage struct {
	ID string `json:"id"`
}

// ImageField can handle both string and object forms;
// we do custom unmarshal for tricky image fields (string vs. object).
type ImageField struct {
	ServerImage
}

func (i *ImageField) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		// If it was just a string, we don't have an ID for an object
		i.ID = ""
		return nil
	}
	// Otherwise unmarshal as image object
	return json.Unmarshal(data, &i.ServerImage)
}

// Basic VM struct used by List operation
type VM struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Detailed VM struct used by Get operation
type VMDetail struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	//	TenantID   string     `json:"tenant_id"`
	//	Host       string     `json:"host,omitempty"`
	PowerState int    `json:"OS-EXT-STS:power_state"`
	TaskState  string `json:"OS-EXT-STS:task_state"`
	Created    string `json:"created"`
	Updated    string `json:"updated,omitempty"`
	//	Progress   int        `json:"progress,omitempty"`
	//	VMState    string     `json:"OS-EXT-STS:vm_state"`
	Image  ImageField `json:"image"`
	Flavor struct {
		ID           string            `json:"id"`
		Ephemeral    int               `json:"ephemeral"`
		RAM          int               `json:"ram"`
		OriginalName string            `json:"original_name"`
		VCPUs        int               `json:"vcpus"`
		Swap         int               `json:"swap"`
		Disk         int               `json:"disk"`
		ExtraSpecs   map[string]string `json:"extra_specs"`
	} `json:"flavor"`
	SecurityGroups                   []SecurityGroup   `json:"security_groups"`
	HCIInfo                          HCIInfo           `json:"hci_info"`
	OSExtendedVolumesVolumesAttached []VmVolume        `json:"os-extended-volumes:volumes_attached"`
	Metadata                         map[string]string `json:"metadata,omitempty"`
}

type SecurityGroup struct {
	Name        string              `json:"name"`
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Rules       []SecurityGroupRule `json:"rules,omitempty"`
}

type SecurityGroupRule struct {
	ID             string `json:"id"`
	Direction      string `json:"direction"`
	Protocol       string `json:"protocol"`
	PortRangeMin   *int   `json:"port_range_min"`
	PortRangeMax   *int   `json:"port_range_max"`
	RemoteIPPrefix string `json:"remote_ip_prefix"`
	EtherType      string `json:"ethertype"`
}

type VmVolume struct {
	ID                  string `json:"id"`
	DeleteOnTermination bool   `json:"delete_on_termination"`
}

type NetworkInfo struct {
	Mac     string `json:"mac"`
	Network struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	} `json:"network"`
}

type HCIInfo struct {
	Network []NetworkInfo `json:"network"`
}

type VMNetworkListResponse struct {
	InterfaceAttachments []struct {
		PortState string `json:"port_state"`
		FixedIPs  []struct {
			IPAddress string `json:"ip_address"`
			SubnetID  string `json:"subnet_id"`
		} `json:"fixed_ips"`
		PortID  string `json:"port_id"`
		NetID   string `json:"net_id"`
		MacAddr string `json:"mac_addr"`
	} `json:"interfaceAttachments"`
}

// VMListResponse represents the JSON structure for the list of VMs.
type VMListResponse struct {
	Servers []VM `json:"servers"`
}

// ActionRequest is used for some actions like "os-stop"
type ActionRequest struct {
	OsStop *struct{} `json:"os-stop,omitempty"`
}

type RebootRequestPayload struct {
	Reboot struct {
		Type string `json:"type"`
	} `json:"reboot"`
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
