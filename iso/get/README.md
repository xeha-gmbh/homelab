# ISO Get Command

This commands helps download the most used ISO images from Internet.

## TLDR;

```bash
$ homelab iso get \
    --flavor ubuntu/bionic64.live \
    --target-dir /tmp \
    --reuse
```

The above command instructs to download the Ubuntu 18.04 LTS Live CD and save it to the `/tmp` directory. But if it
is already saved there, then skip the download.

## Parameters
|Parameter|Required|Default|Value|
|---|---|---|---|
|`--flavor`|yes|--|{`ubuntu/bionic64.live`,`ubuntu/bionic64`,`ubuntu/xenial64`}|
|`--target-dir`|no|`/tmp`|directory to save to|
|`--reuse`|no|`false`|whether to check for existing downloads|