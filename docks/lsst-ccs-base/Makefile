.PHONY: build all

all: build

build-image: mk-arch
	sudo ./mk-arch

build: build-image arch-image.tar.xz
	cat ./arch-image.tar.xz | docker import - lsst-ccs/base

