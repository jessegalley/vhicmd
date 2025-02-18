package cmd

import (
	"fmt"
	"strings"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"set"},
	Short:   "Update resource attributes",
}

var updateVMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Update VM attributes",
}

var vmFlavorCmd = &cobra.Command{
	Use:   "flavor",
	Short: "Manage VM flavor changes",
}

var vmNameCmd = &cobra.Command{
	Use:   "name <vm-id> <new-name>",
	Short: "Update VM name",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		newName := args[1]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		id, err := api.GetVMIDByName(computeURL, tok.Value, vmID)
		if err == nil {
			vmID = id
		}

		err = api.UpdateVMName(computeURL, tok.Value, vmID, newName)
		if err != nil {
			return err
		}

		fmt.Printf("Updated name of VM %s to %s\n", vmID, newName)
		return nil
	},
}

var vmMetadataCmd = &cobra.Command{
	Use:   "metadata <vm-id> <key> <value>",
	Short: "Update VM metadata key-value pair",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		key := args[1]
		value := args[2]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		id, err := api.GetVMIDByName(computeURL, tok.Value, vmID)
		if err == nil {
			vmID = id
		}

		err = api.UpdateVMMetadataItem(computeURL, tok.Value, vmID, key, value)
		if err != nil {
			return err
		}

		fmt.Printf("Updated metadata %s=%s for VM %s\n", key, value, vmID)
		return nil
	},
}

