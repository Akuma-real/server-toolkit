package internal

import ()

// DryRunManager Dry-run 模式管理器
type DryRunManager struct {
	enabled bool
	logger  *Logger
}

// NewDryRunManager 创建新 Dry-run 管理器
func NewDryRunManager(enabled bool, logger *Logger) *DryRunManager {
	return &DryRunManager{
		enabled: enabled,
		logger:  logger,
	}
}

// IsEnabled 返回是否启用 Dry-run 模式
func (m *DryRunManager) IsEnabled() bool {
	return m.enabled
}

// SetEnabled 设置 Dry-run 模式
func (m *DryRunManager) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// LogOperation 记录操作日志
func (m *DryRunManager) LogOperation(op string, args ...interface{}) {
	if m.enabled {
		m.logger.Info("[DRY-RUN] "+op, args...)
	}
}

// LogCommand 记录命令执行
func (m *DryRunManager) LogCommand(cmd string, args ...string) {
	if m.enabled {
		cmdStr := cmd
		for _, arg := range args {
			cmdStr += " " + arg
		}
		m.logger.Info("[DRY-RUN] Would execute: %s", cmdStr)
	}
}

// LogFileWrite 记录文件写入
func (m *DryRunManager) LogFileWrite(path string, content string) {
	if m.enabled {
		m.logger.Info("[DRY-RUN] Would write to file: %s (%d bytes)", path, len(content))
	}
}

// LogFileOperation 记录文件操作
func (m *DryRunManager) LogFileOperation(op string, path string) {
	if m.enabled {
		m.logger.Info("[DRY-RUN] Would %s: %s", op, path)
	}
}

// LogServiceOperation 记录服务操作
func (m *DryRunManager) LogServiceOperation(op string, service string) {
	if m.enabled {
		m.logger.Info("[DRY-RUN] Would %s service: %s", op, service)
	}
}

// WrapCommand 包装命令执行
func (m *DryRunManager) WrapCommand(fn func() error, description string) error {
	if m.enabled {
		m.LogOperation(description)
		return nil
	}
	return fn()
}
