package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Network represents a single network object in the response.
type Network struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	ProjectID string `json:"project_id"`
	//AdminStateUp            bool     `json:"admin_state_up"`
	//MTU                     int      `json:"mtu"`
	Shared            bool     `json:"shared"`
	RouterExternal    bool     `json:"router:external"`
	AvailabilityZones []string `json:"availability_zones"`
	//ProviderNetworkType     string   `json:"provider:network_type"`
	//ProviderPhysicalNetwork string   `json:"provider:physical_network"`
	PortSecurityEnabled bool     `json:"port_security_enabled"`
	SubnetIDs           []string `json:"subnets"`
	//PortIDs                 []string `json:"ports"`
}

// Subnet represents a single subnet object in the response.
type Subnet struct {
	//ID   string `json:"id"`
	Name string `json:"name"`
	CIDR string `json:"cidr"`
}

// AttachNetworkRequest represents the payload for attaching a network to a VM.
type AttachNetworkRequest struct {
	InterfaceAttachment struct {
		NetID    string   `json:"net_id,omitempty"`
		PortID   string   `json:"port_id,omitempty"`
		FixedIPs []IPInfo `json:"fixed_ips,omitempty"`
		//		Tag      string   `json:"tag,omitempty"`
	} `json:"interfaceAttachment"`
}

// IPInfo represents the structure for specifying fixed IPs.
type IPInfo struct {
	IPAddress string `json:"ip_address"`
}

// AttachNetworkResponse represents the response after attaching a network to a VM.
type AttachNetworkResponse struct {
	InterfaceAttachment struct {
		FixedIPs []struct {
			IPAddress string `json:"ip_address"`
			SubnetID  string `json:"subnet_id"`
		} `json:"fixed_ips"`
		MacAddr   string `json:"mac_addr"`
		NetID     string `json:"net_id"`
		PortID    string `json:"port_id"`
		PortState string `json:"port_state"`
		Tag       string `json:"tag,omitempty"`
	} `json:"interfaceAttachment"`
}

// ListNetworks fetches the list of networks available to the project.
func ListNetworks(baseURL, token string, queryParams map[string]string) (struct {
	Networks []Network `json:"networks"`
}, error) {
	//var result NetworkListResponse
	var result struct {
		Networks []Network `json:"networks"`
	}

	// Construct the request URL with query parameters.
	baseURL += "/v2.0/networks"
	if len(queryParams) > 0 {
		params := url.Values{}
		for key, value := range queryParams {
			params.Add(key, value)
		}
		baseURL += "?" + params.Encode()
	}

	// Send a GET request to fetch the networks.
	apiResp, err := callGET(baseURL, token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch networks: %v", err)
	}

	// Check for a successful response code.
	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("list networks request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	// Parse the JSON response.
	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse networks response: %v", err)
	}

	return result, nil
}

// AttachNetworkToVM attaches a network interface to a VM with optional parameters.
func AttachNetworkToVM(networkURL, computeURL, token, vmID, networkID, portID string, fixedIPs []string) (AttachNetworkResponse, error) {
	var result AttachNetworkResponse
	id, err := GetNetworkIDByName(networkURL, token, networkID)
	if err == nil {
		networkID = id
	}

	url := fmt.Sprintf("%s/servers/%s/os-interface", computeURL, vmID)

	request := AttachNetworkRequest{}
	if networkID != "" {
		request.InterfaceAttachment.NetID = networkID
	}
	if portID != "" {
		request.InterfaceAttachment.PortID = portID
	}

	if len(fixedIPs) > 0 {
		for _, ip := range fixedIPs {
			request.InterfaceAttachment.FixedIPs = append(request.InterfaceAttachment.FixedIPs, IPInfo{IPAddress: ip})
		}
	}

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return result, fmt.Errorf("failed to attach network: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("attach network request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse attach network response: %v", err)
	}

	return result, nil
}

// DetachNetworkFromVM detaches a network interface from a VM.
func DetachNetworkFromVM(networkURL, token, vmID, portID string) error {
	url := fmt.Sprintf("%s/servers/%s/os-interface/%s", networkURL, vmID, portID)

	apiResp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to detach network: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return fmt.Errorf("detach network request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// GetSubnetDetails fetches the details of a subnet by its ID.
func GetSubnetDetails(networkURL, token, subnetID string) (Subnet, error) {
	var wrapper struct {
		Subnet Subnet `json:"subnet"`
	}

	url := fmt.Sprintf("%s/v2.0/subnets/%s", networkURL, subnetID)

	apiResp, err := callGET(url, token)
	if err != nil {
		return wrapper.Subnet, fmt.Errorf("failed to fetch subnet details: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return wrapper.Subnet, fmt.Errorf("get subnet details request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &wrapper)
	if err != nil {
		return wrapper.Subnet, fmt.Errorf("failed to parse subnet details response: %v", err)
	}

	return wrapper.Subnet, nil
}

// GetNetworkIDByName fetches the ID of a network by its name.
func GetNetworkIDByName(networkURL, token, networkName string) (string, error) {
	if isUuid(networkName) {
		return networkName, nil
	}

	networks, err := ListNetworks(networkURL, token, nil)
	if err != nil {
		return "", err
	}

	foundNetworks := []Network{}
	for _, network := range networks.Networks {
		if strings.Contains(network.Name, networkName) {
			foundNetworks = append(foundNetworks, network)
		}
	}

	if len(foundNetworks) == 0 {
		return "", fmt.Errorf("no network found for name %s", networkName)
	}

	if len(foundNetworks) > 1 {
		return "", fmt.Errorf("multiple networks found for name %s", networkName)
	}

	return foundNetworks[0].ID, nil
}