var flavorStartCmd = &cobra.Command{
	Use:   "start <vm-id> <flavor>",
	Short: "Step 1: Start VM flavor change process",
	Long: `Step 1: Begin the flavor change process for a VM.
After starting, use 'confirm' to accept or 'revert' to cancel.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		flavor := args[1]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		id, err := api.GetVMIDByName(computeURL, tok.Value, vmID)
		if err == nil {
			vmID = id
		}

		flavorID, err := api.GetFlavorIDByName(computeURL, tok.Value, flavor)
		if err != nil {
			return err
		}

		err = api.ResizeVM(computeURL, tok.Value, vmID, flavorID)
		if err != nil {
			return err
		}

		fmt.Printf("Started flavor change for VM %s to %s\n", vmID, flavor)
		fmt.Printf("Once the change is ready, use either:\n")
		fmt.Printf("  - 'update vm flavor confirm %s' to accept the change\n", vmID)
		fmt.Printf("  - 'update vm flavor revert %s' to cancel the change\n", vmID)
		return nil
	},
}

var flavorConfirmCmd = &cobra.Command{
	Use:   "confirm <vm-id>",
	Short: "Step 2a: Confirm VM flavor change",
	Long: `Step 2a: Accept and finalize the flavor change.
Only use after 'start' command has completed.`,
	Args: cobra.ExactArgs(1),
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

		err = api.ConfirmResize(computeURL, tok.Value, vmID)
		if err != nil {
			return err
		}

		fmt.Printf("Confirmed flavor change for VM %s\n", vmID)
		return nil
	},
}

var flavorRevertCmd = &cobra.Command{
	Use:   "revert <vm-id>",
	Short: "Step 2b: Revert VM flavor change",
	Long: `Step 2b: Cancel the flavor change and restore original size.
Only use after 'start' command has completed.`,
	Args: cobra.ExactArgs(1),
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

		err = api.RevertResize(computeURL, tok.Value, vmID)
		if err != nil {
			return err
		}

		fmt.Printf("Reverted flavor change for VM %s\n", vmID)
		return nil
	},
}

// Volume update command group
var updateVolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Update volume attributes",
}

var volumeTypeAccessCmd = &cobra.Command{
	Use:   "access <volume-type-id> <project-id> <action>",
	Short: "Update volume type access (add/remove)",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeTypeID := args[0]
		projectID := args[1]
		action := args[2]

		if action != "add" && action != "remove" {
			return fmt.Errorf("action must be either 'add' or 'remove'")
		}

		storageURL, err := validateTokenEndpoint(tok, "volumev3")
		if err != nil {
			return err
		}

		if action == "add" {
			err = api.AddVolumeTypeAccess(storageURL, tok.Value, volumeTypeID, projectID)
		} else {
			err = api.RemoveVolumeTypeAccess(storageURL, tok.Value, volumeTypeID, projectID)
		}

		if err != nil {
			return err
		}

		fmt.Printf("%s access for project %s to volume type %s\n",
			action, projectID, volumeTypeID)
		return nil
	},
}

var updateImageCmd = &cobra.Command{
	Use:   "image",
	Short: "Update image attributes",
}

var imageVisibilityCmd = &cobra.Command{
	Use:   "visibility <image> <visibility>",
	Short: "Update image visibility (public, private, shared, community)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageID := args[0]
		visibility := strings.ToLower(args[1])

		if visibility != "public" && visibility != "private" &&
			visibility != "shared" && visibility != "community" {
			return fmt.Errorf("visibility must be one of: public, private, shared, community")
		}

		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		id, err := api.GetImageIDByName(imageURL, tok.Value, imageID)
		if err == nil {
			imageID = id
		}

		err = api.UpdateImageVisibility(imageURL, tok.Value, imageID, visibility)
		if err != nil {
			return err
		}

		fmt.Printf("Updated visibility of image %s to %s\n", imageID, visibility)
		return nil
	},
}

var imageMemberStatusCmd = &cobra.Command{
	Use:   "member <image> <member-id> <status>",
	Short: "Update image member status (accepted, rejected, pending)",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageID := args[0]
		memberID := args[1]
		status := strings.ToLower(args[2])

		if status != "accepted" && status != "rejected" && status != "pending" {
			return fmt.Errorf("status must be one of: accepted, rejected, pending")
		}

		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		id, err := api.GetImageIDByName(imageURL, tok.Value, imageID)
		if err == nil {
			imageID = id
		}

		err = api.UpdateImageMemberStatus(imageURL, tok.Value, imageID, memberID, status)
		if err != nil {
			return err
		}

		fmt.Printf("Updated member %s status to %s for image %s\n", memberID, status, imageID)
		return nil
	},
}

var attachPortCmd = &cobra.Command{
	Use:   "attach-port <vm-id> <port-id>",
	Short: "Attach an existing port to a VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		portID := args[1]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		networkURL, err := validateTokenEndpoint(tok, "network")
		if err != nil {
			return err
		}

		id, err := api.GetVMIDByName(computeURL, tok.Value, vmID)
		if err == nil {
			vmID = id
		}

		resp, err := api.AttachNetworkToVM(networkURL, computeURL, tok.Value, vmID, "", portID, nil)
		if err != nil {
			return err
		}

		fmt.Printf("Attached port %s to VM %s (MAC: %s)\n",
			resp.InterfaceAttachment.PortID, vmID, resp.InterfaceAttachment.MacAddr)
		return nil
	},
}

var detachPortCmd = &cobra.Command{
	Use:   "detach-port <vm-id> <port-id>",
	Short: "Detach a port from a VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmID := args[0]
		portID := args[1]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		// Get VM ID if name was provided
		id, err := api.GetVMIDByName(computeURL, tok.Value, vmID)
		if err == nil {
			vmID = id
		}

		err = api.DetachNetworkFromVM(computeURL, tok.Value, vmID, portID)
		if err != nil {
			return err
		}

		fmt.Printf("Detached port %s from VM %s\n", portID, vmID)
		return nil
	},
}

func init() {
	// VM subcommands
	updateVMCmd.AddCommand(vmNameCmd)
	updateVMCmd.AddCommand(vmMetadataCmd)
	updateVMCmd.AddCommand(vmFlavorCmd)
	updateVMCmd.AddCommand(attachPortCmd)
	updateVMCmd.AddCommand(detachPortCmd)

	// Flavor subcommands
	vmFlavorCmd.AddCommand(flavorStartCmd)
	vmFlavorCmd.AddCommand(flavorConfirmCmd)
	vmFlavorCmd.AddCommand(flavorRevertCmd)

	// Volume subcommands
	updateVolumeCmd.AddCommand(volumeTypeAccessCmd)

	// Image subcommands
	updateImageCmd.AddCommand(imageVisibilityCmd)
	updateImageCmd.AddCommand(imageMemberStatusCmd)

	// Add to update command
	updateCmd.AddCommand(updateVMCmd)
	updateCmd.AddCommand(updateVolumeCmd)
	updateCmd.AddCommand(updateImageCmd)

	rootCmd.AddCommand(updateCmd)
}
