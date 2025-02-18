package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/jessegalley/vhicmd/internal/httpclient"
	"github.com/spf13/viper"
)

// GetImageDetails fetches detailed information about a specific image
func GetImageDetails(imageURL, token, imageID string) (ImageDetails, error) {
	var result ImageDetails

	url := fmt.Sprintf("%s/v2/images/%s", imageURL, imageID)

	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to get image details: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("image details request failed [%d]: %s",
			apiResp.ResponseCode, apiResp.Response)
	}

	// For debugging
	if viper.GetBool("debug") {
		fmt.Printf("\nUnmarshaling response into struct: %s\n", apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse image details: %v", err)
	}

	// For debugging
	if viper.GetBool("debug") {
		fmt.Printf("\nParsed struct: %+v\n", result)
	}

	return result, nil
}

// ListImages fetches the list of images with optional filters and sorting.
func ListImages(imageURL, token string, queryParams map[string]string) (ImageListResponse, error) {
	var result ImageListResponse

	baseURL, err := url.Parse(fmt.Sprintf("%s/v2/images", imageURL))
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
func DeleteImage(imageURL, token, imageID string) error {
	url := fmt.Sprintf("%s/v2/images/%s", imageURL, imageID)

	apiResp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to delete image: %v", err)
	}
	if apiResp.ResponseCode != 204 {
		return fmt.Errorf("delete image request failed [%d]", apiResp.ResponseCode)
	}

	return nil
}

// CreateImage initiates image creation and returns the image ID
func createImage(imageURL, token string, req CreateImageRequest) (string, error) {
	var result Image
	url := fmt.Sprintf("%s/v2/images", imageURL)

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
func uploadImageData(imageURL, token, imageID string, data io.Reader) error {
	url := fmt.Sprintf("%s/v2/images/%s/file", imageURL, imageID)

	if viper.GetBool("debug") {
		fmt.Printf("Attempting upload to URL: %s\n", url)
	}

	resp, err := httpclient.UploadBigFile(url, token, data)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed [%d]: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// CreateAndUploadImage creates an image and uploads the image data
func CreateAndUploadImage(imageURL, token string, req CreateImageRequest, data io.Reader) (string, error) {
	debug := viper.GetBool("debug")

	if debug {
		fmt.Printf("Creating image entry in Glance...\n")
	}

	if req.DiskFmt == "" {
		return "", fmt.Errorf("disk_format must be specified")
	}
	if req.ContainerFmt == "" {
		return "", fmt.Errorf("container_format must be specified")
	}

	imageID, err := createImage(imageURL, token, req)
	if err != nil {
		return imageID, fmt.Errorf("failed to create image: %v", err)
	}

	if debug {
		fmt.Printf("Image entry created with ID: %s\n", imageID)
		fmt.Printf("Waiting for image to be ready for upload...\n")
	}

	// Wait for image to be in ready state
	if err := waitForImageReady(imageURL, token, imageID, debug); err != nil {
		return imageID, fmt.Errorf("image not ready: %v", err)
	}

	url := fmt.Sprintf("%s/v2/images/%s/file", imageURL, imageID)

	resp, err := httpclient.UploadBigFile(url, token, data)
	if err != nil {
		if debug {
			fmt.Printf("Upload failed, cleaning up image entry...\n")
		}
		_ = DeleteImage(imageURL, token, imageID)
		return imageID, fmt.Errorf("failed to upload image data: %v", err)
	}
	defer resp.Body.Close()

	if debug {
		fmt.Printf("Upload completed successfully\n")
	}

	return imageID, nil
}

// GetImageByName fetches the details of an image by its name.
// The image names are not unique, so this function returns the first image if only one is found,
// if multiple images or none are found, it returns an error.
func GetImageIDByName(imageURL, token, imageName string) (string, error) {
	if isUuid(imageName) {
		return imageName, nil
	}

	images, err := ListImages(imageURL, token, nil)
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
func GetImageNameByID(imageURL, token, imageID string) (string, error) {
	images, err := ListImages(imageURL, token, nil)
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
func GetImageByID(imageURL, token, imageID string) (Image, error) {
	images, err := ListImages(imageURL, token, nil)
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
func GetImageSize(imageURL, token, imageID string) (int64, error) {
	image, err := GetImageByID(imageURL, token, imageID)
	if err != nil {
		return 0, err
	}
	return image.Size, nil
}

// waitForImageReady polls the image API until the image is in a ready state
func waitForImageReady(imageURL, token, imageID string, debug bool) error {
	deadline := time.Now().Add(15 * time.Second)

	for time.Now().Before(deadline) {
		image, err := GetImageByID(imageURL, token, imageID)
		if err != nil {
			if debug {
				fmt.Printf("Error checking image status: %v\n", err)
			}
			return fmt.Errorf("failed to check image status: %v", err)
		}

		if debug {
			fmt.Printf("Image status: %s\n", image.Status)
		}

		switch image.Status {
		case "queued":
			// Ready for upload
			return nil
		case "error":
			return fmt.Errorf("image entered error state")
		}

		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("timeout waiting for image to be ready")
}

// UpdateImageVisibility updates the visibility of an image
func UpdateImageVisibility(imageURL, token, imageID, visibility string) error {
	url := fmt.Sprintf("%s/v2/images/%s", imageURL, imageID)

	request := UpdateImageVisibilityRequest{
		Visibility: visibility,
	}

	apiResp, err := callPATCH(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to update image visibility: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return fmt.Errorf("visibility update failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// ListImageMembers lists all members who have access to an image
func ListImageMembers(imageURL, token, imageID string) (ImageMemberList, error) {
	var result ImageMemberList

	url := fmt.Sprintf("%s/v2/images/%s/members", imageURL, imageID)

	apiResp, err := callGET(url, token)
	if err != nil {
		return result, fmt.Errorf("failed to list image members: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("list members request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("failed to parse member list response: %v", err)
	}

	return result, nil
}

// AddImageMember adds a project as a member to an image
func AddImageMember(imageURL, token, imageID, projectID string) error {
	url := fmt.Sprintf("%s/v2/images/%s/members", imageURL, imageID)

	request := AddImageMemberRequest{}
	request.Member.MemberID = projectID

	apiResp, err := callPOST(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to add image member: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return fmt.Errorf("add member request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// RemoveImageMember removes a project's membership from an image
func RemoveImageMember(imageURL, token, imageID, memberID string) error {
	url := fmt.Sprintf("%s/v2/images/%s/members/%s", imageURL, imageID, memberID)

	apiResp, err := callDELETE(url, token)
	if err != nil {
		return fmt.Errorf("failed to remove image member: %v", err)
	}

	if apiResp.ResponseCode != 204 {
		return fmt.Errorf("remove member request failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}

// UpdateImageMemberStatus updates the status of a member's access to an image
func UpdateImageMemberStatus(imageURL, token, imageID, memberID, status string) error {
	url := fmt.Sprintf("%s/v2/images/%s/members/%s", imageURL, imageID, memberID)

	request := UpdateImageMemberStatusRequest{
		Status: status,
	}

	apiResp, err := callPUT(url, token, request)
	if err != nil {
		return fmt.Errorf("failed to update member status: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return fmt.Errorf("update member status failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	return nil
}
