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
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources like VMs or volumes",
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

		name := flagImageName
		if name == "" {
			name = fmt.Sprintf("%s-%s", filepath.Base(flagImageFile), time.Now().Format("20060102-150405"))
		}

		// Get file size for progress
		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat file: %v", err)
		}

		fmt.Printf("Starting upload of %s (%d MB)\n", flagImageFile, info.Size()/1024/1024)
		progressReader := newProgressReader(file, info.Size())

		req := api.CreateImageRequest{
			Name:         name,
			ContainerFmt: "bare",
			DiskFmt:      format,
			Visibility:   "shared",
		}

		imageID, err := api.CreateAndUploadImage(imageURL, tok.Value, req, progressReader)
		if err != nil {
			return fmt.Errorf("failed to create/upload image: %v", err)
		}

		fmt.Printf("\nImage created: ID: %s, Name: %s\n", imageID, name)
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

var createPortCmd = &cobra.Command{
	Use:   "port",
	Short: "Create a network port",
	RunE: func(cmd *cobra.Command, args []string) error {
		networkURL, err := validateTokenEndpoint(tok, "network")
		if err != nil {
			return err
		}

		// Check required network flag
		networkID := flagPortNetwork
		if networkID == "" {
			return fmt.Errorf("network is required: specify with --network flag")
		}

		// Check if network exists by name first
		fmt.Printf("Checking network ID for %s\n", networkID)
		netID, err := api.GetNetworkIDByName(networkURL, tok.Value, networkID)
		if err == nil {
			fmt.Printf("Network found: %s\n", netID)
			networkID = netID
		} else {
			fmt.Printf("Network ID not found by name, using as-is: %s\n", err)
		}

		// Create port
		resp, err := api.CreatePort(networkURL, tok.Value, networkID, flagPortMAC)
		if err != nil {
			return fmt.Errorf("failed to create port: %v", err)
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Printf("Port created successfully:\n")
			fmt.Printf("  ID: %s\n", resp.Port.ID)
			fmt.Printf("  MAC: %s\n", resp.Port.MACAddress)
			fmt.Printf("  Network: %s\n", resp.Port.NetworkID)
			fmt.Printf("  Status: %s\n", resp.Port.Status)
		}

		return nil
	},
}

var (
	flagVolumeName        string
	flagVolumeSize        int
	flagVolumeDescription string
	flagVolumeType        string
	flagImageFile         string
	flagImageName         string
	flagDiskFormat        string
	flagPortNetwork       string
	flagPortMAC           string
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
	createVMCmd.Flags().StringVar(&flagMacAddrCSV, "macaddr", "", "Comma-separated list of MAC addresses")

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

	// Flags for create port
	createPortCmd.Flags().StringVar(&flagPortNetwork, "network", "", "Network ID or name")
	createPortCmd.Flags().StringVar(&flagPortMAC, "mac", "", "MAC address")

	// Add subcommands to the parent create command
	createCmd.AddCommand(createVMCmd)
	createCmd.AddCommand(createVolumeCmd)
	createCmd.AddCommand(createImageCmd)
	createCmd.AddCommand(createPortCmd)

	// Add the create command to the root command
	rootCmd.AddCommand(createCmd)
}
