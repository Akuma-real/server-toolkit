package ssh

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/system"
)

// Source 密钥来源
type Source int

const (
	SourceGitHub Source = iota
	SourceURL
	SourceFile
)

// Manager SSH 密钥管理器
type Manager struct {
	user   string
	dryRun bool
	logger *internal.Logger
	drm    *internal.DryRunManager
}

// NewManager 创建 SSH 密钥管理器
func NewManager(user string, dryRun bool, logger *internal.Logger) *Manager {
	return &Manager{
		user:   user,
		dryRun: dryRun,
		logger: logger,
		drm:    internal.NewDryRunManager(dryRun, logger),
	}
}

// FetchKeys 获取公钥
func (m *Manager) FetchKeys(source Source, value string) ([]string, error) {
	var keys []string
	var err error

	switch source {
	case SourceGitHub:
		keys, err = m.fetchGitHubKeys(value)
	case SourceURL:
		keys, err = m.fetchURLKeys(value)
	case SourceFile:
		keys, err = m.readFileKeys(value)
	default:
		return nil, fmt.Errorf("unknown key source")
	}

	if err != nil {
		return nil, err
	}

	// 过滤和验证密钥
	return m.filterAndValidateKeys(keys)
}

// fetchGitHubKeys 从 GitHub 获取密钥
func (m *Manager) fetchGitHubKeys(username string) ([]string, error) {
	url := fmt.Sprintf("https://github.com/%s.keys", username)

	m.logger.Info("Fetching keys from GitHub: %s", username)

	if m.dryRun {
		m.drm.LogOperation("Would fetch keys from GitHub", username)
		return []string{fmt.Sprintf("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIexample-key-for-%s", username)}, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub returned status %d", resp.StatusCode)
	}

	// 读取所有密钥
	var keys []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		if key != "" {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// fetchURLKeys 从 URL 获取密钥
func (m *Manager) fetchURLKeys(url string) ([]string, error) {
	m.logger.Info("Fetching keys from URL: %s", url)

	if m.dryRun {
		m.drm.LogOperation("Would fetch keys from URL", url)
		return []string{fmt.Sprintf("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIexample-key-from-url")}, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("URL returned status %d", resp.StatusCode)
	}

	// 读取所有密钥
	var keys []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		if key != "" {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// readFileKeys 从文件读取密钥
func (m *Manager) readFileKeys(path string) ([]string, error) {
	m.logger.Info("Reading keys from file: %s", path)

	if m.dryRun {
		m.drm.LogOperation("Would read keys from file", path)
		return []string{fmt.Sprintf("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIexample-key-from-file")}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 读取所有密钥
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

// filterAndValidateKeys 过滤和验证密钥
func (m *Manager) filterAndValidateKeys(keys []string) ([]string, error) {
	var validKeys []string

	for _, key := range keys {
		// 去除 CRLF
		key = strings.TrimRight(key, "\r\n")

		// 跳过空行和注释
		if key == "" || strings.HasPrefix(key, "#") {
			continue
		}

		// 验证密钥格式
		if err := ValidateKey(key); err != nil {
			m.logger.Warn("Skipping invalid key: %v", err)
			continue
		}

		validKeys = append(validKeys, key)
	}

	if len(validKeys) == 0 {
		return nil, fmt.Errorf("no valid keys found")
	}

	return validKeys, nil
}

// Install 安装密钥
func (m *Manager) Install(keys []string, overwrite bool) (int, error) {
	// 获取用户信息
	userInfo, err := system.GetUser(m.user)
	if err != nil {
		return 0, fmt.Errorf("failed to get user info: %w", err)
	}

	// 获取 authorized_keys 路径
	sshDir := fmt.Sprintf("%s/.ssh", userInfo.HomeDir)
	authKeysPath := fmt.Sprintf("%s/authorized_keys", sshDir)

	// 创建 SSH 目录
	if !m.dryRun {
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			return 0, fmt.Errorf("failed to create .ssh directory: %w", err)
		}
	} else {
		m.drm.LogFileOperation("Create directory", sshDir)
	}

	// 备份现有文件
	if _, err := os.Stat(authKeysPath); err == nil {
		if !m.dryRun {
			backupPath, err := system.BackupFile(authKeysPath)
			if err != nil {
				m.logger.Warn("Failed to backup authorized_keys: %v", err)
			} else if backupPath != "" {
				m.logger.Info("Backed up authorized_keys: %s", backupPath)
			}
		} else {
			m.drm.LogFileOperation("Backup file", authKeysPath)
		}
	}

	// 如果 overwrite，先清空文件
	if overwrite {
		if m.dryRun {
			m.drm.LogOperation("Would clear authorized_keys")
		} else {
			if err := os.WriteFile(authKeysPath, []byte{}, 0600); err != nil {
				return 0, fmt.Errorf("failed to clear authorized_keys: %w", err)
			}
		}
	}

	// 读取现有密钥
	existingKeys := make(map[string]bool)
	if !overwrite {
		if _, err := os.Stat(authKeysPath); err == nil {
			file, err := os.Open(authKeysPath)
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

	file, err := os.OpenFile(authKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
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

	// 设置所有权
	if err := os.Chown(authKeysPath, userInfo.UID, userInfo.GID); err != nil {
		m.logger.Warn("Failed to set ownership on authorized_keys: %v", err)
	}

	m.logger.Info("Added %d keys to authorized_keys", added)
	return added, nil
}

// List 列出已安装的密钥
func (m *Manager) List() ([]string, error) {
	// 获取用户信息
	userInfo, err := system.GetUser(m.user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 获取 authorized_keys 路径
	authKeysPath := fmt.Sprintf("%s/.ssh/authorized_keys", userInfo.HomeDir)

	// 读取文件
	file, err := os.Open(authKeysPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open authorized_keys: %w", err)
	}
	defer file.Close()

	// 读取所有密钥
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

// ValidateKey 验证密钥格式
func ValidateKey(key string) error {
	// SSH 密钥格式：key-type base64-data comment
	// 例如：ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... user@host

	parts := strings.Fields(key)
	if len(parts) < 2 {
		return fmt.Errorf("invalid SSH key format")
	}

	// 检查密钥类型
	keyType := parts[0]
	validTypes := []string{
		"ssh-rsa",
		"ssh-dss",
		"ecdsa-sha2-nistp256",
		"ecdsa-sha2-nistp384",
		"ecdsa-sha2-nistp521",
		"ssh-ed25519",
		"sk-ssh-ed25519",
		"sk-ecdsa-sha2-nistp256",
	}

	validType := false
	for _, vt := range validTypes {
		if keyType == vt {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid SSH key type: %s", keyType)
	}

	// 检查 base64 数据（简单检查）
	if !regexp.MustCompile(`^[A-Za-z0-9+/=]+$`).MatchString(parts[1]) {
		return fmt.Errorf("invalid SSH key data")
	}

	return nil
}
