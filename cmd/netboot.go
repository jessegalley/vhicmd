package cmd

import (
	"fmt"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
)

var netbootCmd = &cobra.Command{
	Use:   "netboot",
	Short: "Configure netboot settings",
}

var setNetworkInstallCmd = &cobra.Command{
	Use:   "set <vm-id> <true|false>",
	Short: "Set network_install metadata for a VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		value := args[1]

		if value != "true" && value != "false" {
			return fmt.Errorf("value must be 'true' or 'false'")
		}

		tok, err := api.LoadTokenStruct(vhiHost)
		if err != nil {
			return fmt.Errorf("no valid auth token found; run 'vhicmd auth' first: %v", err)
		}

		computeURL := tok.Endpoints["compute"]
		if computeURL == "" {
			return fmt.Errorf("no 'compute' endpoint found in token; re-auth or check your catalog")
		}

		enabled := value == "true"
		err = api.UpdateNetworkInstall(computeURL, tok.Value, vmID, enabled)
		if err != nil {
			return err
		}

		fmt.Printf("Set network_install=%s for VM %s\n", value, vmID)
		return nil
	},
}

func init() {
	netbootCmd.AddCommand(setNetworkInstallCmd)
	rootCmd.AddCommand(netbootCmd)
}
