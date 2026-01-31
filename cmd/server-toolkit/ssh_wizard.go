package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/i18n"
	sshModule "github.com/Akuma-real/server-toolkit/pkg/modules/ssh"
	"github.com/Akuma-real/server-toolkit/pkg/tui"
)

type sshWizardStep int

const (
	sshWizardStepUser sshWizardStep = iota
	sshWizardStepSource
	sshWizardStepSourceValue
	sshWizardStepOverwriteConfirm
	sshWizardStepApplyConfirm
	sshWizardStepApplying
	sshWizardStepResult
)

type sshKeysResultMsg struct {
	err     error
	summary string
	lines   []string
}

type SSHInstallKeysWizard struct {
	parent tui.MenuModel
	cfg    *internal.Config
	logger *internal.Logger

	step sshWizardStep

	userInput textinput.Model

	sourceCursor  int // 0: GitHub, 1: URL, 2: File
	valueInput    textinput.Model
	overwrite     bool
	confirmCursor int // 0: No, 1: Yes

	status string

	result sshKeysResultMsg
}

func NewSSHInstallKeysWizard(parent tui.MenuModel, cfg *internal.Config, logger *internal.Logger) SSHInstallKeysWizard {
	u := defaultUsername()

	userTI := textinput.New()
	userTI.Width = 50
	userTI.CharLimit = 64
	userTI.SetValue(u)
	userTI.Focus()

	valueTI := textinput.New()
	valueTI.Width = 50
	valueTI.CharLimit = 512

	return SSHInstallKeysWizard{
		parent: parent,
		cfg:    cfg,
		logger: logger,
		step:   sshWizardStepUser,

		userInput: userTI,

		sourceCursor:  0,
		valueInput:    valueTI,
		overwrite:     false,
		confirmCursor: 1,
		status:        "",
	}
}

func (m SSHInstallKeysWizard) Init() tea.Cmd { return textinput.Blink }

