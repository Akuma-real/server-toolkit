# AGENTS.md

本文件用于指导**代码贡献者**与**自动化编程代理（AI Agent）**在本仓库内进行一致、安全、可回滚的开发与变更。

---

## 1. 仓库贡献指南（Repository Guidelines）

### 1.1 项目定位与风险边界（务必阅读）

本项目属于 Linux 服务器运维工具箱，部分模块会修改系统关键配置与服务（例如 `/etc/hosts`、SSH 配置、systemd 服务、用户权限相关文件）。

**强约束：**
- 默认按“可能破坏系统连接/登录”的风险级别对待所有改动。
- 任何涉及 `/etc/*`、SSH、用户/权限、服务启停的变更：
  - 必须具备 **dry-run 不落盘** 的行为路径
  - 必须做到 **幂等（重复执行不产生额外副作用）**
  - 必须有 **清晰的回滚/恢复策略**（至少备份 + 原子替换）
- 优先在 VM/容器中验证，避免在真实生产机器上直接测试。

---

## 2. 项目结构与模块组织

- `internal/`：内部实现（配置、日志、更新检查等），不对外暴露。
- `pkg/`：可复用包
  - `pkg/modules/hostname/`、`pkg/modules/ssh/`：运维模块（可能修改系统配置）。
  - `pkg/system/`：OS/发行版识别与文件/包管理器/服务/用户等封装。
  - `pkg/tui/`：Bubble Tea TUI 组件、消息与样式。
  - `pkg/i18n/`：多语言表（`zh_CN`、`en_US`）与查询方法。
- `scripts/`：构建与安装脚本。
- `bin/`：构建产物（不提交）。

**结构性原则：**
- `pkg/modules/*`：只关心“业务动作”（例如：禁用密码登录），不直接散落系统调用细节。
- `pkg/system/*`：封装 OS 差异与具体系统操作（文件、服务、包管理器、用户等），供 modules 调用。
- `internal/*`：配置、日志、更新检查等“应用级粘合层”。

---

## 3. 构建、测试与开发命令

- `make build`：构建 `bin/server-toolkit`（Makefile 目标假定入口在 `./cmd/server-toolkit`）。
- `make dev`：本地开发运行（`go run`）。
- `make test`：单元测试（带 `-race`）。
- `make test-integration`：在 Debian/AlmaLinux 的 Docker 容器内跑 `go test`。
- `make fmt`：`gofmt` + `goimports` 格式化（导入按：标准库/第三方/本地 分组）。
- `make lint`：运行 `golangci-lint`（本机未安装需先安装）。
- `make deps`：`go mod tidy` + 依赖下载。

**AI Agent 执行顺序建议：**
1) 修改前先通读相关模块目录与 `pkg/system` 的封装能力  
2) 优先补/改单测  
3) `make fmt && make test`  
4) 涉及系统差异时再跑 `make test-integration`

---

## 4. 代码风格与命名约定

- 以 Go 工具为准：缩进/空格/导入顺序交给 `gofmt`/`goimports`。
- 包名短且小写（如 `system`、`ssh`）；导出标识符用 `CamelCase`。
- 文件名小写；测试文件用 `*_test.go`（如 `pkg/system/os_test.go`）。

**错误处理与日志：**
- 优先返回可定位的错误（包含上下文），避免吞错。
- 不在库代码里 `panic`（除非不可恢复且明确）。
- 禁止在库层随意 `fmt.Println`：统一走日志/调用层展示。

---

## 5. 系统变更规范（高风险区域）

### 5.1 Dry-run 语义（强制一致）

- dry-run 下：
  - 不写入文件、不更改权限、不启停服务、不执行破坏性命令
  - 允许做“读取/探测/校验”（例如读取现有配置、检查发行版、验证目标路径是否存在）
  - 需要输出“将要执行的动作摘要”（用于 TUI/日志展示）

### 5.2 文件写入（特别是 `/etc/*`）

建议遵循：
- **先备份**：例如 `xxx.bak` 或带时间戳备份（避免覆盖历史）
- **原子替换**：写到临时文件 -> `fsync` -> `rename`
- **保留权限/属主**：写回后权限、属主与原文件保持一致（或按安全要求更严格）
- **幂等**：重复执行不得产生重复条目/重复配置块

### 5.3 SSH/远程连接安全（避免把用户锁死）

- 修改 SSH 相关配置前，尽量进行语法/可用性校验（例如验证配置格式、关键字段）。
- 禁用密码登录、改端口、改 `PermitRootLogin` 等动作：
  - 必须在 PR 描述中列出风险与验证方式
  - 建议提供“回滚/恢复步骤”（例如恢复备份文件并重启服务）
