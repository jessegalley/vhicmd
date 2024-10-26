/*
Copyright © 2024 jesse galley  
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth <domain> <project>",
	Short: "Get an authentication token from VHI.",
	Long: `Sends credentials from a credentials file or cli flags to 
the VHI API to recieve an authentication token. 
  
The token will be required for all subsequent API calls.

<domain> is the _name_ not the ID. To use ID, pass the -i flag.

For admin access, use "default" and "admin" respectively.`,
  Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
    domain := args[0]
    project := args[1]
    
    var username string 
    if flagUsername == "" {
      username, _ = readUsernameFromStdin()
    } else {
      username = flagUsername 
    }

    var password string 
    if flagPassword == "" && flagAuthFile == "" {
      password, _ = readPasswordFromStdin()
    } else {
      password = flagPassword
    }

    token, err := getAuthToken(domain, project, username, password)
    if err != nil {
      fmt.Printf("ERROR: %v \n", err)
      os.Exit(2)
    }

    fmt.Printf("Authentication OK!\n")
    tokenPath := ".vhitoken.tmp"
    err = writeTokenToFile(token, tokenPath)
    if err != nil {
      fmt.Printf("ERROR: %v \n", err)
      os.Exit(3)
    }

    fmt.Println("Token written to: ", tokenPath)
	},
}

var flagUsername string
var flagPassword string 
var flagUseIds   bool
var flagAuthFile string 

func init() {
	rootCmd.AddCommand(authCmd)

  // we want to take -u and -p together for plain ole insecure cli credentials, or 
  // we want to take a file with the username and password which is a bit more secure 
  // if neither of these options are set, then we'll prompt the user for the creds 
  authCmd.Flags().BoolVarP(&flagUseIds, "id", "i", false, "use domain and project IDs instead of names")
  authCmd.Flags().StringVarP(&flagAuthFile, "passfile", "f", "",  "file containing the password")
	authCmd.Flags().StringVarP(&flagUsername, "username", "u", "",  "username to authenticate with")
	authCmd.Flags().StringVarP(&flagPassword, "password", "p", "",  "password to authenticate with")
  authCmd.MarkFlagsMutuallyExclusive("password", "passfile")
  authCmd.MarkFlagFilename("passfile")
}

// getAuthToken() calls the api to attempt to get an auth token 
// returns an error if any of the underlying api calls or if the 
// authentication fails
func getAuthToken(domain, project, username, password string) (string, error)  {
  var authToken string 
  var authErr   error 
  if flagUseIds {
    authToken, authErr = api.AuthenticateById(domain, project, username, password)
  } else {
    authToken, authErr = api.Authenticate(domain, project, username, password)
  }
  if authErr != nil {
    return "", fmt.Errorf("couldn't get auth token, %v", authErr)
  }

  return authToken, nil
}

// readUsernameFromStdin() prompts the user for a username 
// on stdout and then reads in and returns what they enter 
func readUsernameFromStdin() (string, error) {
  reader := bufio.NewReader(os.Stdin)

  fmt.Print("username: ")
  username, err := reader.ReadString('\n')
  if err != nil {
    log.Fatalf("failed to read username: %v", err)
  }
  username = username[:len(username)-1] // remove trailing newline
  return username, nil
}

// readPasswordFromStdin() prompts the user for a password 
// on stdout and then reads in and returns what they enter 
// does not echo what they type
func readPasswordFromStdin() (string, error) {
  fmt.Print("password: ")
  bytePassword, err := term.ReadPassword(int(syscall.Stdin))
  if err != nil {
    log.Fatalf("failed to read password: %v", err)
  }
  fmt.Println() // add a newline after password input

  password := string(bytePassword)

  return password, nil
}

// writeTokenToFile() writes the given auth token string to a 
// file in the users current working directory 
// returns an error if any of the operating system open or write 
// calls have failed 
// TODO: accept an output flag to send the tokenfile to an arbitrary path 
func writeTokenToFile(token, tokenPath string) (error) {
  // path := "./vhitoken.txt"

  file, err := os.Create(tokenPath)
  if err != nil {
    return fmt.Errorf("can't open %s for writing (%v)", tokenPath, err)
  }

  defer file.Close()

  _, err = file.WriteString(token)
  if err != nil {
    return fmt.Errorf("couldn't write token to %s. (%v)", tokenPath, err)
  }

  // fmt.Println("auth token written to: ", path)  

  return nil 
}
