# Server Toolkit - 代理开发指南

本指南为在此代码库中工作的 AI 代理提供编码标准、工作流程和最佳实践。

## 构建和测试命令

### 构建和测试
- `make build` - 构建当前平台的二进制文件
- `make build-all` - 为多个平台交叉编译（linux-amd64, linux-arm64）
- `make test` - 运行所有测试（带 -race 竞态检测）
- `make test-integration` - 在 Docker 容器中运行集成测试（Debian 12 和 AlmaLinux 9）

### 单个测试运行
- `go test -v -run TestFunctionName ./path/to/package` - 运行特定测试
- `go test -v -run TestFunctionName/Subtest ./path/to/package` - 运行特定子测试
- `go test -v ./... -run TestFunctionName` - 在所有包中运行特定测试

### 代码质量
- `make fmt` - 使用 gofmt 和 goimports 格式化代码
- `make lint` - 使用 golangci-lint 运行代码检查
- `make deps` - 下载并整理依赖项

### 开发
- `make dev` - 使用 `go run` 运行应用程序
- `make run` - 构建并运行二进制文件
- `make clean` - 清理构建产物

## 代码风格指南

### 导入组织
按以下顺序组织导入：
1. 标准库
2. 第三方库
3. 本地包（github.com/Akuma-real/server-toolkit）

使用别名导入以避免冲突：`tea "github.com/charmbracelet/bubbletea"`

### 格式化
- 使用 `gofmt` 自动格式化
- 使用 `goimports` 管理导入
- 始终在提交前运行 `make fmt`

### 命名约定
- **包名**：小写，简短，描述性（如 `system`, `hostname`, `ssh`）
- **类型**：驼峰命名（如 `DistroInfo`, `SystemInfo`）
- **常量**：全大写或驼峰命名（如 `maxHostnameLength`, `DEBUG`）
- **变量**：驼峰命名（如 `userInfo`, `authKeysPath`）
- **函数**：驼峰命名，公共函数大写开头，私有函数小写开头
- **文件名**：小写，下划线分隔（如 `os_test.go`, `ssh_config.go`）

### 错误处理
- 使用 `fmt.Errorf` 和 `%w` 包装错误以保留上下文
- 返回描述性错误消息
- 使用 defer 记录资源释放错误（通常是警告）
- 示例：`return fmt.Errorf("failed to get hostname: %w", err)`

### 测试
- 使用 `testify/assert` 进行断言
- 使用表驱动测试进行多种情况
- 为每个测试用例使用描述性名称
- 使用 `t.Run()` 组织子测试

示例：
```go
func TestDetermineFamily(t *testing.T) {
    tests := []struct {
        id     string
        family DistroFamily
    }{
        {"debian", Debian},
        {"ubuntu", Debian},
    }
    for _, tt := range tests {
        t.Run(tt.id, func(t *testing.T) {
            assert.Equal(t, tt.family, determineFamily(tt.id))
        })
    }
}
```

### Bubble Tea TUI 组件
- 实现标准模型方法：`Init()`, `Update()`, `View()`
- 使用 `tea.Cmd` 返回命令
- 保持不可变模型状态
- 使用自定义消息类型进行组件间通信
- 使用 `lipgloss` 进行样式定义

### 国际化 (i18n)
- 使用 `i18n.T("key")` 获取翻译
- 在 `pkg/i18n/` 中定义翻译键
- 默认语言为简体中文（zh_CN）
- 支持英语（en_US）

### Dry-run 模式
- 检查 `dryRun` 标志并使用 `DryRunManager` 记录操作
- 不要在 dry-run 模式下实际执行操作
- 使用 `drm.LogOperation()`, `drm.LogCommand()` 等方法记录意图

### 结构体标签
- JSON 字段使用 `json` 标签：`json:"language"`
- 遵循 Go 的标准约定

### 注释
- 为所有导出的类型、函数和常量添加注释
- 对复杂的逻辑或 TODO 项目添加注释
- 注释应该是完整句子，以类型/函数名称开头

### 文件结构
- `cmd/server-toolkit/` - 主入口点
- `pkg/` - 公共库代码（按功能组织）
- `internal/` - 内部实现细节（未导出）
- `pkg/modules/` - 功能模块（hostname, ssh 等）
- `pkg/tui/` - TUI 组件和样式
- `pkg/i18n/` - 国际化

### 常量定义
- 使用包级常量定义固定值
- 将相关常量分组在一起
- 使用 iota 定义枚举

示例：
```go
const (
    maxHostnameLength = 253
    maxSegmentLength  = 63
)
```

## 重要注意事项

- 此项目使用 Go 1.25.6
- 主要依赖：Bubble Tea (TUI), Lip Gloss (样式), testify (测试)
- 默认配置位置：`/etc/server-toolkit/config.json`
- 日志位置：`/var/log/server-toolkit.log`
- 在提交前运行 `make lint` 和 `make test`
- 测试覆盖目标是 >80%

## 模块开发

创建新模块时：
1. 在 `pkg/modules/` 下创建新目录
2. 实现适当的验证器和管理器结构
3. 添加单元测试（`*_test.go`）
4. 在主菜单中集成新模块
5. 添加国际化字符串
