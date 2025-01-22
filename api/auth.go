package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var TokenFile string

// TokenStore structure to store tokens per host
type TokenStore struct {
	Tokens map[string]Token `json:"tokens"` // map[hostname]Token
}

// Token structure to store the token, its expiration, the host, and the compute URL
type Token struct {
	Value     string            `json:"value"`
	ExpiresAt time.Time         `json:"expires_at"`
	Host      string            `json:"host"`
	Endpoints map[string]string `json:"endpoints,omitempty"`
}

// AuthPayload is used for the authentication request body
type AuthPayload struct {
	Auth Auth `json:"auth"`
}

type Auth struct {
	Identity Identity `json:"identity"`
	Scope    Scope    `json:"scope"`
}

type Identity struct {
	Methods  []string `json:"methods"`
	Password Password `json:"password"`
}

type Password struct {
	User User `json:"user"`
}

type User struct {
	Name     string `json:"name"`
	Domain   Domain `json:"domain"`
	Password string `json:"password"`
}

type Domain struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Scope struct {
	Project Project `json:"project"`
}

type Project struct {
	Name   string `json:"name,omitempty"`
	Domain Domain `json:"domain"`
}

// Builds the authentication payload
func newAuthPayload(domain Domain, project, user, password string) AuthPayload {
	return AuthPayload{
		Auth: Auth{
			Identity: Identity{
				Methods: []string{"password"},
				Password: Password{
					User: User{
						Name:     user,
						Domain:   domain,
						Password: password,
					},
				},
			},
			Scope: Scope{
				Project: Project{
					Name:   project,
					Domain: domain,
				},
			},
		},
	}
}

