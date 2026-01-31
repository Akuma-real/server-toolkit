.PHONY: build build-all test test-integration clean install release fmt lint deps

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
BINARY := server-toolkit

# 构建
build:
	@echo "Building $(BINARY)..."
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/$(BINARY)

# 交叉编译
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/$(BINARY)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-arm64 ./cmd/$(BINARY)
	@echo "Build complete: bin/"

# 测试
test:
	@echo "Running tests..."
	go test -v -race ./...

# 集成测试
test-integration:
	@echo "Running integration tests..."
	@echo "Testing on Debian 12..."
	docker run --rm -v $(PWD):/app debian:12 sh -c "cd /app && apt-get update -qq && apt-get install -y golang && go test ./..."
	@echo "Testing on AlmaLinux 9..."
	docker run --rm -v $(PWD):/app almalinux:9 sh -c "cd /app && dnf install -y golang && go test ./..."

# 代码格式化
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "Warning: goimports not found; skipping. Install with:"; \
		echo "  go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# 代码检查
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# 依赖管理
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 安装
install: build
	@echo "Installing $(BINARY) to /usr/local/bin..."
	sudo install -m 0755 bin/$(BINARY) /usr/local/bin/
	@echo "Installation complete!"

# 卸载
uninstall:
	@echo "Uninstalling $(BINARY)..."
	sudo rm -f /usr/local/bin/$(BINARY)
	@echo "Uninstall complete!"

# 清理
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out
	@echo "Clean complete!"

# 发布
release: build-all
	@echo "Creating release $(VERSION)..."
	gh release create $(VERSION) bin/$(BINARY)-*
	@echo "Release created!"

# 运行
run: build
	@echo "Running $(BINARY)..."
	./bin/$(BINARY)

# 开发
dev:
	@echo "Running in development mode..."
	go run ./cmd/$(BINARY)
