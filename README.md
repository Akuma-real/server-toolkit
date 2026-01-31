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

\`\`\`bash
bash <(curl -sL https://raw.githubusercontent.com/Akuma-real/server-toolkit/main/scripts/install.sh)
\`\`\`

### 从源码构建

\`\`\`bash
git clone https://github.com/Akuma-real/server-toolkit.git
cd server-toolkit
make build
sudo make install
\`\`\`

## 使用方法

### 启动 TUI 界面

\`\`\`bash
server-toolkit
\`\`\`

### 命令行选项

\`\`\`bash
server-toolkit --help
\`\`\`

## 配置

配置文件位于 \`/etc/server-toolkit/config.json\`：

\`\`\`json
{
  "language": "zh_CN",
  "dry_run": false,
  "log_level": "INFO",
  "auto_update": true,
  "log_path": "/var/log/server-toolkit.log"
}
\`\`\`

### 配置说明

| 选项 | 说明 | 可选值 |
|------|------|--------|
| \`language\` | 界面语言 | \`zh_CN\`, \`en_US\` |
| \`dry_run\` | Dry-run 模式 | \`true\`, \`false\` |
| \`log_level\` | 日志级别 | \`DEBUG\`, \`INFO\`, \`WARN\`, \`ERROR\` |
| \`auto_update\` | 自动更新检查 | \`true\`, \`false\` |
| \`log_path\` | 日志文件路径 | 任意有效路径 |

## 功能模块

### 系统管理

- **设置主机名**: 修改系统主机名
- **配置 /etc/hosts**: 更新 hosts 文件
- **Cloud-init 配置**: 配置 cloud-init preserve_hostname

### SSH 管理

- **安装 SSH 公钥**: 从 GitHub/URL/文件获取并安装
- **列出已安装的密钥**: 查看当前授权密钥
- **禁用密码登录**: 增强安全配置
- **启用 SSH 服务**: 确保 SSH 服务运行

## 开发

### 目录结构

\`\`\`
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
\`\`\`

### 运行测试

\`\`\`bash
# 单元测试
go test -v ./...

# 集成测试
docker run --rm -v $(PWD):/app debian:12 sh -c "cd /app && go test ./..."
docker run --rm -v $(PWD):/app almalinux:9 sh -c "cd /app && go test ./..."

# 覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
\`\`\`

### 构建

\`\`\`bash
# 当前平台
make build

# 交叉编译
make build-all
\`\`\`

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License - see [LICENSE](LICENSE) file for details

## 致谢

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - 优秀的 TUI 框架
- [kejilion.sh](https://github.com/kejilion/sh) - 灵感来源

## 路线图

- [ ] v0.1.0-beta.1 - 初始 Beta 版
- [ ] v0.1.0 - 正式版发布
- [ ] v0.2.0 - Docker 管理模块
- [ ] v0.3.0 - Web 服务管理
- [ ] v1.0.0 - 稳定版发布
