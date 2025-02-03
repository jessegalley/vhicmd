package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		storageURL, err := validateTokenEndpoint(tok, "volumev3")
		if err != nil {
			return err
		}

		imageRef := flagImageRef
		if imageRef == "" {
			imageRef = viper.GetString("image_id")
		}

		// Get required parameters
		flavorRef := flagFlavorRef
		if flavorRef == "" {
			flavorRef = viper.GetString("flavor_id")
		}
		if flavorRef == "" {
			return fmt.Errorf("no flavor specified; provide --flavor flag or set 'flavor_id' in config")
		}

		// Ensure networks are specified
		networks := flagNetworkCSV
		if networks == "" {
			networks = viper.GetString("networks")
		}
		if networks == "" {
			return fmt.Errorf("no networks specified; provide --networks flag or set 'networks' in config")
		}

		// Ensure IPs are specified
		ips := flagIPCSV
		if ips == "" {
			return fmt.Errorf("no IPs specified; provide --ips flag")
		}

		// Split networks and IPs into slices
		networkIDs := strings.Split(networks, ",")
		ipAddresses := strings.Split(ips, ",")

		// Validate that the number of networks matches the number of IPs
		if len(networkIDs) != len(ipAddresses) {
			return fmt.Errorf("the number of networks (%d) must match the number of IPs (%d)", len(networkIDs), len(ipAddresses))
		}

		// Create network mapping with IPs
		//var networkMapping []map[string]string
		//for i, networkID := range networkIDs {
		//	ip := strings.TrimSpace(ipAddresses[i])
		//	if ip == "" {
		//		return fmt.Errorf("IP address for network %s cannot be empty", networkID)
		//	}

		//	networkMapping = append(networkMapping, map[string]string{
		//		"uuid":     strings.TrimSpace(networkID),
		//		"fixed_ip": ip,
		//	})
		//}

		// Calculate volume size
		volumeSize := 10 // Default minimum
		if flagVMSize > 0 {
			volumeSize = flagVMSize
		}

		// Create initial VM request
		var request api.CreateVMRequest
		request.Server.Name = flagVMName
		request.Server.FlavorRef = flavorRef
		//request.Server.Networks = networkMapping
		// Pass "none" to networks, so no interfaces are attached initially**
		request.Server.Networks = "none"

		// Set metadata for network boot if no image is specified
		// netboot is deprecated, use --image flag instead
		// the reason is that VHI does not have good support for netboot,
		// and a custom iPXE rom is required to boot from network
		if flagVMNetboot {
			imageRef = "" // Clear image ref if netboot is enabled
			request.Server.Metadata = map[string]string{
				"network_install": "true",
			}
		}

		// Determine block device mapping
		if imageRef != "" {
			// imageRef exists, create boot volume from image
			request.Server.BlockDeviceMappingV2 = []map[string]interface{}{
				{
					"boot_index":            "0",
					"uuid":                  imageRef,
					"source_type":           "image",
					"destination_type":      "volume",
					"volume_size":           volumeSize,
					"delete_on_termination": true,
					"volume_type":           "nvme_ec7_2",
					"disk_bus":              "sata",
				},
			}
			// cloud-init script
			// VHI calls this user_data, just b64 encoded cloud-init script
			if flagUserData != "" {
				userData, err := readAndEncodeUserData(flagUserData)
				if err != nil {
					return err
				}
				request.Server.UserData = userData
			}
		} else {
			// Create blank volume if no image is specified
			// we get here if --netboot is flagged or --image is not provided
			// and there is no image in the Vyper config
			fmt.Printf("Creating blank boot volume for VM %s...\n", flagVMName)
			volRequest := api.CreateVolumeRequest{}
			volRequest.Volume.Name = fmt.Sprintf("%s-boot", flagVMName)
			volRequest.Volume.Size = volumeSize
			volRequest.Volume.Description = "Boot volume for " + flagVMName
			volRequest.Volume.VolumeType = "nvme_ec7_2"

			volResp, err := api.CreateVolume(storageURL, tok.Value, volRequest)
			if err != nil {
				return fmt.Errorf("failed to create blank boot volume: %v", err)
			}

			fmt.Printf("Waiting for volume to become available...\n")
			err = api.WaitForVolumeStatus(storageURL, tok.Value, volResp.Volume.ID, "available")
			if err != nil {
				return fmt.Errorf("failed waiting for volume: %v", err)
			}

			// Set bootable flag
			err = api.SetVolumeBootable(storageURL, tok.Value, volResp.Volume.ID, true)
			if err != nil {
				return fmt.Errorf("failed to set bootable flag: %v", err)
			}

			request.Server.BlockDeviceMappingV2 = []map[string]interface{}{
				{
					"boot_index":            "0",
					"uuid":                  volResp.Volume.ID,
					"source_type":           "volume",
					"destination_type":      "volume",
					"delete_on_termination": true,
				},
			}
		}

		// Create the VM
		fmt.Printf("Creating VM %s...\n", flagVMName)
		resp, err := api.CreateVM(computeURL, tok.Value, request)
		if err != nil {
			return fmt.Errorf("failed to create VM: %v", err)
		}

		// Wait for VM to become active
		vmDetails, err := api.WaitForStatus(computeURL, tok.Value, resp.Server.ID, "ACTIVE")
		if err != nil {
			return err
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

		// Iterate over user-provided networks and IPs
		for i, networkID := range networkIDs {
			ip := strings.TrimSpace(ipAddresses[i])
			fixedIPs := []string{}
			if ip != "" {
				fixedIPs = append(fixedIPs, ip)
			}

			fmt.Printf("Attaching network %s to VM %s...\n", networkID, resp.Server.ID)

			// Call API to attach the network interface
			interfaceResp, err := api.AttachNetworkToVM(computeURL, tok.Value, resp.Server.ID, networkID, "", "", fixedIPs)
			if err != nil {
				fmt.Printf("Failed to attach network with ip %s, retrying as unmanaged iface\n", ip)

				// Retry without fixed IP (unmanaged interface)
				interfaceResp, err = api.AttachNetworkToVM(computeURL, tok.Value, resp.Server.ID, networkID, "", "", nil)
				if err != nil {
					return fmt.Errorf("failed to attach network %s even without fixed IP", networkID)
				}
				fmt.Printf("Successfully attached unmanaged iface %s.\n", networkID)
			} else {
				fmt.Printf("Successfully attached network %s with fixed IP %s.\n", networkID, ip)
			}

			// Extract MAC address from the response
			macAddress := strings.ToUpper(interfaceResp.InterfaceAttachment.MacAddr)
			if macAddress == "" {
				macAddress = "UNKNOWN"
			}

			// Extract IP from response, if available (otherwise, use user-provided)
			attachedIP := ip // Default to user input
			if len(interfaceResp.InterfaceAttachment.FixedIPs) > 0 {
				attachedIP = interfaceResp.InterfaceAttachment.FixedIPs[0].IPAddress
			}

			// Append to network info
			netInfo = append(netInfo, map[string]interface{}{
				"network_id":  networkID,
				"mac_address": macAddress,
				"ip_address":  attachedIP,
			})
			time.Sleep(10 * time.Second) // sleep to ensure network is attached before next iteration
		}

		// Add network details to output
		if len(netInfo) > 0 {
			details["networks"] = netInfo
		}

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

		// Print netboot command if enabled
		if flagVMNetboot {
			consoleURL := fmt.Sprintf("%s:8800/compute/servers/instances/%s/console", tok.Host, vmDetails.ID)
			fmt.Printf("\nGo to VHI console to complete machine bootup and installation.")
			fmt.Printf("\nVHI console: %s\n", consoleURL)
		}

		return nil
	},
}

