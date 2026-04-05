# 变量定义
BINARY_NAME=klaw
GO_VERSION=1.20
DOCKER_IMAGE=klaw:latest
DOCKER_REGISTRY=kudig-io

# 默认目标
.PHONY: all
all: build

# 构建前端
.PHONY: build-frontend
build-frontend:
	@echo "Building frontend..."
	cd web && npm install && npm run build

# 构建后端
.PHONY: build-backend
build-backend:
	@echo "Building backend..."
	go build -o $(BINARY_NAME) cmd/klaw/main.go

# 构建所有
.PHONY: build
build: build-frontend build-backend
	@echo "Build complete!"

# 运行开发环境
.PHONY: dev
dev:
	@echo "Starting development environment..."
	@make -j2 dev-frontend dev-backend

.PHONY: dev-frontend
dev-frontend:
	@echo "Starting frontend dev server..."
	cd web && npm run dev

.PHONY: dev-backend
dev-backend:
	@echo "Starting backend server..."
	go run cmd/klaw/main.go

# 运行应用
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -rf web/dist
	rm -rf web/node_modules

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# 运行Go测试
.PHONY: test-go
test-go:
	@echo "Running Go tests..."
	go test -v ./...

# 运行前端测试
.PHONY: test-frontend
test-frontend:
	@echo "Running frontend tests..."
	cd web && npm test

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	cd web && npm run lint -- --fix

# 代码检查
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run
	cd web && npm run lint

# Docker相关
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -d \
		-p 8080:8080 \
		-v ~/.kube/config:/root/.kube/config \
		-v $(PWD)/configs/config.yaml:/app/configs/config.yaml \
		$(DOCKER_IMAGE)

.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker container..."
	docker stop $$(docker ps -q --filter ancestor=$(DOCKER_IMAGE))

.PHONY: docker-clean
docker-clean: docker-stop
	@echo "Removing Docker container and image..."
	docker rm $$(docker ps -aq --filter ancestor=$(DOCKER_IMAGE)) || true
	docker rmi $(DOCKER_IMAGE) || true

# Helm相关
.PHONY: helm-install
helm-install:
	@echo "Installing Helm chart..."
	helm install klaw ./helm/klaw

.PHONY: helm-upgrade
helm-upgrade:
	@echo "Upgrading Helm chart..."
	helm upgrade klaw ./helm/klaw

.PHONY: helm-uninstall
helm-uninstall:
	@echo "Uninstalling Helm chart..."
	helm uninstall klaw

.PHONY: helm-package
helm-package:
	@echo "Packaging Helm chart..."
	helm package ./helm/klaw

# 依赖管理
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	cd web && npm install

# 生成依赖
.PHONY: deps-go
deps-go:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

.PHONY: deps-frontend
deps-frontend:
	@echo "Installing frontend dependencies..."
	cd web && npm install

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Build everything (default)"
	@echo "  build            - Build frontend and backend"
	@echo "  build-frontend   - Build only frontend"
	@echo "  build-backend    - Build only backend"
	@echo "  dev              - Start development environment (frontend + backend)"
	@echo "  dev-frontend    - Start frontend dev server"
	@echo "  dev-backend     - Start backend server"
	@echo "  run              - Build and run the application"
	@echo "  clean            - Clean build artifacts"
	@echo "  test             - Run all tests"
	@echo "  test-go          - Run Go tests"
	@echo "  test-frontend    - Run frontend tests"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint code"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  docker-stop      - Stop Docker container"
	@echo "  docker-clean     - Clean Docker container and image"
	@echo "  helm-install     - Install Helm chart"
	@echo "  helm-upgrade     - Upgrade Helm chart"
	@echo "  helm-uninstall   - Uninstall Helm chart"
	@echo "  helm-package     - Package Helm chart"
	@echo "  deps             - Install all dependencies"
	@echo "  deps-go          - Install Go dependencies"
	@echo "  deps-frontend    - Install frontend dependencies"
	@echo "  help             - Show this help message"
