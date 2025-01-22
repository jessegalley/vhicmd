package cmd

import (
	"fmt"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
)

var rebootCmd = &cobra.Command{
	Use:   "reboot",
	Short: "Reboot a virtual machine",
}

var hardRebootCmd = &cobra.Command{
	Use:   "hard <vm-id>",
	Short: "Perform a hard reboot on a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		err = api.RebootVM(computeURL, tok.Value, vmID, "HARD")
		if err != nil {
			return err
		}

		fmt.Printf("Hard reboot initiated for VM %s\nIf using Cobbler after creating a new VM, the VM should now be installing.\n", vmID)
		return nil
	},
}

var softRebootCmd = &cobra.Command{
	Use:   "soft <vm-id>",
	Short: "Perform a soft reboot on a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		err = api.RebootVM(computeURL, tok.Value, vmID, "SOFT")
		if err != nil {
			return err
		}

		fmt.Printf("Soft reboot initiated for VM %s\n", vmID)
		return nil
	},
}

func init() {
	rebootCmd.AddCommand(hardRebootCmd)
	rebootCmd.AddCommand(softRebootCmd)
	rootCmd.AddCommand(rebootCmd)
}
