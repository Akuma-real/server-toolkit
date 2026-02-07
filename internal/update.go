package internal

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Akuma-real/server-toolkit/pkg/i18n"
)

const (
	repo   = "Akuma-real/server-toolkit"
	apiURL = "https://api.github.com/repos/" + repo + "/releases/latest"
	dlBase = "https://github.com/" + repo + "/releases/download"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// Release GitHub Release 信息
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
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
	u.logger.Info("%s", i18n.T("info"))

	release, err := fetchLatestRelease()
	if err != nil {
		return "", false, err
	}

	if release.TagName != u.current {
		u.logger.Info("%s %s %s", i18n.T("info"), i18n.T("settings_title"), release.TagName)
		return release.TagName, true, nil
	}

	u.logger.Info("%s %s", i18n.T("info"), i18n.T("info"))
	return "", false, nil
}

// DoUpdate 执行更新
func (u *Updater) DoUpdate() error {
	release, err := fetchLatestRelease()
	if err != nil {
		return err
	}

	latest := release.TagName
	hasUpdate := latest != u.current

	if !hasUpdate {
		u.logger.Info("%s", i18n.T("info"))
		return nil
	}

	u.logger.Info("%s %s", i18n.T("info"), latest)

	// 构建下载 URL
	osName := runtime.GOOS
	arch := runtime.GOARCH
	downloadURL := fmt.Sprintf("%s/%s/server-toolkit-%s-%s", dlBase, latest, osName, arch)

	u.logger.Info("%s", fmt.Sprintf(i18n.T("log_fetching_url"), downloadURL))

	// 下载文件
	resp, err := httpClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	// 读取下载内容（用于校验 + 写文件）
	binaryData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}

	// 校验 SHA256（若 release 提供 checksums 资产）
	if err := verifyReleaseSHA256(binaryData, fmt.Sprintf("server-toolkit-%s-%s", osName, arch), release, u.logger); err != nil {
		return err
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "server-toolkit-")
	if err != nil {
		return fmt.Errorf(i18n.T("err_operation_failed"), err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入临时文件
	_, err = io.Copy(tmpFile, bytes.NewReader(binaryData))
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

	u.logger.Info("%s", i18n.T("success"))
	return nil
}

// GetCurrentVersion 获取当前版本
func (u *Updater) GetCurrentVersion() string {
	return u.current
}

func verifyReleaseSHA256(binaryData []byte, binaryName string, release Release, logger *Logger) error {
	var checksumURL string
	for _, asset := range release.Assets {
		if asset.Name == "checksums.txt" || asset.Name == "checksums.sha256" {
			checksumURL = asset.BrowserDownloadURL
			break
		}
	}

	if checksumURL == "" {
		return fmt.Errorf("checksum asset not found for release %s", release.TagName)
	}

	resp, err := httpClient.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksum file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksum file: HTTP %s", resp.Status)
	}

	checksumData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksum file: %w", err)
	}

	expected, found := extractChecksumForFile(string(checksumData), binaryName)
	if !found {
		return fmt.Errorf("checksum for %s not found in checksum asset", binaryName)
	}

	actualSum := sha256.Sum256(binaryData)
	actual := fmt.Sprintf("%x", actualSum)
	if !strings.EqualFold(expected, actual) {
		return fmt.Errorf("checksum mismatch for %s", binaryName)
	}

	logger.Info("Checksum verified for %s", binaryName)
	return nil
}

func fetchLatestRelease() (Release, error) {
	resp, err := httpClient.Get(apiURL)
	if err != nil {
		return Release{}, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("HTTP request failed: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("failed to decode release response: %w", err)
	}

	if release.TagName == "" {
		return Release{}, fmt.Errorf("invalid release response: empty tag_name")
	}

	return release, nil
}

func extractChecksumForFile(content, filename string) (string, bool) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		hashValue := fields[0]
		name := strings.TrimPrefix(fields[len(fields)-1], "*")
		if name == filename {
			return hashValue, true
		}
	}
	return "", false
}
