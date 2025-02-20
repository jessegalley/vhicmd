package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/facette/natsort"
	"github.com/jessegalley/vhicmd/api"
	"github.com/jessegalley/vhicmd/internal/responseparser"
	"github.com/spf13/cobra"
)

var flagJsonOutput bool

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List various objects in OpenStack/VHI (domains, projects, etc.)",
	Long:    "List subcommand for domains, projects, or other items in the system. Requires a valid auth token.",
}

var listDomainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "List domains [Req: admin]",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Call the API
		resp, err := api.ListDomains(tok.Host, tok.Value)
		if err != nil {
			return err
		}

		// Check if user passed --json
		if flagJsonOutput {
			// Original JSON approach
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
		} else {
			// Convert resp.Domains into our local Domain type if needed
			// or just pass the raw if the shape matches
			var domainList []responseparser.Domain
			for _, d := range resp.Domains {
				domainList = append(domainList, responseparser.Domain{
					Description: d.Description,
					Enabled:     d.Enabled,
					ID:          d.ID,
					Name:        d.Name,
				})
			}
			// Now call the pretty-print function
			responseparser.PrintDomainsTable(domainList)
		}
		return nil
	},
}

var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		identityUrl, err := validateTokenEndpoint(tok, "identity")
		if err != nil {
			return err
		}

		resp, err := api.ListProjects(identityUrl, tok.Value)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		var projectList []responseparser.Project
		for _, p := range resp.Projects {
			projectList = append(projectList, responseparser.Project{
				ID:       p.ID,
				Name:     p.Name,
				DomainID: p.DomainID,
				Enabled:  p.Enabled,
			})
		}
		sort.Slice(projectList, func(i, j int) bool {
			return natsort.Compare(projectList[i].Name, projectList[j].Name)
		})

		responseparser.PrintProjectsTable(projectList)
		return nil
	},
}

var listFlavorsCmd = &cobra.Command{
	Use:   "flavors",
	Short: "List flavors",
	RunE: func(cmd *cobra.Command, args []string) error {
		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		// Gather optional query parameters
		queryParams := make(map[string]string)
		if projectID, _ := cmd.Flags().GetString("project-id"); projectID != "" {
			queryParams["project_id"] = projectID
		}
		if sortKey, _ := cmd.Flags().GetString("sort-key"); sortKey != "" {
			queryParams["sort_key"] = sortKey
		}
		if sortDir, _ := cmd.Flags().GetString("sort-dir"); sortDir != "" {
			queryParams["sort_dir"] = sortDir
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", limit)
		}
		if marker, _ := cmd.Flags().GetString("marker"); marker != "" {
			queryParams["marker"] = marker
		}
		if minDisk, _ := cmd.Flags().GetInt("min-disk"); minDisk > 0 {
			queryParams["minDisk"] = fmt.Sprintf("%d", minDisk)
		}
		if minRam, _ := cmd.Flags().GetInt("min-ram"); minRam > 0 {
			queryParams["minRam"] = fmt.Sprintf("%d", minRam)
		}
		if isPublic, _ := cmd.Flags().GetString("is-public"); isPublic != "" {
			queryParams["is_public"] = isPublic
		}

		resp, err := api.ListFlavors(computeURL, tok.Value, queryParams)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			// JSON
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		// Table
		var flavorList []responseparser.Flavor
		for _, f := range resp.Flavors {
			flavorList = append(flavorList, responseparser.Flavor{
				ID:          f.ID,
				Name:        f.Name,
				Description: f.Description,
			})
		}
		sort.Slice(flavorList, func(i, j int) bool {
			return natsort.Compare(flavorList[i].Name, flavorList[j].Name)
		})

		responseparser.PrintFlavorsTable(flavorList)
		return nil
	},
}

var listImagesCmd = &cobra.Command{
	Use:     "images",
	Aliases: []string{"img", "image"},
	Short:   "List virtual machine images",
	RunE: func(cmd *cobra.Command, args []string) error {
		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		queryParams := make(map[string]string)
		if visibility, _ := cmd.Flags().GetString("visibility"); visibility != "" {
			queryParams["visibility"] = visibility
		}
		if status, _ := cmd.Flags().GetString("status"); status != "" {
			queryParams["status"] = status
		}
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", limit)
		}
		if marker, _ := cmd.Flags().GetString("marker"); marker != "" {
			queryParams["marker"] = marker
		}

		resp, err := api.ListImages(imageURL, tok.Value, queryParams)
		if err != nil {
			return err
		}

		nameFilter, _ := cmd.Flags().GetString("name")

		var imgList []responseparser.Image
		for _, i := range resp.Images {
			if nameFilter == "" || strings.Contains(
				strings.ToLower(i.Name),
				strings.ToLower(nameFilter),
			) {
				imgList = append(imgList, responseparser.Image{
					ID:         i.ID,
					Name:       i.Name,
					Status:     i.Status,
					Size:       i.Size,
					Owner:      i.Owner,
					MinDisk:    i.MinDisk,
					MinRAM:     i.MinRAM,
					Visibility: i.Visibility,
				})
			}
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(imgList, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		sort.Slice(imgList, func(i, j int) bool {
			return natsort.Compare(imgList[i].Name, imgList[j].Name)
		})

		responseparser.PrintImagesTable(imgList)
		return nil
	},
}

var listImageMembersCmd = &cobra.Command{
	Use:   "image-members <image>",
	Short: "List members with access to a shared image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageID := args[0]

		imageURL, err := validateTokenEndpoint(tok, "image")
		if err != nil {
			return err
		}

		id, err := api.GetImageIDByName(imageURL, tok.Value, imageID)
		if err == nil {
			imageID = id
		}

		resp, err := api.ListImageMembers(imageURL, tok.Value, imageID)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		var memberList []responseparser.ImageMember
		for _, m := range resp.Members {
			memberList = append(memberList, responseparser.ImageMember{
				MemberID:  m.MemberID,
				Status:    m.Status,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			})
		}

		responseparser.PrintImageMembersTable(memberList)
		return nil
	},
}

