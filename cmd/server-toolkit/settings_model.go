package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/i18n"
	"github.com/Akuma-real/server-toolkit/pkg/tui"
)

var saveConfig = internal.Save

type settingsModel struct {
	parent   tui.MenuModel
	cfg      *internal.Config
	logger   *internal.Logger
	rebuild  func() tui.MenuModel
	cursor   int
	items    []string
	status   string
	quitting bool
}

func NewSettingsModel(parent tui.MenuModel, cfg *internal.Config, logger *internal.Logger, rebuild func() tui.MenuModel) tea.Model {
	if cfg == nil {
		cfg = internal.Default()
	}
	return settingsModel{
		parent:  parent,
		cfg:     cfg,
		logger:  logger,
		rebuild: rebuild,
		cursor:  0,
		items: []string{
			"lang",
			"dryrun",
			"loglevel",
			"autoupdate",
			"back",
		},
	}
}

func (m settingsModel) Init() tea.Cmd { return initRefreshTickerCmd(nil) }

func (m settingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.RefreshMenuMsg:
		return m, keepRefreshTickerCmd(msg, nil)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.rebuild != nil {
				return m.rebuild(), nil
			}
			return m.parent, nil
		case tea.KeyUp, tea.KeyShiftTab:
			m.status = ""
			if m.cursor == 0 {
				m.cursor = len(m.items) - 1
			} else {
				m.cursor--
			}
			return m, nil
		case tea.KeyDown, tea.KeyTab:
			m.status = ""
			if m.cursor == len(m.items)-1 {
				m.cursor = 0
			} else {
				m.cursor++
			}
			return m, nil
		case tea.KeyEnter:
			return m.applySelection()
		}
	}

	return m, nil
}

func (m settingsModel) applySelection() (tea.Model, tea.Cmd) {
	selected := m.items[m.cursor]
	if selected == "back" {
		if m.rebuild != nil {
			return m.rebuild(), nil
		}
		return m.parent, nil
	}

	if m.cfg == nil {
		m.cfg = internal.Default()
	}

	nextCfg := *m.cfg

	switch selected {
	case "lang":
		if nextCfg.Language == "zh_CN" {
			nextCfg.Language = "en_US"
		} else {
			nextCfg.Language = "zh_CN"
		}
	case "dryrun":
		nextCfg.DryRun = !nextCfg.DryRun
	case "loglevel":
		nextCfg.LogLevel = nextLogLevel(nextCfg.LogLevel)
	case "autoupdate":
		nextCfg.AutoUpdate = !nextCfg.AutoUpdate
	default:
		return m, nil
	}

	if err := saveConfig(&nextCfg); err != nil {
		m.status = i18n.T("settings_save_failed", err)
		return m, nil
	}

	*m.cfg = nextCfg

	switch selected {
	case "lang":
		i18n.SetLanguage(m.cfg.Language)
	case "loglevel":
		if m.logger != nil {
			m.logger.SetLevel(internal.ParseLevel(m.cfg.LogLevel))
		}
	case "autoupdate":
		if m.cfg.AutoUpdate {
			startAsyncUpdateCheck(m.cfg)
		} else {
			cancelAsyncUpdateCheck()
		}
	}

	m.status = i18n.T("settings_saved")
	return m, nil
}

func (m settingsModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(tui.TitleStyle.Width(60).Render(i18n.T("settings_title")) + "\n\n")

	for i, key := range m.items {
		label := m.itemLabel(key)
		line := "  " + label
		if i == m.cursor {
			line = tui.CursorStyle.Render("> " + label)
		} else {
			line = tui.NormalStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	if m.status != "" {
		if strings.Contains(strings.ToLower(m.status), "failed") || strings.Contains(m.status, "失败") {
			b.WriteString("\n" + tui.ErrorStyle.Render(m.status) + "\n")
		} else {
			b.WriteString("\n" + tui.SuccessStyle.Render(m.status) + "\n")
		}
	}

	b.WriteString("\n" + tui.DimStyle.Render(i18n.T("press_enter")+" / "+i18n.T("press_esc")) + "\n")
	return tui.BorderStyle.Width(62).Render(b.String())
}

func (m settingsModel) itemLabel(key string) string {
	switch key {
	case "lang":
		return fmt.Sprintf("%s: %s", i18n.T("settings_language"), m.cfg.Language)
	case "dryrun":
		return fmt.Sprintf("%s: %s", i18n.T("settings_dryrun"), onOff(m.cfg.DryRun))
	case "loglevel":
		return fmt.Sprintf("%s: %s", i18n.T("settings_loglevel"), m.cfg.LogLevel)
	case "autoupdate":
		return fmt.Sprintf("%s: %s", i18n.T("settings_autoupdate"), onOff(m.cfg.AutoUpdate))
	default:
		return i18n.T("menu_back")
	}
}

func nextLogLevel(level string) string {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return "INFO"
	case "INFO":
		return "WARN"
	case "WARN":
		return "ERROR"
	default:
		return "DEBUG"
	}
}

func onOff(value bool) string {
	if value {
		return i18n.T("yes")
	}
	return i18n.T("no")
}
