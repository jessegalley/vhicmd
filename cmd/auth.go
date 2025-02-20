/*
Copyright © 2024 jesse galley
*/
package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/facette/natsort"
	"github.com/gookit/color"
	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var authCmd = &cobra.Command{
	Use:   "auth [domain] [project]",
	Short: "Get an authentication token from VHI.",
	Long: `Sends credentials to the VHI API to receive an authentication token.
The token will be required for all subsequent API calls.

Project names with spaces don't require quotes:
  vhicmd auth myDomain my project name

Domain is the *name* not the ID. To use ID, pass the -i flag.
For admin access, use "default" and "admin" respectively.`,
	Run: func(cmd *cobra.Command, args []string) {
		domain := viper.GetString("domain")
		project := viper.GetString("project")

		if len(args) > 0 {
			domain = args[0]
		}
		if len(args) > 1 {
			project = strings.Join(args[1:], " ")
		}

		if domain == "" || project == "" {
			fmt.Printf("ERROR: domain and project required via args or config\n")
			os.Exit(2)
		}

		// Track if we prompted for credentials
		didPrompt := false

		// Get username from flag, config, or prompt
		username := flagUsername
		if username == "" {
			username = viper.GetString("username")
		}
		if username == "" {
			didPrompt = true
			var err error
			username, err = readUsernameFromStdin()
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				os.Exit(2)
			}
		}

		// Get password from flag, config, or prompt
		password := flagPassword
		if password == "" {
			password = viper.GetString("password")
		}
		if password == "" {
			didPrompt = true
			var err error
			password, err = readPasswordFromStdin()
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				os.Exit(2)
			}
		}

		host := flagHost // Get host from flag
		if host == "" {
			host = viper.GetString("host")
		}
		if host == "" {
			fmt.Printf("ERROR: no host found in flags or config. Provide --host or set 'host' in .vhirc\n")
			os.Exit(2)
		}

		_, err := doAuth(host, domain, project, username, password)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(2)
		}

		fmt.Printf("Authentication OK!\n")

		// If we prompted for creds, offer to save them
		if didPrompt {
			saveConfig, err := readConfirmation("Save credentials to config? [y/N] ")
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				os.Exit(2)
			}

			if saveConfig {
				viper.Set("host", host)
				viper.Set("username", username)
				viper.Set("password", password)
				viper.Set("domain", domain)
				viper.Set("project", project)
				if err := viper.WriteConfig(); err != nil {
					if err := viper.SafeWriteConfig(); err != nil {
						fmt.Printf("ERROR: failed to save config: %v\n", err)
						os.Exit(2)
					}
				}
			}
		}

		fmt.Printf("You can now run other commands.\n")
	},
}

var switchProjectCmd = &cobra.Command{
	Use:     "switch-project [project]",
	Aliases: []string{"sw"},
	Short:   "Switch to a different project using saved credentials",
	Long: `Switch to a different project using credentials saved in ~/.vhirc
If no project is specified, displays available projects and prompts for selection.
  Project names with spaces don't require quotes.
  Examples:
    vhicmd switch-project           # Interactive project selection
    vhicmd switch-project my project # Direct project switch
    `,
	RunE: func(cmd *cobra.Command, args []string) error {
		user := viper.GetString("username")
		pass := viper.GetString("password")
		if user == "" || pass == "" {
			return fmt.Errorf("no saved credentials found, run 'vhicmd auth' first and save credentials")
		}

		domain := viper.GetString("domain")

		identityUrl, err := validateTokenEndpoint(tok, "identity")
		if err != nil {
			return err
		}

		projects, err := api.ListProjects(identityUrl, tok.Value)
		if err != nil {
			return fmt.Errorf("failed to list projects: %v", err)
		}

		if len(projects.Projects) == 0 {
			return fmt.Errorf("no projects found")
		}

		var project string
		if len(args) > 0 {
			project = strings.Join(args, " ")
			found := false
			for _, p := range projects.Projects {
				if p.Name == project {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("project '%s' not found", project)
			}
		} else {
			fmt.Printf("Current project: %s\n", color.Style{color.Bold}.Sprintf("%s", tok.Project))
			sort.Slice(projects.Projects, func(i, j int) bool {
				if projects.Projects[i].Enabled != projects.Projects[j].Enabled {
					return projects.Projects[i].Enabled
				}
				return natsort.Compare(
					strings.ToLower(projects.Projects[i].Name),
					strings.ToLower(projects.Projects[j].Name),
				)
			})

			displayProjects(projects)

			// Get user selection
			var selection int
			for {
				fmt.Print("\nEnter project number (or Ctrl+C to cancel): ")
				_, err := fmt.Scanf("%d", &selection)
				if err != nil {
					// Clear input buffer
					fmt.Scanf("%s")
					continue
				}
				if selection < 1 || selection > len(projects.Projects) {
					fmt.Printf("Invalid selection. Please enter 1-%d\n", len(projects.Projects))
					continue
				}
				break
			}
			project = projects.Projects[selection-1].Name
		}

		_, err = doAuth(tok.Host, domain, project, user, pass)
		if err != nil {
			return fmt.Errorf("failed to switch project: %v", err)
		}

		viper.Set("project", project)
		if err := viper.WriteConfig(); err != nil {
			if err := viper.SafeWriteConfig(); err != nil {
				fmt.Printf("ERROR: failed to save config: %v\n", err)
				os.Exit(2)
			}
		}

		fmt.Printf("Switched to project: %s\n", color.Style{color.Bold}.Sprintf("%s", project))
		return nil
	},
}

var (
	flagUsername string
	flagPassword string
	flagUseIds   bool
	flagAuthFile string
	flagHost     string
)

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(switchProjectCmd)

	authCmd.Flags().BoolVarP(&flagUseIds, "id", "i", false, "use domain and project IDs instead of names")
	authCmd.Flags().StringVarP(&flagAuthFile, "passfile", "f", "", "file containing the password")
	authCmd.Flags().StringVarP(&flagUsername, "username", "u", "", "username to authenticate with")
	authCmd.Flags().StringVarP(&flagPassword, "password", "p", "", "password to authenticate with")
	authCmd.Flags().StringVarP(&flagHost, "host", "H", "", "VHI host to authenticate against")
	authCmd.MarkFlagsMutuallyExclusive("password", "passfile")
	authCmd.MarkFlagsMutuallyExclusive("password", "passfile")
	authCmd.MarkFlagFilename("passfile") // not really needed with Viper config but left for backward compatibility
}

func doAuth(host, domain, project, username, password string) (string, error) {
	var authToken string
	var authErr error
	if flagUseIds {
		authToken, authErr = api.AuthenticateById(host, domain, project, username, password)
	} else {
		authToken, authErr = api.Authenticate(host, domain, project, username, password)
	}
	if authErr != nil {
		return "", fmt.Errorf("couldn't get auth token, %v", authErr)
	}

	return authToken, nil
}
