package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jessegalley/vhicmd/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 'migrate' parent command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate resources",
}

// 'migrate vm' subcommand
var migrateVMCmd = &cobra.Command{
	Use:   "vm",
	Short: "Migrate a virtual machine from a VMWare VMDK",
	Long: `Example:
  vhicmd migrate vm \
    --name MyVM \
    --vmdk /path/to/disk.vmdk \
    --flavor myflavor \
    --networks netA,netB \
    --mac aa:aa:aa:aa:aa:aa,bb:bb:bb:bb:bb:bb \
    --size 20 \
    --shutdown`,
	RunE: func(cmd *cobra.Command, args []string) error {
		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}
		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}
		networkURL, err := validateTokenEndpoint(tok, "network")
		if err != nil {
			return err
		}

		if migrateFlagVMName == "" {
			return fmt.Errorf("must provide --name for the VM")
		}
		if migrateFlagVMDKPath == "" {
			return fmt.Errorf("must provide --vmdk /path/to/image for migration")
		}

		flavorRef := migrateFlagFlavorRef
		if flavorRef == "" {
			flavorRef = viper.GetString("flavor_id")
		}
		if flavorRef == "" {
			return fmt.Errorf("no flavor specified; provide --flavor or set 'flavor_id' in config")
		}

		networks := migrateFlagNetworkCSV
		if networks == "" {
			networks = viper.GetString("networks")
		}
		if networks == "" {
			return fmt.Errorf("no networks specified; provide --networks or set 'networks' in config")
		}
		macs := migrateFlagMacAddrCSV
		networkIDs := strings.Split(networks, ",")
		macAddresses := strings.Split(macs, ",")
		if len(networkIDs) != len(macAddresses) {
			return fmt.Errorf("the number of networks must match the number of MAC addresses")
		}

		fid, err := api.GetFlavorIDByName(computeURL, tok.Value, flavorRef)
		if err == nil && fid != "" {
			flavorRef = fid
		}

		fmt.Printf("Creating temporary image for VM '%s'...\n", migrateFlagVMName)

		f, err := os.Open(migrateFlagVMDKPath)
		if err != nil {
			return fmt.Errorf("failed to open image file: %v", err)
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat image file: %v", err)
		}
		fileSize := info.Size()

		progressReader := newProgressReader(f, fileSize)

		imgReq := api.CreateImageRequest{
			Name:         fmt.Sprintf("Migrated-%s", migrateFlagVMName),
			ContainerFmt: "bare",
			DiskFmt:      "vmdk",
			Visibility:   "shared",
		}

		imageID, err := api.CreateAndUploadImage(imageURL, tok.Value, imgReq, progressReader)
		if err != nil {
			return fmt.Errorf("failed to create/upload image: %v", err)
		}
		fmt.Printf("\nImage created: %s\n", imageID)

		volumeSize := migrateFlagVMSize
		if volumeSize <= 0 {
			volumeSize = 10 // default
		}

		vmReq := api.CreateVMRequest{}
		vmReq.Server.Name = migrateFlagVMName
		vmReq.Server.FlavorRef = flavorRef
		vmReq.Server.ImageRef = imageID
		vmReq.Server.Networks = "none"

		// Force SATA block device
		// NOTE: This is a bit of a hack to force the use of SATA for the root volume
		// so udev in the VM uses /dev/sda instead of /dev/vda, as with VMWare.
		vmReq.Server.BlockDeviceMappingV2 = []map[string]interface{}{
			{
				"boot_index":            0,
				"uuid":                  imageID,
				"source_type":           "image",
				"destination_type":      "volume",
				"volume_size":           volumeSize,
				"delete_on_termination": true,
				"disk_bus":              "sata",
				"volume_type":           "nvme_ec7_2",
			},
		}

		fmt.Printf("Creating VM '%s'...\n", migrateFlagVMName)
		vmResp, err := api.CreateVM(computeURL, tok.Value, vmReq)
		if err != nil {
			return fmt.Errorf("failed to create VM: %v", err)
		}

		// Wait for ACTIVE
		vmDetails, err := api.WaitForStatus(computeURL, tok.Value, vmResp.Server.ID, "ACTIVE")
		if err != nil {
			return fmt.Errorf("failed waiting for VM to become ACTIVE: %v", err)
		}

		netInfo := make([]map[string]interface{}, 0)
		for i, netNameOrID := range networkIDs {
			netNameOrID = strings.TrimSpace(netNameOrID)
			macAddr := strings.TrimSpace(macAddresses[i])

			// Try to resolve network name->ID
			netID, err := api.GetNetworkIDByName(networkURL, tok.Value, netNameOrID)
			if err == nil && netID != "" {
				netNameOrID = netID
			}

			fmt.Printf("Attaching network '%s' to VM '%s' with MAC '%s'...\n",
				netNameOrID, vmDetails.ID, macAddr)

			// Create an unmanaged port with the custom MAC
			portResp, err := api.CreatePort(networkURL, tok.Value, netNameOrID, macAddr)
			if err != nil {
				return fmt.Errorf("failed to create port on network %s: %v", netNameOrID, err)
			}

			// Attach the port to the VM
			_, err = api.AttachNetworkToVM(networkURL, computeURL, tok.Value, vmDetails.ID, "", portResp.Port.ID, "", nil)
			if err != nil {
				return fmt.Errorf("failed to attach port '%s' to VM '%s': %v", portResp.Port.ID, vmDetails.ID, err)
			}

			netInfo = append(netInfo, map[string]interface{}{
				"network_id":  netNameOrID,
				"mac_address": macAddr,
			})
		}

		if migrateFlagShutdown {
			fmt.Printf("Shutting down VM '%s'...\n", vmDetails.ID)
			if err := api.StopVM(computeURL, tok.Value, vmDetails.ID); err != nil {
				return fmt.Errorf("failed to shut down VM: %v", err)
			}
		}

		fmt.Printf("Deleting temporary image %s...\n", imageID)
		err = api.DeleteImage(imageURL, tok.Value, imageID)
		if err != nil {
			return fmt.Errorf("failed to delete temporary image: %v", err)
		}

		summary := map[string]interface{}{
			"vm_id":   vmDetails.ID,
			"vm_name": vmDetails.Name,
			"power_state": fmt.Sprintf("%d (%s)",
				vmDetails.PowerState,
				getPowerStateString(vmDetails.PowerState)),
			"networks": netInfo,
		}

		data, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println(string(data))

		return nil
	},
}

