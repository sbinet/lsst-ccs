.PHONY: all build

REPO := lsst-ccs/cwrapper
TAG ?= $(shell date +%Y%m%d)

all: build
	@echo "done"

build: build/build-img.sh
	docker run --privileged -v $(PWD)/build:/build debian /build/build-img.sh
	docker commit $(shell docker ps -q -l) $(REPO):$(TAG)
	genisoimage -V "$(REPO)-$(TAG)" -r -o debian-32b-$(TAG).img build/rootfs

	#docker build --rm --tag=$(REPO):$(TAG) .
	#docker tag --force $(REPO):$(TAG) $(REPO):latest

upload: build
	docker push $(REPO):$(TAG)
	docker push $(REPO):latest