func (m SSHInstallKeysWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sshKeysResultMsg:
		m.result = msg
		m.step = sshWizardStepResult
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		switch m.step {
		case sshWizardStepUser:
			switch msg.Type {
			case tea.KeyEsc:
				return m.parent, nil
			case tea.KeyEnter:
				m.status = ""
				if strings.TrimSpace(m.userInput.Value()) == "" {
					m.status = errors.New(i18n.T("err_invalid_input")).Error()
					return m, nil
				}
				m.userInput.Blur()
				m.step = sshWizardStepSource
				return m, nil
			}

		case sshWizardStepSource:
			switch msg.Type {
			case tea.KeyEsc:
				m.userInput.Focus()
				m.step = sshWizardStepUser
				return m, nil
			case tea.KeyUp:
				if m.sourceCursor > 0 {
					m.sourceCursor--
				} else {
					m.sourceCursor = 2
				}
				return m, nil
			case tea.KeyDown:
				if m.sourceCursor < 2 {
					m.sourceCursor++
				} else {
					m.sourceCursor = 0
				}
				return m, nil
			case tea.KeyEnter:
				m.status = ""
				m.valueInput.SetValue("")
				m.valueInput.Focus()
				m.step = sshWizardStepSourceValue
				return m, nil
			}

		case sshWizardStepSourceValue:
			switch msg.Type {
			case tea.KeyEsc:
				m.valueInput.Blur()
				m.step = sshWizardStepSource
				return m, nil
			case tea.KeyEnter:
				m.status = ""
				if strings.TrimSpace(m.valueInput.Value()) == "" {
					m.status = errors.New(i18n.T("err_invalid_input")).Error()
					return m, nil
				}
				m.valueInput.Blur()
				m.confirmCursor = 0 // 默认 No
				m.step = sshWizardStepOverwriteConfirm
				return m, nil
			}

		case sshWizardStepOverwriteConfirm:
			switch msg.Type {
			case tea.KeyEsc:
				m.valueInput.Focus()
				m.step = sshWizardStepSourceValue
				return m, nil
			case tea.KeyLeft, tea.KeyShiftTab:
				m.confirmCursor = 0
				return m, nil
			case tea.KeyRight, tea.KeyTab:
				m.confirmCursor = 1
				return m, nil
			case tea.KeyEnter:
				m.overwrite = (m.confirmCursor == 1)
				m.confirmCursor = 1
				m.step = sshWizardStepApplyConfirm
				return m, nil
			}

		case sshWizardStepApplyConfirm:
			switch msg.Type {
			case tea.KeyEsc:
				return m.parent, nil
			case tea.KeyLeft, tea.KeyShiftTab:
				m.confirmCursor = 0
				return m, nil
			case tea.KeyRight, tea.KeyTab:
				m.confirmCursor = 1
				return m, nil
			case tea.KeyEnter:
				if m.confirmCursor == 0 {
					return m.parent, nil
				}
				m.step = sshWizardStepApplying
				return m, m.applyCmd()
			}

		case sshWizardStepApplying:
			return m, nil

		case sshWizardStepResult:
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				return m.parent, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.step {
	case sshWizardStepUser:
		m.userInput, cmd = m.userInput.Update(msg)
	case sshWizardStepSourceValue:
		m.valueInput, cmd = m.valueInput.Update(msg)
	}
	return m, cmd
}

func (m SSHInstallKeysWizard) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Width(60).Render(i18n.T("ssh_wizard_install_title")) + "\n\n")
	if m.cfg != nil && m.cfg.DryRun {
		b.WriteString(tui.WarningStyle.Render(i18n.T("settings_dryrun_on")) + "\n\n")
	}

	switch m.step {
	case sshWizardStepUser:
		b.WriteString(tui.NormalStyle.Render(i18n.T("ssh_wizard_user_prompt")) + "\n")
		b.WriteString(m.userInput.View() + "\n")

	case sshWizardStepSource:
		b.WriteString(tui.NormalStyle.Render(i18n.T("ssh_wizard_source_prompt")) + "\n\n")
		options := []string{
			i18n.T("ssh_source_github"),
			i18n.T("ssh_source_url"),
			i18n.T("ssh_source_file"),
		}
		for i, opt := range options {
			line := "  " + opt
			if i == m.sourceCursor {
				line = tui.CursorStyle.Render("> " + opt)
			} else {
				line = tui.NormalStyle.Render(line)
			}
			b.WriteString(line + "\n")
		}

	case sshWizardStepSourceValue:
		b.WriteString(tui.NormalStyle.Render(sourceValuePrompt(m.sourceCursor)) + "\n")
		b.WriteString(m.valueInput.View() + "\n")

	case sshWizardStepOverwriteConfirm:
		b.WriteString(tui.NormalStyle.Render(i18n.T("ssh_overwrite")) + "\n\n")
		b.WriteString(renderYesNo(m.confirmCursor) + "\n")

	case sshWizardStepApplyConfirm:
		b.WriteString(tui.SubtitleStyle.Render(i18n.T("ssh_wizard_actions")) + "\n")
		for _, line := range m.actionLines() {
			b.WriteString("  " + line + "\n")
		}
		b.WriteString("\n" + tui.NormalStyle.Render(i18n.T("ssh_wizard_confirm_apply")) + "\n\n")
		b.WriteString(renderYesNo(m.confirmCursor) + "\n")

	case sshWizardStepApplying:
		b.WriteString(tui.InfoStyle.Render(i18n.T("ssh_installing")) + "\n")

	case sshWizardStepResult:
		if m.result.err != nil {
			b.WriteString(tui.ErrorStyle.Render(i18n.T("err_operation_failed", m.result.err)) + "\n")
		} else if m.result.summary != "" {
			b.WriteString(tui.SuccessStyle.Render(m.result.summary) + "\n")
		} else {
			b.WriteString(tui.SuccessStyle.Render(i18n.T("success")) + "\n")
		}
		for _, line := range m.result.lines {
			b.WriteString("\n" + tui.DimStyle.Render(line))
		}
		b.WriteString("\n\n" + tui.DimStyle.Render(i18n.T("ssh_wizard_done")) + "\n")
	}

	if m.status != "" {
		b.WriteString("\n" + tui.ErrorStyle.Render(m.status) + "\n")
	}
	if m.step != sshWizardStepApplying && m.step != sshWizardStepResult {
		b.WriteString("\n" + tui.DimStyle.Render(i18n.T("press_enter")+" / "+i18n.T("press_esc")) + "\n")
	}

	return tui.BorderStyle.Width(62).Render(b.String())
}

