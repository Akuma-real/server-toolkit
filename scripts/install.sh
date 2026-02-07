#!/bin/bash
# server-toolkit 一键安装脚本

set -euo pipefail

REPO="Akuma-real/server-toolkit"
BINARY="server-toolkit"
INSTALL_DIR="/usr/local/bin"

require_cmd() {
    local cmd="$1"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "缺少依赖命令: $cmd"
        exit 1
    fi
}

# 参数
NIGHTLY=false
YES=false
for arg in "$@"; do
    case "$arg" in
        --nightly)
            NIGHTLY=true
            ;;
        -y|--yes)
            YES=true
            ;;
        -h|--help)
            echo "Usage: install.sh [--nightly] [--yes]"
            echo ""
            echo "  --nightly    安装 Nightly（pre-release）版本（仅支持 Linux/amd64）"
            echo "  --yes        非交互环境自动确认（不会再提示 y/N）"
            exit 0
            ;;
        *)
            echo "未知参数: $arg"
            echo "用法: install.sh [--nightly] [--yes]"
            exit 1
            ;;
    esac
done

confirm() {
    local prompt="$1"

    if [ "$YES" = true ]; then
        return 0
    fi

    if [ -t 0 ]; then
        read -p "$prompt" -n 1 -r
    elif [ -r /dev/tty ]; then
        read -p "$prompt" -n 1 -r </dev/tty
    else
        echo "检测到非交互环境且无法读取 /dev/tty，请追加 --yes 以继续。"
        return 1
    fi

    echo
    [[ ${REPLY:-} =~ ^[Yy]$ ]]
}

as_root() {
    if [ "${EUID:-0}" -eq 0 ]; then
        "$@"
        return
    fi

    if command -v sudo >/dev/null 2>&1; then
        sudo "$@"
        return
    fi

    echo "需要 root 权限（请用 root 运行或安装 sudo）。"
    exit 1
}

# 检测系统
require_cmd curl
require_cmd sha256sum

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
if command -v "$BINARY" >/dev/null 2>&1; then
    echo "server-toolkit 已安装"
    if ! confirm "是否要更新？[y/N] "; then
        exit 0
    fi
fi

# 获取最新版本
echo "正在获取最新版本..."
if [ "$NIGHTLY" = true ]; then
    echo "你选择安装 Nightly（pre-release）版本，可能不稳定。"
    if ! confirm "是否继续？[y/N] "; then
        exit 0
    fi
    VERSION="$(
        curl -fsSL "https://api.github.com/repos/${REPO}/releases/tags/nightly" 2>/dev/null \
            | sed -nE 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/p' \
            | head -n 1 \
            || true
    )"
else
    VERSION="$(
        curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null \
            | sed -nE 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/p' \
            | head -n 1 \
            || true
    )"
fi

if [ -z "$VERSION" ]; then
    echo "无法获取版本信息"
    exit 1
fi

echo "最新版本: $VERSION"

# 下载校验文件（checksums.txt 或 checksums.sha256）
CHECKSUM_URL_BASE="https://github.com/${REPO}/releases/download/${VERSION}"
CHECKSUM_FILE=""
TMP_SUM="$(mktemp -t "${BINARY}.sum.XXXXXX")"

if curl -fsSL "${CHECKSUM_URL_BASE}/checksums.txt" -o "$TMP_SUM" 2>/dev/null; then
    CHECKSUM_FILE="checksums.txt"
elif curl -fsSL "${CHECKSUM_URL_BASE}/checksums.sha256" -o "$TMP_SUM" 2>/dev/null; then
    CHECKSUM_FILE="checksums.sha256"
else
    echo "未找到校验文件（checksums.txt/checksums.sha256），停止安装。"
    exit 1
fi

echo "已下载校验文件: ${CHECKSUM_FILE}"

# 下载
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}-${OS}-${ARCH}"
echo "正在下载 $URL..."
TMP_FILE="$(mktemp -t "${BINARY}.XXXXXX")"
cleanup() {
    rm -f "$TMP_FILE"
    rm -f "${TMP_SUM:-}"
}
trap cleanup EXIT

curl -fsSL "$URL" -o "$TMP_FILE" || {
    echo "下载失败"
    exit 1
}

# 校验 SHA256
EXPECTED_SUM="$( (sed -nE "s/^([a-fA-F0-9]{64})[[:space:]]+\*?${BINARY}-${OS}-${ARCH}$/\1/p" "$TMP_SUM" || true) | head -n 1 )"
if [ -z "$EXPECTED_SUM" ]; then
    echo "校验文件中未找到 ${BINARY}-${OS}-${ARCH} 的 SHA256，停止安装。"
    exit 1
fi

ACTUAL_SUM="$(sha256sum "$TMP_FILE" | awk '{print $1}')"
if [ "$EXPECTED_SUM" != "$ACTUAL_SUM" ]; then
    echo "SHA256 校验失败，停止安装。"
    echo "expected: $EXPECTED_SUM"
    echo "actual:   $ACTUAL_SUM"
    exit 1
fi

echo "SHA256 校验通过。"

# 安装
echo "正在安装 $BINARY..."
TARGET_PATH="${INSTALL_DIR}/${BINARY}"
if [ -f "$TARGET_PATH" ]; then
    BACKUP_PATH="${TARGET_PATH}.bak.$(date +%Y%m%d%H%M%S)"
    echo "检测到已存在的二进制，先备份到: $BACKUP_PATH"
    as_root cp -a "$TARGET_PATH" "$BACKUP_PATH"
fi

as_root mkdir -p "$INSTALL_DIR"
as_root install -m 0755 "$TMP_FILE" "$TARGET_PATH"

echo "安装完成！运行 '$BINARY' 启动。"
