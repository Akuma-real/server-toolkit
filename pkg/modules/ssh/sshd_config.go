package ssh

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
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
	dryRun bool
	logger *internal.Logger
	drm    *internal.DryRunManager
}

// NewConfig 创建 SSH 配置
func NewConfig(path string, dryRun bool, logger *internal.Logger) (*Config, error) {
	return &Config{
		path:   path,
		dryRun: dryRun,
		logger: logger,
		drm:    internal.NewDryRunManager(dryRun, logger),
	}, nil
}

// SetGlobalOption 设置全局选项（仅在 Match 之前）
func (c *Config) SetGlobalOption(key, value string) error {
	return c.SetGlobalOptions(map[string]string{key: value})
}

// SetGlobalOptions 设置多个全局选项（仅在 Match 之前），尽量单次读写
func (c *Config) SetGlobalOptions(options map[string]string) error {
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
	found := make(map[string]bool)
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
		if len(fields) > 0 {
			updated := false
			for k, v := range options {
				if strings.EqualFold(fields[0], k) {
					lines = append(lines, fmt.Sprintf("%s %s", k, v))
					found[strings.ToLower(k)] = true
					updated = true
					c.logger.Debug("Updated %s: %s -> %s", k, strings.Join(fields[1:], " "), v)
					break
				}
			}
			if updated {
				continue
			}
		} else {
			lines = append(lines, line)
			continue
		}

		lines = append(lines, line)
	}

	// 如果没有找到选项，添加到文件开头
	var toPrepend []string
	orderedKeys := []string{
		"PubkeyAuthentication",
		"PasswordAuthentication",
		"KbdInteractiveAuthentication",
		"ChallengeResponseAuthentication",
	}
	// 先按常用顺序，避免每次写入顺序飘移
	for _, k := range orderedKeys {
		v, ok := options[k]
		if !ok {
			continue
		}
		if !found[strings.ToLower(k)] {
			toPrepend = append(toPrepend, fmt.Sprintf("%s %s", k, v))
			c.logger.Debug("Added %s: %s", k, v)
		}
	}
	// 再补充其他未排序 key（稳定排序）
	var rest []string
	for k := range options {
		lk := strings.ToLower(k)
		if found[lk] {
			continue
		}
		seen := false
		for _, okk := range orderedKeys {
			if strings.EqualFold(okk, k) {
				seen = true
				break
			}
		}
		if seen {
			continue
		}
		rest = append(rest, k)
	}
	if len(rest) > 0 {
		sort.Strings(rest)
		for _, k := range rest {
			toPrepend = append(toPrepend, fmt.Sprintf("%s %s", k, options[k]))
			c.logger.Debug("Added %s: %s", k, options[k])
		}
	}
	if len(toPrepend) > 0 {
		lines = append(toPrepend, lines...)
	}

	// 写回文件
	content := strings.Join(lines, "\n") + "\n"
	if c.dryRun {
		c.drm.LogFileWrite(c.path, content)
		return nil
	}

	backupPath, err := system.BackupFile(c.path)
	if err != nil {
		return fmt.Errorf("failed to backup sshd_config: %w", err)
	}
	if backupPath != "" {
		c.logger.Info("Backed up: %s -> %s", c.path, backupPath)
	}

	perm := os.FileMode(0644)
	if info, err := os.Stat(c.path); err == nil {
		perm = info.Mode()
	}

	if err := system.SafeWrite(c.path, []byte(content), perm); err != nil {
		return fmt.Errorf("failed to write sshd_config: %w", err)
	}
	_ = system.RestoreSELinuxContext(c.path)

	if err := validateSSHDConfig(c.path); err != nil {
		if backupPath != "" {
			// 尝试回滚
			if restoreErr := restoreFromBackup(c.path, backupPath, perm); restoreErr != nil {
				return fmt.Errorf("sshd_config validation failed: %v (restore failed: %v)", err, restoreErr)
			}
		}
		return fmt.Errorf("sshd_config validation failed: %w", err)
	}

	for k, v := range options {
		c.logger.Info("Set sshd_config option: %s = %s", k, v)
	}
	return nil
}

// DisablePasswordAuth 禁用密码认证
func (c *Config) DisablePasswordAuth() error {
	if err := c.SetGlobalOptions(map[string]string{
		"PubkeyAuthentication":            "yes",
		"PasswordAuthentication":          "no",
		"KbdInteractiveAuthentication":    "no",
		"ChallengeResponseAuthentication": "no",
	}); err != nil {
		return err
	}

	c.logger.Info("Disabled password authentication")
	return nil
}

// Reload 重载 sshd
func Reload() error {
	return ReloadSSHD(false, nil)
}

// Restart 重启 sshd
func Restart() error {
	return restartSSHD(false, nil)
}

// ReloadSSHD 重载 SSH 服务（兼容 service name: sshd/ssh）
func ReloadSSHD(dryRun bool, logger *internal.Logger) error {
	if dryRun {
		if logger == nil {
			return nil
		}
		drm := internal.NewDryRunManager(dryRun, logger)
		drm.LogServiceOperation("reload", "sshd")
		drm.LogServiceOperation("reload", "ssh")
		return nil
	}

	svc := system.NewServiceManager()
	if err := svc.Reload("sshd"); err == nil {
		return nil
	}
	if err := svc.Reload("ssh"); err == nil {
		return nil
	}
	return fmt.Errorf("failed to reload ssh service (tried sshd, ssh)")
}

func restartSSHD(dryRun bool, logger *internal.Logger) error {
	if dryRun {
		if logger == nil {
			return nil
		}
		drm := internal.NewDryRunManager(dryRun, logger)
		drm.LogServiceOperation("restart", "sshd")
		drm.LogServiceOperation("restart", "ssh")
		return nil
	}

	svc := system.NewServiceManager()
	if err := svc.Restart("sshd"); err == nil {
		return nil
	}
	if err := svc.Restart("ssh"); err == nil {
		return nil
	}
	return fmt.Errorf("failed to restart ssh service (tried sshd, ssh)")
}

func validateSSHDConfig(path string) error {
	sshdPath, err := exec.LookPath("sshd")
	if err != nil {
		// 无 sshd 可执行文件：无法做语法校验，跳过
		return nil
	}

	cmd := exec.Command(sshdPath, "-t", "-f", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func restoreFromBackup(dstPath, backupPath string, perm os.FileMode) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}
	return system.SafeWrite(dstPath, data, perm)
}
