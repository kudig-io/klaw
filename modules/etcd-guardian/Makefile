# Image URL to use all building/pushing image targets
IMG ?= etcdguardian/operator:latest
BACKEND_IMG ?= etcdguardian/backend:latest
WEBUI_IMG ?= etcdguardian/web-ui:latest

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE) -w -s"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: version
version: ## Display version information.
	@echo "Version:    $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

##@ Development

.PHONY: manifests
manifests: ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	go run sigs.k8s.io/controller-tools/cmd/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	go run sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint.
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
	golangci-lint run --timeout 5m

.PHONY: test
test: manifests generate fmt vet ## Run tests.
	go test ./... -coverprofile cover.out -v

.PHONY: test-coverage
test-coverage: test ## Generate test coverage report.
	go tool cover -html=cover.out -o cover.html
	@echo "Coverage report generated: cover.html"

.PHONY: test-race
test-race: ## Run tests with race detector.
	go test -race ./...

.PHONY: test-integration
test-integration: ## Run integration tests.
	go test -tags=integration ./...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests.
	go test -tags=e2e ./test/e2e/...

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build $(LDFLAGS) -o bin/manager cmd/manager/main.go

.PHONY: build-all
build-all: build build-backend build-cli ## Build all binaries.

.PHONY: build-backend
build-backend: ## Build backend API binary.
	cd backend && go build $(LDFLAGS) -o ../bin/backend main.go

.PHONY: build-cli
build-cli: ## Build CLI tool.
	go build $(LDFLAGS) -o bin/etcdguardian cmd/cli/main.go

.PHONY: build-cross
build-cross: ## Build binaries for multiple platforms.
	@echo "Building for linux/amd64..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/manager-linux-amd64 cmd/manager/main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/etcdguardian-linux-amd64 cmd/cli/main.go
	@echo "Building for linux/arm64..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/manager-linux-arm64 cmd/manager/main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/etcdguardian-linux-arm64 cmd/cli/main.go
	@echo "Building for darwin/amd64..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/manager-darwin-amd64 cmd/manager/main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/etcdguardian-darwin-amd64 cmd/cli/main.go
	@echo "Building for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/manager-darwin-arm64 cmd/manager/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/etcdguardian-darwin-arm64 cmd/cli/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run cmd/manager/main.go

.PHONY: run-backend
run-backend: ## Run backend API server.
	cd backend && go run main.go

##@ Docker

.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	docker build -t ${IMG} --build-arg VERSION=${VERSION} --build-arg GIT_COMMIT=${GIT_COMMIT} --build-arg BUILD_DATE=${BUILD_DATE} .

.PHONY: docker-build-backend
docker-build-backend: ## Build backend API docker image.
	docker build -f Dockerfile.backend -t ${BACKEND_IMG} --build-arg VERSION=${VERSION} --build-arg GIT_COMMIT=${GIT_COMMIT} --build-arg BUILD_DATE=${BUILD_DATE} .

.PHONY: docker-build-webui
docker-build-webui: ## Build web UI docker image.
	docker build -f Dockerfile.webui -t ${WEBUI_IMG} .

.PHONY: docker-build-all
docker-build-all: docker-build docker-build-backend docker-build-webui ## Build all docker images.

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: docker-push-backend
docker-push-backend: ## Push backend API docker image.
	docker push ${BACKEND_IMG}

.PHONY: docker-push-webui
docker-push-webui: ## Push web UI docker image.
	docker push ${WEBUI_IMG}

.PHONY: docker-push-all
docker-push-all: docker-push docker-push-backend docker-push-webui ## Push all docker images.

.PHONY: docker-tag
docker-tag: ## Tag docker images for release.
	docker tag ${IMG} etcdguardian/operator:$(VERSION)
	docker tag ${BACKEND_IMG} etcdguardian/backend:$(VERSION)
	docker tag ${WEBUI_IMG} etcdguardian/web-ui:$(VERSION)

##@ Kubernetes Deployment

.PHONY: install
install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd

