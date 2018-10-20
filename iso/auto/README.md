# ISO Auto Command

This command helps create an auto installation media out of the original media. As of now, it supports
Ubuntu 18.04 LTS server (using old installer) and Ubuntu 16.04 LTS server.

## Flavor and Provider

As this command may potentially support other operating systems in the future. It leaves the option open using the 
concept of _flavor_. A flavor corresponds to a specific OS. As before, the two supported flavor is `ubuntu/bionic64` and
`ubuntu/xenial64`. A provider handles one or more flavor.

#### Bionic and Xenial

The provider that currently handles the `ubuntu/bionic64` and `ubuntu/xenial64`
uses an external script and seed file to repack the installation media. The process is largely inspired by the 
[ubuntu-unattended](https://github.com/netson/ubuntu-unattended) project.

The command accepts user input, uses those input to parse the [preseed](https://www.debian.org/releases/wheezy/example-preseed.txt) 
file and places it the workspace. The rest of the work is handed to the [script](https://github.com/imulab/homelab/blob/master/iso/auto/ubuntu/ubuntu-auto.sh). 
The script downloads the flavor image if necessary, mounts it and copies it to a new location where real changes are made. 
Then, the preseed file is copied in, updates the installation menu to accept the preseed file and repacks it as a new iso.

**If preseed file is present, the script can be used directly:**

```bash
$ sudo -s
$ chmod +x ./scripts/ubuntu-preseed.sh
$ ./scripts/ubuntu-preseed.sh \
    --seed=/tmp/imulab.seed \
    --flavor=bionic64 \
    --workspace=/tmp \
    --input=/tmp/ubuntu-18.04-lts.iso \
    --output=/tmp/ubuntu-autoinstall.iso \
    --bootable \
    --debug \
    --reuse
```

|Parameter|Shorthand|Required|Value|
|---|---|---|---|
|`--seed`|`-s`|yes|path to preseed file|
|`--flavor`|`-v`|no|`{bionic64,xenial64}`, default `bionic64`|
|`--workspace`|`-w`|yes|directory to place temp files|
|`--input`|`-i`|no|`ubuntu.iso`|
|`--output`|`-o`|no|`ubuntu-auto.iso`|
|`--bootable`|`-b`|no|`{y,n}`, default `n`|
|`--reuse`|`-r`|no|`{y,n}`, default `n`. Whether to reuse assets.|
|`--debug`|`-d`|no|`{y,n}`, default `n`. Whether to print debug messages.|

**The command is much easier to use if there's no preseed file:**

```bash
$ sudo -s
$ homelab iso auto \
    --flavor=ubuntu/bionic64 \
    --workspace=/tmp \
    --input-iso=/tmp/ubuntu-18.04-lts.iso \
    --output-iso=/tmp/ubuntu-autoinstall.iso \
    --timezone=America/Toronto \
    --username=imulab \
    --password=s3cret \
    --hostname=test \
    --domain=home.local \
    --ip-address=192.168.100.30 \
    --net-mask=255.255.255.0 \
    --gateway=192.168.100.1 \
    --name-servers=8.8.8.8 \
    --usb-boot \
    --reuse \
    --debug \
    --output-format=json
```

The above command downloads a new Ubuntu 18.04 LTS server image, or reuses one from workspace if it exists. It then configures
user account, network, and uses all disk as one volume. Makes the new image USB bootable and prints out debug messages in the
process. Finally, a new image is placed at `/tmp/ubuntu-auto.iso`.

Parameters are described as follows:

|Flag|Required|Default|Content|
|---|---|---|---|
|`--flavor`|no|`ubuntu/bionic64`|Flavor of the OS. {ubuntu/bionic64, ubuntu/xenial64} is supported.|
|`--workspace`|no|`/tmp`|Workspace of the command. All temp files and the resulting ISO is placed here.|
|`--input-iso`|yes|--|Path to the downloaded iso file|
|`--output-iso`|yes|--|Path to the converted iso file|
|`--timezone`|no|`America/Toronto`|Timezone of the system|
|`--username`|no|`imulab`|Username of the new user.|
|`--password`|yes|--|Password of the new user. Right now, the seed file does not crypt, and uses plain text.|
|`--hostname`|yes|--|Host name of the system|
|`--domain`|no|`home.local`|Domain of the system|
|`--ip-address`|no|--|Ip address, if configuring fixed network. If not specified, all network related flags are ignored, installation will use DHCP.|
|`--net-mask`|no|`255.255.255.0`|Net mask|
|`--gateway`|no|--|Gateway, required only if ip address is specified.|
|`--name-servers`|no|`8.8.8.8`|Comma delimited DNS servers, required only if ip address is specified.|
|`--usb-boot`|no|`false`|Whether to make the remastered ISO usb bootable.|
|`--reuse`|no|`false`|Whether to reuse existing original ISO in the workspace. If false, will download new one every time.|
|`--debug`|no|`false`|Whether to print debug messages.|
|`--output-format`|no|`text`|Format for the print out. {`text`,`json`}|

**Note** Ubuntu 18.04 LTS now uses [Subiquity](https://github.com/CanonicalLtd/subiquity) as the default 
[Live Server](http://releases.ubuntu.com/bionic/) installer. The original pressed file no longer works. Instead, Subiquity
uses a mechanism of `answers.yaml` to prepare answers to all installation questions. The file is supposed to be placed in
`squashfs-root/subiquity_config/`. The `squashfs-root` can be obtained by un-squashing `casper/filesystem.squashfs` using 
`unsquashfs`. When modification is done, re-squash it using `mksquashfs`. This process should work, however, the subiquity
project didn't provide examples as to how to configure static network. Hence, as of now, we are using a 
[CD image download of Ubuntu 18.04 LTS](http://cdimage.ubuntu.com/ubuntu/releases/18.04/release/ubuntu-18.04.1-server-amd64.iso) which still ships with the old installer. 