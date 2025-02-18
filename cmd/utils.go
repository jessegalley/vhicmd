package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/facette/natsort"
	"github.com/jessegalley/vhicmd/api"
	"github.com/jessegalley/vhicmd/internal/responseparser"
	"golang.org/x/term"
)

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

// findSingleVMDK() searches for a single VMDK file in /mnt/vmdk
// if multiple matches are found, an error is returned
func findSingleVMDK(pattern string) (string, error) {
	matches, err := findVMDKs(pattern)
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no matching VMDK files found")
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple matching VMDK files found, be more specific")
	}

	return matches[0], nil
}

// findVMDKs launches one goroutine per datastore in /mnt/vmdk
func findVMDKs(pattern string) ([]string, error) {
	rootDir := "/mnt/vmdk"
	var matches []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	entries, err := os.ReadDir(rootDir) // Get all top-level directories (datastores)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %v", rootDir, err)
	}

	// Channel for collecting results
	results := make(chan string, 1000) // Buffered to prevent blocking

	// Start searching each datastore in its own goroutine
	for _, entry := range entries {
		if entry.IsDir() {
			wg.Add(1)
			go func(datastore string) {
				defer wg.Done()
				storePath := filepath.Join(rootDir, datastore)
				storeMatches := findVMDKsInPath(storePath, pattern)

				// Lock before appending to shared slice
				mu.Lock()
				matches = append(matches, storeMatches...)
				mu.Unlock()

				// Send matches to channel (optional)
				for _, match := range storeMatches {
					results <- match
				}
			}(entry.Name())
		}
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	wg.Wait()

	// Sort matches using natural sorting
	sort.Slice(matches, func(i, j int) bool {
		return natsort.Compare(strings.ToLower(matches[i]), strings.ToLower(matches[j]))
	})

	return matches, nil
}

// findVMDKsInPath recursively searches a specific datastore path
func findVMDKsInPath(rootPath, pattern string) []string {
	var matches []string
	filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Warning: Cannot access %s: %v\n", path, err)
			return nil
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), "-flat.vmdk") {
			baseName := strings.TrimSuffix(d.Name(), "-flat.vmdk")
			if strings.Contains(strings.ToLower(baseName), strings.ToLower(pattern)) {
				matches = append(matches, path)
			}
		}
		return nil
	})
	return matches
}

func displayProjects(response api.ProjectListResponse) {
	var displayProjects []responseparser.Project
	for _, project := range response.Projects {
		displayProjects = append(displayProjects, responseparser.Project{
			ID:       project.ID,
			DomainID: project.DomainID,
			Name:     project.Name,
			Enabled:  project.Enabled,
		})
	}
	fmt.Println("\nAvailable projects:")
	responseparser.PrintProjectsSelectionTable(displayProjects)
}
