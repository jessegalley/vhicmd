package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/jessegalley/vhicmd/api"
	"golang.org/x/term"
)

type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
}

func newProgressReader(reader io.Reader, total int64) *progressReader {
	return &progressReader{
		reader: reader,
		total:  total,
	}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)

	// Calculate percentage
	percent := float64(pr.downloaded) * 100 / float64(pr.total)
	fmt.Printf("\rUploading: %.2f%% (%d/%d MB)", percent, pr.downloaded/1024/1024, pr.total/1024/1024)

	if err == io.EOF {
		fmt.Println() // New line on completion
	}

	return n, err
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

// readConfirmation() prompts the user for a yes/no confirmation
// on stdout and then reads in and returns their response
// returns true for yes, false for no
func readConfirmation(prompt string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %v", err)
	}
	confirm = confirm[:len(confirm)-1] // remove trailing newline
	return confirm == "y" || confirm == "Y", nil
}

// writeTokenToFile() writes the given auth token string to a
// file in the users current working directory
// returns an error if any of the operating system open or write
// calls have failed
func writeTokenToFile(token, tokenPath string) error {
	file, err := os.Create(tokenPath)
	if err != nil {
		return fmt.Errorf("can't open %s for writing (%v)", tokenPath, err)
	}

	defer file.Close()

	_, err = file.WriteString(token)
	if err != nil {
		return fmt.Errorf("couldn't write token to %s. (%v)", tokenPath, err)
	}

	return nil
}

func loadAuthFile(path string) (Credentials, error) {
	var creds Credentials
	file, err := os.Open(path)
	if err != nil {
		return creds, fmt.Errorf("failed to open auth file: %v", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&creds)
	if err != nil {
		return creds, fmt.Errorf("failed to parse auth file: %v", err)
	}

	return creds, nil
}

// getPowerStateString() returns a string representation of the
// given power state integer
func getPowerStateString(state int) string {
	switch state {
	case 0:
		return "NOSTATE"
	case 1:
		return "RUNNING"
	case 3:
		return "PAUSED"
	case 4:
		return "SHUTDOWN"
	case 6:
		return "CRASHED"
	case 7:
		return "SUSPENDED"
	default:
		return "UNKNOWN"
	}
}

// stringOrNone() returns the given string if it's not empty
func stringOrNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

// validateTokenEndpoint() checks if the given token has an endpoint
// for the given service and returns the URL if it does
func validateTokenEndpoint(tok api.Token, endpoint string) (string, error) {
	url, exists := tok.Endpoints[endpoint]
	if !exists || url == "" {
		return "", fmt.Errorf("no '%s' endpoint found in token; re-auth or check your catalog", endpoint)
	}
	return url, nil
}

// readAndEncodeUserData() reads the user data file at the given path
// Commonly used for cloud-init scripts
func readAndEncodeUserData(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read user data file: %v", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func validateMacAddr(mac string) error {
	if mac == "" || mac == "auto" {
		return nil
	}
	if _, err := net.ParseMAC(mac); err != nil {
		return fmt.Errorf("invalid MAC address: %v", err)
	}
	return nil
}
