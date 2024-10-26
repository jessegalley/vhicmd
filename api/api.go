package api

import (
	"io"
	"log"

	"github.com/jessegalley/vhicmd/internal/httpclient"
)

// payload defines the interface for all API payloads, ensuring they can be marshaled to JSON.
type Payload interface {
	MarshalJSON() ([]byte, error)
}

type ApiResponse struct {
  TokenHeader   string 
  ResponseCode  int
  Response      string
}

func call(url string, payload Payload) (ApiResponse, error) {
  apiResp := ApiResponse{}

  jsonData, err := payload.MarshalJSON()
  if err != nil {
    log.Fatalf("error marshalling json payload: %v", err)
  }

  resp, err := httpclient.SendRequest(url, jsonData)
  if err != nil {
    log.Fatalf("error making HTTP request: %v", err)
  }
  defer resp.Body.Close()

  apiResp.ResponseCode = resp.StatusCode

  if token := resp.Header.Get("X-Subject-Token"); token != "" {
    apiResp.TokenHeader = token
  }
	
	body, err := io.ReadAll(resp.Body) // Read the full body
	if err != nil {
		log.Fatalf("Error reading body: %v", err)
	}
  apiResp.Response = string(body)

  // var authResp AuthResponse 
  // if resp.StatusCode == http.StatusOK {
  //   err = json.Unmarshal(body, &authResp)
  //   if err != nil {
  //     log.Fatalf("error unmarshalling json into auth resp, %v", err)
  //   }
  //   spew.Dump(authResp)
  //   fmt.Println("auth response okay!")
  //   return "auth response okay", nil
  // }
  //
  //
  // var authErr AuthError
  // if resp.StatusCode  != http.StatusOK {
  //   err = json.Unmarshal(body, &authErr)
  //   if err != nil {
  //     log.Fatalf("error unmarshalling json into auth error, %v", err)
  //   }
  //   spew.Dump(authErr)
  //   fmt.Println("auth error okay!")
  //   return "auth error dokay", nil
  // }

  return apiResp, nil
}
