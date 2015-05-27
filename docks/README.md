docks
=====

`docks` is a set of `Docker` images for `LSST` `FCS` camera subsystem.

## lsst-ccs-base
`lsst-ccs-base` is a mini `ArchLinux` base image.
This has been extracted from
[nfty/arch-mini](https://github.com/nfnty/dockerfiles/tree/master/images/arch-mini)

## lsst-ccs-fcs
Development of `FCS` requires `Java-8`, `Maven` and (optionally) `NetBeans`.
The `lsst-ccs-fcs-base` image tries to provide all that (sans `NetBeans`) based
on an `ArchLinux` image.

Typical session would look like:

```sh
> docker run -it -v `pwd`/org-lsst-ccs-fcs:/opt/lsst lsst-ccs-fcs-base
lsst> cd /opt/lsst
lsst> mvn clean && mvn install
```

ie: one would mount the local development tree under `/opt/lsst`.

