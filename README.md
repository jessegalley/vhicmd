# vhicmd

A command-line utility for interacting with VHI (Virtual Hosting Infrastructure) APIs. This tool provides a streamlined interface for managing virtual machines, volumes, networks.

## Features

- Resource listing (VMs, volumes, networks, flavors, images)
- Detailed resource information
- VM creation
- Volume creation
- Reboot VM (hard, soft)
- TODO: Delete & update resources

## Build & Installation

```bash
make
cd bin && mv vhicmd /usr/local/bin
```

## Configuration

vhicmd uses a configuration file located at `~/.vhirc` (YAML format by default).

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

Explanation:
- `host`: VHI host URL
- `domain`: VHI domain
- `project`: VHI project
- `username`: VHI username
- `password`: VHI password
- `networks`: Default networks to use for VM creation
- `flavor_id`: Default flavor to use for VM creation
- `image_id`: Default image to use for VM creation

Configuration can be managed using:
```bash
vhicmd config list                # Show current config
vhicmd config get <key>          # Get specific value
vhicmd config set <key> <value>  # Set specific value
```

Configuration can also be set via environment variables by prefixing with `VHI_`:
```bash
export VHI_HOST=panel-vhi1.yourhost.com
export VHI_USERNAME=user
export VHI_PASSWORD=pass
```

## Authentication

```bash
# Using config file values
vhicmd auth

# Prompt for username and password
# * Note: domain and project are required, this will also prompt to save the values to `~/.vhirc` for future use.
vhicmd auth <domain> <project> --host <vhi host>

# Override with command line
vhicmd auth <domain> <project> -u username -p password
```

Tokens are saved to `~/.vhicmd.token` and can be refreshed with `vhicmd auth`.

## Basic Commands

After authentication, you can start using the tool.
Note that UUIDs are used to identify resources (VMs, volumes, networks, flavors, images).

Show catalog:
```bash
vhicmd catalog
```

Reboot VM:
```bash
vhicmd reboot <vm-id> <soft/hard>
```

List resources:
```bash
vhicmd list vms
vhicmd list volumes
vhicmd list networks
vhicmd list flavors
vhicmd list images
```

Get detailed information:
```bash
vhicmd details vm <vm-id>
```

Create resources:
```bash
# Create VM
# * Note: networks is a comma-separated list of UUIDs
vhicmd create vm --name test-vm --flavor <flavor-id> --image <image-id> --networks <network-ids> --ips <ips-csv> --size <size-in-GB>

# Create VM with netboot enabled, this will create a blank volume instead of using an image (deprecated)
vhicmd create vm --name test-vm --flavor <flavor-id> --networks <network-ids> --ips <ip-csv> --size <size-in-GB> --netboot true

# Create VM with config values from `~/.vhirc`
vhicmd create vm --name test-vm --size <size-in-GB> --ips <ips-csv>

# Create Volume
vhicmd create volume --name test-vol --size 10
```

Make volume bootable:
```bash
vhicmd bootable <volume-id> true/false
```

After creating a VM, `vhicmd` will print the VM ID, IP/MAC addresses, and other relevant information.

Manage netboot:
```bash
vhicmd netboot set <vm-id> true/false
```

## Global Flags

- `-H, --host`: Override the VHI host
- `--json`: Output in JSON format instead of tables
