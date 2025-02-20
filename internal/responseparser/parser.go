package responseparser

import (
	"fmt"
	"os"
	"text/tabwriter"

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

func PrintProjectsSelectionTable(projects []Project) {
	var enabledProjects []Project
	for _, p := range projects {
		if p.Enabled {
			enabledProjects = append(enabledProjects, p)
		}
	}

	if len(enabledProjects) == 0 {
		fmt.Println("No enabled projects found")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	totalProjects := len(enabledProjects)
	useColumns := totalProjects > 10
	rows := totalProjects
	if useColumns {
		rows = (totalProjects + 1) / 2
	}

	for i := 0; i < rows; i++ {
		number := color.Style{color.OpBold}.Sprintf("%2d", i+1)

		if useColumns && i+rows < totalProjects {
			number2 := color.Style{color.OpBold}.Sprintf("%2d", i+rows+1)
			fmt.Fprintf(w, "%s) %s\t\t\t\t%s) %s\n", number, enabledProjects[i].Name,
				number2, enabledProjects[i+rows].Name)
		} else {
			fmt.Fprintf(w, "%s) %s\n", number, enabledProjects[i].Name)
		}
	}
	w.Flush()
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
	ID         string
	Name       string
	Status     string
	Size       int64
	Owner      string
	MinDisk    int
	MinRAM     int
	Visibility string
}

type ImageDetails struct {
	ID               string
	Name             string
	Status           string
	Visibility       string
	Size             int64
	VirtualSize      int64
	MinDisk          int
	MinRAM           int
	DiskFormat       string
	ContainerFormat  string
	CreatedAt        string
	UpdatedAt        string
	Protected        bool
	Checksum         string
	OsHashAlgo       string
	OsHashValue      string
	OsHidden         bool
	Owner            string
	Tags             []string
	DirectURL        string
	File             string
	Self             string
	Schema           string
	HwQemuGuestAgent string
	OsType           string
	OsDistro         string
	ImageValidated   string
}

type ImageMember struct {
	MemberID  string
	Status    string
	CreatedAt string
	UpdatedAt string
}

func PrintImageMembersTable(members []ImageMember) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Project ID", "Status", "Created", "Updated"})
	applyTableStyle(table)

	for _, m := range members {
		table.Append([]string{
			m.MemberID,
			colorStyleStatus(m.Status),
			m.CreatedAt,
			stringOrNA(m.UpdatedAt),
		})
	}
	table.Render()
}

func PrintImagesTable(images []Image) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name", "ID", "STATUS", "VIS", "SIZE (B)", "MinDisk", "MinRAM",
	})

	applyTableStyle(table)

	for _, i := range images {
		vis := ""
		switch i.Visibility {
		case "public":
			vis = "pub"
		case "private":
			vis = "priv"
		case "shared":
			vis = "shrd"
		case "community":
			vis = "comm"
		}
		table.Append([]string{
			color.Style{color.FgGreen}.Render(i.Name),
			i.ID,
			colorStyleStatus(i.Status),
			vis,
			fmt.Sprintf("%d", i.Size),
			fmt.Sprintf("%d", i.MinDisk),
			fmt.Sprintf("%d", i.MinRAM),
		})
	}
	table.Render()
}

