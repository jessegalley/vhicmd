package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jessegalley/vhicmd/api"
	"github.com/jessegalley/vhicmd/internal/responseparser"
	"github.com/spf13/cobra"
)

var detailsCmd = &cobra.Command{
	Use:     "details",
	Aliases: []string{"show"},
	Short:   "Show details of resources",
}

var vmDetailsCmd = &cobra.Command{
	Use:   "vm [vm_id]",
	Short: "Show details of a specific VM",
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

		vm, err := api.GetVMDetails(computeURL, tok.Value, vmID)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(vm, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		// Extract security groups
		var secGroups []responseparser.SecurityGroupDetail
		for _, sg := range vm.SecurityGroups {
			secGroup := responseparser.SecurityGroupDetail{
				ID:          sg.ID,
				Name:        sg.Name,
				Description: sg.Description,
			}
			for _, rule := range sg.Rules {
				secGroup.Rules = append(secGroup.Rules, responseparser.SecurityGroupRule{
					ID:             rule.ID,
					Direction:      rule.Direction,
					Protocol:       rule.Protocol,
					PortRangeMin:   rule.PortRangeMin,
					PortRangeMax:   rule.PortRangeMax,
					RemoteIPPrefix: rule.RemoteIPPrefix,
					EtherType:      rule.EtherType,
				})
			}
			secGroups = append(secGroups, secGroup)
		}

		details := responseparser.VMDetails{
			ID:             vm.ID,
			Name:           vm.Name,
			Status:         vm.Status,
			PowerState:     vm.PowerState,
			Task:           vm.TaskState,
			Created:        vm.Created,
			Updated:        vm.Updated,
			ImageID:        vm.Image.ID,
			SecurityGroups: secGroups,
			Flavor: responseparser.FlavorDetail{
				ID:         vm.Flavor.ID,
				Name:       vm.Flavor.OriginalName,
				RAM:        vm.Flavor.RAM,
				VCPUs:      vm.Flavor.VCPUs,
				Disk:       vm.Flavor.Disk,
				Ephemeral:  vm.Flavor.Ephemeral,
				Swap:       vm.Flavor.Swap,
				ExtraSpecs: vm.Flavor.ExtraSpecs,
			},
			Metadata: vm.Metadata,
		}

		// Fetch network details (for managed networks)
		networkPorts, err := api.GetVMNetworks(computeURL, tok.Value, vmID)
		if err != nil {
			return err
		}

		// Track MACs to avoid duplication
		seenMACs := make(map[string]bool)

		// Process HCI networks and match with NetID from GetVMNetworks
		for _, hciNet := range vm.HCIInfo.Network {
			if seenMACs[hciNet.Mac] {
				continue
			}

			netDetail := responseparser.NetworkDetail{
				Name:    hciNet.Network.Label,
				UUID:    hciNet.Network.ID,
				MacAddr: hciNet.Mac,
				PortID:  "N/A",
			}

			// Match with networkPorts.InterfaceAttachments using NetID and MAC
			for _, port := range networkPorts.InterfaceAttachments {
				if port.NetID == hciNet.Network.ID && port.MacAddr == hciNet.Mac {
					netDetail.PortID = port.PortID
					// Add IPs if they exist
					for _, ip := range port.FixedIPs {
						netDetail.IPs = append(netDetail.IPs, responseparser.IPDetail{
							Address: ip.IPAddress,
						})
					}
					break
				}
			}

			// If there are no IPs, it's an unmanaged network
			if len(netDetail.IPs) == 0 {
				netDetail.IPs = []responseparser.IPDetail{{Address: "N/A"}}
			}

			seenMACs[hciNet.Mac] = true
			details.Networks = append(details.Networks, netDetail)
		}

		// Process volumes
		for _, vol := range vm.OSExtendedVolumesVolumesAttached {
			details.Volumes = append(details.Volumes, responseparser.VolumeDetail{
				ID:                  vol.ID,
				DeleteOnTermination: vol.DeleteOnTermination,
			})
		}

		responseparser.PrintVMDetailsTable([]responseparser.VMDetails{details})
		return nil
	},
}

var portDetailsCmd = &cobra.Command{
	Use:   "port [port_id]",
	Short: "Show details of a specific port",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		portID := args[0]

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		networkURL, err := validateTokenEndpoint(tok, "network")
		if err != nil {
			return err
		}

		port, err := api.GetPortDetails(networkURL, tok.Value, portID)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(port, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		// Convert fixed IPs to string array for display
		ips := make([]string, 0)
		for _, ip := range port.FixedIPs {
			ips = append(ips, ip.IPAddress)
		}

		vmName, err := api.GetVMNameByID(computeURL, tok.Value, port.DeviceID)
		if err != nil {
			vmName = port.DeviceID
		}

		details := responseparser.PortDetails{
			ID:              port.ID,
			MACAddress:      port.MACAddress,
			NetworkID:       port.NetworkID,
			DeviceID:        vmName,
			DeviceOwner:     port.DeviceOwner,
			Status:          port.Status,
			FixedIPs:        ips,
			SecurityGroups:  port.SecurityGroups,
			AdminStateUp:    port.AdminStateUp,
			BindingHostID:   port.BindingHostID,
			BindingVnicType: port.BindingVnicType,
			DNSDomain:       port.DNSDomain,
			DNSName:         port.DNSName,
			CreatedAt:       port.CreatedAt,
			UpdatedAt:       port.UpdatedAt,
		}

		responseparser.PrintPortDetailsTable(details)
		return nil
	},
}

