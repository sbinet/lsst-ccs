##
## lsst-ccs/fcs
## A container where all dependencies for FCS are installed.
##
FROM lsst-ccs/base
MAINTAINER Sebastien Binet "binet@cern.ch"

USER root
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

## update
RUN pacman-key --init && \
	pacman-key --populate archlinux && \
	pacman -Syu --noconfirm \
	;

## install fcs deps 
RUN pacman -S --noconfirm \
	awk \
	bash-completion \
	jdk8-openjdk \
	maven \
	mysql \
	openssh \
	sed \
	subversion \
	which \
	xorg-server xorg-xclock xorg-xhost libxtst \
	;

## create lsst user
RUN useradd -m -g users -G wheel -s /bin/bash lsst
USER lsst
ENV HOME /home/lsst

## modify maven local repository default location
ADD maven-settings.xml /home/lsst/.m2/settings.xml

## CANOpen will need this port
EXPOSE 50000

## JGroups will need this port
EXPOSE 45566

WORKDIR /opt/lsst

## make the whole container seamlessly executable
CMD ["/bin/bash"]

