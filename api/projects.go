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
		IsDomain bool   `json:"is_domain"`
		ParentID string `json:"parent_id"`
	} `json:"projects"`
}

// ListProjects calls GET /v3/projects using the token for authentication.
func ListProjects(host, token string) (ProjectListResponse, error) {
	var result ProjectListResponse

	url := fmt.Sprintf("https://%s:5000/v3/projects", host)
	apiResp, err := callGET(url, token)
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
