#!/usr/bin/env bash

set -euo pipefail

# update repos & pkgs
sudo apt update -y
sudo apt upgrade -y

# grab needed pkgs
sudo apt install -y build-essential make git ssh wget curl linux-headers-5.4.0-26-generic vim

# tell grub to boot the right entry
echo 'GRUB_DEFAULT="1>GNU/Linux, with Linux 5.4.0-26-generic"' | sudo tee -a /etc/default/grub > /dev/null
sudo update-grub
