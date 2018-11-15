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
* It is assumed that you run `packer` as an unprivileged user and that `doas` can be used to start `vmctl`.
* For now, instances will only be started with a local network interface. The ability to attach a network interface to a virtual switch will be added later.
* The HTTP server, [provisioning](https://www.packer.io/docs/provisioners/index.html) and [post-processors](https://www.packer.io/docs/post-processors/index.html) are current unsupported, so everything has to be packed into the `boot_command`.
* The [Alpine example](examples/alpine.json) needs VM networking to be set up for local interfaces. The [vmctl manpage section about local interfaces](http://man.openbsd.org/vmctl#LOCAL_INTERFACES) has some examples on getting that working.

## Example
This project has a couple of example packer configurations in the [examples](examples) directory, like [an Alpine Linux 3.8.1 installation](examples/alpine.json) and [an OpenBSD 6.4 installation](examples/openbsd.json). Let's demonstrate what building the OpenBSD example looks like. First, issue the `packer build` command:

```
packer build examples/openbsd.json
```

The build process will create a logfile of what's happening behind the scenes, that you can follow by tailing it in another terminal. This is useful to detect timing errors (commands being issued too soon, etc) or configuration errors in your build script:

```
tail -f output-vmm/openbsd-example.log
```

When the build is done, packer should have left you a `output-vmm/openbsd-example.qcow2` bootable disk image of a clean OpenBSD installation, which you can immediately boot:

```
vmctl start openbsd-example -c -d output-vmm/openbsd-example.qcow2 -L
```