// Subcommand: create Image (and upload data)

var createImageCmd = &cobra.Command{
	Use:   "image",
	Short: "Create a new image",
	Long:  "Create a new image from a VM snapshot (.qcow2, .raw, .vmdk, .iso)",
	RunE: func(cmd *cobra.Command, args []string) error {
		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		// Validate file exists
		if _, err := os.Stat(flagImageFile); os.IsNotExist(err) {
			return fmt.Errorf("image file not found: %s", flagImageFile)
		}

		// Determine format from flag or file extension if not specified
		format := flagDiskFormat
		if format == "" {
			ext := strings.ToLower(filepath.Ext(flagImageFile))
			switch ext {
			case ".qcow2":
				format = "qcow2"
			case ".raw":
				format = "raw"
			case ".vmdk":
				format = "vmdk"
			case ".iso":
				format = "iso"
			default:
				return fmt.Errorf("unsupported image format %s, must specify --format flag", ext)
			}
		}

		// Validate format
		switch format {
		case "qcow2", "raw", "vmdk", "iso":
			// Valid formats
		default:
			return fmt.Errorf("unsupported format %s, must be qcow2, raw, vmdk or iso", format)
		}

		file, err := os.Open(flagImageFile)
		if err != nil {
			return fmt.Errorf("failed to open image file: %v", err)
		}
		defer file.Close()

		req := api.CreateImageRequest{
			Name:         flagImageName,
			ContainerFmt: "bare",
			DiskFmt:      format,
			Visibility:   "shared",
		}

		image, err := api.CreateAndUploadImage(imageURL, tok.Value, req, []byte(data))
		if err != nil {
			return fmt.Errorf("failed to create/upload image: %v", err)
		}

		fmt.Printf("Image created: ID: %s, Name: %s\n", image.ID, image.Name)
		return nil
	},
}

