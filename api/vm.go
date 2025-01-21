package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// CreateVMRequest defines the payload structure for creating a VM.
type CreateVMRequest struct {
	Server struct {
		Name                 string                   `json:"name"`
		FlavorRef            string                   `json:"flavorRef"`
		ImageRef             string                   `json:"imageRef,omitempty"`
		Networks             []map[string]string      `json:"networks"`
		BlockDeviceMappingV2 []map[string]interface{} `json:"block_device_mapping_v2,omitempty"`
		KeyName              string                   `json:"key_name,omitempty"`
		AdminPass            string                   `json:"adminPass,omitempty"`
		Metadata             map[string]string        `json:"metadata,omitempty"`
		AvailabilityZone     string                   `json:"availability_zone,omitempty"`
		OSDCF                string                   `json:"OS-DCF:diskConfig,omitempty"`
	} `json:"server"`
}

// MarshalJSON implements the Payload interface for CreateVMRequest.
func (r CreateVMRequest) MarshalJSON() ([]byte, error) {
	type Alias CreateVMRequest // Avoid recursion by using an alias
	return json.Marshal(Alias(r))
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

type ServerImage struct {
	ID string `json:"id"`
}

// ImageField handles both string and object image fields
type ImageField struct {
	ServerImage
}

func (i *ImageField) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		i.ID = ""
		return nil
	}
	return json.Unmarshal(data, &i.ServerImage)
}

// NetworkAddress represents a single network interface address
type NetworkAddress struct {
	OSEXTIPSMACAddr string `json:"OS-EXT-IPS-MAC:mac_addr"`
	Version         int    `json:"version"`
	Addr            string `json:"addr"`
	OSEXTIPSType    string `json:"OS-EXT-IPS:type"`
	NetworkUUID     string // Will be populated from parent network info
}

// Basic VM struct used by List operation
type VM struct {
	ID         string                      `json:"id"`
	Name       string                      `json:"name"`
	Status     string                      `json:"status"`
	PowerState int                         `json:"OS-EXT-STS:power_state"`
	TaskState  string                      `json:"OS-EXT-STS:task_state"`
	Addresses  map[string][]NetworkAddress `json:"addresses"`
	Created    string                      `json:"created"`
	Updated    string                      `json:"updated,omitempty"`
	Progress   int                         `json:"progress,omitempty"`
	VMState    string                      `json:"OS-EXT-STS:vm_state"`
	Image      ImageField                  `json:"image"`
	Flavor     struct {
		ID string `json:"id"`
	} `json:"flavor"`
}

// Detailed VM struct used by Get operation
type VMDetail struct {
	ID         string                      `json:"id"`
	Name       string                      `json:"name"`
	Status     string                      `json:"status"`
	Host       string                      `json:"host,omitempty"`
	PowerState int                         `json:"OS-EXT-STS:power_state"`
	TaskState  string                      `json:"OS-EXT-STS:task_state"`
	Addresses  map[string][]NetworkAddress `json:"addresses"`
	Created    string                      `json:"created"`
	Updated    string                      `json:"updated,omitempty"`
	Progress   int                         `json:"progress,omitempty"`
	VMState    string                      `json:"OS-EXT-STS:vm_state"`
	Image      ImageField                  `json:"image"`
	Flavor     struct {
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
	OSExtendedVolumesVolumesAttached []Volume          `json:"os-extended-volumes:volumes_attached"`
	Metadata                         map[string]string `json:"metadata,omitempty"` // Add this field
}

// Support structs
type SecurityGroup struct {
	Name        string              `json:"name"`
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Rules       []SecurityGroupRule `json:"rules,omitempty"`
}

type SecurityGroupRule struct {
	ID             string `json:"id"`
	Direction      string `json:"direction"`      // ingress or egress
	Protocol       string `json:"protocol"`       // tcp, udp, icmp
	PortRangeMin   *int   `json:"port_range_min"` // pointer since it can be null
	PortRangeMax   *int   `json:"port_range_max"` // pointer since it can be null
	RemoteIPPrefix string `json:"remote_ip_prefix"`
	EtherType      string `json:"ethertype"` // IPv4 or IPv6
}

type Volume struct {
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

// VMListResponse represents the JSON structure for the list of VMs.
type VMListResponse struct {
	Servers []VM `json:"servers"`
}

// ActionRequest represents a power state action request
type ActionRequest struct {
	OsStop *struct{} `json:"os-stop,omitempty"`
}

func (r ActionRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OsStop *struct{} `json:"os-stop,omitempty"`
	}{r.OsStop})
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
		baseURL = baseURL[:len(baseURL)-1] // Remove trailing &
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
	// A simple check is if RAM, VCPUs, and Disk are all zero.
	if vm.Flavor.RAM == 0 && vm.Flavor.VCPUs == 0 && vm.Flavor.Disk == 0 {
		flavorID := vm.Flavor.ID
		if flavorID != "" {
			flv, err := GetFlavorDetails(computeURL, token, flavorID)
			if err == nil {
				vm.Flavor.RAM = flv.Flavor.RAM
				vm.Flavor.VCPUs = flv.Flavor.VCPUs
				vm.Flavor.Disk = flv.Flavor.Disk
				vm.Flavor.Ephemeral = flv.Flavor.Ephemeral
				// If needed, parse swap. For example, if flv.Flavor.Swap = "" then 0
				if flv.Flavor.Swap != "" {
					// optionally parse int, or just store as string
				}
				vm.Flavor.Swap = 0 // set to int 0 if that's your usage
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
	request := ActionRequest{
		OsStop: &struct{}{},
	}

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

// WaitForActive waits for a VM to become active or returns an error
func WaitForActive(computeURL, token, vmID string) (VMDetail, error) {
	maxAttempts := 30 // 5 minutes total (10 second intervals)
	attempts := 0

	for {
		if attempts >= maxAttempts {
			return VMDetail{}, fmt.Errorf("timeout waiting for VM to become active")
		}

		vmDetails, err := GetVMDetails(computeURL, token, vmID)
		if err != nil {
			return VMDetail{}, fmt.Errorf("failed to get VM details: %v", err)
		}

		if vmDetails.Status == "ERROR" {
			return VMDetail{}, fmt.Errorf("VM creation failed: status ERROR")
		}

		if vmDetails.Status == "ACTIVE" {
			return vmDetails, nil
		}

		time.Sleep(10 * time.Second)
		attempts++
	}
}
