package cmd

import (
	"fmt"
	"strings"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
)

var bootableCmd = &cobra.Command{
	Use:   "bootable <volumeid> <true|false>",
	Short: "Set the bootable flag for a volume",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]
		bootableStr := strings.ToLower(args[1])

		bootable := false
		switch bootableStr {
		case "true":
			bootable = true
		case "false":
			bootable = false
		default:
			return fmt.Errorf("invalid value for bootable: must be 'true' or 'false'")
		}

		storageURL, err := validateTokenEndpoint(tok, "volumev3")
		if err != nil {
			return err
		}

		if err := api.SetVolumeBootable(storageURL, tok.Value, volumeID, bootable); err != nil {
			return fmt.Errorf("failed to set bootable flag: %v", err)
		}

		fmt.Printf("Successfully set bootable=%v for volume %s\n", bootable, volumeID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bootableCmd)
}
