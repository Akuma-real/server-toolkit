package system

import (
	"fmt"
	"os/exec"
)

// ServiceManager 服务管理器
type ServiceManager struct{}

// NewServiceManager 创建服务管理器
func NewServiceManager() *ServiceManager {
	return &ServiceManager{}
}

// EnableAndStart 启用并启动服务
func (m *ServiceManager) EnableAndStart(serviceName string) error {
	// 检查 systemctl 是否存在
	if _, err := exec.LookPath("systemctl"); err == nil {
		// 启用服务
		cmd := exec.Command("systemctl", "enable", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to enable service %s: %w", serviceName, err)
		}

		// 启动服务
		cmd = exec.Command("systemctl", "start", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", serviceName, err)
		}

		return nil
	}

	// 尝试 rc-service（OpenRC）
	if _, err := exec.LookPath("rc-service"); err == nil {
		cmd := exec.Command("rc-service", serviceName, "start")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start service %s: %w", serviceName, err)
		}
		return nil
	}

	return fmt.Errorf("no service manager found")
}

// Restart 重启服务
func (m *ServiceManager) Restart(serviceName string) error {
	// 检查 systemctl 是否存在
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "restart", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart service %s: %w", serviceName, err)
		}
		return nil
	}

	// 尝试 rc-service
	if _, err := exec.LookPath("rc-service"); err == nil {
		cmd := exec.Command("rc-service", serviceName, "restart")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to restart service %s: %w", serviceName, err)
		}
		return nil
	}

	return fmt.Errorf("no service manager found")
}

// Reload 重载服务
func (m *ServiceManager) Reload(serviceName string) error {
	// 检查 systemctl 是否存在
	if _, err := exec.LookPath("systemctl"); err == nil {
		// 尝试 reload
		cmd := exec.Command("systemctl", "reload", serviceName)
		if err := cmd.Run(); err != nil {
			// reload 失败，尝试 restart
			return m.Restart(serviceName)
		}
		return nil
	}

	// 尝试 rc-service
	if _, err := exec.LookPath("rc-service"); err == nil {
		cmd := exec.Command("rc-service", serviceName, "restart")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to reload service %s: %w", serviceName, err)
		}
		return nil
	}

	return fmt.Errorf("no service manager found")
}

// Stop 停止服务
func (m *ServiceManager) Stop(serviceName string) error {
	// 检查 systemctl 是否存在
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "stop", serviceName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
		}
		return nil
	}

	// 尝试 rc-service
	if _, err := exec.LookPath("rc-service"); err == nil {
		cmd := exec.Command("rc-service", serviceName, "stop")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
		}
		return nil
	}

	return fmt.Errorf("no service manager found")
}

// IsActive 检查服务是否激活
func (m *ServiceManager) IsActive(serviceName string) (bool, error) {
	// 检查 systemctl 是否存在
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "is-active", serviceName)
		err := cmd.Run()
		return err == nil, nil
	}

	// 尝试 rc-service
	if _, err := exec.LookPath("rc-service"); err == nil {
		// rc-service 没有直接的 is-active 命令
		// 需要使用其他方法
		return true, nil
	}

	return false, fmt.Errorf("no service manager found")
}

// IsEnabled 检查服务是否启用
func (m *ServiceManager) IsEnabled(serviceName string) (bool, error) {
	// 检查 systemctl 是否存在
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "is-enabled", serviceName)
		err := cmd.Run()
		return err == nil, nil
	}

	return false, fmt.Errorf("no service manager found")
}

// GetServiceStatus 获取服务状态
func (m *ServiceManager) GetServiceStatus(serviceName string) (string, error) {
	// 检查 systemctl 是否存在
	if _, err := exec.LookPath("systemctl"); err == nil {
		cmd := exec.Command("systemctl", "status", serviceName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to get service status: %w", err)
		}
		return string(output), nil
	}

	return "", fmt.Errorf("no service manager found")
}

// 便捷函数
func EnableAndStart(serviceName string) error {
	m := NewServiceManager()
	return m.EnableAndStart(serviceName)
}

func Restart(serviceName string) error {
	m := NewServiceManager()
	return m.Restart(serviceName)
}

func Reload(serviceName string) error {
	m := NewServiceManager()
	return m.Reload(serviceName)
}

func Stop(serviceName string) error {
	m := NewServiceManager()
	return m.Stop(serviceName)
}
