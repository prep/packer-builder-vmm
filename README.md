packer-builder-vmm
==================
This is an OpenBSD VMM plugin for [packer](https://www.packer.io/).

## Installation
Execute the following command to install this plugin:

```
go get -u github.com/prep/packer-builder-vmm/cmd/packer-builder-vmm
```

Then move the newly built `~/go/bin/packer-builder-vmm` binary to either `~/.packer.d/plugins` or in place it in the directory you're going to run `packer build` from.

## Notes
* It is assumed that you run `packer` as an unprivileged user and that `doas` can be used to start both `vmctl` and `tee`. The `tee` command is used to get write-only access to the pseudo TTY to send the build process commands.
* For now, instances will only be started with a local network interface. The ability to attach a network interface to a virtual switch will be added later.
* The HTTP server, [provisioning](https://www.packer.io/docs/provisioners/index.html) and [post-processors](https://www.packer.io/docs/post-processors/index.html) are current unsupported, so everything has to be packed into the `boot_command`.

## Example
This project has a couple of example packer configurations in the [examples](examples) directory, like [an Alpine Linux 3.8.0 installation](examples/alpine.json) and [an OpenBSD 6.4-beta installation](examples/openbsd.json). Note that you need to update the SHA256 hash of the OpenBSD installation yourself, as it references a snapshot installation file that gets updated regularly.

Before any build is started, poll _vmctl_ in a separate terminal for the specific VM that _packer build_ is going to spin up so that we can follow what's happening. In this example, we're going to build the OpenBSD configuration from the [examples](examples) directory whose `vm_name` is `openbsd-example`, so lets start with that:

```
while true; do vmctl console openbsd-example; sleep 3; done
```

Now that we have something in place to monitor the build, start the _packer_ build operation:

```
packer build examples/openbsd.json
```

This should leave you with an `output-vmm/openbsd-example.raw` bootable disk image of a clean OpenBSD installation.
