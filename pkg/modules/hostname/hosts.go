package hostname

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/system"
)

const (
	hostsFile = "/etc/hosts"
)

// UpdateMode 更新模式
type UpdateMode int

const (
	Replace127   UpdateMode = iota // 替换 127.0.1.1 行
	ReplaceToken                   // 替换旧 hostname token
	InsertAfter                    // 插入到 127.0.0.1 后
)

// UpdateHosts 更新 /etc/hosts
func UpdateHosts(oldName, newName, fqdn string, mode UpdateMode, dryRun bool, logger *internal.Logger) error {
	drm := internal.NewDryRunManager(dryRun, logger)

	// 读取 /etc/hosts
	file, err := os.Open(hostsFile)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", hostsFile, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// 备份文件
	if !dryRun {
		backupPath, err := system.BackupFile(hostsFile)
		if err != nil {
			return fmt.Errorf("failed to backup %s: %w", hostsFile, err)
		}
		if backupPath != "" {
			logger.Info("Backed up: %s -> %s", hostsFile, backupPath)
		}
	}

	// 构建新行
	newLine := "127.0.1.1"
	if fqdn != "" {
		newLine += " " + fqdn
	}
	newLine += " " + newName

	// 根据模式更新
	var newLines []string
	updated := false

	if mode == Replace127 {
		// 替换 127.0.1.1 行
		for _, line := range lines {
			if strings.HasPrefix(line, "127.0.1.1") {
				newLines = append(newLines, newLine)
				updated = true
				logger.Info("Updated %s: replaced 127.0.1.1 line", hostsFile)
			} else {
				newLines = append(newLines, line)
			}
		}
	} else if mode == ReplaceToken && oldName != "" {
		// 替换旧 hostname token
		oldLower := strings.ToLower(oldName)
		newLower := strings.ToLower(newName)
		for _, line := range lines {
			fields := strings.Fields(line)
			found := false
			for i, field := range fields {
				if strings.EqualFold(field, oldLower) {
					fields[i] = newLower
					found = true
					break
				}
			}
			if found {
				newLines = append(newLines, strings.Join(fields, " "))
				updated = true
				logger.Info("Updated %s: replaced hostname token", hostsFile)
			} else {
				newLines = append(newLines, line)
			}
		}
	}

	// 如果没有更新，插入新行
	if !updated {
		for i, line := range lines {
			if strings.HasPrefix(line, "127.0.0.1") {
				newLines = append(newLines[:i+1], append([]string{newLine}, newLines[i+1:]...)...)
				updated = true
				logger.Info("Updated %s: inserted after 127.0.0.1", hostsFile)
				break
			}
		}
	}

	// 如果还是没有更新，追加到末尾
	if !updated {
		newLines = append(newLines, newLine)
		logger.Info("Updated %s: appended to end", hostsFile)
	}

	// 写回文件
	data := []byte(strings.Join(newLines, "\n") + "\n")
	if dryRun {
		drm.LogFileWrite(hostsFile, string(data))
		return nil
	}

	if err := system.SafeWrite(hostsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", hostsFile, err)
	}

	logger.Info("Written to %s", hostsFile)
	return nil
}

// GetHostsEntries 解析 /etc/hosts
func GetHostsEntries() ([]string, error) {
	file, err := os.Open(hostsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", hostsFile, err)
	}
	defer file.Close()

	var entries []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			entries = append(entries, line)
		}
	}

	return entries, nil
}

// FindHostnameEntry 在 /etc/hosts 中查找主机名
func FindHostnameEntry(hostname string) (string, bool) {
	entries, err := GetHostsEntries()
	if err != nil {
		return "", false
	}

	hostnameLower := strings.ToLower(hostname)
	for _, entry := range entries {
		fields := strings.Fields(entry)
		if len(fields) > 1 {
			for _, field := range fields[1:] {
				if strings.EqualFold(field, hostnameLower) {
					return entry, true
				}
			}
		}
	}

	return "", false
}
