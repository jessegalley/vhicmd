package cmd

import (
	"fmt"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
}

var deleteVMCmd = &cobra.Command{
	Use:   "vm <vm_id>",
	Short: "Delete a VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		id, err := api.GetVMIDByName(computeURL, tok.Value, vmID)
		if err == nil {
			vmID = id
		}

		err = api.DeleteVM(computeURL, tok.Value, vmID)
		if err != nil {
			return err
		}

		fmt.Printf("VM %s deleted\n", vmID)

		return nil
	},
}

var deleteImageCmd = &cobra.Command{
	Use:   "image <image_id>",
	Short: "Delete an image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageID := args[0]

		computeURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		img, err := api.GetImageIDByName(computeURL, tok.Value, imageID)
		if err == nil {
			imageID = img
		}

		err = api.DeleteImage(computeURL, tok.Value, imageID)
		if err != nil {
			return err
		}

		fmt.Printf("Image %s deleted\n", imageID)

		return nil
	},
}

var deleteVolumeCmd = &cobra.Command{
	Use:   "volume <volume_id>",
	Short: "Delete a volume",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

		blockURL, err := validateTokenEndpoint(tok, "block")
		if err != nil {
			return err
		}

		err = api.DeleteVolume(blockURL, tok.Value, volumeID)
		if err != nil {
			return err
		}

		fmt.Printf("Volume %s deleted\n", volumeID)

		return nil
	},
}

var deletePortCmd = &cobra.Command{
	Use:   "port <port_id>",
	Short: "Delete a port",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		portID := args[0]

		networkURL, err := validateTokenEndpoint(tok, "network")
		if err != nil {
			return err
		}

		err = api.DeletePort(networkURL, tok.Value, portID)
		if err != nil {
			return err
		}

		fmt.Printf("Port %s deleted\n", portID)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteVMCmd)
	deleteCmd.AddCommand(deleteImageCmd)
	deleteCmd.AddCommand(deleteVolumeCmd)
	deleteCmd.AddCommand(deletePortCmd)
}
