#!/bin/bash
# server-toolkit 一键安装脚本

set -e

REPO="Akuma-real/server-toolkit"
BINARY="server-toolkit"
INSTALL_DIR="/usr/local/bin"

# 检测系统
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    armv7l) ARCH="armv7" ;;
    *) echo "不支持的架构: $ARCH"; exit 1 ;;
esac

# 检查是否已安装
if command -v $BINARY >/dev/null 2>&1; then
    echo "server-toolkit 已安装"
    read -p "是否要更新？[y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
fi

# 获取最新版本
echo "正在获取最新版本..."
VERSION=$(curl -s https://api.github.com/repos/${REPO}/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "无法获取版本信息"
    exit 1
fi

echo "最新版本: $VERSION"

# 下载
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}-${OS}-${ARCH}"
echo "正在下载 $URL..."
curl -fsSL "$URL" -o /tmp/$BINARY || {
    echo "下载失败"
    exit 1
}

# 安装
echo "正在安装 $BINARY..."
sudo install -m 0755 /tmp/$BINARY $INSTALL_DIR/
rm -f /tmp/$BINARY

echo "安装完成！运行 '$BINARY' 启动。"
