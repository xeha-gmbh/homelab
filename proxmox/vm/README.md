# Proxmox VM Command

This command package manages VM on Proxmox KVM. It deals with endpoints at `"/api2/json/nodes/$node/qemu"`.

## VM Creation

This command creates a QEMU vm. Proxmox offers a wide range of configuration parameters, which is impractical to include
all of them as command parameters. Hence, this command adopts the concept of an _archetype_. An archetype is an opinionated
set of parameters with limited option for configuration. It is similar to the Maven archetype. As of now, only one archetype
is offered.

#### Basic Archetype

The basic archetype has the following features:
* Creates Linux VM with 2.6/3.X or later Kernal.
* Creates VM with one CPU socket and configurable number of cores.
* Creates one hard drive with SCSI format with VirtIO PCI driver.
* Creates bridged network with VirtIO driver on the configurable interface.
* Installs OS using ISO image mounted as CD-ROM.
* NUMA support is turned on.

**Example:**

```bash
$ homelab proxmox vm create basic \
    --node=pve \
    --id=103 \
    --name=test-vm \
    --iso-storage=local \
    --iso-image=ubuntu.iso \
    --drive-storage=local-data \
    --drive-size=64 \
    --core=2 \
    --memory=2048 \
    --iface=vmbr0 \
    --start
```

The above snippet uses ubuntu.iso from the local storage device as the installation media, creates a single drive of 
64G in the local-data storage device as the system drive, uses the vmbr0 interface as the default network interface and
creates a VM of 2 CPU cores and 2048M memory. In the end, it starts the VM upon creation.

**Parameters:**

|Flag|Required|Default|Content|
|---|---|---|---|
|`--node`|no|`pve`|The node in the Proxmox cluster to create vm|
|`--id`|yes|--|Id of the new vm, must be unique|
|`--name`|yes|--|Name of the new vm|
|`--iso-storage`|no|`local`|Storage device of the installation media|
|`--iso-image`|yes|--|Image name in the image storage device|
|`--drive-storage`|yes|--|Storage device of the system drive|
|`--drive-size`|no|`64`|Size of the system drive in GB|
|`--core`|no|`2`|Number of virtual CPU cores|
|`--memory`|no|`2048`|Size of virtual memory in MB|
|`--iface`|no|`vmbr0`|Default network interface for the vm|
|`--start`|no|`false`|Whether to start VM on successful creation|

_As of now, all communications to Proxmox endpoints skip TLS verification._