func (m SSHInstallKeysWizard) actionLines() []string {
	targetUser := strings.TrimSpace(m.userInput.Value())
	srcName := []string{i18n.T("ssh_source_github"), i18n.T("ssh_source_url"), i18n.T("ssh_source_file")}[m.sourceCursor]
	val := strings.TrimSpace(m.valueInput.Value())
	overwriteLabel := i18n.T("no")
	if m.overwrite {
		overwriteLabel = i18n.T("yes")
	}
	lines := []string{
		fmt.Sprintf("%s: %s", i18n.T("ssh_wizard_action_user"), targetUser),
		fmt.Sprintf("%s: %s", i18n.T("ssh_wizard_action_source"), srcName),
		fmt.Sprintf("%s: %s", i18n.T("ssh_wizard_action_value"), val),
		fmt.Sprintf("%s: %s", i18n.T("ssh_wizard_action_overwrite"), overwriteLabel),
	}
	return lines
}

func (m SSHInstallKeysWizard) applyCmd() tea.Cmd {
	targetUser := strings.TrimSpace(m.userInput.Value())
	source := m.sourceCursor
	val := strings.TrimSpace(m.valueInput.Value())
	overwrite := m.overwrite
	dryRun := m.cfg != nil && m.cfg.DryRun
	logger := m.logger

	return func() tea.Msg {
		mgr := sshModule.NewManager(targetUser, dryRun, logger)

		var src sshModule.Source
		switch source {
		case 0:
			src = sshModule.SourceGitHub
		case 1:
			src = sshModule.SourceURL
		default:
			src = sshModule.SourceFile
		}

		keys, err := mgr.FetchKeys(src, val)
		if err != nil {
			return sshKeysResultMsg{err: err}
		}

		added, err := mgr.Install(keys, overwrite)
		if err != nil {
			return sshKeysResultMsg{err: err}
		}

		return sshKeysResultMsg{
			summary: i18n.T("ssh_added", added),
		}
	}
}

type SSHListKeysModel struct {
	parent tui.MenuModel
	cfg    *internal.Config
	logger *internal.Logger

	step      sshWizardStep
	userInput textinput.Model

	status string
	result sshKeysResultMsg
}

func NewSSHListKeysModel(parent tui.MenuModel, cfg *internal.Config, logger *internal.Logger) SSHListKeysModel {
	u := defaultUsername()
	userTI := textinput.New()
	userTI.Width = 50
	userTI.CharLimit = 64
	userTI.SetValue(u)
	userTI.Focus()

	return SSHListKeysModel{
		parent:    parent,
		cfg:       cfg,
		logger:    logger,
		step:      sshWizardStepUser,
		userInput: userTI,
		status:    "",
	}
}

func (m SSHListKeysModel) Init() tea.Cmd { return textinput.Blink }

func (m SSHListKeysModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sshKeysResultMsg:
		m.result = msg
		m.step = sshWizardStepResult
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch m.step {
		case sshWizardStepUser:
			switch msg.Type {
			case tea.KeyEsc:
				return m.parent, nil
			case tea.KeyEnter:
				m.status = ""
				if strings.TrimSpace(m.userInput.Value()) == "" {
					m.status = errors.New(i18n.T("err_invalid_input")).Error()
					return m, nil
				}
				m.userInput.Blur()
				m.step = sshWizardStepApplying
				return m, m.listCmd()
			}
		case sshWizardStepApplying:
			return m, nil
		case sshWizardStepResult:
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				return m.parent, nil
			}
		}
	}

	var cmd tea.Cmd
	if m.step == sshWizardStepUser {
		m.userInput, cmd = m.userInput.Update(msg)
	}
	return m, cmd
}

