docks
=====

`docks` is a set of `Docker` images for the `LSST` `FCS` camera subsystem.

## lsst-ccs/base
`lsst-ccs/base` is a mini `ArchLinux` base image.
This has been extracted from
[nfty/arch-mini](https://github.com/nfnty/dockerfiles/tree/master/images/arch-mini)

## lsst-ccs/fcs
Development of `FCS` requires `Java-8`, `Maven` and (optionally) `NetBeans`.
The `lsst-ccs/fcs` image tries to provide all that (sans `NetBeans`) based
on an `ArchLinux` image.

Typical session would look like:

```sh
> docker run -it -v `pwd`/org-lsst-ccs-fcs:/opt/lsst lsst-ccs/fcs
lsst> cd /opt/lsst
lsst> mvn clean && mvn install
```

ie: one would mount the local development tree under `/opt/lsst`.

One may also pass an additional `-v $HOME/.ssh:/root/.ssh` to make the proper
ssh-keys available.

### fcs-dev
A helper command is provided to ease the configuration and launching of the `lsst-css/fcs` image.

```sh
> go build fcs-dev.go
> ./fcs-dev --lsst=/path/to/lsst/sw
lsst> mvn clean && mvn install
lsst> fcs-mgr lpc
```
