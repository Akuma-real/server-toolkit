package system

import (
	"fmt"
	"os/exec"
)

// PackageManager 包管理器接口
type PackageManager interface {
	Update() error
	Install(packages ...string) error
	Remove(packages ...string) error
	IsInstalled(pkg string) (bool, error)
}

// AptManager apt 包管理器
type AptManager struct{}

func (m *AptManager) Update() error {
	cmd := exec.Command("apt-get", "update", "-y")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *AptManager) Install(packages ...string) error {
	args := []string{"install", "-y"}
	args = append(args, packages...)
	cmd := exec.Command("apt-get", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *AptManager) Remove(packages ...string) error {
	args := []string{"remove", "-y"}
	args = append(args, packages...)
	cmd := exec.Command("apt-get", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *AptManager) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("dpkg", "-s", pkg)
	err := cmd.Run()
	return err == nil, nil
}

// DnfManager dnf 包管理器
type DnfManager struct{}

func (m *DnfManager) Update() error {
	cmd := exec.Command("dnf", "update", "-y")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *DnfManager) Install(packages ...string) error {
	args := []string{"install", "-y"}
	args = append(args, packages...)
	cmd := exec.Command("dnf", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *DnfManager) Remove(packages ...string) error {
	args := []string{"remove", "-y"}
	args = append(args, packages...)
	cmd := exec.Command("dnf", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *DnfManager) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("rpm", "-q", pkg)
	err := cmd.Run()
	return err == nil, nil
}

// YumManager yum 包管理器（兼容旧系统）
type YumManager struct{}

func (m *YumManager) Update() error {
	cmd := exec.Command("yum", "update", "-y")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *YumManager) Install(packages ...string) error {
	args := []string{"install", "-y"}
	args = append(args, packages...)
	cmd := exec.Command("yum", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *YumManager) Remove(packages ...string) error {
	args := []string{"remove", "-y"}
	args = append(args, packages...)
	cmd := exec.Command("yum", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *YumManager) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("rpm", "-q", pkg)
	err := cmd.Run()
	return err == nil, nil
}

// DetectPackageManager 检测包管理器
func DetectPackageManager(family DistroFamily) (PackageManager, error) {
	switch family {
	case Debian:
		return &AptManager{}, nil
	case RedHat:
		// 检测是 dnf 还是 yum
		if _, err := exec.LookPath("dnf"); err == nil {
			return &DnfManager{}, nil
		}
		return &YumManager{}, nil
	default:
		return nil, fmt.Errorf("unsupported package manager for family: %d", family)
	}
}

// InstallPackage 安装单个包
func InstallPackage(pkg string) error {
	distro, err := DetectDistro()
	if err != nil {
		return err
	}

	mgr, err := DetectPackageManager(distro.Family)
	if err != nil {
		return err
	}

	// 检查是否已安装
	installed, _ := mgr.IsInstalled(pkg)
	if installed {
		return nil
	}

	// 安装包
	return mgr.Install(pkg)
}