func PrintImageDetailsTable(details ImageDetails) {
	fmt.Println("\nImage Info:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Field", "Value"})
	applyTableStyle(table)

	// Core Image Details
	table.Append([]string{"ID", details.ID})
	table.Append([]string{"Name", color.Style{color.FgGreen}.Render(details.Name)})
	table.Append([]string{"Status", colorStyleStatus(details.Status)})
	visColor := color.FgYellow
	if details.Visibility == "shared" {
		visColor = color.FgCyan
	} else if details.Visibility == "public" {
		visColor = color.FgGreen
	}
	table.Append([]string{"Visibility", color.Style{visColor, color.OpBold}.Render(details.Visibility)})
	table.Append([]string{"Size", formatBytes(details.Size)})

	if details.VirtualSize > 0 {
		table.Append([]string{"Virtual Size", formatBytes(details.VirtualSize)})
	}

	table.Append([]string{"Min Disk", fmt.Sprintf("%d GB", details.MinDisk)})
	table.Append([]string{"Min RAM", fmt.Sprintf("%d MB", details.MinRAM)})
	table.Append([]string{"Disk Format", stringOrNA(details.DiskFormat)})
	table.Append([]string{"Container Format", stringOrNA(details.ContainerFormat)})
	table.Append([]string{"Created", stringOrNA(details.CreatedAt)})
	table.Append([]string{"Updated", stringOrNA(details.UpdatedAt)})
	table.Append([]string{"Protected", colorStyleBool(details.Protected)})
	table.Append([]string{"Hidden", colorStyleBool(details.OsHidden)})

	// Hash Information
	if details.Checksum != "" {
		table.Append([]string{"Checksum", details.Checksum})
	}
	if details.OsHashAlgo != "" {
		table.Append([]string{"Hash Algorithm", details.OsHashAlgo})
		table.Append([]string{"Hash Value", details.OsHashValue})
	}

	// Image Properties
	if details.HwQemuGuestAgent != "" {
		table.Append([]string{"QEMU Guest Agent", details.HwQemuGuestAgent})
	}
	if details.OsType != "" {
		table.Append([]string{"OS Type", details.OsType})
	}
	if details.OsDistro != "" {
		table.Append([]string{"OS Distribution", details.OsDistro})
	}
	if details.ImageValidated != "" {
		table.Append([]string{"Validated", details.ImageValidated})
	}

	// Owner and URLs
	table.Append([]string{"Owner", stringOrNA(details.Owner)})
	if details.DirectURL != "" {
		table.Append([]string{"Direct URL", details.DirectURL})
	}
	if details.File != "" {
		table.Append([]string{"File Path", details.File})
	}
	if details.Self != "" {
		table.Append([]string{"Self Link", details.Self})
	}

	table.Render()

	// Tags Table
	if len(details.Tags) > 0 {
		fmt.Println("\nTags:")
		tagTable := tablewriter.NewWriter(os.Stdout)
		tagTable.SetHeader([]string{"Tag"})
		applyTableStyle(tagTable)

		for _, tag := range details.Tags {
			tagTable.Append([]string{tag})
		}
		tagTable.Render()
	}
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

// VolumeDetails represents formatted volume details for display
type VolumeDetails struct {
	ID                 string
	Name               string
	Status             string
	Size               int
	VolumeType         string
	Bootable           string
	Multiattach        bool
	Encrypted          bool
	AvailabilityZone   string
	CreatedAt          string
	UpdatedAt          string
	Description        string
	ReplicationStatus  string
	SnapshotID         string
	SourceVolID        string
	GroupID            string
	ConsistencyGroupID string
	ConsumesQuota      bool
	Attachments        []VolumeAttachment
	Metadata           map[string]string
}

type VolumeAttachment struct {
	ServerID     string
	ServerName   string // Resolved from ID if possible
	Device       string
	AttachedAt   string
	AttachmentID string
}

