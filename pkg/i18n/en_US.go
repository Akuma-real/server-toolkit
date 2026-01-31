package i18n

var enUS = map[string]string{
	// App
	"app_title":     "Server Toolkit",
	"app_version":   "Version: %s",
	"system_info":   "System Information",
	"menu_main":     "Main Menu",
	"menu_system":   "System Management",
	"menu_ssh":      "SSH Management",
	"menu_settings": "Settings",
	"menu_exit":     "Exit",
	"menu_back":     "Back to Main Menu",

	// Hostname
	"hostname_title":     "Hostname Setup",
	"hostname_current":   "Current hostname: %s",
	"hostname_new":       "New hostname: ",
	"hostname_fqdn":      "FQDN (optional): ",
	"hostname_success":   "Hostname updated to: %s",
	"hostname_error":     "Hostname update failed: %v",
	"hostname_invalid":   "Invalid hostname format",
	"hostname_setting":   "Set Hostname",
	"hostname_hosts":     "Configure /etc/hosts",
	"hostname_cloudinit": "Cloud-init Configuration",

	// SSH
	"ssh_title":           "SSH Management",
	"ssh_target_user":     "Target user: %s",
	"ssh_keys_count":      "Installed keys: %d",
	"ssh_install_keys":    "Install SSH Public Keys",
	"ssh_list_keys":       "List Installed Keys",
	"ssh_disable_pwd":     "Disable Password Login",
	"ssh_enable_service":  "Enable SSH Service",
	"ssh_source_github":   "Fetch from GitHub",
	"ssh_source_url":      "Fetch from URL",
	"ssh_source_file":     "Read from File",
	"ssh_github_username": "GitHub username: ",
	"ssh_url":             "Key URL: ",
	"ssh_file":            "Key file path: ",
	"ssh_overwrite":       "Overwrite existing keys?",
	"ssh_added":           "Added %d keys",
	"ssh_success":         "SSH configuration completed",
	"ssh_error":           "SSH operation failed: %v",
	"ssh_fetching":        "Fetching keys...",
	"ssh_installing":      "Installing keys...",
	"ssh_reloading":       "Reloading SSH configuration...",

	// Settings
	"settings_title":      "Settings",
	"settings_language":   "Language",
	"settings_dryrun":     "Dry-run Mode",
	"settings_loglevel":   "Log Level",
	"settings_autoupdate": "Auto Update",
	"settings_lang_curr":  "Current: %s",
	"settings_dryrun_on":  "Enable to preview operations without executing",
	"settings_dryrun_off": "Disable to actually execute operations",

	// Common
	"confirm":      "Confirm this operation?",
	"yes":          "Yes",
	"no":           "No",
	"loading":      "Loading...",
	"success":      "Success",
	"error":        "Error",
	"warning":      "Warning",
	"info":         "Info",
	"press_enter":  "Press Enter to continue",
	"press_esc":    "Press Esc to go back",
	"press_ctrl_c": "Press Ctrl+C to exit",

	// System Information
	"os_info":     "OS: %s",
	"cpu_info":    "CPU: %d cores",
	"memory_info": "Memory: %.1f GB / %.1f GB",
	"disk_info":   "Disk: %.1f GB / %.1f GB",
	"uptime_info": "Uptime: %s",

	// Operation Log
	"log_backing_up":     "Backing up: %s -> %s",
	"log_writing_file":   "Writing file: %s",
	"log_executing_cmd":  "Executing command: %s",
	"log_fetching_url":   "Fetching from URL: %s",
	"log_installing_pkg": "Installing package: %s",
	"log_reloading_svc":  "Reloading service: %s",

	// Errors
	"err_permission_denied": "Permission denied",
	"err_file_not_found":    "File not found: %s",
	"err_invalid_input":     "Invalid input",
	"err_operation_failed":  "Operation failed: %v",
	"err_unknown_os":        "Unsupported operating system",
}

func init() {
	RegisterLanguage("en_US", enUS)
}
