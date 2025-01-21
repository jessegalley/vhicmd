package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources like VMs or volumes",
}

// Subcommand: create vm
var createVMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Create a new virtual machine",
	RunE: func(cmd *cobra.Command, args []string) error {
		host := viper.GetString("host")
		tok, err := api.LoadTokenStruct(host)
		if err != nil {
			return fmt.Errorf("no valid auth token found; run 'vhicmd auth' first: %v", err)
		}

		computeURL := tok.Endpoints["compute"]
		if computeURL == "" {
			return fmt.Errorf("no 'compute' endpoint found in token; re-auth or check your catalog")
		}

		// Use flag value if set, otherwise fall back to viper config
		flavorRef := flagFlavorRef
		if flavorRef == "" {
			flavorRef = viper.GetString("flavor_id")
		}

		// Get flavor details
		flavorDetails, err := api.GetFlavorDetails(computeURL, tok.Value, flavorRef)
		if err != nil {
			return fmt.Errorf("failed to fetch flavor details: %v", err)
		}

		// Create the VM request
		var request api.CreateVMRequest
		request.Server.Name = flagVMName
		request.Server.FlavorRef = flavorRef
		request.Server.Metadata = map[string]string{
			"network_install": "true",
		}

		// Use flag value if set, otherwise fall back to viper config
		imageRef := flagImageRef
		if imageRef == "" {
			imageRef = viper.GetString("image_id")
		}

		// Handle disk setup based on flavor
		if flavorDetails.Flavor.Disk == 0 {
			request.Server.BlockDeviceMappingV2 = []map[string]interface{}{
				{
					"boot_index":            "0",
					"uuid":                  imageRef,
					"source_type":           "image",
					"destination_type":      "volume",
					"volume_size":           10,
					"delete_on_termination": true,
					"volume_type":           "nvme_ec7_2",
				},
			}
		} else {
			request.Server.ImageRef = imageRef
		}

		// Use flag value if set, otherwise fall back to viper config
		networks := flagNetworkCSV
		if networks == "" {
			networks = viper.GetString("networks")
		}
		if networks == "" {
			return fmt.Errorf("no networks specified; provide via --networks flag or config")
		}

		networkIDs := strings.Split(networks, ",")
		for _, networkID := range networkIDs {
			request.Server.Networks = append(request.Server.Networks, map[string]string{
				"uuid": strings.TrimSpace(networkID),
			})
		}

		// Create VM
		resp, err := api.CreateVM(computeURL, tok.Value, request)
		if err != nil {
			return fmt.Errorf("failed to create VM: %v", err)
		}

		fmt.Printf("Creating VM %s (%s)\n", flagVMName, resp.Server.ID)

		// Wait for VM to become active
		vmDetails, err := api.WaitForActive(computeURL, tok.Value, resp.Server.ID)
		if err != nil {
			return err
		}

		// Stop the VM if power-on is false
		if !flagPowerOn {
			fmt.Printf("VM created, setting netboot and stopping VM... almost done\n\n")
			err = api.StopVM(computeURL, tok.Value, resp.Server.ID)
			if err != nil {
				return fmt.Errorf("failed to stop VM: %v", err)
			}
			// Get final state
			vmDetails, err = api.GetVMDetails(computeURL, tok.Value, resp.Server.ID)
			if err != nil {
				return fmt.Errorf("failed to get final VM state: %v", err)
			}
		}

		// Prepare output details
		details := map[string]interface{}{
			"power_state": getPowerStateString(vmDetails.PowerState),
			"name":        vmDetails.Name,
			"id":          vmDetails.ID,
			"metadata":    vmDetails.Metadata,
		}

		// Add network info
		netInfo := make([]map[string]interface{}, 0)
		for netName, addrs := range vmDetails.Addresses {
			if len(addrs) > 0 {
				net := map[string]interface{}{
					"name":        netName,
					"mac_address": addrs[0].OSEXTIPSMACAddr,
					"ip_address":  addrs[0].Addr,
				}
				netInfo = append(netInfo, net)
			}
		}
		if len(netInfo) > 0 {
			details["networks"] = netInfo
		}

		// Output results
		if flagJsonOutput {
			jsonBytes, err := json.MarshalIndent(details, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal VM details to JSON: %v", err)
			}
			fmt.Println(string(jsonBytes))
		} else {
			yamlBytes, err := yaml.Marshal(details)
			if err != nil {
				return fmt.Errorf("failed to marshal VM details to YAML: %v", err)
			}
			fmt.Println(string(yamlBytes))
		}

		return nil
	},
}

