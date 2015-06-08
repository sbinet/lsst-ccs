fcs-boot
========

`fcs-boot` runs the lsst-ccs/fcs docker image with all needed options.

## Example

```sh
sh> fcs-boot fcs-mgr lpc
2015/06/08 07:18:19 pom-data: main.POM{XMLName:xml.Name{Space:"http://maven.apache.org/POM/4.0.0", Local:"project"}, Name:"LSST Camera Control Software - Subsystem FCS", Version:"1.6.2-SNAPSHOT"}
2015/06/08 07:18:19 creating DISTRIB [/opt/lsst/DISTRIB]...
2015/06/08 07:18:19 Starting c-wrapper on PC-104... (listen for 134.158.120.94:50000)
2015/06/08 07:18:19 c-wrapper command: [ssh -X root@clrlsstemb01.in2p3.fr startCWrapper --host=134.158.120.94 --port=50000]
2015/06/08 07:18:20 creating DISTRIB [/opt/lsst/DISTRIB]... [done]
3 INFO: ### Registering testbenchLPC for buses LOG  (org.lsst.ccs.bus.jgroups.JGroupsBusMessagingLayer register) [Mon Jun 08 07:18:21 GMT 2015]
5 INFO: adding ====== testbenchLPC to LOG (org.lsst.ccs.bus.jgroups.JGroupsBusMessagingLayer updateMapAddressesFromView) [Mon Jun 08 07:18:25 GMT 2015]
6 INFO: main:Real Hardware (org.lsst.ccs.subsystems.fcs.MainModule initModule) [Mon Jun 08 07:18:25 GMT 2015]
```
