package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	configDir  = "/etc/server-toolkit"
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
		// 如果文件不存在，返回默认配置
		return Default(), nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		// 解析失败，返回默认配置
		return Default(), nil
	}

	return &cfg, nil
}

// Save 保存配置
func Save(cfg *Config) error {
	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(configDir, configFile)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
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
