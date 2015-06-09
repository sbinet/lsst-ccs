fcs-mgr
=======

`fcs-mgr` eases the development and use of the `FCS` subsystem development
environment.

## Installation

```sh
sh> go get github.com/sbinet/lsst-ccs/...
```

## Usage

### `fcs-mgr init`
`fcs-mgr init` creates a new workspace for `CCS` and the `FCS` subsystem.

```sh
sh> fcs-mgr init
2015/06/09 17:01:10 init-dir="${PWD}"
2015/06/09 17:01:10 updating repo [org-lsst-ccs-subsystem-fcs]...
2015/06/09 17:01:10 updating repo [org-lsst-ccs-localdb]...
```

### `fcs-mgr update`
`fcs-mgr update` updates the local `git-svn` repositories of
`org-lsst-ccs-localdb` and `org-lsst-ccs-subsystem-fcs`.

```sh
sh> fcs-mgr update
2015/06/09 17:36:54 updating repo [org-lsst-ccs-subsystem-fcs]...
2015/06/09 17:36:54 updating repo [org-lsst-ccs-localdb]...
```

### `fcs-mgr build`
`fcs-mgr build` builds and installs the `org-lsst-ccs-localdb` and
`org-lsst-ccs-subsystem-fcs` repositories.

```sh
sh> fcs-mgr build
2015/06/09 17:39:31 building repo [org-lsst-ccs-subsystem-fcs]...
2015/06/09 17:39:31 building repo [org-lsst-ccs-localdb]...
2015/06/09 17:39:56 building repo [org-lsst-ccs-localdb]... [ok] (time=24.384273959s)
2015/06/09 17:40:28 building repo [org-lsst-ccs-subsystem-fcs]... [ok] (time=56.480276874s)
```

### `fcs-mgr dist`
`fcs-mgr dist` creates a `CCS` distribution from the 2
`org-lsst-ccs-{localdb,subsystem-fcs}` repositories.

```sh
sh> fcs-mgr dist
2015/06/09 17:41:21 creating distribution for repo [org-lsst-ccs-subsystem-fcs]...
2015/06/09 17:41:21 creating distribution for repo [org-lsst-ccs-localdb]...

sh> ll DISTRIB/
total 0
drwxr-xr-x 1 binet binet 22 Jun  9 17:41 org-lsst-ccs-localdb-config-1.4.1-SNAPSHOT
drwxr-xr-x 1 binet binet 22 Jun  9 17:41 org-lsst-ccs-localdb-jar-1.4.1-SNAPSHOT
drwxr-xr-x 1 binet binet 22 Jun  9 17:41 org-lsst-ccs-localdb-main-1.4.1-SNAPSHOT
drwxr-xr-x 1 binet binet 22 Jun  9 17:41 org-lsst-ccs-localdb-war-1.4.1-SNAPSHOT
drwxr-xr-x 1 binet binet 22 Jun  9 17:41 org-lsst-ccs-subsystem-fcs-buses-1.6.2-SNAPSHOT
drwxr-xr-x 1 binet binet 22 Jun  9 17:41 org-lsst-ccs-subsystem-fcs-gui-1.6.2-SNAPSHOT
drwxr-xr-x 1 binet binet 22 Jun  9 17:41 org-lsst-ccs-subsystem-fcs-main-1.6.2-SNAPSHOT
```

### `fcs-mgr run`
`fcs-mgr run` runs a command inside a `docker` container with the `CCS`
distribution mounted under `/opt/lsst`, via the `fcs-run` command from
`github.com/sbinet/lsst-ccs/fcs-mgr/fcs-run`.