var listNetworksCmd = &cobra.Command{
	Use:     "networks",
	Aliases: []string{"nets", "net", "network"},
	Short:   "List virtual networks",
	RunE: func(cmd *cobra.Command, args []string) error {
		networkURL, err := validateTokenEndpoint(tok, "network")
		if err != nil {
			return err
		}

		queryParams := make(map[string]string)
		if projectID, _ := cmd.Flags().GetString("project-id"); projectID != "" {
			queryParams["project_id"] = projectID
		}
		if status, _ := cmd.Flags().GetString("status"); status != "" {
			queryParams["status"] = status
		}

		resp, err := api.ListNetworks(networkURL, tok.Value, queryParams)
		if err != nil {
			return err
		}

		// Get the name filter
		nameFilter, _ := cmd.Flags().GetString("name")

		// Filter networks based on name containing the filter string
		var filteredNetworks []responseparser.Network
		for _, n := range resp.Networks {
			if nameFilter == "" || strings.Contains(strings.ToLower(n.Name), strings.ToLower(nameFilter)) {
				CIDRs := ""
				for _, subnetID := range n.SubnetIDs {
					subnet, _ := api.GetSubnetDetails(networkURL, tok.Value, subnetID)
					CIDRs += subnet.CIDR
					if subnetID != n.SubnetIDs[len(n.SubnetIDs)-1] {
						CIDRs += ","
					}
				}

				filteredNetworks = append(filteredNetworks, responseparser.Network{
					ID:       n.ID,
					Name:     n.Name,
					Status:   n.Status,
					Project:  n.ProjectID,
					Shared:   n.Shared,
					External: n.RouterExternal,
					PortSec:  n.PortSecurityEnabled,
					CIDRs:    CIDRs,
				})
			}
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(filteredNetworks, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		sort.Slice(filteredNetworks, func(i, j int) bool {
			return natsort.Compare(filteredNetworks[i].Name, filteredNetworks[j].Name)
		})
		responseparser.PrintNetworksTable(filteredNetworks)
		return nil
	},
}

var listPortsCmd = &cobra.Command{
	Use:     "ports",
	Aliases: []string{"nics", "interfaces", "nic", "interface"},
	Short:   "List network ports",
	RunE: func(cmd *cobra.Command, args []string) error {
		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}
		networkURL, err := validateTokenEndpoint(tok, "network")
		if err != nil {
			return err
		}

		queryParams := make(map[string]string)
		if deviceID, _ := cmd.Flags().GetString("device-id"); deviceID != "" {
			queryParams["device_id"] = deviceID
		}
		if networkID, _ := cmd.Flags().GetString("network-id"); networkID != "" {
			queryParams["network_id"] = networkID
		}
		if macAddress, _ := cmd.Flags().GetString("mac-address"); macAddress != "" {
			queryParams["mac_address"] = macAddress
		}
		if status, _ := cmd.Flags().GetString("status"); status != "" {
			queryParams["status"] = status
		}

		resp, err := api.ListPorts(networkURL, tok.Value, queryParams)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		var portList []responseparser.Port
		for _, p := range resp.Ports {
			vmName, err := api.GetVMNameByID(computeURL, tok.Value, p.DeviceID)
			if err != nil {
				vmName = p.DeviceID
			}

			ips := make([]string, 0)
			for _, ip := range p.FixedIPs {
				ips = append(ips, ip.IPAddress)
			}

			portList = append(portList, responseparser.Port{
				ID:          p.ID,
				MACAddress:  p.MACAddress,
				NetworkID:   p.NetworkID,
				DeviceID:    vmName,
				DeviceOwner: p.DeviceOwner,
				Status:      p.Status,
				FixedIPs:    strings.Join(ips, ", "),
			})
		}
		sort.Slice(portList, func(i, j int) bool {
			return natsort.Compare(portList[i].DeviceID, portList[j].DeviceID)
		})

		responseparser.PrintPortsTable(portList)
		return nil
	},
}

var listVmCmd = &cobra.Command{
	Use:     "vms",
	Aliases: []string{"vm", "instances", "servers"},
	Short:   "List virtual machines",
	Long:    "Fetches and displays a list of virtual machines in the project (determined by auth).",
	RunE: func(cmd *cobra.Command, args []string) error {
		computeURL, err := validateTokenEndpoint(tok, "compute")
		if err != nil {
			return err
		}

		queryParams := make(map[string]string)
		if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
			queryParams["limit"] = fmt.Sprintf("%d", limit)
		}
		if marker, _ := cmd.Flags().GetString("marker"); marker != "" {
			queryParams["marker"] = marker
		}

		resp, err := api.ListVMs(computeURL, tok.Value, queryParams)
		if err != nil {
			return err
		}

		nameFilter, _ := cmd.Flags().GetString("name")
		var vmList []responseparser.VM
		for _, v := range resp.Servers {
			if nameFilter == "" || strings.Contains(
				strings.ToLower(v.Name),
				strings.ToLower(nameFilter),
			) {
				vmList = append(vmList, responseparser.VM{
					ID:   v.ID,
					Name: v.Name,
				})
			}
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(vmList, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		sort.Slice(vmList, func(i, j int) bool {
			return natsort.Compare(vmList[i].Name, vmList[j].Name)
		})
		responseparser.PrintVMsTable(vmList)
		return nil
	},
}

var listVolumesCmd = &cobra.Command{
	Use:     "volumes",
	Aliases: []string{"vol", "vols", "storage", "volume"},
	Short:   "List storage volumes",
	RunE: func(cmd *cobra.Command, args []string) error {
		storageURL, err := validateTokenEndpoint(tok, "volumev3")
		if err != nil {
			return err
		}

		queryParams := make(map[string]string)
		resp, err := api.ListVolumes(storageURL, tok.Value, queryParams)
		if err != nil {
			return err
		}

		if flagJsonOutput {
			b, _ := json.MarshalIndent(resp.Volumes, "", "  ")
			fmt.Println(string(b))
			return nil
		}

		var volumeList []responseparser.Volume
		for _, v := range resp.Volumes {
			volumeList = append(volumeList, responseparser.Volume{
				ID:     v.ID,
				Name:   v.Name,
				Size:   v.Size,
				Status: v.Status,
			})
		}
		sort.Slice(volumeList, func(i, j int) bool {
			return natsort.Compare(volumeList[i].Name, volumeList[j].Name)
		})
		responseparser.PrintVolumesTable(volumeList)
		return nil
	},
}

func init() {
	listCmd.PersistentFlags().BoolVar(&flagJsonOutput, "json", false, "Output in JSON format")

	listImagesCmd.Flags().String("name", "", "Filter by image name")
	listImagesCmd.Flags().String("visibility", "", "Filter by visibility (public, private, etc.)")
	listImagesCmd.Flags().String("status", "", "Filter by image status")
	listImagesCmd.Flags().Int("limit", 0, "Limit the number of images returned")
	listImagesCmd.Flags().String("marker", "", "Marker for pagination")

	listNetworksCmd.Flags().String("name", "", "Filter networks by name")
	listNetworksCmd.Flags().String("status", "", "Filter networks by status (e.g., ACTIVE)")
	listNetworksCmd.Flags().String("project-id", "", "Filter networks by project ID")

	listPortsCmd.Flags().String("name", "", "Filter ports by name")

	listVmCmd.Flags().String("name", "", "Filter by VM name")
	listVmCmd.Flags().String("status", "", "Filter by VM status")
	listVmCmd.Flags().Int("limit", 0, "Limit the number of VMs returned")
	listVmCmd.Flags().String("marker", "", "Marker for pagination")

	listFlavorsCmd.Flags().String("project-id", "", "Project ID")
	listFlavorsCmd.Flags().String("sort-key", "", "Sort key for flavors")
	listFlavorsCmd.Flags().String("sort-dir", "", "Sort direction (asc or desc)")
	listFlavorsCmd.Flags().Int("limit", 0, "Limit the number of flavors returned")
	listFlavorsCmd.Flags().String("marker", "", "Marker for pagination")
	listFlavorsCmd.Flags().Int("min-disk", 0, "Minimum disk size (GiB)")
	listFlavorsCmd.Flags().Int("min-ram", 0, "Minimum RAM size (MiB)")
	listFlavorsCmd.Flags().String("is-public", "", "Filter by public/private flavors")

	// Mark project-id as required
	//listFlavorsCmd.MarkFlagRequired("project-id")

	listCmd.AddCommand(listDomainsCmd)
	listCmd.AddCommand(listProjectsCmd)
	listCmd.AddCommand(listNetworksCmd)
	listCmd.AddCommand(listPortsCmd)
	listCmd.AddCommand(listFlavorsCmd)
	listCmd.AddCommand(listVmCmd)
	listCmd.AddCommand(listImagesCmd)
	listCmd.AddCommand(listVolumesCmd)
	listCmd.AddCommand(listImageMembersCmd)
	rootCmd.AddCommand(listCmd)
}
