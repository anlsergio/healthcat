MODULE_NAME="wiley.com/do-k8s-cluster-health-check"
BINARY_NAME := chc
VERSION_VAR := $(MODULE_NAME)/version.Version
GIT_VAR := $(MODULE_NAME)/version.GitCommit
BUILD_DATE_VAR := $(MODULE_NAME)/version.BuildDate
REPO_VERSION := $$(git describe --abbrev=0 --tags)
BUILD_DATE := $$(date +%Y-%m-%d-%H:%M)
GIT_HASH := $$(git rev-parse --short HEAD)
GOBUILD_VERSION_ARGS := -ldflags "-X $(VERSION_VAR)=$(REPO_VERSION) -X $(GIT_VAR)=$(GIT_HASH) -X $(BUILD_DATE_VAR)=$(BUILD_DATE)"
# useful for other docker repos
DOCKER_REPO ?= 681504496077.dkr.ecr.us-east-1.amazonaws.com
IMAGE_NAME := $(DOCKER_REPO)/$(BINARY_NAME)
ARCH ?= darwin
GOLANGCI_LINT_VERSION ?= v1.23.8
GOLANGCI_LINT_CONCURRENCY ?= 4
GOLANGCI_LINT_DEADLINE ?= 180
# useful for passing --build-arg http_proxy :)
DOCKER_BUILD_FLAGS :=

all: build

build: *.go 
	go build -v -o build/bin/$(ARCH)/$(BINARY_NAME) $(GOBUILD_VERSION_ARGS)

run: build
	./build/bin/$(ARCH)/$(BINARY_NAME) -i testdev

# Install just performs a normal `go install` which builds the source
# files from the package at `./` .
install: test
	go install -v

docker:
	docker build -t $(IMAGE_NAME):$(GIT_HASH) . $(DOCKER_BUILD_FLAGS)

docker-dev: docker
	docker tag $(IMAGE_NAME):$(GIT_HASH) $(IMAGE_NAME):dev
	docker push $(IMAGE_NAME):dev

release: check test docker
	docker push $(IMAGE_NAME):$(GIT_HASH)
	docker tag $(IMAGE_NAME):$(GIT_HASH) $(IMAGE_NAME):$(REPO_VERSION)
	docker push $(IMAGE_NAME):$(REPO_VERSION)
ifeq (, $(findstring -rc, $(REPO_VERSION)))
	docker tag $(IMAGE_NAME):$(GIT_HASH) $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):latest
endif

test:
	go test ./...

test-race:
	go test -race ./...

bench:
	go test -bench=. ./...

bench-race:
	go test -race -bench=. ./...

clean:
	rm -fr bin/*

fmt:
	go fmt ./...

# removes unneeded dependencies
tidy:
	go mod tidy

version:
	@echo $(REPO_VERSION)

.PHONY: build version
