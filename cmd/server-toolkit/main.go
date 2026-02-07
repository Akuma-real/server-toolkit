package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/i18n"
	"github.com/Akuma-real/server-toolkit/pkg/system"
	"github.com/Akuma-real/server-toolkit/pkg/tui"
)

var version = "dev"

func main() {
	i18n.Init()

	cfg, err := internal.Load()
	if err != nil || cfg == nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load config, using defaults: %v\n", err)
		}
		cfg = internal.Default()
	}
	i18n.SetLanguage(cfg.Language)

	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}

	logger := newLogger(cfg)

	subtitle := buildSubtitle()
	model := buildMainMenu(subtitle, cfg, logger)

	if _, err := tea.NewProgram(model).Run(); err != nil {
		logger.Error("TUI exited with error: %v", err)
		os.Exit(1)
	}
}

func newLogger(cfg *internal.Config) *internal.Logger {
	level := internal.ParseLevel(cfg.LogLevel)

	out := os.Stdout
	if cfg.LogPath != "" {
		f, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			out = f
		}
	}

	return internal.NewLogger(level, out)
}

func buildSubtitle() string {
	var parts []string

	parts = append(parts, i18n.T("app_version", version))

	if distro, err := system.DetectDistro(); err == nil && distro != nil && distro.Pretty != "" {
		parts = append(parts, i18n.T("os_info", distro.Pretty))
	}

	parts = append(parts, fmt.Sprintf("GOOS=%s GOARCH=%s", runtime.GOOS, runtime.GOARCH))

	return strings.Join(parts, "\n")
}

func buildMainMenu(subtitle string, cfg *internal.Config, logger *internal.Logger) tui.MenuModel {
	unimplemented := i18n.T("menu_unimplemented")

	systemMenu := tui.NewMenu(
		i18n.T("menu_system"),
		"",
		[]tui.MenuItem{
			{ID: "hostname", Label: i18n.T("hostname_setting"), Next: func(parent tui.MenuModel) tea.Model {
				return NewHostnameWizard(parent, cfg, logger, true, true)
			}},
			{ID: "back", Label: i18n.T("menu_back"), Action: func() tea.Cmd { return func() tea.Msg { return tui.ParentMenuMsg{} } }},
		},
	).SetUnimplementedMessage(unimplemented)

	sshMenu := tui.NewMenu(
		i18n.T("menu_ssh"),
		"",
		[]tui.MenuItem{
			{ID: "install_keys", Label: i18n.T("ssh_install_keys"), Next: func(parent tui.MenuModel) tea.Model {
				return NewSSHInstallKeysWizard(parent, cfg, logger)
			}},
			{ID: "list_keys", Label: i18n.T("ssh_list_keys"), Next: func(parent tui.MenuModel) tea.Model {
				return NewSSHListKeysModel(parent, cfg, logger)
			}},
			{ID: "disable_pwd", Label: i18n.T("ssh_disable_pwd"), Next: func(parent tui.MenuModel) tea.Model {
				return NewSSHDisablePasswordModel(parent, cfg, logger)
			}},
			{ID: "back", Label: i18n.T("menu_back"), Action: func() tea.Cmd { return func() tea.Msg { return tui.ParentMenuMsg{} } }},
		},
	).SetUnimplementedMessage(unimplemented)

	settingsMenu := tui.NewMenu(
		i18n.T("menu_settings"),
		"",
		[]tui.MenuItem{
			{ID: "lang", Label: i18n.T("settings_language")},
			{ID: "dryrun", Label: i18n.T("settings_dryrun")},
			{ID: "loglevel", Label: i18n.T("settings_loglevel")},
			{ID: "autoupdate", Label: i18n.T("settings_autoupdate")},
			{ID: "back", Label: i18n.T("menu_back"), Action: func() tea.Cmd { return func() tea.Msg { return tui.ParentMenuMsg{} } }},
		},
	).SetUnimplementedMessage(unimplemented)

	mainMenu := tui.NewMenu(
		i18n.T("app_title"),
		subtitle,
		[]tui.MenuItem{
			{ID: "system", Label: i18n.T("menu_system"), Submenu: &systemMenu},
			{ID: "ssh", Label: i18n.T("menu_ssh"), Submenu: &sshMenu},
			{ID: "settings", Label: i18n.T("menu_settings"), Submenu: &settingsMenu},
			{ID: "exit", Label: i18n.T("menu_exit"), Action: func() tea.Cmd { return tea.Quit }},
		},
	)

	return mainMenu
}