type AuthResponse struct {
	Token struct {
		Methods []string `json:"methods"`
		User    struct {
			Domain struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"domain"`
			ID                string `json:"id"`
			Name              string `json:"name"`
			PasswordExpiresAt any    `json:"password_expires_at"`
		} `json:"user"`
		AuditIds  []string  `json:"audit_ids"`
		ExpiresAt time.Time `json:"expires_at"`
		IssuedAt  time.Time `json:"issued_at"`
		Project   struct {
			Domain struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"domain"`
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		IsDomain bool `json:"is_domain"`
		Roles    []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"roles"`
		Catalog []struct {
			Endpoints []struct {
				ID        string `json:"id"`
				Interface string `json:"interface"`
				RegionID  string `json:"region_id"`
				URL       string `json:"url"`
				Region    string `json:"region"`
			} `json:"endpoints"`
			ID   string `json:"id"`
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"catalog"`
	} `json:"token"`
}

// SaveToken saves or updates a token in the token store
func SaveToken(host, token string, expiresAt time.Time, endpoints map[string]string) error {
	store, err := loadTokenStore()
	if err != nil {
		store = TokenStore{Tokens: make(map[string]Token)}
	}

	store.Tokens[host] = Token{
		Value:     token,
		ExpiresAt: expiresAt,
		Host:      host,
		Endpoints: endpoints,
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token store: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(TokenFile), 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %v", err)
	}

	return os.WriteFile(TokenFile, data, 0600)
}

// LoadToken loads the token for a specific host
func LoadTokenStruct(host string) (Token, error) {
	var t Token

	store, err := loadTokenStore()
	if err != nil {
		return t, err
	}
	tokenObj, exists := store.Tokens[host]
	if !exists {
		return t, fmt.Errorf("no token found for host %s", host)
	}

	// Check expiration
	if time.Now().After(tokenObj.ExpiresAt) {
		return t, fmt.Errorf("token for %s is expired", host)
	}

	return tokenObj, nil
}

func loadTokenStore() (TokenStore, error) {
	var store TokenStore
	data, err := os.ReadFile(TokenFile)
	if err != nil {
		return store, fmt.Errorf("failed to read token file: %v", err)
	}

	err = json.Unmarshal(data, &store)
	if err != nil {
		return store, fmt.Errorf("failed to unmarshal token data: %v", err)
	}

	return store, nil
}

func LoadToken(host string) (string, error) {
	t, err := LoadTokenStruct(host)
	if err != nil {
		return "", err
	}
	return t.Value, nil
}

// Authenticate uses domain/project names, calls the auth token API, and returns the token on success.
func Authenticate(host, domain, project, username, password string) (string, error) {
	// Attempt to load an existing token if it's valid
	existingToken, err := LoadToken(host)
	if err == nil {
		return existingToken, nil
	}

	// Not found or expired -> do a fresh authentication
	url := fmt.Sprintf("https://%s:5000/v3/auth/tokens", host)
	payload := newAuthPayload(Domain{Name: domain}, project, username, password)
	apiResp, err := callPOST(url, "", payload)
	if err != nil {
		return "", fmt.Errorf("authentication request failed: %v", err)
	}

	if apiResp.ResponseCode != 201 {
		return "", fmt.Errorf("authentication failed: %v", apiResp.Response)
	}
	if apiResp.TokenHeader == "" {
		return "", fmt.Errorf("no token found in the response")
	}

	// Parse the auth response
	var authResponse AuthResponse
	err = json.Unmarshal([]byte(apiResp.Response), &authResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse auth response: %v", err)
	}
	expiresAt := authResponse.Token.ExpiresAt

	// Extract "public" endpoints we care about
	endpoints := make(map[string]string)
	for _, svc := range authResponse.Token.Catalog {
		for _, ep := range svc.Endpoints {
			if ep.Interface == "public" {
				endpoints[svc.Type] = ep.URL
			}
		}
	}

	// Save token + endpoints
	err = SaveToken(host, apiResp.TokenHeader, expiresAt, endpoints)
	if err != nil {
		return "", fmt.Errorf("failed to save token: %v", err)
	}

	return apiResp.TokenHeader, nil
}

// AuthenticateById uses domain ID instead of name
func AuthenticateById(host, domainID, project, username, password string) (string, error) {
	// Try existing token first
	existingToken, err := LoadToken(host)
	if err == nil {
		return existingToken, nil
	}

	url := fmt.Sprintf("https://%s:5000/v3/auth/tokens", host)
	payload := AuthPayload{
		Auth: Auth{
			Identity: Identity{
				Methods: []string{"password"},
				Password: Password{
					User: User{
						Name:     username,
						Domain:   Domain{ID: domainID},
						Password: password,
					},
				},
			},
			Scope: Scope{
				Project: Project{
					Name:   project,
					Domain: Domain{ID: domainID},
				},
			},
		},
	}

	apiResp, err := callPOST(url, "", payload)
	if err != nil {
		return "", fmt.Errorf("authentication request failed: %v", err)
	}
	if apiResp.ResponseCode != 201 {
		return "", fmt.Errorf("authentication failed: %v", apiResp.Response)
	}
	if apiResp.TokenHeader == "" {
		return "", fmt.Errorf("no token found in the response")
	}

	// Parse the auth response
	var authResponse AuthResponse
	err = json.Unmarshal([]byte(apiResp.Response), &authResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse auth response: %v", err)
	}
	expiresAt := authResponse.Token.ExpiresAt

	// Extract "public" endpoints
	endpoints := make(map[string]string)
	for _, svc := range authResponse.Token.Catalog {
		for _, ep := range svc.Endpoints {
			if ep.Interface == "public" {
				endpoints[svc.Type] = ep.URL
			}
		}
	}

	// Save
	err = SaveToken(host, apiResp.TokenHeader, expiresAt, endpoints)
	if err != nil {
		return "", fmt.Errorf("failed to save token: %v", err)
	}

	return apiResp.TokenHeader, nil
}

// Initialize TokenFile path on module load
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user home directory: %v", err))
	}
	TokenFile = filepath.Join(homeDir, ".vhicmd.token")
}
