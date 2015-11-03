## recipe to cross-compile go programs for PC-104 using gccgo
from debian:jessie
maintainer Sebastien Binet binet@cern.ch

run mkdir -p /go/src

run apt-get update -y
run apt-get install -y git golang
run apt-get install -y gccgo-multilib libc6-dev-i386

env GOPATH /go
env GOARCH 386
env GO386  387

