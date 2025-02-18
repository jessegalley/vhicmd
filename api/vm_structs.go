package api

import "encoding/json"

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

// SecurityGroup represents a security group.
type SecurityGroup struct {
	Name        string              `json:"name"`
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Rules       []SecurityGroupRule `json:"rules,omitempty"`
}

// SecurityGroupRule represents a security group rule.
type SecurityGroupRule struct {
	ID             string `json:"id"`
	Direction      string `json:"direction"`
	Protocol       string `json:"protocol"`
	PortRangeMin   *int   `json:"port_range_min"`
	PortRangeMax   *int   `json:"port_range_max"`
	RemoteIPPrefix string `json:"remote_ip_prefix"`
	EtherType      string `json:"ethertype"`
}

// VmVolume represents a volume attached to a VM.
type VmVolume struct {
	ID                  string `json:"id"`
	DeleteOnTermination bool   `json:"delete_on_termination"`
}

// NetworkInfo represents network information for a VM.
type NetworkInfo struct {
	Mac     string `json:"mac"`
	Network struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	} `json:"network"`
}

// HCIInfo represents HCI information for a VM.
type HCIInfo struct {
	Network []NetworkInfo `json:"network"`
}

// VMNetworkListResponse represents the JSON structure for the list of VM networks.
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

// RebootRequestPayload is used for rebooting a VM
type RebootRequestPayload struct {
	Reboot struct {
		Type string `json:"type"`
	} `json:"reboot"`
}

// UpdateVMRequest represents a request to update VM properties
type UpdateVMRequest struct {
	Server struct {
		Name        string            `json:"name,omitempty"`
		Description string            `json:"description,omitempty"`
		Metadata    map[string]string `json:"metadata,omitempty"`
	} `json:"server"`
}

// UpdateMetadataRequest represents a request to update VM metadata
type UpdateMetadataRequest struct {
	Metadata map[string]string `json:"metadata"`
}

// UpdateMetadataItemRequest represents a request to update a single metadata item
type UpdateMetadataItemRequest struct {
	Meta struct {
		Key   string `json:"-"` // Used in URL
		Value string `json:"value"`
	} `json:"meta"`
}
