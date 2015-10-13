#!/bin/bash

set -x
set -e

apt-get update -y
apt-get install -y debootstrap

export debdist=jessie
export rootfs=/build/rootfs

mkdir -p $rootfs

debootstrap \
		--arch i386 \
		$debdist \
		$rootfs \
		http://http.debian.net/debian/

mount -t proc  /proc $rootfs/proc
mount -t sysfs /sys  $rootfs/sys
cp -f /etc/hosts     $rootfs/etc/.

