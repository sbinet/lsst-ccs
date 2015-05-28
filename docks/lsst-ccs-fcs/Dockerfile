##
## lsst-ccs/fcs
## A container where all dependencies for FCS are installed.
##
FROM lsst-ccs/base
MAINTAINER Sebastien Binet "binet@cern.ch"

USER root
ENV USER root
ENV HOME /root

## update
RUN pacman-key --init && \
	pacman-key --populate archlinux && \
	pacman -Syu --noconfirm \
	;

## install fcs deps 
RUN pacman -S --noconfirm \
	bash-completion \
	jdk8-openjdk \
	maven \
	openssh \
	subversion \
	which \
	;

## create lsst user
RUN useradd -m -g users -G wheel -s /bin/bash lsst
USER lsst
ENV HOME /home/lsst

## modify maven local repository default location
ADD maven-settings.xml /home/lsst/.m2/settings.xml

## CANOpen will need this port
EXPOSE 50000

WORKDIR /opt/lsst

## make the whole container seamlessly executable
CMD ["/bin/bash"]

