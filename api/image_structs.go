package api

// ImageMember represents a member who has access to an image
type ImageMember struct {
	MemberID  string `json:"member_id"` // Project ID of the member
	Status    string `json:"status"`    // Can be pending, accepted, rejected
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	ImageID   string `json:"image_id"`
	SchemaURI string `json:"schema"`
}

// ImageMemberList represents the response for listing image members
type ImageMemberList struct {
	Members []ImageMember `json:"members"`
	Schema  string        `json:"schema"`
}

// UpdateImageVisibilityRequest represents the request to update image visibility
type UpdateImageVisibilityRequest struct {
	Visibility string `json:"visibility"` // private, public, shared, community
}

// UpdateImageMemberStatusRequest represents the request to update member status
type UpdateImageMemberStatusRequest struct {
	Status string `json:"status"` // accepted, rejected, pending
}

// AddImageMemberRequest represents the request to add a member
type AddImageMemberRequest struct {
	Member struct {
		MemberID string `json:"member_id"`
	} `json:"member"`
}

// Image represents a virtual machine image.
type Image struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Size       int64  `json:"size"`
	MinDisk    int    `json:"min_disk"`
	MinRAM     int    `json:"min_ram"`
	Owner      string `json:"owner"`
	Visibility string `json:"visibility"`
}

type CreateImageRequest struct {
	Name         string   `json:"name"`
	ContainerFmt string   `json:"container_format"` // bare, ovf, ova, aki, ari, ami
	DiskFmt      string   `json:"disk_format"`      // raw, vhd, vmdk, vdi, iso, qcow2, aki, ari, ami
	MinDisk      int      `json:"min_disk,omitempty"`
	MinRAM       int      `json:"min_ram,omitempty"`
	Protected    bool     `json:"protected,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Visibility   string   `json:"visibility,omitempty"`
}

// ImageListResponse is the structure for the API response.
type ImageListResponse struct {
	Images []Image `json:"images"`
	Schema string  `json:"schema"`
	First  string  `json:"first"`
	Next   string  `json:"next,omitempty"`
}

type ImageDetails struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	Visibility       string   `json:"visibility"`
	Size             int64    `json:"size"`
	VirtualSize      int64    `json:"virtual_size"`
	MinDisk          int      `json:"min_disk"`
	MinRAM           int      `json:"min_ram"`
	DiskFormat       string   `json:"disk_format"`
	ContainerFormat  string   `json:"container_format"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
	Protected        bool     `json:"protected"`
	Checksum         string   `json:"checksum"`
	OsHashAlgo       string   `json:"os_hash_algo"`
	OsHashValue      string   `json:"os_hash_value"`
	OsHidden         bool     `json:"os_hidden"`
	Owner            string   `json:"owner"`
	Tags             []string `json:"tags"`
	DirectURL        string   `json:"direct_url"`
	File             string   `json:"file"`
	Self             string   `json:"self"`
	Schema           string   `json:"schema"`
	HwQemuGuestAgent string   `json:"hw_qemu_guest_agent"`
	OsType           string   `json:"os_type"`
	OsDistro         string   `json:"os_distro"`
	ImageValidated   string   `json:"image_validated"`
}
