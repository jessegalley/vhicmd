package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// Image represents a virtual machine image.
type Image struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Size    int64  `json:"size"`
	MinDisk int    `json:"min_disk"`
	MinRAM  int    `json:"min_ram"`
	Owner   string `json:"owner"`
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

	apiResp, err := callPOST(url, token, req)
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

// GetImageByName fetches the details of an image by its name.
// The image names are not unique, so this function returns the first image if only one is found,
// if multiple images or none are found, it returns an error.
func GetImageIDByName(computeURL, token, imageName string) (string, error) {
	images, err := ListImages(computeURL, token, nil)
	if err != nil {
		return "", err
	}

	foundImages := []Image{}
	for _, image := range images.Images {
		if strings.Contains(image.Name, imageName) {
			foundImages = append(foundImages, image)
		}
	}

	if len(foundImages) == 0 {
		return "", fmt.Errorf("no images found for name %s", imageName)
	}
	if len(foundImages) > 1 {
		return "", fmt.Errorf("multiple images found for name %s", imageName)
	}

	return foundImages[0].ID, nil
}

// GetImageNameByID fetches the name of an image by its ID.
func GetImageNameByID(computeURL, token, imageID string) (string, error) {
	images, err := ListImages(computeURL, token, nil)
	if err != nil {
		return "", err
	}
	imageName := ""
	for _, image := range images.Images {
		if image.ID == imageID {
			imageName = image.Name
			break
		}
	}
	if imageName == "" {
		return "", fmt.Errorf("no image found for ID %s", imageID)
	}
	return imageName, nil
}

// GetImageByID fetches the details of an image by its ID.
func GetImageByID(computeURL, token, imageID string) (Image, error) {
	images, err := ListImages(computeURL, token, nil)
	if err != nil {
		return Image{}, err
	}
	for _, image := range images.Images {
		if image.ID == imageID {
			return image, nil
		}
	}
	return Image{}, fmt.Errorf("no image found for ID %s", imageID)
}

// GetImageSize fetches the size of an image by its ID.
func GetImageSize(computeURL, token, imageID string) (int64, error) {
	image, err := GetImageByID(computeURL, token, imageID)
	if err != nil {
		return 0, err
	}
	return image.Size, nil
}
