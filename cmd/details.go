package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jessegalley/vhicmd/api"
	"github.com/jessegalley/vhicmd/internal/responseparser"
	"github.com/spf13/cobra"
)

var detailsCmd = &cobra.Command{
	Use:   "details",
	Short: "Show details of resources",
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
				PortID:  "N/A", // Default, updated if found
			}

			// Match with networkPorts.InterfaceAttachments using NetID
			for _, port := range networkPorts.InterfaceAttachments {
				if port.NetID == hciNet.Network.ID {
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

			// If there are no IPs, it's an unmanaged network, but still has a Port ID
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

func init() {
	detailsCmd.AddCommand(vmDetailsCmd)
	rootCmd.AddCommand(detailsCmd)
	detailsCmd.PersistentFlags().BoolVar(&flagJsonOutput, "json", false, "Output in JSON format")
}
