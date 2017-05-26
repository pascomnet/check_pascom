# check_pascom

check_pascom is a Nagios (Icinga) check, written in go, to check the pascom
cloudstack. It connects via ssh to the target system, performs automatically
a number of helthchecks and returns an overall state plus multiline details.

## Installation

In order to install check_pascom at your Nagios host you can either compile
the source code or use the precompiled binary for linux out of the `bin/linux_amd64`
directory of this repository and copy it to your Nagios plugin directory.

## Prepare ssh keys

check_pascom connects to the pascom cloudstack using ssh keys. Three steps are
nescessary:

### Nagios host

Login via ssh to your Nagios host. Become the user which performs the
checks (e.g. nagios):

`su nagios -s /bin/bash`

Check if this user has already a public ssh key:

`cat ~/.ssh/id_rsa.pub`

If so, copy it to your clipboard. If not, create one:

`ssh-keygen`

Add the keys (if not already done) to the authentication agent:

`ssh-add`

### pascom cloudstack

Login as admin on the cloudstack via ssh and become root.

Copy the ssh public key (from the previous step) to the file `/root/.ssh/authorized_keys`
on the pascom cloudstack.

### Nagios host (again)

Now you should be able to login via ssh to the cloudstack as root without a password.
Test it with:

`ssh root@my.cloudstack.hostname`

If this works close the shell again and execute check_pascom on the Nagios host:

`check_pascom -h my.cloudstack.hostname`

You should get something like:

```bash
OK: All checks are ok
Memory is OK
Swap is OK
Ramdisk / is OK
Disk /SYSTEM is OK
Load is OK
Container pg is OK
Container controller is OK
Container mem controller is OK
Container mem pg is OK
Container mem proxy is OK
...
```

## Usage

As check_pascom directly for help:

`check_pascom -h`
