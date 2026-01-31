#!/bin/bash
# 交叉编译脚本

set -e

VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
LDFLAGS="-ldflags \"-s -w -X main.version=${VERSION} -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)\""

echo "Building server-toolkit ${VERSION}..."

# 创建 bin 目录
mkdir -p bin

# 构建当前平台
echo "Building for current platform..."
go build $LDFLAGS -o bin/server-toolkit ./cmd/server-toolkit

# 交叉编译
echo "Cross-compiling..."

# Linux AMD64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $LDFLAGS -o bin/server-toolkit-linux-amd64 ./cmd/server-toolkit

# Linux ARM64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $LDFLAGS -o bin/server-toolkit-linux-arm64 ./cmd/server-toolkit

# Linux ARMv7
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build $LDFLAGS -o bin/server-toolkit-linux-armv7 ./cmd/server-toolkit

echo "Build complete!"
ls -lh bin/
