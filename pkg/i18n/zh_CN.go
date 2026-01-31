package i18n

var zhCN = map[string]string{
	// 应用
	"app_title":          "服务器运维工具箱",
	"app_version":        "版本: %s",
	"system_info":        "系统信息",
	"menu_main":          "主菜单",
	"menu_system":        "系统管理",
	"menu_ssh":           "SSH 管理",
	"menu_settings":      "设置",
	"menu_exit":          "退出",
	"menu_back":          "返回主菜单",
	"menu_unimplemented": "该功能尚未实现",

	// 主机名
	"hostname_title":     "主机名设置",
	"hostname_current":   "当前主机名: %s",
	"hostname_new":       "新主机名: ",
	"hostname_fqdn":      "FQDN（可选）: ",
	"hostname_success":   "主机名已更新为: %s",
	"hostname_error":     "主机名更新失败: %v",
	"hostname_invalid":   "主机名格式无效",
	"hostname_setting":   "设置主机名",
	"hostname_hosts":     "配置 /etc/hosts",
	"hostname_cloudinit": "Cloud-init 配置",

	// SSH
	"ssh_title":           "SSH 管理",
	"ssh_target_user":     "目标用户: %s",
	"ssh_keys_count":      "已安装密钥: %d 个",
	"ssh_install_keys":    "安装 SSH 公钥",
	"ssh_list_keys":       "列出已安装的密钥",
	"ssh_disable_pwd":     "禁用密码登录",
	"ssh_enable_service":  "启用 SSH 服务",
	"ssh_source_github":   "从 GitHub 获取",
	"ssh_source_url":      "从 URL 获取",
	"ssh_source_file":     "从文件读取",
	"ssh_github_username": "GitHub 用户名: ",
	"ssh_url":             "密钥 URL: ",
	"ssh_file":            "密钥文件路径: ",
	"ssh_overwrite":       "覆盖现有密钥？",
	"ssh_added":           "已添加 %d 个密钥",
	"ssh_success":         "SSH 配置完成",
	"ssh_error":           "SSH 操作失败: %v",
	"ssh_fetching":        "正在获取密钥...",
	"ssh_installing":      "正在安装密钥...",
	"ssh_reloading":       "正在重载 SSH 配置...",

	// 设置
	"settings_title":      "设置",
	"settings_language":   "语言设置",
	"settings_dryrun":     "Dry-run 模式",
	"settings_loglevel":   "日志级别",
	"settings_autoupdate": "自动更新",
	"settings_lang_curr":  "当前: %s",
	"settings_dryrun_on":  "开启后只显示操作，不实际执行",
	"settings_dryrun_off": "关闭后将实际执行操作",

	// 通用
	"confirm":      "确认执行此操作？",
	"yes":          "是",
	"no":           "否",
	"loading":      "加载中...",
	"success":      "操作成功",
	"error":        "操作失败",
	"warning":      "警告",
	"info":         "信息",
	"press_enter":  "按 Enter 继续",
	"press_esc":    "按 Esc 返回",
	"press_ctrl_c": "按 Ctrl+C 退出",

	// 系统信息
	"os_info":     "OS: %s",
	"cpu_info":    "CPU: %d cores",
	"memory_info": "内存: %.1f GB / %.1f GB",
	"disk_info":   "磁盘: %.1f GB / %.1f GB",
	"uptime_info": "运行时间: %s",

	// 操作日志
	"log_backing_up":     "备份: %s -> %s",
	"log_writing_file":   "写入文件: %s",
	"log_executing_cmd":  "执行命令: %s",
	"log_fetching_url":   "从 URL 获取: %s",
	"log_installing_pkg": "安装包: %s",
	"log_reloading_svc":  "重载服务: %s",

	// 错误
	"err_permission_denied": "权限不足",
	"err_file_not_found":    "文件不存在: %s",
	"err_invalid_input":     "无效输入",
	"err_operation_failed":  "操作失败: %v",
	"err_unknown_os":        "不支持的操作系统",
}

func init() {
	RegisterLanguage("zh_CN", zhCN)
}
