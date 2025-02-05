package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Port represents a Neutron port
type Port struct {
	AdminStateUp          bool                   `json:"admin_state_up,omitempty"`
	AllowedAddressPairs   []AllowedAddressPair   `json:"allowed_address_pairs,omitempty"`
	BindingHostID         string                 `json:"binding:host_id,omitempty"`
	BindingProfile        map[string]interface{} `json:"binding:profile,omitempty"`
	BindingVifDetails     map[string]interface{} `json:"binding:vif_details,omitempty"`
	BindingVifType        string                 `json:"binding:vif_type,omitempty"`
	BindingVnicType       string                 `json:"binding:vnic_type,omitempty"`
	CreatedAt             string                 `json:"created_at,omitempty"`
	DataPlaneStatus       string                 `json:"data_plane_status,omitempty"`
	Description           string                 `json:"description,omitempty"`
	DeviceID              string                 `json:"device_id,omitempty"`
	DeviceOwner           string                 `json:"device_owner,omitempty"`
	DNSAssignment         []DNSAssignment        `json:"dns_assignment,omitempty"`
	DNSDomain             string                 `json:"dns_domain,omitempty"`
	DNSName               string                 `json:"dns_name,omitempty"`
	ExtraDHCPOpts         []ExtraDHCPOpt         `json:"extra_dhcp_opts,omitempty"`
	FixedIPs              []IPInfo               `json:"fixed_ips,omitempty"`
	ID                    string                 `json:"id,omitempty"`
	IPAllocation          string                 `json:"ip_allocation,omitempty"`
	MACAddress            string                 `json:"mac_address,omitempty"`
	Name                  string                 `json:"name,omitempty"`
	NetworkID             string                 `json:"network_id"`
	NumaAffinityPolicy    string                 `json:"numa_affinity_policy,omitempty"`
	PortSecurityEnabled   *bool                  `json:"port_security_enabled,omitempty"`
	ProjectID             string                 `json:"project_id,omitempty"`
	QoSNetworkPolicyID    string                 `json:"qos_network_policy_id,omitempty"`
	QoSPolicyID           string                 `json:"qos_policy_id,omitempty"`
	RevisionNumber        int                    `json:"revision_number,omitempty"`
	SecurityGroups        []string               `json:"security_groups,omitempty"`
	Status                string                 `json:"status,omitempty"`
	Tags                  []string               `json:"tags,omitempty"`
	TenantID              string                 `json:"tenant_id,omitempty"`
	UpdatedAt             string                 `json:"updated_at,omitempty"`
	PropagateUplinkStatus bool                   `json:"propagate_uplink_status,omitempty"`
	MacLearningEnabled    *bool                  `json:"mac_learning_enabled,omitempty"`
	PortTrustedVif        *bool                  `json:"port_trusted_vif,omitempty"`
}

type AllowedAddressPair struct {
	IPAddress  string `json:"ip_address"`
	MACAddress string `json:"mac_address,omitempty"`
}

type DNSAssignment struct {
	Hostname  string `json:"hostname"`
	IPAddress string `json:"ip_address"`
	FQDN      string `json:"fqdn"`
}

type ExtraDHCPOpt struct {
	OptValue  string `json:"opt_value"`
	IPVersion int    `json:"ip_version,omitempty"`
	OptName   string `json:"opt_name"`
}

// PortCreateRequest represents the request body for port creation
type PortCreateRequest struct {
	Port Port `json:"port"`
}

// PortCreateResponse represents the response from port creation
type PortCreateResponse struct {
	Port Port `json:"port"`
}

// PortListResponse represents the response for listing ports
type PortListResponse struct {
	Ports []Port `json:"ports"`
}

// CreatePort creates a new port with specified parameters
func CreatePort(baseURL, token string, networkID, macAddress string) (PortCreateResponse, error) {
	var result PortCreateResponse

	url := fmt.Sprintf("%s/v2.0/ports", baseURL)

	request := PortCreateRequest{
		Port: Port{
			NetworkID:  networkID,
			MACAddress: macAddress,
		},
	}

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return result, fmt.Errorf("failed to create port: %v", err)
	}

	if apiResp.ResponseCode != 201 {
		return result, fmt.Errorf("create port request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse create port response: %v", err)
	}

	return result, nil
}

// ListPorts fetches list of ports with optional query parameters
func ListPorts(baseURL, token string, queryParams map[string]string) (PortListResponse, error) {
	var result PortListResponse

	url := fmt.Sprintf("%s/v2.0/ports", baseURL)
	if len(queryParams) > 0 {
		url += "?"
		for key, value := range queryParams {
			url += fmt.Sprintf("%s=%s&", key, value)
		}
		url = strings.TrimSuffix(url, "&") // Remove trailing &
	}

	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to list ports: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("list ports request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse list ports response: %v", err)
	}

	return result, nil
}

// GetPortDetails fetches details of a specific port by ID
func GetPortDetails(baseURL, token, portID string) (Port, error) {
	var wrapper struct {
		Port Port `json:"port"`
	}

	url := fmt.Sprintf("%s/v2.0/ports/%s", baseURL, portID)

	apiResp, err := callGET(url, token)
	if err != nil {
		return wrapper.Port, fmt.Errorf("failed to fetch port details: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return wrapper.Port, fmt.Errorf("get port details request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &wrapper)
	if err != nil {
		return wrapper.Port, fmt.Errorf("failed to parse port details response: %v", err)
	}

	return wrapper.Port, nil
}

// DeletePort deletes a port by ID
func DeletePort(baseURL, token, portID string) error {
	url := fmt.Sprintf("%s/v2.0/ports/%s", baseURL, portID)

	apiResp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to delete port: %v", err)
	}

	if apiResp.ResponseCode != 204 {
		return fmt.Errorf("delete port request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}
