# Proxmox Login Command

This command calls out to the specified Proxmox API server endpoint at `/api2/json/access/ticket` and obtains
a ticket and csrf token. The acquired credentials, along with the API server url will be saved in a `.proxmox` file
in the user's home directory.Any subsequent API calls to Proxmox server will parse this file and use its content as the credential.

By default, this command performs no-op when a `.proxmox` cache is already there, unless the user specifies `--force` option.

_As of now, all communications to Proxmox endpoints skip TLS verification._

## TLDR;

```bash
$ homelab proxmox login \
    --username=root \
    --password=s3cret \
    --realm=pam \
    --api-server=https://proxmox:8006 \
    --force  
```

## Parameters

|Flag|Required|Default|Content|
|---|---|---|---|
|`--username`|no|`root`|Login username|
|`--password`|yes|--|Login password|
|`--realm`|no|`pam`|Proxmox realm to log into|
|`--api-server`|yes|--|Proxmox server url. e.g., `https://192.168.100.111:8006`|
|`--force`|no|`false`|Whether to ignore any ticket cache (see below)|

