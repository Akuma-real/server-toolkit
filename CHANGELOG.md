# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0-beta.1] - 2025-01-31

### Added
- 初始版本发布
- 主机名管理功能
  - 设置主机名（支持 hostnamectl 和 hostname）
  - 配置 /etc/hosts（替换/插入模式）
  - Cloud-init 配置（preserve_hostname）
  - 主机名验证（RFC 1123）
- SSH 管理功能
  - SSH 公钥获取（GitHub/URL/文件）
  - authorized_keys 管理（追加/覆盖）
  - SSH 服务管理（启动/重启）
  - SSH 安全加固（禁用密码登录）
- 系统底层功能
  - OS 检测（Debian/RedHat 系）
  - 用户管理
  - 文件操作（备份/权限/SELinux）
  - 包管理器支持（apt/dnf）
  - 服务管理（systemd）
- TUI 界面
  - 交互式菜单（Bubble Tea）
  - 输入组件
  - 确认对话框
  - 进度条
  - 消息提示
- 国际化支持
  - 简体中文（zh_CN）
  - 英语（en_US）
- 配置管理
  - 配置持久化（/etc/server-toolkit/config.json）
  - 语言设置
  - Dry-run 模式
  - 日志级别
  - 自动更新
- 日志系统
  - 分级日志（DEBUG/INFO/WARN/ERROR）
  - 文件日志
  - Dry-run 日志
- 自动更新
  - GitHub API 检查更新
  - 一键更新功能

### Supported Platforms
- Debian 12 (bookworm)
- Ubuntu 22.04 LTS / 24.04 LTS
- AlmaLinux 9
- Rocky Linux 9
- CentOS 9 Stream

### Known Issues
- 仅支持 systemd 系统
- 部分功能需要 root 权限

[Unreleased]: https://github.com/Akuma-real/server-toolkit/compare/v0.1.0-beta.1...HEAD
[0.1.0-beta.1]: https://github.com/Akuma-real/server-toolkit/releases/tag/v0.1.0-beta.1
