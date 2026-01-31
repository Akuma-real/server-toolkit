package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// BackupFile 备份文件（.bak.timestamp 格式）
func BackupFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，不需要备份
			return "", nil
		}
		return "", err
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", path)
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102-150405")
	backupPath := path + ".bak." + timestamp

	// 复制文件
	src, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// 复制内容
	_, err = dst.ReadFrom(src)
	if err != nil {
		return "", err
	}

	// 保留权限
	srcInfo, _ := os.Stat(path)
	dst.Chmod(srcInfo.Mode())

	return backupPath, nil
}

// SetPermissions 设置权限
func SetPermissions(path string, mode os.FileMode, uid, gid int) error {
	// 设置权限
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("failed to chmod %s: %w", path, err)
	}

	// 设置所有权
	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("failed to chown %s: %w", path, err)
	}

	return nil
}

// SafeWrite 安全写入（原子操作）
func SafeWrite(path string, data []byte, perm os.FileMode) error {
	// 创建临时文件
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+".tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// 确保临时文件被清理
	defer func() {
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	// 写入数据
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// 关闭文件
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// 设置权限
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}

	// 原子性重命名
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// RestoreSELinuxContext 恢复 SELinux 上下文
func RestoreSELinuxContext(path string) error {
	// 检查 restorecon 是否存在
	if _, err := exec.LookPath("restorecon"); err != nil {
		// restorecon 不存在，跳过
		return nil
	}

	// 执行 restorecon
	cmd := exec.Command("restorecon", "-R", path)
	if err := cmd.Run(); err != nil {
		// SELinux 可能未启用，忽略错误
		return nil
	}

	return nil
}

// ReadFile 读取文件
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile 写入文件
func WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory 检查路径是否为目录
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CreateDir 创建目录
func CreateDir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// EnsureDir 确保目录存在
func EnsureDir(path string, perm os.FileMode) error {
	if !FileExists(path) {
		return CreateDir(path, perm)
	}
	return nil
}

// Mkdir 创建目录
func Mkdir(path string) error {
	return os.MkdirAll(path, 0755)
}
