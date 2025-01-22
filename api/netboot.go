package api

import (
	"fmt"
)

// UpdateNetworkInstallRequest represents the metadata update request
type UpdateNetworkInstallRequest struct {
	Metadata map[string]string `json:"metadata"`
}

// UpdateNetworkInstall sets the network_install metadata for a VM
func UpdateNetworkInstall(computeURL, token, vmID string, enabled bool) error {
	url := fmt.Sprintf("%s/servers/%s/metadata", computeURL, vmID)

	request := UpdateNetworkInstallRequest{
		Metadata: map[string]string{
			"network_install": fmt.Sprintf("%v", enabled),
		},
	}

	resp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to update network_install: %v", err)
	}

	if resp.ResponseCode != 200 {
		return fmt.Errorf("failed to update network_install: %s", resp.Response)
	}

	return nil
}