func PrintVolumeDetailsTable(details VolumeDetails) {
	fmt.Println("\nVolume Info:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Field", "Value"})
	applyTableStyle(table)

	table.Append([]string{"ID", details.ID})
	table.Append([]string{"Name", color.Style{color.FgGreen}.Render(details.Name)})
	table.Append([]string{"Status", colorStyleVolAvailability(details.Status)})
	table.Append([]string{"Size", fmt.Sprintf("%d GB", details.Size)})
	table.Append([]string{"Type", stringOrNA(details.VolumeType)})
	table.Append([]string{"Bootable", details.Bootable})
	table.Append([]string{"Multiattach", colorStyleBool(details.Multiattach)})
	table.Append([]string{"Encrypted", colorStyleBool(details.Encrypted)})
	if details.AvailabilityZone != "" {
		table.Append([]string{"Availability Zone", details.AvailabilityZone})
	}
	table.Append([]string{"Created", details.CreatedAt})
	if details.UpdatedAt != "" {
		table.Append([]string{"Updated", details.UpdatedAt})
	}
	if details.Description != "" {
		table.Append([]string{"Description", details.Description})
	}
	if details.ReplicationStatus != "" {
		table.Append([]string{"Replication Status", details.ReplicationStatus})
	}
	if details.SnapshotID != "" {
		table.Append([]string{"Created from Snapshot", details.SnapshotID})
	}
	if details.SourceVolID != "" {
		table.Append([]string{"Created from Volume", details.SourceVolID})
	}
	if details.GroupID != "" {
		table.Append([]string{"Group ID", details.GroupID})
	}
	if details.ConsistencyGroupID != "" {
		table.Append([]string{"Consistency Group", details.ConsistencyGroupID})
	}
	table.Append([]string{"Consumes Quota", colorStyleBool(details.ConsumesQuota)})

	table.Render()

	// Show attachments if any exist
	if len(details.Attachments) > 0 {
		fmt.Println("\nAttachments:")
		attachTable := tablewriter.NewWriter(os.Stdout)
		attachTable.SetHeader([]string{"Server", "Device", "Attached At", "Attachment ID"})
		applyTableStyle(attachTable)

		for _, att := range details.Attachments {
			serverDisplay := att.ServerID
			if att.ServerName != "" {
				serverDisplay = fmt.Sprintf("%s (%s)", att.ServerName, att.ServerID)
			}

			attachTable.Append([]string{
				serverDisplay,
				att.Device,
				att.AttachedAt,
				att.AttachmentID,
			})
		}
		attachTable.Render()
	}

	// Show metadata if any exists
	if len(details.Metadata) > 0 {
		fmt.Println("\nMetadata:")
		metaTable := tablewriter.NewWriter(os.Stdout)
		metaTable.SetHeader([]string{"Key", "Value"})
		applyTableStyle(metaTable)

		for k, v := range details.Metadata {
			metaTable.Append([]string{k, v})
		}
		metaTable.Render()
	}
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
	table.SetHeader([]string{"NAME", "ID", "STATUS", "MANAGED", "CIDRs"})

	applyTableStyle(table)

	for _, n := range nets {
		table.Append([]string{
			n.Name,
			n.ID,
			colorStyleStatus(n.Status),
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
// PORTS
// -------------------------------------------------------------------

type Port struct {
	ID          string
	MACAddress  string
	NetworkID   string
	DeviceID    string
	DeviceOwner string
	Status      string
	FixedIPs    string
}

type PortDetails struct {
	ID              string
	MACAddress      string
	NetworkID       string
	DeviceID        string
	DeviceOwner     string
	Status          string
	FixedIPs        []string
	SecurityGroups  []string
	AdminStateUp    bool
	BindingHostID   string
	BindingVnicType string
	DNSDomain       string
	DNSName         string
	CreatedAt       string
	UpdatedAt       string
}

func PrintPortDetailsTable(details PortDetails) {
	ips := details.FixedIPs
	if len(ips) == 0 {
		ips = []string{"Unmanaged"}
	} else {
		for i, ip := range ips {
			ips[i] = ip
		}
	}
	fmt.Println("\nPort Info:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Field", "Value"})
	applyTableStyle(table)

	table.Append([]string{"ID", details.ID})
	table.Append([]string{"MAC Address", color.Style{color.FgGreen}.Render(details.MACAddress)})
	for _, ip := range ips {
		if ip == "Unmanaged" {
			table.Append([]string{"IP Address", color.Style{color.FgRed}.Render(ip)})
		} else {
			table.Append([]string{"IP Address", ip})
		}
	}
	table.Append([]string{"Network ID", details.NetworkID})
	table.Append([]string{"Attached VM", stringOrNA(details.DeviceID)})
	table.Append([]string{"Device Owner", stringOrNA(details.DeviceOwner)})
	table.Append([]string{"Status", colorStyleStatus(details.Status)})
	table.Append([]string{"Admin State", colorStyleBool(details.AdminStateUp)})
	table.Append([]string{"Binding Host", stringOrNA(details.BindingHostID)})
	table.Append([]string{"VNIC Type", details.BindingVnicType})
	table.Append([]string{"DNS Domain", stringOrNA(details.DNSDomain)})
	table.Append([]string{"DNS Name", stringOrNA(details.DNSName)})
	table.Append([]string{"Created", details.CreatedAt})
	table.Append([]string{"Updated", stringOrNA(details.UpdatedAt)})
	table.Render()

	if len(details.FixedIPs) > 0 {
		fmt.Println("\nFixed IPs:")
		ipTable := tablewriter.NewWriter(os.Stdout)
		ipTable.SetHeader([]string{"IP Address"})
		applyTableStyle(ipTable)

		for _, ip := range details.FixedIPs {
			ipTable.Append([]string{ip})
		}
		ipTable.Render()
	}

	if len(details.SecurityGroups) > 0 {
		fmt.Println("\nSecurity Groups:")
		sgTable := tablewriter.NewWriter(os.Stdout)
		sgTable.SetHeader([]string{"ID"})
		applyTableStyle(sgTable)

		for _, sg := range details.SecurityGroups {
			sgTable.Append([]string{sg})
		}
		sgTable.Render()
	}
}

func PrintPortsTable(ports []Port) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "MAC", "NETWORK", "DEVICE", "STATUS", "IPS"})

	applyTableStyle(table)

	for _, p := range ports {
		table.Append([]string{
			p.ID,
			color.Style{color.FgGreen}.Render(p.MACAddress),
			p.NetworkID,
			stringOrNA(p.DeviceID),
			colorStyleStatus(p.Status),
			stringOrNA(p.FixedIPs),
		})
	}
	table.Render()
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
