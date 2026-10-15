SCRIPT_TARGETS := ci clean package release test validate
ARCH ?= amd64
export ARCH
PLATFORM ?= linux/$(ARCH)

BUILDX_BUILDER := machine
IMAGE_BUILDER ?= docker buildx
DEFAULT_PLATFORMS := linux/amd64,linux/arm64
BUILDX_ARGS ?= --provenance=false --sbom=false

$(SCRIPT_TARGETS):
	./scripts/$@

.PHONY: buildx
buildx:
	@bash -c 'set -euo pipefail; \
	source ./scripts/version; \
	BUILDX_PLATFORM="$${BUILDX_PLATFORM:-$${TARGET_PLATFORMS:-$(PLATFORM)}}"; \
	$(IMAGE_BUILDER) build \
		$${IID_FILE_FLAG:-} \
		--file package/Dockerfile \
		$${BUILDX_BUILDER_FLAG:+--builder $${BUILDX_BUILDER_FLAG}} \
		--build-arg VERSION="$${VERSION}" \
		--build-arg COMMIT="$${COMMIT}" \
		$${BUILDX_TARGET:+--target $${BUILDX_TARGET}} \
		$${BUILDX_PLATFORM:+--platform $${BUILDX_PLATFORM}} \
		$${BUILDX_OUTPUT:+--output $${BUILDX_OUTPUT}} \
		$${BUILDX_TAG_ARGS:-} \
		$${BUILDX_EXTRA_ARGS:-} \
		$${BUILDX_PUSH:+--push} \
		.'

.PHONY: build
build:
ifeq ($(CROSS),1)
	@./scripts/build
else
	@mkdir -p bin
	@$(MAKE) --no-print-directory buildx \
		BUILDX_TARGET=machine-binary \
		BUILDX_PLATFORM="$(PLATFORM)" \
		BUILDX_OUTPUT="type=local,dest=./bin"
	@chmod +x ./bin/rancher-machine
	@# Ignore exec format errors when ARCH differs from the host.
	@./bin/rancher-machine --version || true
endif

.PHONY: buildx-builder
buildx-builder:
	@docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1 || docker buildx create --name=$(BUILDX_BUILDER) --platform=$(DEFAULT_PLATFORMS)

.PHONY: push-image
push-image: buildx-builder
	@echo "--- Building and Pushing Image ---"
	@bash -c 'set -euo pipefail; \
	source ./scripts/version; \
	IMAGE="$${REPO}/machine:$${TAG}"; \
	$(MAKE) --no-print-directory buildx \
		BUILDX_BUILDER_FLAG="$(BUILDX_BUILDER)" \
		BUILDX_TAG_ARGS="--tag $${IMAGE}" \
		BUILDX_EXTRA_ARGS="$(BUILDX_ARGS)" \
		BUILDX_PUSH=1 \
		IID_FILE_FLAG="$${IID_FILE_FLAG:-}"; \
	echo "Pushed $${IMAGE}"'

.PHONY: push-prime-image
push-prime-image:
	@$(MAKE) --no-print-directory push-image \
		BUILDX_ARGS="--sbom=true --attest type=provenance,mode=max"

.DEFAULT_GOAL := ci

.PHONY: $(SCRIPT_TARGETS)

# Local and CI quality checks share the pinned Podman environment.
PODMAN ?= podman
DEV_IMAGE ?= localhost/ioplane-rancher-machine:dev
LINT_IMAGE ?= localhost/ioplane-rancher-machine:lint
GIT_COMMON_DIR := $(shell git rev-parse --path-format=absolute --git-common-dir)
GIT_DIR := $(shell git rev-parse --absolute-git-dir)
DEV_RUN = $(PODMAN) run --rm -v "$(CURDIR):/src:z" -w /src \
	-v ioplane-rancher-machine-go-modules:/cache/modules -v ioplane-rancher-machine-go-build:/cache/build
GIT_MOUNTS = -v "$(GIT_COMMON_DIR):$(GIT_COMMON_DIR):ro,z" $(if $(filter-out $(GIT_COMMON_DIR),$(GIT_DIR)),-v "$(GIT_DIR):$(GIT_DIR):ro") \
	--env GIT_DIR=$(GIT_DIR) --env GIT_WORK_TREE=/src
NEW_GO = commands/create_recovery_test.go libmachine/drivers/create_recovery.go \
	libmachine/drivers/rpc/create_recovery.go libmachine/drivers/rpc/create_recovery_test.go \
	libmachine/crashreport/crash_report_logger_test.go
.PHONY: dev-build lint-build check check-test check-vet check-build check-lint check-format check-mod check-vuln

dev-build:
	$(PODMAN) build --target dev --build-arg SOURCE_REVISION=$(shell git rev-parse HEAD) \
		--build-arg BUILD_CREATED=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) -t $(DEV_IMAGE) -f Containerfile.dev .
lint-build:
	$(PODMAN) build --target lint --build-arg SOURCE_REVISION=$(shell git rev-parse HEAD) \
		--build-arg BUILD_CREATED=$(shell date -u +%Y-%m-%dT%H:%M:%SZ) -t $(LINT_IMAGE) -f Containerfile.dev .
check-test:
	$(DEV_RUN) $(DEV_IMAGE) go test -race -shuffle=on -covermode=atomic -timeout=5m -run "TestCreateRecoveryUsesCurrentHost|TestExistingUnsupportedHostKeepsOriginalError|TestExistingUserdataHostKeepsOriginalError|TestPrepareCreateRecoveryRPC|TestLoggerPrintfForwardsArguments" ./commands ./libmachine/drivers/rpc ./libmachine/crashreport
check-vet:
	$(DEV_RUN) $(DEV_IMAGE) go vet ./...
check-build:
	$(DEV_RUN) $(DEV_IMAGE) env CGO_ENABLED=0 go build -trimpath -o bin/rancher-machine ./cmd/rancher-machine
check-lint:
	$(DEV_RUN) $(GIT_MOUNTS) $(LINT_IMAGE) golangci-lint run ./...
check-format:
	$(DEV_RUN) $(LINT_IMAGE) sh -c 'golangci-lint fmt --diff $(NEW_GO) > /tmp/format.patch; status=$$?; cat /tmp/format.patch; test "$$status" -eq 0 && test ! -s /tmp/format.patch'
	$(DEV_RUN) $(DEV_IMAGE) sh -ec 'gofmt -l commands/create.go libmachine/drivers/rpc/client_driver.go libmachine/crashreport/crash_report_logger.go > /tmp/gofmt-files; cat /tmp/gofmt-files; test ! -s /tmp/gofmt-files'
check-mod:
	$(DEV_RUN) $(DEV_IMAGE) sh -ec 'cp go.mod /tmp/go.mod; cp go.sum /tmp/go.sum; GOFLAGS=-buildvcs=false go mod tidy; cmp /tmp/go.mod go.mod; cmp /tmp/go.sum go.sum; go mod verify'
check-vuln:
	$(DEV_RUN) $(DEV_IMAGE) go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
check: check-mod check-test check-vet check-build check-lint check-format check-vuln
