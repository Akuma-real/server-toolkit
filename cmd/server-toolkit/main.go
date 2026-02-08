package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/i18n"
	"github.com/Akuma-real/server-toolkit/pkg/system"
	"github.com/Akuma-real/server-toolkit/pkg/tui"
)

var version = "dev"

type updateStatus struct {
	Available   bool
	Latest      string
	CheckFailed bool
}

type updateStateSnapshot struct {
	generation int64
	status     updateStatus
}

var (
	updateStateMu         sync.RWMutex
	currentUpdateSnapshot updateStateSnapshot
	updateCheckGeneration atomic.Int64
)

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
	startAsyncUpdateCheck(cfg)

	model := buildMainMenu(cfg, logger)

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

func buildSubtitle(updateState updateStatus) string {
	var parts []string

	parts = append(parts, i18n.T("app_version", version))

	if distro, err := system.DetectDistro(); err == nil && distro != nil && distro.Pretty != "" {
		parts = append(parts, i18n.T("os_info", distro.Pretty))
	}

	parts = append(parts, fmt.Sprintf("GOOS=%s GOARCH=%s", runtime.GOOS, runtime.GOARCH))

	if updateState.Available {
		parts = append(parts, i18n.T("update_available", updateState.Latest))
	} else if updateState.CheckFailed {
		parts = append(parts, i18n.T("update_check_failed"))
	}

	return strings.Join(parts, "\n")
}

func buildMainMenu(cfg *internal.Config, logger *internal.Logger) tui.MenuModel {
	unimplemented := i18n.T("menu_unimplemented")
	subtitle := buildSubtitle(getUpdateStatus())

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

	mainMenu := tui.NewMenu(
		i18n.T("app_title"),
		subtitle,
		[]tui.MenuItem{
			{ID: "system", Label: i18n.T("menu_system"), Submenu: &systemMenu},
			{ID: "ssh", Label: i18n.T("menu_ssh"), Submenu: &sshMenu},
			{ID: "settings", Label: i18n.T("menu_settings"), Next: func(parent tui.MenuModel) tea.Model {
				return NewSettingsModel(parent, cfg, logger, func() tui.MenuModel {
					return buildMainMenu(cfg, logger)
				})
			}},
			{ID: "exit", Label: i18n.T("menu_exit"), Action: func() tea.Cmd { return tea.Quit }},
		},
	).SetSubtitleProvider(func() string {
		return buildSubtitle(getUpdateStatus())
	}).SetInitCmd(tui.RefreshMenuTickerCmd())

	return mainMenu
}

func maybeCheckUpdates(logger *internal.Logger) updateStatus {
	if version == "" || version == "dev" {
		return updateStatus{}
	}

	updater := internal.NewUpdater(version, logger)
	latest, hasUpdate, err := updater.Check()
	if err != nil {
		logger.Warn("Auto update check failed: %v", err)
		return updateStatus{CheckFailed: true}
	}

	if hasUpdate {
		logger.Warn("Update available: %s", latest)
		return updateStatus{Available: true, Latest: latest}
	}

	return updateStatus{}
}

func getUpdateStatus() updateStatus {
	updateStateMu.RLock()
	defer updateStateMu.RUnlock()
	return currentUpdateSnapshot.status
}

func setUpdateStatus(status updateStatus) {
	updateStateMu.Lock()
	defer updateStateMu.Unlock()
	currentUpdateSnapshot.status = status
}

func clearUpdateStatus() {
	setUpdateStatus(updateStatus{})
}

func setUpdateStatusIfGeneration(expected int64, status updateStatus) bool {
	updateStateMu.Lock()
	defer updateStateMu.Unlock()
	if currentUpdateSnapshot.generation != expected {
		return false
	}
	currentUpdateSnapshot.status = status
	return true
}

func newUpdateCheckLogger() *internal.Logger {
	return internal.NewLogger(internal.ERROR, os.Stdout)
}

func startAsyncUpdateCheck(cfg *internal.Config) {
	if cfg == nil || !cfg.AutoUpdate {
		cancelAsyncUpdateCheck()
		return
	}

	if version == "" || version == "dev" {
		cancelAsyncUpdateCheck()
		return
	}

	generation := updateCheckGeneration.Add(1)
	updateStateMu.Lock()
	currentUpdateSnapshot.generation = generation
	currentUpdateSnapshot.status = updateStatus{}
	updateStateMu.Unlock()

	logger := newUpdateCheckLogger()

	go func(expected int64) {
		status := maybeCheckUpdates(logger)
		setUpdateStatusIfGeneration(expected, status)
	}(generation)
}

func cancelAsyncUpdateCheck() {
	generation := updateCheckGeneration.Add(1)
	updateStateMu.Lock()
	currentUpdateSnapshot.generation = generation
	currentUpdateSnapshot.status = updateStatus{}
	updateStateMu.Unlock()
}
