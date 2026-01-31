#!/bin/bash
# server-toolkit 一键安装脚本

set -e

REPO="Akuma-real/server-toolkit"
BINARY="server-toolkit"
INSTALL_DIR="/usr/local/bin"

# 参数
NIGHTLY=false
for arg in "$@"; do
    case "$arg" in
        --nightly)
            NIGHTLY=true
            ;;
        -h|--help)
            echo "Usage: install.sh [--nightly]"
            echo ""
            echo "  --nightly    安装 Nightly（pre-release）版本（仅支持 Linux/amd64）"
            exit 0
            ;;
        *)
            echo "未知参数: $arg"
            echo "用法: install.sh [--nightly]"
            exit 1
            ;;
    esac
done

# 检测系统
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$OS" != "linux" ]; then
    echo "当前仅支持 Linux（检测到: $OS）"
    exit 1
fi

if [ "$ARCH" != "x86_64" ]; then
    echo "当前仅支持 amd64/x86_64（检测到: $ARCH）"
    exit 1
fi

ARCH="amd64"

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
if [ "$NIGHTLY" = true ]; then
    echo "你选择安装 Nightly（pre-release）版本，可能不稳定。"
    read -p "是否继续？[y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
    VERSION=$(curl -s https://api.github.com/repos/${REPO}/releases/tags/nightly | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
else
    VERSION=$(curl -s https://api.github.com/repos/${REPO}/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
fi

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
