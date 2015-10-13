## lsst-ccs/cwrapper
## a container used to build the image for the PC104's smart-card.
from debian

maintainer binet@cern.ch

env debdist jessie

run apt-get update -y && \
	apt-get install -y debootstrap

run mkdir -p /build/rootfs && \
	debootstrap \
		--arch i386 \
		$debdist \
		http://http.debian.net/debian/${debdist}/Release

