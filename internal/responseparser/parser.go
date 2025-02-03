package responseparser

import (
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
)

// -------------------------------------------------------------------
// DOMAINS
// -------------------------------------------------------------------

type Domain struct {
	ID          string
	Name        string
	Enabled     bool
	Description string
}

func PrintDomainsTable(domains []Domain) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "ID", "ENABLED", "DESCRIPTION"})

	applyTableStyle(table)

	for _, d := range domains {
		table.Append([]string{
			d.Name,
			d.ID,
			colorStyleBool(d.Enabled),
			d.Description,
		})
	}
	table.Render()
}

// -------------------------------------------------------------------
// PROJECTS
// -------------------------------------------------------------------

type Project struct {
	ID       string
	Name     string
	DomainID string
	Enabled  bool
}

func PrintProjectsTable(projects []Project) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "ID", "DOMAIN_ID", "ENABLED"})

	applyTableStyle(table)

	for _, p := range projects {
		table.Append([]string{
			color.Style{color.FgGreen}.Render(p.Name),
			p.ID,
			p.DomainID,
			colorStyleBool(p.Enabled),
		})
	}
	table.Render()
}

// -------------------------------------------------------------------
// FLAVORS
// -------------------------------------------------------------------

type Flavor struct {
	ID          string
	Name        string
	Description string
}

func PrintFlavorsTable(flavors []Flavor) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "ID", "DESCRIPTION"})

	applyTableStyle(table)

	for _, f := range flavors {
		table.Append([]string{
			color.Style{color.FgGreen}.Render(f.Name),
			f.ID,
			f.Description,
		})
	}
	table.Render()
}

// -------------------------------------------------------------------
// IMAGES
// -------------------------------------------------------------------

type Image struct {
	ID      string
	Name    string
	Status  string
	Size    int64
	Owner   string
	MinDisk int
	MinRAM  int
}

func PrintImagesTable(images []Image) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name", "ID", "STATUS", "SIZE (Bytes)", "OWNER", "MinDisk", "MinRAM",
	})

	applyTableStyle(table)

	for _, i := range images {
		table.Append([]string{
			color.Style{color.FgGreen}.Render(i.Name),
			i.ID,
			colorStyleStatus(i.Status),
			fmt.Sprintf("%d", i.Size),
			i.Owner,
			fmt.Sprintf("%d", i.MinDisk),
			fmt.Sprintf("%d", i.MinRAM),
		})
	}
	table.Render()
}

// -------------------------------------------------------------------
// VOLUMES
// -------------------------------------------------------------------
// Volume represents a single volume object in the response.
type Volume struct {
	ID     string
	Name   string
	Size   int
	Status string
}

// PrintVolumesTable prints a table of volumes.
func PrintVolumesTable(volumes []Volume) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "ID", "SIZE", "STATUS"})

	applyTableStyle(table)

	for _, v := range volumes {
		table.Append([]string{
			color.Style{color.FgGreen}.Render(v.Name),
			v.ID,
			fmt.Sprintf("%d GB", v.Size),
			colorStyleVolAvailability(v.Status),
		})
	}
	table.Render()
}

// -------------------------------------------------------------------
// NETWORKS
// -------------------------------------------------------------------

type Network struct {
	ID       string
	Name     string
	Status   string
	Project  string
	Shared   bool
	External bool
	PortSec  bool
	CIDRs    string
}

func PrintNetworksTable(nets []Network) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "ID", "STATUS", "PROJECT", "MANAGED", "CIDRs"})

	applyTableStyle(table)

	for _, n := range nets {
		table.Append([]string{
			n.Name,
			n.ID,
			colorStyleStatus(n.Status),
			n.Project,
			colorStyleBool(n.PortSec),
			n.CIDRs,
		})
	}
	table.Render()
}

// -------------------------------------------------------------------
// VMs
// -------------------------------------------------------------------

type VM struct {
	ID   string
	Name string
}

