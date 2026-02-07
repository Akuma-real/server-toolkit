# Repository Guidelines

## 个人项目定位
这是一个我自己长期维护的 Linux 运维工具项目。文档目标不是“团队流程完整”，而是“下次回来能快速继续开发、少踩坑、可回滚”。

## 项目结构（快速索引）
- `cmd/server-toolkit/`：程序入口与向导编排。
- `pkg/modules/`：功能模块（`hostname`、`ssh`）。
- `pkg/system/`：系统调用与 OS 差异封装。
- `pkg/tui/`：Bubble Tea 交互组件与样式。
- `pkg/i18n/`：多语言文案（`zh_CN`、`en_US`）。
- `internal/`：配置、日志、更新检查、dry-run。
- `scripts/`、`.github/workflows/`：安装脚本与 CI/发布流程。

原则：业务逻辑放 `pkg/modules`，系统细节放 `pkg/system`。

## 常用命令
- `make dev`：本地开发运行。
- `make build`：构建当前平台二进制到 `bin/`。
- `make test`：单元测试（含 race）。
- `make test-integration`：容器内跨发行版测试。
- `make fmt`：格式化（`go fmt` + `goimports`）。
- `make lint`：静态检查。

## 编码与命名约定
- Go 代码统一交给 `make fmt`，不手调格式。
- 包名小写短词；导出用 `CamelCase`，非导出用 `camelCase`。
- 测试文件命名为 `*_test.go`，优先表驱动 + `t.Run()`。
- 用户可见文案必须走 `pkg/i18n`，中英文键保持一致。

## 个人提交流程（替代重型 PR 流程）
- 提交信息继续用 Conventional Commits：`feat(ssh): ...`、`fix(hostname): ...`。
- 每次改动尽量小而清晰，不把重构和功能混在一个提交。
- 推送前自检：`make fmt && make test`；涉及系统行为再跑 `make test-integration`。
- 若改动 CLI/安装行为，同步更新 `README.md` 或 `scripts/install.sh` 说明。

## 安全与运维红线
- 涉及 `/etc/*`、SSH、服务重启的改动，必须支持 `dry-run`。
- 写配置前先备份，写入逻辑保持幂等，避免重复条目。
- 日志中禁止输出私钥、Token、完整密钥内容。
- 默认配置：`/etc/server-toolkit/config.json`；默认日志：`/var/log/server-toolkit.log`。