// Subcommand: create volume
var createVolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Create a new storage volume",
	RunE: func(cmd *cobra.Command, args []string) error {
		tok, err := api.LoadTokenStruct(vhiHost)
		if err != nil {
			return fmt.Errorf("no valid auth token found; run 'vhicmd auth' first: %v", err)
		}

		storageURL := tok.Endpoints["volumev3"]
		if storageURL == "" {
			return fmt.Errorf("no 'volumev3' endpoint found in token; re-auth or check your catalog")
		}

		var request api.CreateVolumeRequest
		request.Volume.Name = flagVolumeName
		request.Volume.Size = flagVolumeSize
		request.Volume.Description = flagVolumeDescription
		request.Volume.VolumeType = flagVolumeType

		resp, err := api.CreateVolume(storageURL, tok.Value, request)
		if err != nil {
			return err
		}

		fmt.Printf("Volume created: ID: %s, Name: %s, Size: %d GB\n", resp.Volume.ID, resp.Volume.Name, resp.Volume.Size)
		return nil
	},
}

var (
	flagVMName            string
	flagFlavorRef         string
	flagImageRef          string
	flagNetworkCSV        string
	flagVolumeName        string
	flagVolumeSize        int
	flagVolumeDescription string
	flagVolumeType        string
	flagPowerOn           bool
)

func init() {
	// Flags for create vm
	createVMCmd.Flags().StringVar(&flagVMName, "name", "", "Name of the virtual machine")
	createVMCmd.Flags().StringVar(&flagFlavorRef, "flavor", "", "Flavor ID for the virtual machine")
	createVMCmd.Flags().StringVar(&flagImageRef, "image", "", "Image ID for the virtual machine")
	createVMCmd.Flags().StringVar(&flagNetworkCSV, "networks", "", "Comma-separated list of network UUIDs")
	createVMCmd.Flags().BoolVar(&flagJsonOutput, "json", false, "Output in JSON format (default: YAML)")
	createVMCmd.Flags().BoolVar(&flagPowerOn, "power-on", false, "Power on the VM after creation (default: true)")

	// Bind flags to viper
	viper.BindPFlag("flavor_id", createVMCmd.Flags().Lookup("flavor"))
	viper.BindPFlag("image_id", createVMCmd.Flags().Lookup("image"))
	viper.BindPFlag("networks", createVMCmd.Flags().Lookup("networks"))

	createVMCmd.MarkFlagRequired("name")

	// Flags for create volume
	createVolumeCmd.Flags().StringVar(&flagVolumeName, "name", "", "Name of the volume")
	createVolumeCmd.Flags().IntVar(&flagVolumeSize, "size", 0, "Size of the volume in GB")
	createVolumeCmd.Flags().StringVar(&flagVolumeDescription, "description", "", "Description of the volume")
	createVolumeCmd.Flags().StringVar(&flagVolumeType, "type", "nvme_ec7_2", "Type of the volume: nvme_ec7_2, replica3")

	createVolumeCmd.MarkFlagRequired("name")
	createVolumeCmd.MarkFlagRequired("size")

	// Add subcommands to the parent create command
	createCmd.AddCommand(createVMCmd)
	createCmd.AddCommand(createVolumeCmd)

	// Add the create command to the root command
	rootCmd.AddCommand(createCmd)
}
