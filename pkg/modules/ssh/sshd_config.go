package ssh

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/system"
)

const (
	sshdConfigPath = "/etc/ssh/sshd_config"
)

// Config SSH 配置
type Config struct {
	path   string
	logger *internal.Logger
	drm    *internal.DryRunManager
}

// NewConfig 创建 SSH 配置
func NewConfig(path string, logger *internal.Logger) (*Config, error) {
	return &Config{
		path:   path,
		logger: logger,
		drm:    internal.NewDryRunManager(false, logger),
	}, nil
}

// SetGlobalOption 设置全局选项（仅在 Match 之前）
func (c *Config) SetGlobalOption(key, value string) error {
	if _, err := os.Stat(c.path); err != nil {
		return fmt.Errorf("sshd_config not found: %w", err)
	}

	// 读取文件
	file, err := os.Open(c.path)
	if err != nil {
		return fmt.Errorf("failed to open sshd_config: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	inMatch := false
	found := false
	matchLineRegex := regexp.MustCompile(`^\s*Match\s+`)

	for scanner.Scan() {
		line := scanner.Text()

		// 检查是否进入 Match 块
		if matchLineRegex.MatchString(line) {
			inMatch = true
		}

		// 如果在 Match 块中，不做修改
		if inMatch {
			lines = append(lines, line)
			continue
		}

		// 检查是否为目标选项
		fields := strings.Fields(line)
		if len(fields) > 0 && strings.EqualFold(fields[0], key) {
			// 替换选项值
			lines = append(lines, fmt.Sprintf("%s %s", key, value))
			found = true
			c.logger.Debug("Updated %s: %s -> %s", key, strings.Join(fields[1:], " "), value)
		} else {
			lines = append(lines, line)
		}
	}

	// 如果没有找到选项，添加到文件开头
	if !found {
		lines = append([]string{fmt.Sprintf("%s %s", key, value)}, lines...)
		c.logger.Debug("Added %s: %s", key, value)
	}

	// 写回文件
	content := strings.Join(lines, "\n") + "\n"
	if err := system.SafeWrite(c.path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write sshd_config: %w", err)
	}

	c.logger.Info("Set sshd_config option: %s = %s", key, value)
	return nil
}

// DisablePasswordAuth 禁用密码认证
func (c *Config) DisablePasswordAuth() error {
	// 设置公钥认证
	if err := c.SetGlobalOption("PubkeyAuthentication", "yes"); err != nil {
		return err
	}

	// 禁用密码认证
	if err := c.SetGlobalOption("PasswordAuthentication", "no"); err != nil {
		return err
	}

	// 禁用键盘交互认证
	if err := c.SetGlobalOption("KbdInteractiveAuthentication", "no"); err != nil {
		return err
	}

	// 禁用挑战响应认证
	if err := c.SetGlobalOption("ChallengeResponseAuthentication", "no"); err != nil {
		return err
	}

	c.logger.Info("Disabled password authentication")
	return nil
}

// Reload 重载 sshd
func Reload() error {
	svc := system.NewServiceManager()
	return svc.Reload("sshd")
}

// Restart 重启 sshd
func Restart() error {
	svc := system.NewServiceManager()
	return svc.Restart("sshd")
}

// EnableService 启用并启动 SSH 服务
func EnableService() error {
	svc := system.NewServiceManager()
	return svc.EnableAndStart("sshd")
}

// EnsureService 确保 SSH 服务运行
func EnsureService(logger *internal.Logger) error {
	// 检查 openssh-server 是否安装
	if !isSSHDInstalled() {
		logger.Info("openssh-server not installed")
		// TODO: 提供安装选项
		return fmt.Errorf("openssh-server not installed")
	}

	// 检查服务是否运行
	svc := system.NewServiceManager()
	active, err := svc.IsActive("sshd")
	if err != nil {
		return err
	}

	if !active {
		logger.Info("Starting SSH service...")
		return svc.EnableAndStart("sshd")
	}

	logger.Info("SSH service is already running")
	return nil
}

// isSSHDInstalled 检查 sshd 是否安装
func isSSHDInstalled() bool {
	paths := []string{
		"/usr/sbin/sshd",
		"/sbin/sshd",
		"/usr/bin/sshd",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}
