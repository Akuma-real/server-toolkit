package ssh

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Akuma-real/server-toolkit/internal"
)

// AuthKeysManager authorized_keys 管理器
type AuthKeysManager struct {
	path   string
	dryRun bool
	logger *internal.Logger
	drm    *internal.DryRunManager
}

// NewAuthKeysManager 创建 authorized_keys 管理器
func NewAuthKeysManager(path string, dryRun bool, logger *internal.Logger) *AuthKeysManager {
	return &AuthKeysManager{
		path:   path,
		dryRun: dryRun,
		logger: logger,
		drm:    internal.NewDryRunManager(dryRun, logger),
	}
}

// Append 追加密钥，返回新增数量
func (m *AuthKeysManager) Append(keys []string, overwrite bool) (int, error) {
	// 备份文件
	if _, err := os.Stat(m.path); err == nil {
		if !m.dryRun {
			backupPath, err := BackupAuthKeysFile(m.path)
			if err != nil {
				m.logger.Warn("Failed to backup authorized_keys: %v", err)
			} else if backupPath != "" {
				m.logger.Info("Backed up authorized_keys: %s", backupPath)
			}
		} else {
			m.drm.LogFileOperation("Backup file", m.path)
		}
	}

	// 如果 overwrite，先清空文件
	if overwrite {
		if m.dryRun {
			m.drm.LogOperation("Would clear authorized_keys")
		} else {
			if err := os.WriteFile(m.path, []byte{}, 0600); err != nil {
				return 0, fmt.Errorf("failed to clear authorized_keys: %w", err)
			}
		}
	}

	// 读取现有密钥
	existingKeys := make(map[string]bool)
	if !overwrite {
		if _, err := os.Stat(m.path); err == nil {
			file, err := os.Open(m.path)
			if err == nil {
				defer file.Close()
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					key := strings.TrimSpace(scanner.Text())
					if key != "" {
						existingKeys[key] = true
					}
				}
			}
		}
	}

	// 追加新密钥
	added := 0
	var newKeys []string

	for _, key := range keys {
		if !existingKeys[key] {
			newKeys = append(newKeys, key)
			added++
		}
	}

	// 如果没有新密钥需要添加
	if added == 0 {
		m.logger.Info("All keys already exist in authorized_keys")
		return 0, nil
	}

	// 写入新密钥
	if m.dryRun {
		m.drm.LogOperation("Would append %d keys to authorized_keys", added)
		return added, nil
	}

	file, err := os.OpenFile(m.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return 0, fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	for _, key := range newKeys {
		if _, err := fmt.Fprintln(file, key); err != nil {
			m.logger.Warn("Failed to write key: %v", err)
			added--
		}
	}

	m.logger.Info("Added %d keys to authorized_keys", added)
	return added, nil
}

// List 列出所有密钥
func (m *AuthKeysManager) List() ([]string, error) {
	file, err := os.Open(m.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	var keys []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		if key != "" {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Count 统计密钥数量
func (m *AuthKeysManager) Count() (int, error) {
	keys, err := m.List()
	if err != nil {
		return 0, err
	}
	return len(keys), nil
}

// Clear 清空所有密钥
func (m *AuthKeysManager) Clear() error {
	if m.dryRun {
		m.drm.LogFileOperation("Clear file", m.path)
		return nil
	}

	return os.WriteFile(m.path, []byte{}, 0600)
}

// RemoveKey 移除指定密钥
func (m *AuthKeysManager) RemoveKey(key string) error {
	keys, err := m.List()
	if err != nil {
		return err
	}

	var newKeys []string
	for _, k := range keys {
		if k != key {
			newKeys = append(newKeys, k)
		}
	}

	if m.dryRun {
		m.drm.LogOperation("Would remove key from authorized_keys")
		return nil
	}

	content := strings.Join(newKeys, "\n")
	return os.WriteFile(m.path, []byte(content), 0600)
}

// BackupAuthKeysFile 备份 authorized_keys 文件
func BackupAuthKeysFile(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// 生成备份文件名
	timestamp := fmt.Sprintf("%d", os.Getegid()) // 简化的时间戳
	backupPath := path + ".bak." + timestamp

	// 写入备份
	if err := os.WriteFile(backupPath, file, 0600); err != nil {
		return "", err
	}

	return backupPath, nil
}