.PHONY: uninstall
uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/crd

.PHONY: deploy
deploy: manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd
	kubectl apply -f config/rbac
	kubectl apply -f config/manager

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/manager
	kubectl delete -f config/rbac
	kubectl delete -f config/crd

.PHONY: deploy-helm
deploy-helm: ## Deploy using Helm chart.
	helm upgrade --install etcdguardian charts/etcdguardian \
		--namespace etcd-guardian-system \
		--create-namespace \
		--set image.repository=${IMG} \
		--set image.tag=$(VERSION)

.PHONY: undeploy-helm
undeploy-helm: ## Undeploy Helm release.
	helm uninstall etcdguardian --namespace etcd-guardian-system

##@ Docker Compose

.PHONY: compose-up
compose-up: ## Start all services using docker-compose.
	docker-compose up -d

.PHONY: compose-down
compose-down: ## Stop all services using docker-compose.
	docker-compose down

.PHONY: compose-dev
compose-dev: ## Start development environment using docker-compose.
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

.PHONY: compose-logs
compose-logs: ## View docker-compose logs.
	docker-compose logs -f

##@ Dependencies

.PHONY: tidy
tidy: ## Tidy go dependencies.
	go mod tidy
	go mod download

.PHONY: deps
deps: ## Install dependencies.
	go mod download

.PHONY: deps-update
deps-update: ## Update dependencies.
	go get -u ./...
	go mod tidy

.PHONY: deps-verify
deps-verify: ## Verify dependencies.
	go mod verify

##@ Security

.PHONY: security-scan
security-scan: ## Run security vulnerability scan.
	@which trivy > /dev/null || brew install trivy
	trivy image --severity HIGH,CRITICAL ${IMG}

.PHONY: security-check
security-check: ## Run security checks.
	go list -json -m all | nancy sleuth

.PHONY: govulncheck
govulncheck: ## Run Go vulnerability check.
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

##@ Release

.PHONY: release
release: ## Create a new release.
	@echo "Creating release $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

.PHONY: release-snapshot
release-snapshot: ## Create a snapshot release.
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --clean

##@ Clean

.PHONY: clean
clean: ## Clean build artifacts.
	@rm -rf bin/
	@rm -rf dist/
	@rm -f cover.out cover.html

.PHONY: clean-all
clean-all: clean compose-down ## Clean all artifacts and stop services.
	@rm -rf vendor/
	@docker system prune -f

##@ Documentation

.PHONY: docs
docs: ## Generate documentation.
	go run golang.org/x/pkgsite/cmd/pkgsite@latest

.PHONY: swagger
swagger: ## Generate Swagger documentation.
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	cd backend && swag init

##@ Utilities

.PHONY: check
check: fmt vet lint test ## Run all checks.

.PHONY: pre-commit
pre-commit: check security-check ## Run pre-commit checks.

.PHONY: ci
ci: check test-race govulncheck ## Run CI pipeline locally.

.PHONY: watch
watch: ## Watch for changes and rebuild.
	@which reflex > /dev/null || go install github.com/cespare/reflex@latest
	reflex -r '\.go$$' -s -- sh -c "make build"

##@ CLI

.PHONY: cli
cli: build-cli ## Build CLI tool (alias for build-cli).

.PHONY: cli-install
cli-install: ## Install CLI tool to system.
	go install cmd/cli/main.go

##@ Code Generation

.PHONY: mock
mock: ## Generate mocks for testing.
	@which mockgen > /dev/null || go install github.com/golang/mock/mockgen@latest
	go generate ./... -type=Mock

.PHONY: proto
proto: ## Generate protobuf code.
	@which protoc > /dev/null || echo "protoc not installed"
	protoc --go_out=. --go_opt=paths=source_relative proto/*.proto

##@ Performance

.PHONY: benchmark
benchmark: ## Run benchmarks.
	go test -bench=. -benchmem ./...

.PHONY: profile
profile: ## Generate CPU and memory profiles.
	go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	go tool pprof -http=:8080 cpu.prof