// Subcommand: create volume
var createVolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Create a new storage volume",
	RunE: func(cmd *cobra.Command, args []string) error {
		storageURL, err := validateTokenEndpoint(tok, "volumev3")
		if err != nil {
			return err
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
	flagIPCSV             string
	flagVolumeName        string
	flagVolumeSize        int
	flagVolumeDescription string
	flagVolumeType        string
	flagVMSize            int
	flagVMNetboot         bool
	flagUserData          string
	flagImageFile         string
	flagImageName         string
	flagDiskFormat        string
)

func init() {
	// Flags for create vm
	createVMCmd.Flags().StringVar(&flagVMName, "name", "", "Name of the virtual machine")
	createVMCmd.Flags().StringVar(&flagFlavorRef, "flavor", "", "Flavor ID for the virtual machine")
	createVMCmd.Flags().StringVar(&flagImageRef, "image", "", "Image ID for the virtual machine")
	createVMCmd.Flags().StringVar(&flagNetworkCSV, "networks", "", "Comma-separated list of network UUIDs")
	createVMCmd.Flags().StringVar(&flagIPCSV, "ips", "", "Comma-separated list of IP addresses")
	createVMCmd.Flags().BoolVar(&flagJsonOutput, "json", false, "Output in JSON format (default: YAML)")
	createVMCmd.Flags().IntVar(&flagVMSize, "size", 0, "Size in GB of boot volume")
	createVMCmd.Flags().BoolVar(&flagVMNetboot, "netboot", false, "Enable network boot with blank volume (deprecated, use --image)")
	createVMCmd.Flags().StringVar(&flagUserData, "user-data", "", "User data for cloud-init (file path)")

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

	// Flags for create image
	createImageCmd.Flags().StringVar(&flagImageFile, "file", "", "Path to the image file")
	createImageCmd.Flags().StringVar(&flagImageName, "name", "", "Name of the image")
	createImageCmd.Flags().StringVar(&flagDiskFormat, "format", "", "Disk format (qcow2, raw, vmdk)")

	// Add subcommands to the parent create command
	createCmd.AddCommand(createVMCmd)
	createCmd.AddCommand(createVolumeCmd)
	createCmd.AddCommand(createImageCmd)

	// Add the create command to the root command
	rootCmd.AddCommand(createCmd)
}
