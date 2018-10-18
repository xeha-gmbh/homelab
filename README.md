# Home Lab
Bootstrap my Home Lab.

> **My Home Lab**: Single master Kubernetes cluster running various systems with GlusterFS as the persistence provider. Optionally with a Concourse CI VM equipped with Vault as the secret keeper. All systems running on Proxmox KVM.

This project aims at bootstraping my Home Lab environment with as little effort as possible. The idea is to use a single
YAML configuration to provide all the data needed to:
* **1)** make auto-install server images
* **2)** login to Proxmox 
* **3)** upload the auto-install images
* **4)** create VM with the necessary parameters and the auto-install image mounted
* **5)** start the VM to kick start installation

The rest of the provisioning work can be handed over to Ansible. Although

## TLDR;

```bash
$ sudo -s
$ homelab bootstrap -c config.yaml
```

## Commands

The `bootstrap` command uses several sub-commands to achieve the overall effect:
* [homelab iso auto](https://github.com/imulab/homelab/tree/master/iso/auto)
* [homelab proxmox login](https://github.com/imulab/homelab/tree/master/proxmox/login)
* [homelab proxmox upload](https://github.com/imulab/homelab/tree/master/proxmox/upload)
* [homelab proxmox vm create](https://github.com/imulab/homelab/tree/master/proxmox/vm)

## Develop

Development requires Go 1.11 environment. 

The [homelab iso auto](https://github.com/imulab/homelab/tree/master/iso/auto)
command needs to be executed in a ubuntu environment. In case of debugging, `vagrant up` to setup the ubuntu environment.

The [homelab proxmox](https://github.com/imulab/homelab/tree/master/proxmox) series commands obviously need a running Proxmox cluster.