type VMDetails struct {
	ID             string
	Name           string
	Status         string
	PowerState     int
	Task           string
	Created        string
	Updated        string
	ImageID        string
	SecurityGroups []SecurityGroupDetail
	Networks       []NetworkDetail
	Volumes        []VolumeDetail
	Flavor         FlavorDetail
	Metadata       map[string]string
}

type FlavorDetail struct {
	ID         string
	Name       string
	RAM        int
	VCPUs      int
	Disk       int
	Ephemeral  int
	Swap       int
	ExtraSpecs map[string]string
}

type NetworkDetail struct {
	Name    string
	UUID    string
	IPs     []IPDetail
	MacAddr string
	Type    string // e.g. "fixed" or "floating"
	PortID  string
}

type IPDetail struct {
	Address string
	Type    string
}

type VolumeDetail struct {
	ID                  string
	DeleteOnTermination bool
}

type SecurityGroupDetail struct {
	ID          string
	Name        string
	Description string
	Rules       []SecurityGroupRule
}

type SecurityGroupRule struct {
	ID             string
	Direction      string
	Protocol       string
	PortRangeMin   *int
	PortRangeMax   *int
	RemoteIPPrefix string
	EtherType      string
}

func PrintVMsTable(vms []VM) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "ID"})

	applyTableStyle(table)

	for _, vm := range vms {
		table.Append([]string{
			color.Style{color.FgGreen}.Render(vm.Name),
			vm.ID,
		})
	}
	table.Render()
}