var imageDetailsCmd = &cobra.Command{
	Use:     "image [image_id]",
	Aliases: []string{"img"},
	Short:   "Show details of a specific image",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageID := args[0]

		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		identityURL, err := validateTokenEndpoint(tok, "identity")
		if err != nil {
			return err
		}

		// Try to resolve name to ID first
		id, err := api.GetImageIDByName(imageURL, tok.Value, imageID)
		if err == nil {
			imageID = id
		}

		details, err := api.GetImageDetails(imageURL, tok.Value, imageID)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(details, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		ownerName, err := api.GetProjectNameByID(identityURL, tok.Value, details.Owner)
		if err != nil {
			ownerName = details.Owner
		} else {
			ownerName = fmt.Sprintf("%s (%s)", ownerName, details.Owner)
		}

		displayDetails := responseparser.ImageDetails{
			ID:              details.ID,
			Name:            details.Name,
			Status:          details.Status,
			Visibility:      details.Visibility,
			Size:            details.Size,
			VirtualSize:     details.VirtualSize,
			MinDisk:         details.MinDisk,
			MinRAM:          details.MinRAM,
			DiskFormat:      details.DiskFormat,
			ContainerFormat: details.ContainerFormat,
			CreatedAt:       details.CreatedAt,
			UpdatedAt:       details.UpdatedAt,
			Protected:       details.Protected,
			//Checksum:         details.Checksum,
			//OsHashAlgo:       details.OsHashAlgo,
			//OsHashValue:      details.OsHashValue,
			OsHidden:  details.OsHidden,
			Owner:     ownerName,
			Tags:      details.Tags,
			DirectURL: details.DirectURL,
			File:      details.File,
			Self:      details.Self,
			//Schema:           details.Schema,
			//HwQemuGuestAgent: details.HwQemuGuestAgent,
			//OsType:           details.OsType,
			//OsDistro:         details.OsDistro,
			ImageValidated: details.ImageValidated,
		}

		responseparser.PrintImageDetailsTable(displayDetails)
		return nil
	},
}

var volumeDetailsCmd = &cobra.Command{
	Use:     "volume [volume_id]",
	Aliases: []string{"vol"},
	Short:   "Show details of a specific volume",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

		storageURL, err := validateTokenEndpoint(tok, "volumev3")
		if err != nil {
			return err
		}

		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		v, err := api.GetVolumeIDByName(storageURL, tok.Value, volumeID)
		if err == nil {
			volumeID = v
		}

		volume, err := api.GetVolumeDetails(storageURL, tok.Value, volumeID)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(volume, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		// Convert API response to display format
		details := responseparser.VolumeDetails{
			ID:                 volume.ID,
			Name:               volume.Name,
			Status:             volume.Status,
			Size:               volume.Size,
			VolumeType:         volume.VolumeType,
			Bootable:           volume.Bootable,
			Multiattach:        volume.Multiattach,
			Encrypted:          volume.Encrypted,
			AvailabilityZone:   volume.AvailabilityZone,
			CreatedAt:          volume.CreatedAt,
			UpdatedAt:          volume.UpdatedAt,
			Description:        volume.Description,
			ReplicationStatus:  volume.ReplicationStatus,
			SnapshotID:         volume.SnapshotID,
			SourceVolID:        volume.SourceVolID,
			GroupID:            volume.GroupID,
			ConsistencyGroupID: volume.ConsistencyGroupID,
			ConsumesQuota:      volume.ConsumesQuota,
			Metadata:           volume.Metadata,
		}

		// Convert and resolve attachments
		for _, att := range volume.Attachments {
			attachment := responseparser.VolumeAttachment{
				ServerID:     att.ServerID,
				Device:       att.Device,
				AttachedAt:   att.AttachedAt,
				AttachmentID: att.AttachmentID,
			}

			// Try to resolve server name
			vmName, err := api.GetVMNameByID(computeURL, tok.Value, att.ServerID)
			if err == nil {
				attachment.ServerName = vmName
			}

			details.Attachments = append(details.Attachments, attachment)
		}

		responseparser.PrintVolumeDetailsTable(details)
		return nil
	},
}

func init() {
	detailsCmd.AddCommand(vmDetailsCmd)
	detailsCmd.AddCommand(portDetailsCmd)
	detailsCmd.AddCommand(imageDetailsCmd)
	detailsCmd.AddCommand(volumeDetailsCmd)
	rootCmd.AddCommand(detailsCmd)
	detailsCmd.PersistentFlags().BoolVar(&flagJsonOutput, "json", false, "Output in JSON format")
}
