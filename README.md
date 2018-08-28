packer-builder-vmm
==================
This is an OpenBSD VMM plugin for [packer](https://www.packer.io/).

1. Installation
---------------
Execute the following command to install this plugin:

```
go -u github.com/prep/packer-builder-vmm/cmd/packer-builder-vmm
```

Then move the newly built `~/go/bin/packer-builder-vmm` binary to either `~/.packer.d/plugins` or in place it in the directory you're going to run `packer build` from.

2. Notes
--------
* It is assumed that you run `packer` as an unprivileged user and that `doas` can be used to start both `vmctl` and `tee`. The `tee` command is used to get write-only access to the pseudo TTY to send the build process commands.
* For now, instances will only be started with a local network interface. The ability to attach a network interface to a virtual switch will be added later.
* Due to the above constraint and the inability to determine both the host and client IP address, the HTTP server functionality doesn't work.
* Because the client IP address cannot be determined, provisioning isn't implemented yet so everything has to be packed into the `boot_command`.

3. Example
----------
This project has an example packer configuration in [examples/openbsd.json](examples/openbsd.json) that references some version of OpenBSD's 6.4-beta _install64.fs_. Note that you'll probably get an error on the SHA256 hash, which you need to change yourself.

Before any build is started, poll _vmctl_ in a separate terminal for the specific VM that the build is going to start:

```
while true; do vmctl console openbsd-example; sleep 3; done
```

Then start the actual _packer_ build:

```
packer build examples/openbsd.json
```

This should leave you with an `output-vmm/openbsd-example.raw` bootable disk image of a clean OpenBSD installation.