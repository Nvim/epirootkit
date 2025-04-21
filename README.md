# Epirootkit

some sneaky rootkit with a cool program to manage it remotely

## Setup

### Victim machine

- grab ubuntu 24.04 LTS
- install it with virt-manager+qemu/kvm setup that i'll document one day
- edit grub config to set a positive timeout
- shutdown
- goto vm settings & create a filesystem hardware device to share the git repo
- start vm
- apt update/upgrade, get vim
- mount the shared filesystem `sudo mount -t virtiofs TAG ./some/local/path`
- cd into the rootkit dir, run make and insmod it
