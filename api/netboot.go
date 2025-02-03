// note: this is likely headed for deprecated status
// netboot isn't 100% viable with VHI, at least
// not in the way it was originally intended

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

	//fmt.Printf("Full req, URL: %s, Token: %s, Request: %s\n", url, token,
	//	func() string {
	//		j, _ := json.Marshal(request)
	//		return string(j)
	//	}())

	resp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to update network_install: %v", err)
	}

	if resp.ResponseCode != 200 {
		return fmt.Errorf("failed to update network_install: %s", resp.Response)
	}

	fmt.Printf("Full resp, ResponseCode: %d, Response: %s\n", resp.ResponseCode, resp.Response)

	return nil
}
