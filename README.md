# Server Toolkit

一个功能强大的 Linux 服务器运维工具箱，使用 Go 和 Bubble Tea 构建。

![Version](https://img.shields.io/badge/version-v0.1.0--beta.1-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## 功能特性

- ✅ 主机名管理
- ✅ SSH 密钥管理
- ✅ SSH 安全加固
- ✅ Cloud-init 配置
- ✅ 交互式 TUI 界面
- ✅ 多语言支持（中文、英文）
- ✅ Dry-run 模式
- ✅ 详细操作日志
- ✅ 自动更新检查
- ✅ 配置持久化

## 系统要求

- Linux (Debian, Ubuntu, AlmaLinux, Rocky, CentOS)
- Go 1.22+ (仅编译时)

## 快速开始

### 一键安装

> 当前一键安装脚本仅支持 **Linux/amd64（x86_64）**。
```bash
bash <(curl -sL https://raw.githubusercontent.com/Akuma-real/server-toolkit/main/scripts/install.sh)
```

> 如果你在非交互环境（例如 CI、无 TTY）安装/更新，建议加 `--yes` 跳过确认提示：
```bash
bash <(curl -sL https://raw.githubusercontent.com/Akuma-real/server-toolkit/main/scripts/install.sh) --yes
```

### 安装 Nightly（pre-release）

Nightly 会随 `main` 分支更新，可能不稳定，建议仅用于测试验证。

> 由于 `bash <(curl ...)` 不便传参，Nightly 推荐用 pipe 方式传入 `--nightly`：

```bash
curl -fsSL https://raw.githubusercontent.com/Akuma-real/server-toolkit/main/scripts/install.sh | bash -s -- --nightly
```

> 若希望非交互自动确认（不提示 y/N），追加 `--yes`：
```bash
curl -fsSL https://raw.githubusercontent.com/Akuma-real/server-toolkit/main/scripts/install.sh | bash -s -- --nightly --yes
```

### 从源码构建

```bash
git clone https://github.com/Akuma-real/server-toolkit.git
cd server-toolkit
make build
sudo make install
```

## 使用方法

### 启动 TUI 界面

```bash
server-toolkit
```

### 命令行选项

```bash
server-toolkit --help
```

## 配置

配置文件位于 `/etc/server-toolkit/config.json`：

```json
{
  "language": "zh_CN",
  "dry_run": false,
  "log_level": "INFO",
  "auto_update": true,
  "log_path": "/var/log/server-toolkit.log"
}
```

### 配置说明

| 选项 | 说明 | 可选值 |
|------|------|--------|
| `language` | 界面语言 | `zh_CN`, `en_US` |
| `dry_run` | Dry-run 模式 | `true`, `false` |
| `log_level` | 日志级别 | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `auto_update` | 自动更新检查 | `true`, `false` |
| `log_path` | 日志文件路径 | 任意有效路径 |

## 功能模块

### 系统管理

> 说明：涉及写入 `/etc/*`、调用 `hostnamectl` 的操作通常需要 root 权限；建议使用 `sudo server-toolkit` 运行。

- **设置主机名（一步式向导）**（系统管理仅保留此入口）：
  - 输入短主机名（short）与可选 FQDN
  - 预览将执行的动作后确认执行（支持 dry-run：仅展示计划，不落盘）
  - 若检测到 cloud-init，会提示是否写入 `preserve_hostname: true`（默认 **否**），用于防止重启后被 cloud-init 覆盖
  - 执行内容包含：设置主机名（`hostnamectl` + 写入 `/etc/hostname`）与更新 `/etc/hosts`

#### 回滚/恢复

- `/etc/hostname`、`/etc/hosts`、`/etc/cloud/cloud.cfg.d/99-hostname-preserve.cfg` 均会在写入前生成 `*.bak.YYYYMMDD-HHMMSS` 备份文件。
- 回滚时可将对应备份文件覆盖回原路径（建议使用原子替换/复制），并按需执行 `hostnamectl set-hostname <旧值>`。

### SSH 管理

- **安装 SSH 公钥**: 从 GitHub/URL/文件获取并安装
- **列出已安装的密钥**: 查看当前授权密钥
- **禁用密码登录**: 增强安全配置

## 开发

### 目录结构

```
server-toolkit/
├── cmd/                    # 命令行入口
├── pkg/                    # 公共包
│   ├── tui/               # TUI 组件
│   ├── modules/           # 功能模块
│   ├── system/            # 系统底层
│   └── i18n/             # 国际化
├── internal/              # 内部包
├── scripts/               # 构建脚本
├── test/                  # 测试
└── .github/              # CI/CD
```

### 运行测试

```bash
# 单元测试
go test -v ./...

# 集成测试
docker run --rm -v $(PWD):/app debian:12 sh -c "cd /app && go test ./..."
docker run --rm -v $(PWD):/app almalinux:9 sh -c "cd /app && go test ./..."

# 覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 构建

```bash
# 当前平台
make build

# 交叉编译
make build-all
```

## 贡献

欢迎提交 Issue 和 Pull Request！

### 提交信息规范（Conventional Commits）

为便于生成清晰的提交历史与自动化发布/Changelog，建议提交信息遵循 Conventional Commits：<https://www.conventionalcommits.org/>。

基本格式：

```text
<type>[optional scope][!]: <description>
```

常用 `type`：`feat`、`fix`、`docs`、`refactor`、`test`、`chore`、`ci`、`build`、`perf`、`style`、`revert`。

示例：

- `docs: update install instructions`
- `fix(ssh): avoid duplicating authorized_keys entries`
- `feat(tui)!: redesign main menu`

## 许可证

MIT License - see [LICENSE](LICENSE) file for details

## 致谢

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - 优秀的 TUI 框架
- [kejilion.sh](https://github.com/kejilion/sh) - 灵感来源