func (m SSHListKeysModel) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Width(60).Render(i18n.T("ssh_wizard_list_title")) + "\n\n")
	if m.cfg != nil && m.cfg.DryRun {
		b.WriteString(tui.WarningStyle.Render(i18n.T("settings_dryrun_on")) + "\n\n")
	}

	switch m.step {
	case sshWizardStepUser:
		b.WriteString(tui.NormalStyle.Render(i18n.T("ssh_wizard_user_prompt")) + "\n")
		b.WriteString(m.userInput.View() + "\n")
		b.WriteString("\n" + tui.DimStyle.Render(i18n.T("press_enter")+" / "+i18n.T("press_esc")) + "\n")
	case sshWizardStepApplying:
		b.WriteString(tui.InfoStyle.Render(i18n.T("loading")) + "\n")
	case sshWizardStepResult:
		if m.result.err != nil {
			b.WriteString(tui.ErrorStyle.Render(i18n.T("err_operation_failed", m.result.err)) + "\n")
		} else {
			b.WriteString(tui.InfoStyle.Render(m.result.summary) + "\n\n")
			for _, line := range m.result.lines {
				b.WriteString(tui.NormalStyle.Render("  "+line) + "\n")
			}
		}
		b.WriteString("\n" + tui.DimStyle.Render(i18n.T("ssh_wizard_done")) + "\n")
	}

	if m.status != "" {
		b.WriteString("\n" + tui.ErrorStyle.Render(m.status) + "\n")
	}
	return tui.BorderStyle.Width(62).Render(b.String())
}

func (m SSHListKeysModel) listCmd() tea.Cmd {
	targetUser := strings.TrimSpace(m.userInput.Value())
	logger := m.logger
	return func() tea.Msg {
		mgr := sshModule.NewManager(targetUser, true, logger)
		keys, err := mgr.List()
		if err != nil {
			return sshKeysResultMsg{err: err}
		}
		var lines []string
		for _, k := range keys {
			lines = append(lines, redactKeyForDisplay(k))
		}
		return sshKeysResultMsg{
			summary: i18n.T("ssh_keys_count", len(keys)),
			lines:   lines,
		}
	}
}

type SSHDisablePasswordModel struct {
	parent tui.MenuModel
	cfg    *internal.Config
	logger *internal.Logger

	step sshWizardStep

	userInput     textinput.Model
	confirmCursor int
	status        string

	result sshKeysResultMsg
}

func NewSSHDisablePasswordModel(parent tui.MenuModel, cfg *internal.Config, logger *internal.Logger) SSHDisablePasswordModel {
	u := defaultUsername()
	userTI := textinput.New()
	userTI.Width = 50
	userTI.CharLimit = 64
	userTI.SetValue(u)
	userTI.Focus()

	return SSHDisablePasswordModel{
		parent:        parent,
		cfg:           cfg,
		logger:        logger,
		step:          sshWizardStepUser,
		userInput:     userTI,
		confirmCursor: 0, // 默认 No
	}
}

func (m SSHDisablePasswordModel) Init() tea.Cmd { return textinput.Blink }

func (m SSHDisablePasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sshKeysResultMsg:
		m.result = msg
		m.step = sshWizardStepResult
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		switch m.step {
		case sshWizardStepUser:
			switch msg.Type {
			case tea.KeyEsc:
				return m.parent, nil
			case tea.KeyEnter:
				m.status = ""
				if strings.TrimSpace(m.userInput.Value()) == "" {
					m.status = errors.New(i18n.T("err_invalid_input")).Error()
					return m, nil
				}
				m.userInput.Blur()
				m.confirmCursor = 0
				m.step = sshWizardStepApplyConfirm
				return m, nil
			}

		case sshWizardStepApplyConfirm:
			switch msg.Type {
			case tea.KeyEsc:
				return m.parent, nil
			case tea.KeyLeft, tea.KeyShiftTab:
				m.confirmCursor = 0
				return m, nil
			case tea.KeyRight, tea.KeyTab:
				m.confirmCursor = 1
				return m, nil
			case tea.KeyEnter:
				if m.confirmCursor == 0 {
					return m.parent, nil
				}
				m.step = sshWizardStepApplying
				return m, m.applyCmd()
			}

		case sshWizardStepApplying:
			return m, nil

		case sshWizardStepResult:
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				return m.parent, nil
			}
		}
	}

	var cmd tea.Cmd
	if m.step == sshWizardStepUser {
		m.userInput, cmd = m.userInput.Update(msg)
	}
	return m, cmd
}

