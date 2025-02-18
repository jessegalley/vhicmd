package api

// Volume represents a block storage volume.
type Volume struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int    `json:"size"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// AttachVolumeRequest represents a request to attach a volume to a server
type AttachVolumeRequest struct {
	VolumeAttachment struct {
		VolumeID string `json:"volumeId"`
	} `json:"volumeAttachment"`
}

// VolumeListResponse represents the response for listing volumes.
type VolumeListResponse struct {
	Volumes []Volume `json:"volumes"`
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
	Volume Volume `json:"volume"`
}

// SetBootableRequest represents the request to set a volume bootable
type SetBootableRequest struct {
	OsSetBootable struct {
		Bootable bool `json:"bootable"`
	} `json:"os-set_bootable"`
}

// VolumeAttachment represents volume attachment information
type VolumeAttachment struct {
	ServerID     string `json:"server_id"`
	AttachmentID string `json:"attachment_id"`
	AttachedAt   string `json:"attached_at"`
	HostName     string `json:"host_name"`
	VolumeID     string `json:"volume_id"`
	Device       string `json:"device"`
	ID           string `json:"id"`
}

// Link represents a volume API link
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// VolumeDetail represents detailed information about a volume
type VolumeDetail struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Status             string             `json:"status"`
	Size               int                `json:"size"`
	AvailabilityZone   string             `json:"availability_zone"`
	CreatedAt          string             `json:"created_at"`
	UpdatedAt          string             `json:"updated_at"`
	Attachments        []VolumeAttachment `json:"attachments"`
	Description        string             `json:"description"`
	VolumeType         string             `json:"volume_type"`
	VolumeTypeID       string             `json:"volume_type_id"`
	SnapshotID         string             `json:"snapshot_id"`
	SourceVolID        string             `json:"source_volid"`
	Bootable           string             `json:"bootable"`
	Multiattach        bool               `json:"multiattach"`
	Encrypted          bool               `json:"encrypted"`
	ReplicationStatus  string             `json:"replication_status"`
	ConsistencyGroupID string             `json:"consistencygroup_id"`
	Metadata           map[string]string  `json:"metadata"`
	MigrationStatus    string             `json:"migration_status"`
	Links              []Link             `json:"links"`
	UserID             string             `json:"user_id"`
	ServiceUUID        string             `json:"service_uuid"`
	SharedTargets      bool               `json:"shared_targets"`
	ClusterName        string             `json:"cluster_name"`
	GroupID            string             `json:"group_id"`
	ProviderID         string             `json:"provider_id"`
	ConsumesQuota      bool               `json:"consumes_quota"`
}

// VolumeTransfer represents a volume transfer request
type VolumeTransfer struct {
	ID              string `json:"id"`
	CreatedAt       string `json:"created_at"`
	Name            string `json:"name"`
	VolumeID        string `json:"volume_id"`
	AuthKey         string `json:"auth_key,omitempty"`
	DestProjectID   string `json:"destination_project_id,omitempty"`
	SourceProjectID string `json:"source_project_id"`
	Expires         string `json:"expires_at,omitempty"`
}

// CreateTransferRequest represents the request to create a volume transfer
type CreateTransferRequest struct {
	Transfer struct {
		VolumeID string `json:"volume_id"`
		Name     string `json:"name,omitempty"`
	} `json:"transfer"`
}

// CreateTransferResponse represents the response from creating a transfer
type CreateTransferResponse struct {
	Transfer VolumeTransfer `json:"transfer"`
}

// AcceptTransferRequest represents the request to accept a volume transfer
type AcceptTransferRequest struct {
	Accept struct {
		AuthKey string `json:"auth_key"`
	} `json:"accept"`
}

// VolumeTypeAccess represents access to a volume type
type VolumeTypeAccess struct {
	ProjectID    string `json:"project_id"`
	VolumeTypeID string `json:"volume_type_id"`
}

// VolumeTypeAccessList represents a list of volume type access entries
type VolumeTypeAccessList struct {
	VolumeTypeAccess []VolumeTypeAccess `json:"volume_type_access"`
}

// AddProjectAccessRequest represents the request to add project access to a volume type
type AddProjectAccessRequest struct {
	AddProjectAccess struct {
		Project string `json:"project"`
	} `json:"addProjectAccess"`
}