- 写入 `authorized_keys` 时：
  - 权限应收紧（通常 `~/.ssh` 为 700、`authorized_keys` 为 600）
  - 不写入私钥、不记录密钥全文到日志

### 5.4 服务管理

- 通过 `pkg/system` 的服务封装实现启停/重载，避免散落 `systemctl` 调用。
- 对于重启类操作：优先 reload；必须 restart 时写明原因。
- dry-run 下仅输出计划动作。

---

## 6. TUI 开发约定（Bubble Tea）

- Model 的 `Update` 中不要直接做阻塞/重 IO 的副作用：
  - 使用 `tea.Cmd` 异步执行，再通过消息回传结果
- UI 文案禁止硬编码中文或英文：
  - 一律走 `pkg/i18n`（见下一节）
- 样式统一放 `pkg/tui`，避免各处自建一套风格。

---

## 7. 国际化（i18n）约定

- 新增/修改任何用户可见字符串时：
  - 必须同时更新 `zh_CN` 与 `en_US`
  - key 保持稳定、语义明确（避免 `text1/text2`）
- 不允许在 modules/system 层直接拼接 UI 文案；传递结构化信息到 TUI 层再决定展示。

---

## 8. 测试指南

- 优先表驱动测试，并使用 `testify/assert` 断言。
- 测试与代码放在同一包内，命名 `TestXxx`，用 `t.Run()` 组织子用例。
- 覆盖率：`go test -coverprofile=coverage.out ./...` 然后 `go tool cover -html=coverage.out`。

**建议的可测性实践：**
- 涉及系统操作的逻辑优先通过接口/封装注入（便于 mock/fake）。
- 避免单测依赖 root 权限、真实系统服务或真实网络。
- 集成测试才验证跨发行版/容器行为。

---

## 9. 提交与 Pull Request 规范

### 9.1 约定式提交（Conventional Commits）

- 现有提交已使用 `docs:` 前缀；建议统一采用 Conventional Commits（详见：<https://www.conventionalcommits.org/>）。
- 基本格式：
  - `<type>[optional scope][!]: <description>`
- `type`（推荐）：
  - `feat` / `fix` / `docs` / `refactor` / `test` / `chore` / `ci` / `build` / `perf` / `style` / `revert`
- `scope`（可选）：用小写名词描述模块范围，例如 `ssh`、`hostname`、`system`、`tui`、`i18n`。
- `description`：一句话概括改动，尽量简短，使用祈使语气（不需要句号）。
- 破坏性变更（BREAKING CHANGE）：
  - 在标题中使用 `!`：`feat(api)!: ...`
  - 或在脚注中使用：`BREAKING CHANGE: ...`
- 示例：
  - `docs: update install instructions`
  - `fix(ssh): avoid duplicating authorized_keys entries`
  - `feat(tui)!: redesign main menu`

> 说明：若仓库采用 squash merge，请确保 PR 标题同样符合以上格式，便于生成清晰历史与自动化发布/Changelog。

### 9.2 Pull Request 说明要求

- PR 至少包含：
  1) 做了什么 / 为什么  
  2) 风险点（尤其是 `/etc/*` 或 SSH 相关变更）  
  3) 如何验证（命令、环境、是否 dry-run）  
  4) 必要时更新 `CHANGELOG.md` 的 `[Unreleased]`

---

## 10. 安全与配置提示

- 本仓库代码可能修改 `/etc/hosts`、SSH 配置与系统服务：
  - 优先在 VM/容器里开发验证
  - 确保 dry-run 行为不做真实变更
- 默认配置：`/etc/server-toolkit/config.json`
- 默认日志：`/var/log/server-toolkit.log`

**敏感信息：**
- 日志中避免输出私钥、token、完整公钥内容（最多展示指纹/前后截断）。
- 任何下载/导入外部内容（例如从 URL 拉取密钥）都应进行最小信任处理：
  - 校验格式、限制大小、失败时给出清晰错误

---

## 11. AI Agent 额外行为准则（面向自动化编程代理）

- 修改范围最小化：避免“顺手重构”造成无关 diff。
- 每次任务完成前必须：
  - `make fmt`
  - `make test`
  - 若涉及 OS/发行版差异或系统操作：补充/更新集成测试或至少说明为何无需更新
- 不新增“隐式行为”：
  - 新开关/默认值变化必须写入配置说明或变更日志
- 对高风险变更（SSH、用户权限、系统服务）：
  - 代码中写清楚保护性检查（guard）
  - PR 描述里写明验证与回滚步骤
