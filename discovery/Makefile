PROJECT = gimbal
REGISTRY ?= gcr.io/heptio-images
IMAGE := $(REGISTRY)/$(PROJECT)
SRCDIRS := ./cmd ./pkg
# PHONY = gencerts

TAG_LATEST ?= false

GIT_REF = $(shell git rev-parse --short=8 --verify HEAD)
VERSION ?= $(GIT_REF)

export GO111MODULE=on

test: install
	go test -mod=readonly ./...

vet: | test
	go vet ./...

check: test vet gofmt staticcheck misspell unconvert unparam ineffassign

install:
	go install -mod=readonly -v -tags "oidc gcp" ./...

download:
	go mod download

container:
	docker build . -t $(IMAGE):$(VERSION)

push: container
	docker push $(IMAGE):$(VERSION)
ifeq ($(TAG_LATEST), true)
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest
	docker push $(IMAGE):latest
endif

staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck
	staticcheck \
		-checks all,-ST1003 \
		./cmd/... ./internal/...

misspell:
	go install github.com/client9/misspell/cmd/misspell
	misspell \
		-i clas \
		-locale US \
		-error \
		cmd/* internal/* docs/* design/* *.md

unconvert:
	go install github.com/mdempsky/unconvert
	unconvert -v ./cmd/... ./internal/...

ineffassign:
	go install github.com/gordonklaus/ineffassign
	find $(SRCDIRS) -name '*.go' | xargs ineffassign

pedantic: check errcheck

unparam:
	go install mvdan.cc/unparam
	unparam -exported ./cmd/... ./internal/...

errcheck:
	go install github.com/kisielk/errcheck
	errcheck ./...

gofmt:
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(SRCDIRS) | tee /dev/stderr)"