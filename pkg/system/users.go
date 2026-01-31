package system

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// UserInfo 用户信息
type UserInfo struct {
	Username string
	UID      int
	GID      int
	HomeDir  string
	Shell    string
}

// GetUserFromPasswd 从 /etc/passwd 获取用户信息
func GetUserFromPasswd(username string) (*UserInfo, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		if parts[0] == username {
			uid, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}

			gid, err := strconv.Atoi(parts[3])
			if err != nil {
				return nil, err
			}

			return &UserInfo{
				Username: username,
				UID:      uid,
				GID:      gid,
				HomeDir:  parts[5],
				Shell:    parts[6],
			}, nil
		}
	}

	return nil, fmt.Errorf("user not found: %s", username)
}

// GetUser 获取用户信息
func GetUser(username string) (*UserInfo, error) {
	// 优先从 /etc/passwd 读取
	return GetUserFromPasswd(username)
}

// LookupUID 通过 UID 获取用户信息
func LookupUID(uid int) (*UserInfo, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 7 {
			continue
		}

		fileUID, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		if fileUID == uid {
			gid, err := strconv.Atoi(parts[3])
			if err != nil {
				return nil, err
			}

			return &UserInfo{
				Username: parts[0],
				UID:      uid,
				GID:      gid,
				HomeDir:  parts[5],
				Shell:    parts[6],
			}, nil
		}
	}

	return nil, fmt.Errorf("user not found with UID: %d", uid)
}

// GetPrimaryGID 获取用户的主组 ID
func GetPrimaryGID(uid int) (int, error) {
	u, err := LookupUID(uid)
	if err != nil {
		return 0, err
	}
	return u.GID, nil
}

// CurrentUser 获取当前用户
func CurrentUser() (*UserInfo, error) {
	uid := os.Getuid()
	return LookupUID(uid)
}

// IsRoot 检查是否为 root
func IsRoot() bool {
	return os.Geteuid() == 0
}

// GetEffectiveUser 获取有效用户
func GetEffectiveUser() (*UserInfo, error) {
	uid := os.Geteuid()
	return LookupUID(uid)
}

// GetUserHomeDir 获取用户主目录
func GetUserHomeDir(username string) (string, error) {
	u, err := GetUser(username)
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}

// ChangeOwnership 更改文件所有权
func ChangeOwnership(path string, uid, gid int) error {
	return os.Chown(path, uid, gid)
}

// GetFileOwnership 获取文件所有权
func GetFileOwnership(path string) (int, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, nil
	}

	return int(stat.Uid), int(stat.Gid), nil
}
