#!/usr/bin/env bash

set -euo pipefail

echo "Updating repos..."
sudo apt update -y

echo "Upgrading packages..."
sudo apt upgrade -y

echo "Installing essential tools..."
sudo apt install -y wget make curl build-essential

echo "Downloading Go..."
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz

echo "Installing Go..."
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.bashrc
source ~/.bashrc
rm go1.24.4.linux-amd64.tar.gz

echo "Installing a Nerd Font..."
curl -fsSL https://raw.githubusercontent.com/getnf/getnf/main/install.sh | bash
~/.local/bin/getnf -i JetBrainsMono

echo "Go version: $(go version)"

echo "Setup complete"

