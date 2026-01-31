package hostname

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/system"
)

const (
	hostnameFile = "/etc/hostname"
)

// Manager 主机名管理器
type Manager struct {
	dryRun bool
	logger *internal.Logger
	drm    *internal.DryRunManager
}

// NewManager 创建主机名管理器
func NewManager(dryRun bool, logger *internal.Logger) *Manager {
	return &Manager{
		dryRun: dryRun,
		logger: logger,
		drm:    internal.NewDryRunManager(dryRun, logger),
	}
}

// SetHostname 设置主机名
func (m *Manager) SetHostname(short, fqdn string) error {
	// 设置主机名
	if err := m.setHostname(short); err != nil {
		return err
	}

	// 写入 /etc/hostname
	if err := m.writeHostnameFile(short); err != nil {
		return err
	}

	m.logger.Info("Hostname set to: %s", short)
	if fqdn != "" {
		m.logger.Info("FQDN set to: %s", fqdn)
	}

	return nil
}

// setHostname 设置主机名（使用 hostnamectl 或 hostname）
func (m *Manager) setHostname(name string) error {
	if m.dryRun {
		m.drm.LogCommand("hostnamectl", "set-hostname", name)
		return nil
	}

	// 优先使用 hostnamectl
	if _, err := exec.LookPath("hostnamectl"); err == nil {
		cmd := exec.Command("hostnamectl", "set-hostname", name)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("hostnamectl failed: %w", err)
		}
		return nil
	}

	// 降级到 hostname
	if _, err := exec.LookPath("hostname"); err == nil {
		cmd := exec.Command("hostname", name)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("hostname failed: %w", err)
		}
		return nil
	}

	return fmt.Errorf("no hostname command found")
}

// writeHostnameFile 写入 /etc/hostname
func (m *Manager) writeHostnameFile(name string) error {
	// 备份文件
	if !m.dryRun {
		backupPath, err := system.BackupFile(hostnameFile)
		if err != nil {
			return fmt.Errorf("failed to backup %s: %w", hostnameFile, err)
		}
		if backupPath != "" {
			m.logger.Info("Backed up: %s -> %s", hostnameFile, backupPath)
		}
	}

	// 写入文件
	data := []byte(name + "\n")
	if m.dryRun {
		m.drm.LogFileWrite(hostnameFile, string(data))
		return nil
	}

	if err := system.SafeWrite(hostnameFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", hostnameFile, err)
	}

	m.logger.Info("Written to %s: %s", hostnameFile, name)
	return nil
}

// GetHostname 获取当前主机名
func (m *Manager) GetHostname() (string, error) {
	// 优先使用 hostnamectl
	if _, err := exec.LookPath("hostnamectl"); err == nil {
		cmd := exec.Command("hostnamectl", "--static")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}

	// 降级到 hostname
	if _, err := exec.LookPath("hostname"); err == nil {
		cmd := exec.Command("hostname")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}

	return "", fmt.Errorf("no hostname command found")
}

// ReadHostnameFile 读取 /etc/hostname
func (m *Manager) ReadHostnameFile() (string, error) {
	data, err := system.ReadFile(hostnameFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