func (m SSHDisablePasswordModel) View() string {
	var b strings.Builder
	b.WriteString(tui.TitleStyle.Width(60).Render(i18n.T("ssh_wizard_disable_pwd_title")) + "\n\n")
	if m.cfg != nil && m.cfg.DryRun {
		b.WriteString(tui.WarningStyle.Render(i18n.T("settings_dryrun_on")) + "\n\n")
	}

	switch m.step {
	case sshWizardStepUser:
		b.WriteString(tui.NormalStyle.Render(i18n.T("ssh_wizard_user_prompt")) + "\n")
		b.WriteString(m.userInput.View() + "\n")
		b.WriteString("\n" + tui.DimStyle.Render(i18n.T("press_enter")+" / "+i18n.T("press_esc")) + "\n")
	case sshWizardStepApplyConfirm:
		b.WriteString(tui.WarningStyle.Render(i18n.T("ssh_wizard_disable_pwd_warning")) + "\n\n")
		b.WriteString(tui.NormalStyle.Render(i18n.T("ssh_wizard_confirm_apply")) + "\n\n")
		b.WriteString(renderYesNo(m.confirmCursor) + "\n")
	case sshWizardStepApplying:
		b.WriteString(tui.InfoStyle.Render(i18n.T("ssh_reloading")) + "\n")
	case sshWizardStepResult:
		if m.result.err != nil {
			b.WriteString(tui.ErrorStyle.Render(i18n.T("err_operation_failed", m.result.err)) + "\n")
		} else if m.result.summary != "" {
			b.WriteString(tui.SuccessStyle.Render(m.result.summary) + "\n")
		} else {
			b.WriteString(tui.SuccessStyle.Render(i18n.T("success")) + "\n")
		}
		b.WriteString("\n" + tui.DimStyle.Render(i18n.T("ssh_wizard_done")) + "\n")
	}

	if m.status != "" {
		b.WriteString("\n" + tui.ErrorStyle.Render(m.status) + "\n")
	}
	return tui.BorderStyle.Width(62).Render(b.String())
}

func (m SSHDisablePasswordModel) applyCmd() tea.Cmd {
	targetUser := strings.TrimSpace(m.userInput.Value())
	dryRun := m.cfg != nil && m.cfg.DryRun
	logger := m.logger
	return func() tea.Msg {
		// 保护性检查：目标用户至少存在 1 个 key（否则默认仍允许继续，但错误信息更清晰）
		mgr := sshModule.NewManager(targetUser, true, logger)
		keys, err := mgr.List()
		if err != nil {
			return sshKeysResultMsg{err: err}
		}
		if len(keys) == 0 {
			return sshKeysResultMsg{err: errors.New(i18n.T("ssh_wizard_no_keys"))}
		}

		cfg, err := sshModule.NewConfig("/etc/ssh/sshd_config", dryRun, logger)
		if err != nil {
			return sshKeysResultMsg{err: err}
		}
		if err := cfg.DisablePasswordAuth(); err != nil {
			return sshKeysResultMsg{err: err}
		}
		if err := sshModule.ReloadSSHD(dryRun, logger); err != nil {
			return sshKeysResultMsg{err: err}
		}
		return sshKeysResultMsg{summary: i18n.T("ssh_success")}
	}
}

func defaultUsername() string {
	if v := strings.TrimSpace(os.Getenv("SUDO_USER")); v != "" {
		return v
	}
	if u, err := user.Current(); err == nil && u != nil && u.Username != "" {
		// 兼容 some systems: Username may contain domain; keep as-is.
		return u.Username
	}
	return "root"
}

func sourceValuePrompt(cursor int) string {
	switch cursor {
	case 0:
		return i18n.T("ssh_github_username")
	case 1:
		return i18n.T("ssh_url")
	default:
		return i18n.T("ssh_file")
	}
}

func redactKeyForDisplay(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	parts := strings.Fields(key)
	if len(parts) < 2 {
		return key
	}
	typ := parts[0]
	data := parts[1]
	comment := ""
	if len(parts) >= 3 {
		comment = strings.Join(parts[2:], " ")
	}
	if len(data) > 24 {
		data = data[:12] + "..." + data[len(data)-8:]
	}
	if comment != "" {
		return fmt.Sprintf("%s %s %s", typ, data, comment)
	}
	return fmt.Sprintf("%s %s", typ, data)
}
