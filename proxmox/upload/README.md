# Proxmox Upload Command

This command calls out to the specified Proxmox API server endpoint at `/api2/json/nodes/$node/storage/$storage/upload` 
and uploads the image. If user didn't specify the storage device, it will call out to `/api2/json/nodes/%s/storage` and
obtain the first storage device that can accept the specified format.

This command requires authentication. Unless a ticket cache is already saved, use [Proxmox Login Command](https://github.com/imulab/homelab/tree/master/proxmox/login) first.

Attempts were made to utilize native Go API (namely the `multipart` package) to do the upload directly, however, it didn't
seem to play well with the Proxmox Upload API. Hence, the currently implementation uses `curl` command in insecure mode.

_As of now, all communications to Proxmox endpoints skip TLS verification._

## TLDR;

```bash
$ homelab proxmox upload \
    --node=pve \
    --storage=local \
    --file=/my/downloads/ubuntu.iso \
    --format=iso
```

## Parameters

|Flag|Required|Default|Content|
|---|---|---|---|
|`--node`|no|`pve`|The node in the Proxmox cluster to upload to|
|`--storage`|yes|--|The storage device to save to|
|`--file`|yes|--|The absolute path to the file to upload|
|`--format`|no|`iso`|The format of the file to upload|

