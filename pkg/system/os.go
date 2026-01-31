package system

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DistroFamily Linux 发行版家族
type DistroFamily int

const (
	Debian DistroFamily = iota
	RedHat
	Arch
	Alpine
	Gentoo
	Unknown
)

// DistroInfo 发行版信息
type DistroInfo struct {
	Family  DistroFamily
	ID      string
	Version string
	Pretty  string
}

// DetectDistro 检测 Linux 发行版
func DetectDistro() (*DistroInfo, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := &DistroInfo{
		Family: Unknown,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

		switch key {
		case "ID":
			info.ID = value
		case "VERSION_ID":
			info.Version = value
		case "PRETTY_NAME":
			info.Pretty = value
		}
	}

	// 判断发行版家族
	info.Family = determineFamily(info.ID)

	return info, nil
}

// determineFamily 根据发行版 ID 判断家族
func determineFamily(id string) DistroFamily {
	switch id {
	case "debian", "ubuntu", "linuxmint", "pop":
		return Debian
	case "rhel", "centos", "fedora", "almalinux", "rocky", "ol":
		return RedHat
	case "arch", "manjaro":
		return Arch
	case "alpine":
		return Alpine
	case "gentoo":
		return Gentoo
	default:
		return Unknown
	}
}

// IsDebianBased 判断是否为 Debian 系
func IsDebianBased() bool {
	info, err := DetectDistro()
	if err != nil {
		return false
	}
	return info.Family == Debian
}

// IsRedHatBased 判断是否为 RedHat 系
func IsRedHatBased() bool {
	info, err := DetectDistro()
	if err != nil {
		return false
	}
	return info.Family == RedHat
}

// GetHostname 获取主机名
func GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return hostname, nil
}

// GetSystemInfo 获取系统信息
type SystemInfo struct {
	OS     string
	CPU    string
	Memory string
	Disk   string
	Uptime string
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// 获取 OS 信息
	distro, err := DetectDistro()
	if err != nil {
		info.OS = "Unknown"
	} else {
		info.OS = distro.Pretty
	}

	// TODO: 实现其他系统信息的获取
	info.CPU = "N/A"
	info.Memory = "N/A"
	info.Disk = "N/A"
	info.Uptime = "N/A"

	return info, nil
}
