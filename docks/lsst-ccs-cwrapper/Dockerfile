## lsst-ccs/cwrapper
## a container used to build the image for the PC104's smart-card.
from debian

maintainer binet@cern.ch

run apt-get update -y && \
	apt-get install -y debootstrap

env debdist jessie
env rootfs  /build/rootfs

run mkdir -p $rootfs && \
	debootstrap \
		--arch i386 \
		$debdist \
		$rootfs \
		http://http.debian.net/debian/

