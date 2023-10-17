$(include .env)
FILES := $(wildcard *.go)
BUILD := patching-automation
BUILD.WINDOWS := $(BUILD).windows
BUILD.LINUX := $(BUILD).linux
BUILD.OSX := $(BUILD).darwin

VERSION ?= UNSET
PRERELEASE_TAG ?=
VERSION_REGEX := ^([0-9]{1,}\.){2}[0-9]{1,}$$

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)
EDITOR ?= code

# Get the git commit info for version tag
GIT_BRANCH := $(shell git symbolic-ref --short HEAD 2>/dev/null)
GIT_DIRTY := $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_IMPORT := github.com/tjm/puppet-patching-automation/version
LDFLAGS ?= -s -w -extldflags '-static'
GOLDFLAGS = -X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY) $(LDFLAGS)

# Parse version/version.go for *current* version information
VERSION_GO_APP := $(shell awk '/Version = / {print $$4}' version/version.go | tr -d [\"])
VERSION_GO_PRE := $(shell awk '/VersionPrerelease = / {print $$4}' version/version.go | tr -d [\"])
VERSION_FULL := $(if VERSION_GO_PRE,$(VERSION_GO_APP)-$(VERSION_GO_PRE),$(VERSION_GO_APP))

DOCKER_IMAGE ?= $(BUILD)
DOCKER_TAG ?= $(VERSION_FULL)
DOCKER_REGISTRY=ghcr.io/tjm/puppet-patching-automation

HELM_CHART=helmchart/$(BUILD)
HELM_CHART_VERSION ?= $(shell awk '/^version: / {print $$2}' $(HELM_CHART)/Chart.yaml)
HELM_APP_VERSION ?= $(shell awk '/^appVersion: / {print $$2}' $(HELM_CHART)/Chart.yaml)
# HELM_BITNAMI_REPO ?= https://raw.githubusercontent.com/bitnami/charts/pre-2022/bitnami
HELM_BITNAMI_REPO ?= https://charts.bitnami.com/bitnami
HELM_PACKAGE=$(BUILD)-$(HELM_CHART_VERSION).tgz
HELM_PUBLISH_URL ?= https://artifactory.example.com/artifactory/helm-local
HELM_PUBLISH_AUTH ?= # anonymous
#HELM_PUBLISH_AUTH ?=-u $(USER) # prompt for password
#HELM_PUBLISH_AUTH ?=-u $(USER):$(PASS) # expose password

CURL_FAIL := $(shell curl --help all | grep -q fail-with-body && echo "--fail-with-body" || echo "--fail")

# By default, build for the "current" OS/Arch (or whatever is specified by GOOS/GOARCH env vars)
all: lint sec tidy vendor build test format $(BUILD)

PHONY+= clean
clean:
	rm -fr vendor $(BUILD) $(BUILD).$(GOOS)

PHONY+= realclean
realclean: clean cleandb
	rm -fr vendor $(BUILD.LINUX) $(BUILD.OSX) $(BUILD.WINDOWS)

PHONY+= cleandb
cleandb:
	@echo "Moving database aside, not removing."
	test -f db/padb.db && mv -f db/padb.db db/padb-backup-`date +'%Y-%m-%d-%H-%M-%S'`.db

PHONY+= test
test:
	@echo "🔘 Running unit tests... (`date '+%H:%M:%S'`)"
	@go test $(TESTFLAGS) ./...

# Run go mod tidy and check go.sum is unchanged
PHONY+= tidy
tidy:
	@echo "🔘 Checking that go mod tidy does not make a change..."
	@cp go.sum go.sum.bak
	@go mod tidy
	@diff go.sum go.sum.bak && rm go.sum.bak || (echo "🔴 go mod tidy would make a change, exiting"; exit 1)
	@echo "✅ Checking go mod tidy complete"

# Format go code and error if any changes are made
PHONY+= format
format:
	@echo "🔘 Checking that go fmt does not make any changes..."
	@test -z $$(go fmt ./...) || (echo "🔴 go fmt would make a change, exiting"; exit 1)
	@echo "✅ Checking go fmt complete"

PHONY+= lint
lint: clean # golint doesn't have exclude for vendors
lint: $(GOPATH)/bin/golangci-lint $(GOPATH)/bin/golint
	@echo "🔘 Linting $(1) (`date '+%H:%M:%S'`)"
	@$(GOPATH)/bin/golint -set_exit_status ./...
	@go vet ./...
	@$(GOPATH)/bin/golangci-lint run \
		-E asciicheck \
		-E bodyclose \
		-E exhaustive \
		-E exportloopref \
		-E gci \
		-E gofmt \
		-E goimports \
		-E gosec \
		-E noctx \
		-E nolintlint \
		-E rowserrcheck \
		-E sqlclosecheck \
		-E stylecheck \
		-E unconvert \
		-E unparam
	@echo "✅ Lint-free (`date '+%H:%M:%S'`)"

PHONY+= sec
sec: $(GOPATH)/bin/gosec
	@echo "🔘 Checking for security problems ... (`date '+%H:%M:%S'`)"
	@$(GOPATH)/bin/gosec ./...
	@echo "✅ No problems found (`date '+%H:%M:%S'`)";

## BUILD

PHONY+= build
build: export CGO_ENABLED=0
build: vendor
	git status
	@echo "🔘 Building - $(1) (`date '+%H:%M:%S'`)"
	@go build -mod=vendor -ldflags "$(GOLDFLAGS)" -o $(BUILD).$(GOOS) $(FILES)
	@echo "✅ Build complete - $(1) (`date '+%H:%M:%S'`)"

$(BUILD): $(BUILD).$(GOOS)
	ln -sf $(BUILD).$(GOOS) $(BUILD)

# legacy static endpoint (same as build)
PHONY+=static
static: build

## Cross-Platform Builds
$(BUILD.OSX): export GOOS=darwin
$(BUILD.OSX): build

$(BUILD.LINUX): export GOOS=linux
$(BUILD.LINUX): build

$(BUILD.WINDOWS): export GOOS=windows
$(BUILD.WINDOWS): build

PHONY+=build_osx
build_osx: $(BUILD.OSX)

PHONY+=build_linux
build_linux: $(BUILD.LINUX)

PHONY+=build_windows
build_windows: $(BUILD.WINDOWS)

PHONY+=build_all
build_all: build_osx build_linux build_windows

## Build Tools
$(GOPATH)/bin/golangci-lint:
	@echo "🔘 Installing golangci-lint... (`date '+%H:%M:%S'`)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin

$(GOPATH)/bin/golint:
	@echo "🔘 Installing golint ... (`date '+%H:%M:%S'`)"
	@GO111MODULE=off go get -u golang.org/x/lint/golint

$(GOPATH)/bin/gosec:
	@echo "🔘 Installing gosec ... (`date '+%H:%M:%S'`)"
	@curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(GOPATH)/bin

PHONY+= update-tools
update-tools: delete-tools $(GOPATH)/bin/golangci-lint $(GOPATH)/bin/golint $(GOPATH)/bin/gosec

PHONY+= delete-tools
delete-tools:
	@rm $(GOPATH)/bin/golangci-lint
	@rm $(GOPATH)/bin/golint
	@rm $(GOPATH)/bin/gosec

PHONY+= run
run: clean vendor
	go run -ldflags "$(GOLDFLAGS)" main.go

PHONY+= vendor
vendor:
	@echo "🔘 Running go mod vendor - $(1) (`date '+%H:%M:%S'`)"
	@go mod vendor
	@echo "✅ go mod vendor complete - $(1) (`date '+%H:%M:%S'`)"

### DOCKER - no actual file produced for docker
PHONY+= docker
docker:
	docker build -t $(DOCKER_IMAGE) .
	docker tag $(DOCKER_IMAGE) $(DOCKER_IMAGE):$(DOCKER_TAG)

PHONY+= dockerpush
dockerpush:
ifeq ($(strip $(VERSION_GO_PRE)),)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
else
	$(info "Refusing to tag/push prelease image, check version/version.go")
endif

### HELM
PHONY+= helm
helm: $(HELM_PACKAGE)

PHONY+= helmpush
helmpush: $(HELM_PACKAGE)
	curl $(CURL_FAIL) $(HELM_PUBLISH_AUTH) -T $(HELM_PACKAGE) $(HELM_PUBLISH_URL)/$(notdir $(HELM_PACKAGE))

$(HELM_PACKAGE):
ifneq ($(VERSION_FULL), $(HELM_APP_VERSION))
	$(error "Helm appVersion ($(HELM_APP_VERSION)) does not match App Version ($(VERSION_FULL))")
endif
	helm repo list 2>/dev/null | grep -q $(HELM_BITNAMI_REPO) || helm repo add bitnami $(HELM_BITNAMI_REPO)
	helm dependency build $(HELM_CHART)
	helm package $(HELM_CHART)


PHONY+= new_branch_from_master
new_branch_from_master:
	@if test -z "$(BRANCH)"; then \
		echo 'BRANCH "$(BRANCH)" must be set!' >&2; \
		exit 99; \
	fi
	@echo "Creating new branch '$(BRANCH)' from **master** branch!"
	git checkout master
	git pull
	git checkout -b $(BRANCH)

PHONY+= update_version
update_version: RELEASE_FULL_VERSION = $(if $(PRERELEASE_TAG),$(VERSION)-$(PRERELEASE_TAG),$(VERSION))
update_version:
ifeq ($(shell echo $(VERSION) | egrep '$(VERSION_REGEX)'),)
	@echo "VERSION=\"$(VERSION)\" was not set or is not complaint with symver: x.y.z (REGEX: \"$(VERSION_REGEX)\")"
	exit 99
else
	@echo "Update version: $(RELEASE_FULL_VERSION)"
	@mkdir -p tmp
	sed -e 's/^const Version =.*/const Version = "$(VERSION)"/;s/^const VersionPrerelease =.*/const VersionPrerelease = "$(PRERELEASE_TAG)"/' version/version.go > tmp/version.go
	\mv tmp/version.go version/version.go
	sed -e 's/^version: .*/version: $(RELEASE_FULL_VERSION)/;s/^appVersion: .*/appVersion: $(RELEASE_FULL_VERSION)/'  $(HELM_CHART)/Chart.yaml > tmp/Chart.yaml
	\mv tmp/Chart.yaml $(HELM_CHART)/Chart.yaml
endif


PHONY+= release
release: export BRANCH=$(subst .,-,build/$(VERSION))
release: new_branch_from_master update_version
	git status
	git commit -m "build: Release $(VERSION)" version/version.go $(HELM_CHART)/Chart.yaml
	git push -o merge_request.create -u origin $(BRANCH)


PHONY+= prepare_for_dev
prepare_for_dev: export BRANCH = $(subst .,-,chore/prepare-dev-$(VERSION))
prepare_for_dev: PRERELEASE_TAG = dev
prepare_for_dev: new_branch_from_master update_version
	git status
	git commit -m "chore: Prepare for dev $(VERSION)-$(PRERELEASE_TAG)" version/version.go $(HELM_CHART)/Chart.yaml
	git push -o merge_request.create -u origin $(BRANCH)


### List all phony (that don't create a target file) builds
.PHONY: $(PHONY)
