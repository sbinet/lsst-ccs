lsst-ccs-cwrapper-xgo
=====================

`lsst-ccs-cwrapper-xgo` is a small `debian` based container where `gccgo-multilib` is installed to build 32bits `go` based programs for the pc104 small factor PC.

## Example

```sh
sh> docker run --rm -v `pwd`:/go lsst-ccx/cwrapper-xgo \
	go install -compiler=gccgo -x -v github.com/go-lsst/fcs-lpc-bench
[...]
cp $WORK/github.com/go-lsst/fcs-lpc-bench/_obj/exe/a.out /go/bin/linux_386/fcs-lpc-bench

sh> scp ./bin/linux_386/fcs-lpc-bench  root@clrlsstemb01.in2p3.fr:/root/.
fcs-lpc-bench                                              100%  714KB 714.2KB/s   00:00
```
