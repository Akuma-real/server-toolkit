package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var (
	configDir = "/etc/server-toolkit"
)

const (
	configFile = "config.json"
)

// Config 配置结构
type Config struct {
	Language   string `json:"language"`
	DryRun     bool   `json:"dry_run"`
	LogLevel   string `json:"log_level"`
	AutoUpdate bool   `json:"auto_update"`
	LogPath    string `json:"log_path"`
}

// Load 加载配置
func Load() (*Config, error) {
	path := filepath.Join(configDir, configFile)
	data, err := os.ReadFile(path)
	if err != nil {
		// 文件不存在：使用默认配置
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Default(), fmt.Errorf("failed to read config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	return &cfg, nil
}

// Save 保存配置
func Save(cfg *Config) error {
	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir %s: %w", configDir, err)
	}

	path := filepath.Join(configDir, configFile)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config %s: %w", path, err)
	}

	return nil
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Language:   "zh_CN",
		DryRun:     false,
		LogLevel:   "INFO",
		AutoUpdate: true,
		LogPath:    "/var/log/server-toolkit.log",
	}
}
