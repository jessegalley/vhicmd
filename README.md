# vhicmd

A command-line utility for interacting with VHI (Virtuozzo Hybrid Infrastructure) APIs. This tool provides a streamlined interface for managing virtual machines, volumes, networks, and other OpenStack resources.

## Features

- Authentication and project management
- Resource listing (VMs, volumes, networks, flavors, images, ports)
- Detailed resource information
- VM creation, deletion, and management
- Volume creation and management
- Network and port management
- Image management and sharing
- VM migration from VMware environments

## Build & Installation

```bash
make
cd bin && mv vhicmd /usr/local/bin
```

## Configuration

vhicmd uses a configuration file located at `~/.vhirc` (YAML format).

Example configuration:
```yaml
host: panel-vhi1.yourhost.com
domain: yourdomain
project: yourproject
username: youruser
password: yourpassword
networks: uuid1,uuid2
flavor_id: flavor-uuid
image_id: image-uuid
```

Configuration Options:
- `host`: VHI host URL
- `domain`: VHI domain
- `project`: VHI project
- `username`: VHI username
- `password`: VHI password (optional)
- `networks`: Default networks for VM creation
- `flavor_id`: Default flavor for VM creation
- `image_id`: Default image for VM creation

Manage configuration:
```bash
vhicmd config list                # Show current config
vhicmd config get <key>           # Get specific value
vhicmd config set <key> <value>   # Set specific value
```

## Authentication

```bash
# Using config file values
vhicmd auth

# Interactive authentication with prompts
vhicmd auth <domain> <project> --host <vhi host>

# Command line authentication
vhicmd auth <domain> <project> -u username -p password

# Switch between projects
vhicmd switch-project [project]   # Interactive if no project specified
```

## Resource Management Commands

### Basic Operations

View service catalog:
```bash
vhicmd catalog [--interface public|admin|internal]
```

List resources:
```bash
vhicmd list domains              # Admin only
vhicmd list projects
vhicmd list vms [--name filter]
vhicmd list volumes
vhicmd list networks [--name filter]
vhicmd list ports
vhicmd list flavors
vhicmd list images [--name filter] [--visibility public|private|shared]
vhicmd list image-members <image>
```

Get detailed information:
```bash
vhicmd details vm <vm-id>
vhicmd details volume <volume-id>
vhicmd details image <image-id>
vhicmd details port <port-id>
```

### Virtual Machine Management

Create VM:
```bash
vhicmd create vm --name <name> \
  --flavor <flavor-id> \
  --image <image-id> \
  --networks <network-ids> \
  --ips <ip-addresses> \
  --size <size-GB> \
  --user-data <cloud-init-file> \
  --macaddr <mac-addresses>
```

Delete VM:
```bash
vhicmd delete vm <vm-id>
```

Update VM:
```bash
vhicmd update vm name <vm-id> <new-name>
vhicmd update vm metadata <vm-id> <key> <value>
```

Manage VM flavor:
```bash
vhicmd update vm flavor start <vm-id> <new-flavor>     # Start flavor change
vhicmd update vm flavor confirm <vm-id>                # Confirm change
vhicmd update vm flavor revert <vm-id>                 # Revert change
```

Reboot VM:
```bash
vhicmd reboot soft <vm-id>
vhicmd reboot hard <vm-id>
```

Network interfaces:
```bash
vhicmd update vm attach-port <vm-id> <port-id>
vhicmd update vm detach-port <vm-id> <port-id>
```

### Storage Management

Create volume:
```bash
vhicmd create volume --name <name> \
  --size <size-GB> \
  --type <volume-type> \
  --description <description>
```

Delete volume:
```bash
vhicmd delete volume <volume-id>
```

Manage bootable flag:
```bash
vhicmd bootable <volume-id> true|false
```

### Image Management

Create image:
```bash
vhicmd create image --file <path> \
  --name <name> \
  [--format qcow2|raw|vmdk|iso]
```

Delete image:
```bash
vhicmd delete image <image-id>
```

Image sharing:
```bash
vhicmd add image-member <image> <project-id>      # Grant access
vhicmd delete image-member <image> <project-id>   # Revoke access
vhicmd update image member <image> <member-id> <status>  # Update member status
vhicmd update image visibility <image> <visibility>       # Update visibility
```

### Network Management

Create port:
```bash
vhicmd create port --network <network-id> [--mac <mac-address>]
```

Delete port:
```bash
vhicmd delete port <port-id>
```

### VM Migration

Migrate from VMware:
```bash
vhicmd migrate vm \
  --name <name> \
  --vmdk <path> \
  --flavor <flavor> \
  --networks <networks> \
  --mac <mac-addresses> \
  --size <size-GB> \
  [--shutdown] \
  [--disk-bus sata|scsi|virtio]
```

Find VMDK files:
```bash
vhicmd migrate find <pattern> [--single]
```

## Global Flags

- `-H, --host`: Override the VHI host
- `--config`: Specify alternate config file
- `--debug`: Enable debug mode
- `--json`: Output in JSON format (available for list/details commands)

## Notes

- Most commands accept either resource IDs or names
- `vhicmd` supports networks with IPAM disabled which allows manually specifying MAC addresses as a normal user when creating a port
