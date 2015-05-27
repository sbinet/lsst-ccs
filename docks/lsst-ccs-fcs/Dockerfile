##
## lsst-ccs/fcs
## A container with all dependencies for FCS are installed.
##
FROM lsst-ccs/base
MAINTAINER Sebastien Binet "binet@cern.ch"

USER root
ENV USER root
ENV HOME /root

## update
RUN pacman -Syu --noconfirm && \
	pacman-key --init && \
	pacman-key --populate archlinux \
	;

## install fcs deps 
RUN pacman -S --noconfirm \
	jdk8-openjdk \
	maven \
	subversion \
	which \
	;

## modify maven local repository default location
ADD maven-settings.xml /root/.m2/settings.xml

WORKDIR /opt/lsst

## make the whole container seamlessly executable
CMD ["/bin/bash"]

