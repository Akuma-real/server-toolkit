package hostname

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Akuma-real/server-toolkit/internal"
)

const (
	cloudInitDir        = "/etc/cloud"
	cloudInitCfgDir     = "/etc/cloud/cloud.cfg.d"
	preserveHostnameCfg = "99-hostname-preserve.cfg"
	hostsTemplateDir    = "/etc/cloud/templates"
)

// IsPresent 检测 cloud-init 是否存在
func IsPresent() bool {
	// 检查 /etc/cloud 目录
	if _, err := os.Stat(cloudInitDir); err != nil {
		return false
	}

	// 检查 cloud-init 命令
	if _, err := exec.LookPath("cloud-init"); err != nil {
		return false
	}

	return true
}

// SetPreserveHostname 设置 cloud-init preserve_hostname
func SetPreserveHostname(dryRun bool, logger *internal.Logger) error {
	drm := internal.NewDryRunManager(dryRun, logger)

	if !IsPresent() {
		logger.Info("cloud-init not found, skipping")
		return nil
	}

	cfgPath := filepath.Join(cloudInitCfgDir, preserveHostnameCfg)

	// 检查文件是否已存在且配置正确
	if _, err := os.Stat(cfgPath); err == nil {
		file, err := os.Open(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", cfgPath, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "preserve_hostname:") &&
				strings.Contains(line, "true") {
				logger.Info("preserve_hostname already set in %s", cfgPath)
				return nil
			}
		}
	}

	// 备份文件
	if !dryRun {
		if _, err := os.Stat(cfgPath); err == nil {
			backupPath, err := BackupCloudInitFile(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to backup %s: %w", cfgPath, err)
			}
			if backupPath != "" {
				logger.Info("Backed up: %s -> %s", cfgPath, backupPath)
			}
		}
	}

	// 写入配置
	content := fmt.Sprintf("# written by server-toolkit: prevent cloud-init from overriding hostname on reboot\npreserve_hostname: true\n")

	if dryRun {
		drm.LogFileWrite(cfgPath, content)
		return nil
	}

	if err := os.MkdirAll(cloudInitCfgDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", cloudInitCfgDir, err)
	}

	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", cfgPath, err)
	}

	logger.Info("Written preserve_hostname: true to %s", cfgPath)
	return nil
}

// PatchHostsTemplates 修补 cloud-init hosts 模板
func PatchHostsTemplates(shortName, fqdn string, dryRun bool, logger *internal.Logger) error {
	drm := internal.NewDryRunManager(dryRun, logger)

	if !IsPresent() {
		logger.Info("cloud-init not found, skipping")
		return nil
	}

	// 检查模板目录
	if _, err := os.Stat(hostsTemplateDir); err != nil {
		if os.IsNotExist(err) {
			logger.Info("Cloud-init templates directory not found: %s", hostsTemplateDir)
			return nil
		}
		return fmt.Errorf("failed to stat %s: %w", hostsTemplateDir, err)
	}

	// 构建新行
	newLine := "127.0.1.1"
	if fqdn != "" {
		newLine += " " + fqdn
	}
	newLine += " " + shortName

	// 遍历模板文件
	pattern := filepath.Join(hostsTemplateDir, "hosts.*.tmpl")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob templates: %w", err)
	}

	if len(matches) == 0 {
		logger.Info("No cloud-init hosts templates found")
		return nil
	}

	patched := 0
	for _, templateFile := range matches {
		// 检查模板是否包含 127.0.1.1
		has127, err := checkTemplateFor127(templateFile)
		if err != nil {
			logger.Warn("Failed to check template %s: %v", templateFile, err)
			continue
		}

		if !has127 {
			continue
		}

		// 备份模板
		if !dryRun {
			backupPath, err := BackupCloudInitFile(templateFile)
			if err != nil {
				logger.Warn("Failed to backup %s: %v", templateFile, err)
				continue
			}
			if backupPath != "" {
				logger.Info("Backed up: %s -> %s", templateFile, backupPath)
			}
		}

		// 更新模板
		if err := patchTemplate(templateFile, newLine, dryRun, drm); err != nil {
			logger.Warn("Failed to patch template %s: %v", templateFile, err)
			continue
		}

		logger.Info("Patched template: %s", templateFile)
		patched++
	}

	if patched > 0 {
		logger.Info("Patched %d cloud-init hosts templates", patched)
	} else {
		logger.Info("No cloud-init hosts templates needed patching")
	}

	return nil
}

// checkTemplateFor127 检查模板是否包含 127.0.1.1
func checkTemplateFor127(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "127.0.1.1") {
			return true, nil
		}
	}

	return false, nil
}

// patchTemplate 修补模板
func patchTemplate(path, newLine string, dryRun bool, drm *internal.DryRunManager) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	updated := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "127.0.1.1") && !updated {
			lines = append(lines, newLine)
			updated = true
		} else {
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")

	if dryRun {
		drm.LogFileWrite(path, content)
		return nil
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// BackupCloudInitFile 备份 cloud-init 文件
func BackupCloudInitFile(path string) (string, error) {
	// 复制文件内容
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// 生成备份文件名
	timestamp := fmt.Sprintf("%d", os.Getegid()) // 简化的时间戳
	backupPath := path + ".bak." + timestamp

	// 写入备份
	if err := os.WriteFile(backupPath, file, 0644); err != nil {
		return "", err
	}

	return backupPath, nil
}