func PrintVMDetailsTable(details []VMDetails) {
	d := details[0]

	// Basic Info Table
	fmt.Println("\nVM Info:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "ID", "Status", "State", "Task", "Created", "Updated"})
	applyTableStyle(table)

	table.Append([]string{
		color.Style{color.FgGreen}.Render(d.Name),
		d.ID,
		colorStyleStatus(d.Status),
		getPowerStateString(d.PowerState),
		stringOrNA(d.Task),
		d.Created,
		stringOrNA(d.Updated),
	})
	table.Render()

	// Image Details
	if d.ImageID != "" {
		fmt.Println("\nImage Info:")
		imgTable := tablewriter.NewWriter(os.Stdout)
		imgTable.SetHeader([]string{"Image ID"})
		applyTableStyle(imgTable)
		imgTable.Append([]string{d.ImageID})
		imgTable.Render()
	}

	// Network Details Table
	fmt.Println("\nNetworks:")
	netTable := tablewriter.NewWriter(os.Stdout)
	netTable.SetHeader([]string{"Network Name", "IP Address", "MAC Address", "Network UUID", "Port ID"})
	applyTableStyle(netTable)

	for _, net := range d.Networks {
		netName := color.Style{color.FgGreen}.Render(net.Name)
		if len(net.IPs) == 0 {
			// Show unmanaged networks
			netTable.Append([]string{
				netName,
				"N/A",
				"N/A",
				"N/A",
				net.MacAddr,
				net.UUID,
			})
		} else {
			// Show networks with IPs
			for _, ip := range net.IPs {
				netTable.Append([]string{
					netName,
					ip.Address,
					net.MacAddr,
					net.UUID,
					net.PortID,
				})
			}
		}
	}
	netTable.Render()

	// Volume Details
	if len(d.Volumes) > 0 {
		fmt.Println("\nVolumes:")
		volTable := tablewriter.NewWriter(os.Stdout)
		volTable.SetHeader([]string{"Volume ID", "Delete on Termination"})
		applyTableStyle(volTable)

		for _, vol := range d.Volumes {
			volTable.Append([]string{
				vol.ID,
				colorStyleBool(vol.DeleteOnTermination),
			})
		}
		volTable.Render()
	}

	// Security Groups
	if len(d.SecurityGroups) > 0 {
		fmt.Println("\nSecurity Groups:")
		secTable := tablewriter.NewWriter(os.Stdout)
		secTable.SetHeader([]string{"Name", "ID", "Description"})
		applyTableStyle(secTable)

		for _, sg := range d.SecurityGroups {
			secTable.Append([]string{
				color.Style{color.FgGreen}.Render(sg.Name),
				sg.ID,
				stringOrNA(sg.Description),
			})
		}
		secTable.Render()

		// Only show rules table if there are actually rules
		rulesExist := false
		for _, sg := range d.SecurityGroups {
			if len(sg.Rules) > 0 {
				rulesExist = true
				break
			}
		}

		if rulesExist {
			fmt.Println("\nSecurity Group Rules:")
			ruleTable := tablewriter.NewWriter(os.Stdout)
			ruleTable.SetHeader([]string{"Group Name", "Direction", "Protocol", "Ports", "Remote CIDR", "EtherType"})
			applyTableStyle(ruleTable)

			for _, sg := range d.SecurityGroups {
				for _, rule := range sg.Rules {
					ports := "Any"
					if rule.PortRangeMin != nil {
						if rule.PortRangeMax != nil && *rule.PortRangeMin == *rule.PortRangeMax {
							ports = fmt.Sprintf("%d", *rule.PortRangeMin)
						} else if rule.PortRangeMax != nil {
							ports = fmt.Sprintf("%d-%d", *rule.PortRangeMin, *rule.PortRangeMax)
						}
					}

					ruleTable.Append([]string{
						color.Style{color.FgGreen}.Render(sg.Name),
						rule.Direction,
						stringOrNA(rule.Protocol),
						ports,
						stringOrNA(rule.RemoteIPPrefix),
						rule.EtherType,
					})
				}
			}
			ruleTable.Render()
		}
	}

	// Flavor Details
	fmt.Println("\nFlavor Details:")
	flavorTable := tablewriter.NewWriter(os.Stdout)
	flavorTable.SetHeader([]string{"Name", "ID", "RAM (MB)", "VCPUs", "Disk (GB)", "Ephemeral", "Swap"})
	applyTableStyle(flavorTable)

	flavorTable.Append([]string{
		color.Style{color.FgGreen}.Render(stringOrNA(d.Flavor.Name)),
		d.Flavor.ID,
		fmt.Sprintf("%d", d.Flavor.RAM),
		fmt.Sprintf("%d", d.Flavor.VCPUs),
		fmt.Sprintf("%d", d.Flavor.Disk),
		fmt.Sprintf("%d", d.Flavor.Ephemeral),
		fmt.Sprintf("%d", d.Flavor.Swap),
	})
	flavorTable.Render()

	if len(d.Flavor.ExtraSpecs) > 0 {
		fmt.Println("\nFlavor Extra Specs:")
		extraTable := tablewriter.NewWriter(os.Stdout)
		extraTable.SetHeader([]string{"Key", "Value"})
		applyTableStyle(extraTable)

		for k, v := range d.Flavor.ExtraSpecs {
			extraTable.Append([]string{k, v})
		}
		extraTable.Render()
	}

	// Display Metadata
	if len(d.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		metaTable := tablewriter.NewWriter(os.Stdout)
		metaTable.SetHeader([]string{"Key", "Value"})
		applyTableStyle(metaTable)

		for k, v := range d.Metadata {
			metaTable.Append([]string{k, v})
		}
		metaTable.Render()
	}
}

// -------------------------------------------------------------------
// CATALOG
// -------------------------------------------------------------------

type CatalogEntry struct {
	Type      string
	Name      string
	Interface string
	Region    string
	URL       string
}

func PrintCatalogTable(entries []CatalogEntry) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "TYPE", "INTERFACE", "REGION", "URL"})

	applyTableStyle(table)

	for _, e := range entries {
		table.Append([]string{
			color.Style{color.FgGreen}.Render(e.Name),
			e.Type,
			e.Interface,
			e.Region,
			e.URL,
		})
	}
	table.Render()
}
