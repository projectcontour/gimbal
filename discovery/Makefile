# Copyright 2016 The Kubernetes Authors.
#
# Modifications Copyright 2017 the Heptio contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This repo's root import path (under GOPATH).
PKG := github.com/heptio/gimbal/discovery

# Where to push the docker image.
REGISTRY ?= gcr.io/heptio-images

# Which architecture to build - see $(ALL_ARCH) for options.
ARCH ?= linux-amd64

VERSION ?= master

###
### These variables should not need tweaking.
###

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

CLI_PLATFORMS := linux-amd64 darwin-amd64
CONTAINER_PLATFORMS := linux-amd64

platform_temp = $(subst -, ,$(ARCH))
GOOS = $(word 1, $(platform_temp))
GOARCH = $(word 2, $(platform_temp))

IMAGE := $(REGISTRY)/gimbal-discoverer

BUILD_IMAGE ?= gcr.io/heptio-images/golang:1.9-alpine3.6

PARENTDIR ?= $(dirname "$(pwd)")

GIT_SHA ?= $(git --git-dir=../git describe --tags --always)
GIT_DIRTY ?= $(git --git-dir=../git status --porcelain 2> /dev/null)

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: build

build-%:
	@$(MAKE) --no-print-directory ARCH=$* build

all-build: $(addprefix build-, $(CLI_PLATFORMS))

build:
	@$(MAKE) BIN=kubernetes-discoverer build_bin
	@$(MAKE) BIN=openstack-discoverer build_bin

build_bin: _output/bin/$(GOOS)/$(GOARCH)/$(BIN)

_output/bin/$(GOOS)/$(GOARCH)/$(BIN): build-dirs
	@echo "building: $@"
	$(MAKE) shell CMD="-c '\
		GOOS=$(GOOS) \
		GOARCH=$(GOARCH) \
		VERSION=$(VERSION) \
		PKG=$(PKG) \
		BIN=$(BIN) \
		OUTPUT_DIR=/output/$(GOOS)/$(GOARCH) \
		GIT_SHA=$(GIT_SHA) \
		GIT_DIRTY=$(GIT_DIRTY) \
		./hack/build.sh'"

TTY := $(shell tty -s && echo "-t")

# Example: make shell CMD="date > datefile"
shell: build-dirs
	@docker run \
		-i $(TTY) \
		--rm \
		-u $$(id -u):$$(id -g) \
		-v "$$(pwd)/.go/pkg:/go/pkg" \
		-v "$$(pwd)/.go/src:/go/src" \
		-v "$$(pwd)/.go/std:/go/std" \
		-v "$$(pwd)/.go/git:/git" \
		-v "$$(pwd):/go/src/$(PKG)" \
		-v "$$(pwd)/_output/bin:/output" \
		-v "$$(pwd)/.go/std/$(GOOS)/$(GOARCH):/usr/local/go/pkg/$(GOOS)_$(GOARCH)_static" \
		-w /go/src/$(PKG) \
		$(BUILD_IMAGE) \
		/bin/sh $(CMD)

# DOTFILE_IMAGE = $(subst :,_,$(subst /,_,$(IMAGE))-$(VERSION))

container: verify test build
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)

SKIP_TESTS ?=
test: build-dirs
ifneq ($(SKIP_TESTS), 1)
	@$(MAKE) shell CMD="-c 'hack/test.sh $(SRC_DIRS)'"
endif

verify:
ifneq ($(SKIP_TESTS), 1)
	@$(MAKE) shell CMD="-c 'hack/verify-all.sh'"
endif

update:
	@$(MAKE) shell CMD="-c 'hack/update-all.sh'"

release: all-tar-bin checksum

checksum:
	@cd _output/release; \
	sha256sum *.tar.gz > CHECKSUM; \
	cat CHECKSUM; \
	sha256sum CHECKSUM

all-tar-bin: $(addprefix tar-bin-, $(CLI_PLATFORMS))

tar-bin-%:
	$(MAKE) ARCH=$* VERSION=$(VERSION) tar-bin

GIT_DESCRIBE = $(shell git describe --tags --always --dirty)
tar-bin: build
	mkdir -p _output/release

# We do the subshell & wildcard ls so we can pick up $(BIN).exe for windows
	(cd _output/bin/$(GOOS)/$(GOARCH) && ls kubernetes-discoverer* openstack-discoverer*) | \
		tar \
			-C _output/bin/$(GOOS)/$(GOARCH) \
			--files-from=- \
			-zcf _output/release/gimbal-discoverer-$(GIT_DESCRIBE)-$(GOOS)-$(GOARCH).tar.gz

build-dirs:
	@mkdir -p _output/bin/$(GOOS)/$(GOARCH)
	@mkdir -p .go/src/$(PKG) .go/pkg .go/bin .go/std/$(GOOS)/$(GOARCH)

clean: container-clean bin-clean

container-clean:
	rm -rf .container-* _output/.dockerfile-* .push-*

bin-clean:
	rm -rf .go _output

ci:
	hack/verify-all.sh
	hack/test.sh $(SRC_DIRS)
	GOOS=$(GOOS) GOARCH=$(GOARCH) VERSION=$(VERSION) PKG=$(PKG) GIT_SHA=$(GIT_SHA) GIT_DIRTY=$(GIT_DIRTY) BIN=kubernetes-discoverer ./hack/build.sh
	GOOS=$(GOOS) GOARCH=$(GOARCH) VERSION=$(VERSION) PKG=$(PKG) GIT_SHA=$(GIT_SHA) GIT_DIRTY=$(GIT_DIRTY) BIN=openstack-discoverer ./hack/build.sh