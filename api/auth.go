package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// auth json request structure:
// the domain ID can be domain NAME instead if the param key
// is changed to "name"
// {
//   "auth": {
//     "identity": {
//       "methods": [
//         "password"
//       ],
//       "password": {
//         "user": {
//           "name": "user1",
//           "domain": {
//             "id": "cdc759b962e34e67997f59f8b1c21027"
//           },
//           "password": "user1_password"
//         }
//       }
//     },
//     "scope": {
//       "project": {
//         "name": "project1",
//         "domain": {
//           "id": "cdc759b962e34e67997f59f8b1c21027"
//         }
//       }
//     }
//   }
// }
// the response header will contain the token in the X-Subject-Token header. Pass it in the X-Auth-Token header in all requests

// Authenticate() takes in the credentials and domain/project names, calls
// the auth token API, and returns the resulting token on sucess.
// Returns an error if the authentication or underlying api calls failed.
func Authenticate(domain, project, username, password string) (string, error) {
  url := "https://panel-vhi1.mia1.oniaas.io:5000/v3/auth/tokens"
  payload := newAuthPayload(Domain{Name: domain}, project, username, password)
  apiResp, _ := call(url, payload)
  
  if apiResp.ResponseCode != 201 {
    return "", fmt.Errorf("authentication failed: %v", apiResp.Response)
  }

  var token string
  if apiResp.TokenHeader != ""  {
    token = apiResp.TokenHeader
  }

  return token, nil 
} 

// AuthenticateById() takes in the credentials and domain/project IDs,calls 
// the auth token API, and returns the resulting token on sucess.
// Returns an error if the authentication or underlying api calls failed.
func AuthenticateById(domainid, project, username, password string) (string, error) {
  payload := newAuthPayload(Domain{ID: domainid}, project, username, password)
  call("https://myurl.com/auth", payload)
  token := "9ijqc8uj1nwef1o"
  return token, nil
}

type AuthPayload struct {
	Auth Auth `json:"auth"`
}

func (a AuthPayload) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(struct {
		Auth Auth `json:"auth"`
	}{a.Auth}, " ", "  ")
}

type Auth struct {
	Identity Identity `json:"identity"`
	Scope    Scope    `json:"scope"`
}

type Identity struct {
	Methods  []string  `json:"methods"`
	Password Password  `json:"password"`
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


func newAuthPayload(domain Domain, project string, user string, password string) AuthPayload {
  	payload := AuthPayload{
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

  return payload
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

type AuthError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Title   string `json:"title"`
	} `json:"error"`
}