// Flags for migrate vm
var (
	migrateFlagVMName     string
	migrateFlagVMDKPath   string
	migrateFlagFlavorRef  string
	migrateFlagNetworkCSV string
	migrateFlagMacAddrCSV string
	migrateFlagVMSize     int
	migrateFlagShutdown   bool
)

func init() {
	migrateVMCmd.Flags().StringVar(&migrateFlagVMName, "name", "", "Name of the VM")
	migrateVMCmd.Flags().StringVar(&migrateFlagVMDKPath, "vmdk", "", "Local path to VMDK file")
	migrateVMCmd.Flags().StringVar(&migrateFlagFlavorRef, "flavor", "", "Flavor name or ID")
	migrateVMCmd.Flags().StringVar(&migrateFlagNetworkCSV, "networks", "", "Comma-separated network names/IDs")
	migrateVMCmd.Flags().StringVar(&migrateFlagMacAddrCSV, "mac", "", "Comma-separated MAC addresses (one per network)")
	migrateVMCmd.Flags().IntVar(&migrateFlagVMSize, "size", 10, "Size of the root volume in GB if extending the image")
	migrateVMCmd.Flags().BoolVar(&migrateFlagShutdown, "shutdown", false, "Shut down the VM after creation")

	migrateCmd.AddCommand(migrateVMCmd)

	rootCmd.AddCommand(migrateCmd)
}
