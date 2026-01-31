package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/Akuma-real/server-toolkit/pkg/i18n"
)

const (
	repo   = "Akuma-real/server-toolkit"
	apiURL = "https://api.github.com/repos/" + repo + "/releases/latest"
	dlBase = "https://github.com/" + repo + "/releases/download"
)

// Release GitHub Release 信息
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// Updater 更新器
type Updater struct {
	current string
	logger  *Logger
}

// NewUpdater 创建新更新器
func NewUpdater(current string, logger *Logger) *Updater {
	return &Updater{
		current: current,
		logger:  logger,
	}
}

// Check 检查更新
func (u *Updater) Check() (string, bool, error) {
	u.logger.Info(i18n.T("info"))

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", false, fmt.Errorf(i18n.T("err_operation_failed"), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", false, fmt.Errorf("failed to decode response: %w", err)
	}

	if release.TagName != u.current {
		u.logger.Info(i18n.T("info"), i18n.T("settings_title"), release.TagName)
		return release.TagName, true, nil
	}

	u.logger.Info(i18n.T("info"), i18n.T("info"))
	return "", false, nil
}

// DoUpdate 执行更新
func (u *Updater) DoUpdate() error {
	latest, hasUpdate, err := u.Check()
	if err != nil {
		return err
	}

	if !hasUpdate {
		u.logger.Info(i18n.T("info"))
		return nil
	}

	u.logger.Info(i18n.T("info"), latest)

	// 构建下载 URL
	osName := runtime.GOOS
	arch := runtime.GOARCH
	downloadURL := fmt.Sprintf("%s/%s/server-toolkit-%s-%s", dlBase, latest, osName, arch)

	u.logger.Info(i18n.T("log_fetching_url"), downloadURL)

	// 下载文件
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "server-toolkit-")
	if err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入临时文件
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}
	tmpFile.Close()

	// 设置可执行权限
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}

	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}

	// 备份当前文件
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}

	// 移动新文件
	if err := os.Rename(tmpFile.Name(), execPath); err != nil {
		// 失败时恢复备份
		os.Rename(backupPath, execPath)
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}

	u.logger.Info(i18n.T("success"))
	return nil
}

// GetCurrentVersion 获取当前版本
func (u *Updater) GetCurrentVersion() string {
	return u.current
}
