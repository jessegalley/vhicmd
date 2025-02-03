package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Image represents a virtual machine image.
type Image struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	Visibility    string   `json:"visibility"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	Size          int64    `json:"size"`
	MinDisk       int      `json:"min_disk"`
	MinRAM        int      `json:"min_ram"`
	Owner         string   `json:"owner"`
	ContainerFmt  string   `json:"container_format"`
	DiskFmt       string   `json:"disk_format"`
	Tags          []string `json:"tags"`
	Protected     bool     `json:"protected"`
	DirectURL     string   `json:"direct_url,omitempty"`
	TraitRequired string   `json:"trait:CUSTOM_HCI_122E856B9E9C4D80A0F8C21591B5AFCB,omitempty"`
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

// ListImages fetches the list of images with optional filters and sorting.
func ListImages(computeURL, token string, queryParams map[string]string) (ImageListResponse, error) {
	var result ImageListResponse

	baseURL, err := url.Parse(fmt.Sprintf("%s/v2/images", computeURL))
	if err != nil {
		return result, fmt.Errorf("failed to parse URL: %v", err)
	}

	query := baseURL.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	baseURL.RawQuery = query.Encode()

	apiResp, err := callGET(baseURL.String(), token)
	if err != nil {
		return result, fmt.Errorf("failed to fetch images: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("image list request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse image list response: %v", err)
	}

	return result, nil
}

// DeleteImage deletes an image by ID.
func DeleteImage(computeURL, token, imageID string) error {
	url := fmt.Sprintf("%s/v2/images/%s", computeURL, imageID)

	apiResp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to delete image: %v", err)
	}
	if apiResp.ResponseCode != 204 {
		return fmt.Errorf("delete image request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// CreateImage initiates image creation and returns the image ID
func createImage(computeURL, token string, req CreateImageRequest) (string, error) {
	var result Image
	url := fmt.Sprintf("%s/v2/images", computeURL)

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %v", err)
	}

	apiResp, err := callPOST(url, token, string(body))
	if err != nil {
		return "", fmt.Errorf("create failed: %v", err)
	}

	if apiResp.ResponseCode != 201 {
		return "", fmt.Errorf("create failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return "", fmt.Errorf("unmarshal failed: %v", err)
	}

	return result.ID, nil
}

// UploadImageData uploads the actual image data
func uploadImageData(computeURL, token, imageID string, data io.Reader) error {
	url := fmt.Sprintf("%s/v2/images/%s/file", computeURL, imageID)
	apiResp, err := callBigPUT(url, token, data)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	if apiResp.ResponseCode != 204 {
		return fmt.Errorf("upload failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// CreateAndUploadImage creates an image and uploads the image data
func CreateAndUploadImage(computeURL, token string, req CreateImageRequest, data io.Reader) (string, error) {
	imageID, err := createImage(computeURL, token, req)
	if err != nil {
		return imageID, fmt.Errorf("failed to create image: %v", err)
	}

	err = uploadImageData(computeURL, token, imageID, data)
	if err != nil {
		return imageID, fmt.Errorf("failed to upload image data: %v", err)
	}

	return imageID, nil
}
