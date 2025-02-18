package api

import (
	"encoding/json"
	"fmt"
)

// ProjectListResponse is the JSON structure returned by GET /v3/projects
type ProjectListResponse struct {
	Projects []struct {
		ID       string `json:"id"`
		DomainID string `json:"domain_id"`
		Name     string `json:"name"`
		Enabled  bool   `json:"enabled"`
	} `json:"projects"`
}

// ListProjects calls GET /v3/projects using the token for authentication.
func ListProjects(identityUrl, token string) (ProjectListResponse, error) {
	var result ProjectListResponse

	apiResp, err := callGET(fmt.Sprintf("%s/projects", identityUrl), token)
	if err != nil {
		return result, fmt.Errorf("failed to list projects: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return result, fmt.Errorf("list projects failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return result, fmt.Errorf("error unmarshalling project list: %v", err)
	}

	return result, nil
}

// Get project Name by ID
func GetProjectNameByID(identityUrl, token, projectID string) (string, error) {
	projectList, err := ListProjects(identityUrl, token)
	if err != nil {
		return "", fmt.Errorf("failed to get project name: %v", err)
	}

	for _, project := range projectList.Projects {
		if project.ID == projectID {
			return project.Name, nil
		}
	}

	return "", fmt.Errorf("no project found for ID %s", projectID)
}

// Get project ID by Name
func GetProjectIDByName(identityUrl, token, projectName string) (string, error) {
	var result ProjectListResponse

	apiResp, err := callGET(fmt.Sprintf("%s/projects?name=%s", identityUrl, projectName), token)
	if err != nil {
		return "", fmt.Errorf("failed to get project ID: %v", err)
	}

	if apiResp.ResponseCode != 200 {
		return "", fmt.Errorf("get project ID failed [%d]: %s", apiResp.ResponseCode, apiResp.Response)
	}

	err = json.Unmarshal([]byte(apiResp.Response), &result)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling project list: %v", err)
	}

	if len(result.Projects) == 0 {
		return "", fmt.Errorf("no project found for name %s", projectName)
	}

	return result.Projects[0].ID, nil
}
