package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/i18n"
	hostnameModule "github.com/Akuma-real/server-toolkit/pkg/modules/hostname"
	"github.com/Akuma-real/server-toolkit/pkg/tui"
)

type hostnameWizardStep int

const (
	hostnameWizardStepShort hostnameWizardStep = iota
	hostnameWizardStepFQDN
	hostnameWizardStepCloudInitConfirm
	hostnameWizardStepApplyConfirm
	hostnameWizardStepApplying
	hostnameWizardStepResult
)

type hostnameWizardAppliedMsg struct {
	err     error
	summary string
}

// HostnameWizardModel 一步式：设置主机名（可选）+ 更新 /etc/hosts（可选）+ 可选 cloud-init preserve
type HostnameWizardModel struct {
	parent tui.MenuModel
	cfg    *internal.Config
	logger *internal.Logger

	doHostname bool
	doHosts    bool

	step hostnameWizardStep

	shortInput textinput.Model
	fqdnInput  textinput.Model

	short string
	fqdn  string

	cloudInitPresent  bool
	preserveCloudInit bool

	confirmCursor int // 0: No, 1: Yes
	status        string

	resultErr     error
	resultSummary string
}

func NewHostnameWizard(parent tui.MenuModel, cfg *internal.Config, logger *internal.Logger, doHostname, doHosts bool) HostnameWizardModel {
	shortTI := textinput.New()
	shortTI.Placeholder = "my-host"
	shortTI.CharLimit = 253
	shortTI.Width = 50
	shortTI.Focus()

	fqdnTI := textinput.New()
	fqdnTI.Placeholder = "my-host.example.com"
	fqdnTI.CharLimit = 253
	fqdnTI.Width = 50

	return HostnameWizardModel{
		parent: parent,
		cfg:    cfg,
		logger: logger,

		doHostname: doHostname,
		doHosts:    doHosts,

		step: hostnameWizardStepShort,

		shortInput: shortTI,
		fqdnInput:  fqdnTI,

		cloudInitPresent:  hostnameModule.IsPresent(),
		preserveCloudInit: false,

		confirmCursor: 1, // 默认 Yes（执行确认）
		status:        "",
	}
}

func (m HostnameWizardModel) Init() tea.Cmd {
	return initRefreshTickerCmd(textinput.Blink)
}

func (m HostnameWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case hostnameWizardAppliedMsg:
		m.resultErr = msg.err
		m.resultSummary = msg.summary
		m.step = hostnameWizardStepResult
		return m, nil

	case tea.KeyMsg:
		// 全局：Ctrl+C 退出
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		switch m.step {
		case hostnameWizardStepShort:
			switch msg.Type {
			case tea.KeyEsc:
				return m.parent, nil
			case tea.KeyEnter:
				m.status = ""
				if err := m.validateAndStoreShort(); err != nil {
					m.status = err.Error()
					return m, nil
				}
				m.shortInput.Blur()
				m.fqdnInput.Focus()
				m.step = hostnameWizardStepFQDN
				return m, nil
			}

		case hostnameWizardStepFQDN:
			switch msg.Type {
			case tea.KeyEsc:
				m.status = ""
				m.fqdnInput.Blur()
				m.shortInput.Focus()
				m.step = hostnameWizardStepShort
				return m, nil
			case tea.KeyEnter:
				m.status = ""
				if err := m.validateAndStoreFQDN(); err != nil {
					m.status = err.Error()
					return m, nil
				}

				// 仅在“设置主机名”流程里提示 cloud-init
				if m.doHostname && m.cloudInitPresent {
					m.confirmCursor = 0 // 默认 No，避免误改云环境策略
					m.step = hostnameWizardStepCloudInitConfirm
				} else {
					m.confirmCursor = 1
					m.step = hostnameWizardStepApplyConfirm
				}
				return m, nil
			}

		case hostnameWizardStepCloudInitConfirm:
			switch msg.Type {
			case tea.KeyEsc:
				m.confirmCursor = 1
				m.step = hostnameWizardStepFQDN
				return m, nil
			case tea.KeyLeft, tea.KeyShiftTab:
				m.confirmCursor = 0
				return m, nil
			case tea.KeyRight, tea.KeyTab:
				m.confirmCursor = 1
				return m, nil
			case tea.KeyEnter:
				m.preserveCloudInit = (m.confirmCursor == 1)
				m.confirmCursor = 1
				m.step = hostnameWizardStepApplyConfirm
				return m, nil
			}

		case hostnameWizardStepApplyConfirm:
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
				m.step = hostnameWizardStepApplying
				return m, m.applyCmd()
			}

		case hostnameWizardStepApplying:
			// 忽略按键（除 Ctrl+C）
			return m, nil

		case hostnameWizardStepResult:
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				return m.parent, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.step {
	case hostnameWizardStepShort:
		m.shortInput, cmd = m.shortInput.Update(msg)
	case hostnameWizardStepFQDN:
		m.fqdnInput, cmd = m.fqdnInput.Update(msg)
	}
	return m, keepRefreshTickerCmd(msg, cmd)
}

func (m HostnameWizardModel) View() string {
	var content strings.Builder

	content.WriteString(tui.TitleStyle.Width(60).Render(i18n.T("hostname_title")) + "\n\n")

	if m.cfg != nil && m.cfg.DryRun {
		content.WriteString(tui.WarningStyle.Render(i18n.T("settings_dryrun_on")) + "\n\n")
	}

	switch m.step {
	case hostnameWizardStepShort:
		content.WriteString(tui.NormalStyle.Render(i18n.T("hostname_new")) + "\n")
		content.WriteString(m.shortInput.View() + "\n")

	case hostnameWizardStepFQDN:
		content.WriteString(tui.NormalStyle.Render(i18n.T("hostname_fqdn")) + "\n")
		content.WriteString(m.fqdnInput.View() + "\n")

	case hostnameWizardStepCloudInitConfirm:
		content.WriteString(tui.NormalStyle.Render(i18n.T("hostname_wizard_cloudinit_prompt")) + "\n\n")
		content.WriteString(renderYesNo(m.confirmCursor) + "\n")

	case hostnameWizardStepApplyConfirm:
		content.WriteString(tui.SubtitleStyle.Render(i18n.T("hostname_wizard_actions")) + "\n")
		for _, line := range m.actionLines() {
			content.WriteString("  " + line + "\n")
		}
		content.WriteString("\n" + tui.NormalStyle.Render(i18n.T("hostname_wizard_confirm_apply")) + "\n\n")
		content.WriteString(renderYesNo(m.confirmCursor) + "\n")

	case hostnameWizardStepApplying:
		content.WriteString(tui.InfoStyle.Render(i18n.T("hostname_wizard_applying")) + "\n")

	case hostnameWizardStepResult:
		if m.resultErr != nil {
			content.WriteString(tui.ErrorStyle.Render(i18n.T("err_operation_failed", m.resultErr)) + "\n")
		} else {
			if m.resultSummary != "" {
				content.WriteString(tui.SuccessStyle.Render(m.resultSummary) + "\n")
			} else {
				content.WriteString(tui.SuccessStyle.Render(i18n.T("success")) + "\n")
			}
		}
		content.WriteString("\n" + tui.DimStyle.Render(i18n.T("hostname_wizard_done")) + "\n")
	}

	if m.status != "" {
		content.WriteString("\n" + tui.ErrorStyle.Render(m.status) + "\n")
	}

	if m.step == hostnameWizardStepShort || m.step == hostnameWizardStepFQDN || m.step == hostnameWizardStepApplyConfirm || m.step == hostnameWizardStepCloudInitConfirm {
		content.WriteString("\n" + tui.DimStyle.Render(i18n.T("press_enter")+" / "+i18n.T("press_esc")) + "\n")
	}

	return tui.BorderStyle.Width(62).Render(content.String())
}

func (m *HostnameWizardModel) validateAndStoreShort() error {
	raw := strings.TrimSpace(m.shortInput.Value())
	if raw == "" {
		return errors.New(i18n.T("hostname_invalid"))
	}

	raw = hostnameModule.NormalizeHostname(raw)
	if strings.Contains(raw, ".") && strings.TrimSpace(m.fqdnInput.Value()) == "" {
		// 用户把 FQDN 填在“主机名”里：自动拆分，减少反复输入
		m.fqdnInput.SetValue(raw)
		raw = hostnameModule.GetShortHostname(raw)
	}

	if err := hostnameModule.ValidateHostname(raw); err != nil {
		return errors.New(i18n.T("hostname_invalid"))
	}

	m.short = raw
	return nil
}

func (m *HostnameWizardModel) validateAndStoreFQDN() error {
	raw := strings.TrimSpace(m.fqdnInput.Value())
	raw = hostnameModule.NormalizeHostname(raw)
	if raw == "" {
		m.fqdn = ""
		return nil
	}

	if err := hostnameModule.ValidateFQDN(raw); err != nil {
		return errors.New(i18n.T("hostname_invalid"))
	}

	m.fqdn = raw
	return nil
}

func (m HostnameWizardModel) actionLines() []string {
	var lines []string
	target := m.short
	if m.fqdn != "" {
		target = fmt.Sprintf("%s (%s)", m.short, m.fqdn)
	}

	if m.doHostname {
		lines = append(lines, i18n.T("hostname_wizard_action_set", target))
	}
	if m.doHosts {
		lines = append(lines, i18n.T("hostname_wizard_action_hosts", target))
	}
	if m.doHostname && m.cloudInitPresent && m.preserveCloudInit {
		lines = append(lines, i18n.T("hostname_wizard_action_cloudinit"))
	}
	return lines
}

func renderYesNo(cursor int) string {
	no := fmt.Sprintf("[ %s ]", i18n.T("no"))
	yes := fmt.Sprintf("[ %s ]", i18n.T("yes"))

	var noStyle, yesStyle string
	if cursor == 0 {
		noStyle = tui.CursorStyle.Render(no)
		yesStyle = tui.NormalStyle.Render(yes)
	} else {
		noStyle = tui.NormalStyle.Render(no)
		yesStyle = tui.CursorStyle.Render(yes)
	}

	return "  " + noStyle + "  " + yesStyle
}

func (m HostnameWizardModel) applyCmd() tea.Cmd {
	short := m.short
	fqdn := m.fqdn
	doHostname := m.doHostname
	doHosts := m.doHosts
	doCloudInit := m.doHostname && m.cloudInitPresent && m.preserveCloudInit
	dryRun := m.cfg != nil && m.cfg.DryRun
	logger := m.logger

	return func() tea.Msg {
		var summaryParts []string
		var oldName string

		if doHosts {
			// 读取现有主机名用于 hosts 更新（即使当前实现主要用 Replace127，也保留上下文）
			mgr := hostnameModule.NewManager(true, logger)
			if n, err := mgr.GetHostname(); err == nil {
				oldName = n
			}
		}

		if doHostname {
			mgr := hostnameModule.NewManager(dryRun, logger)
			if err := mgr.SetHostname(short, fqdn); err != nil {
				return hostnameWizardAppliedMsg{err: err}
			}
			summaryParts = append(summaryParts, i18n.T("hostname_success", short))
		}

		if doHosts {
			if err := hostnameModule.UpdateHosts(oldName, short, fqdn, hostnameModule.Replace127, dryRun, logger); err != nil {
				return hostnameWizardAppliedMsg{err: err}
			}
			summaryParts = append(summaryParts, i18n.T("hostname_wizard_action_hosts", targetForSummary(short, fqdn)))
		}

		if doCloudInit {
			if err := hostnameModule.SetPreserveHostname(dryRun, logger); err != nil {
				return hostnameWizardAppliedMsg{err: err}
			}
			summaryParts = append(summaryParts, i18n.T("hostname_wizard_action_cloudinit"))
		}

		return hostnameWizardAppliedMsg{summary: strings.Join(summaryParts, "\n")}
	}
}

func targetForSummary(short, fqdn string) string {
	if fqdn == "" {
		return short
	}
	return fmt.Sprintf("%s (%s)", short, fqdn)
}

type cloudInitPreserveStep int

const (
	cloudInitPreserveStepConfirm cloudInitPreserveStep = iota
	cloudInitPreserveStepApplying
	cloudInitPreserveStepResult
)

type cloudInitPreserveAppliedMsg struct{ err error }

type CloudInitPreserveModel struct {
	parent tui.MenuModel
	cfg    *internal.Config
	logger *internal.Logger

	present bool

	step   cloudInitPreserveStep
	cursor int // 0: No, 1: Yes

	resultErr error
}

func NewCloudInitPreserveModel(parent tui.MenuModel, cfg *internal.Config, logger *internal.Logger) CloudInitPreserveModel {
	return CloudInitPreserveModel{
		parent:  parent,
		cfg:     cfg,
		logger:  logger,
		present: hostnameModule.IsPresent(),
		step:    cloudInitPreserveStepConfirm,
		cursor:  0, // 默认 No
	}
}

func (m CloudInitPreserveModel) Init() tea.Cmd { return initRefreshTickerCmd(nil) }

func (m CloudInitPreserveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cloudInitPreserveAppliedMsg:
		m.resultErr = msg.err
		m.step = cloudInitPreserveStepResult
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		switch m.step {
		case cloudInitPreserveStepConfirm:
			switch msg.Type {
			case tea.KeyEsc:
				return m.parent, nil
			case tea.KeyLeft, tea.KeyShiftTab:
				m.cursor = 0
				return m, nil
			case tea.KeyRight, tea.KeyTab:
				m.cursor = 1
				return m, nil
			case tea.KeyEnter:
				if !m.present || m.cursor == 0 {
					return m.parent, nil
				}
				m.step = cloudInitPreserveStepApplying
				return m, m.applyCmd()
			}
		case cloudInitPreserveStepApplying:
			return m, nil
		case cloudInitPreserveStepResult:
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				return m.parent, nil
			}
		}
	}

	return m, keepRefreshTickerCmd(msg, nil)
}

func (m CloudInitPreserveModel) View() string {
	var content strings.Builder

	content.WriteString(tui.TitleStyle.Width(60).Render(i18n.T("hostname_cloudinit")) + "\n\n")

	if m.cfg != nil && m.cfg.DryRun {
		content.WriteString(tui.WarningStyle.Render(i18n.T("settings_dryrun_on")) + "\n\n")
	}

	if !m.present {
		content.WriteString(tui.InfoStyle.Render(i18n.T("cloudinit_not_found")) + "\n\n")
		content.WriteString(tui.DimStyle.Render(i18n.T("press_enter")) + "\n")
		return tui.BorderStyle.Width(62).Render(content.String())
	}

	switch m.step {
	case cloudInitPreserveStepConfirm:
		content.WriteString(tui.NormalStyle.Render(i18n.T("cloudinit_preserve_confirm")) + "\n\n")
		content.WriteString(renderYesNo(m.cursor) + "\n")
		content.WriteString("\n" + tui.DimStyle.Render(i18n.T("press_enter")+" / "+i18n.T("press_esc")) + "\n")
	case cloudInitPreserveStepApplying:
		content.WriteString(tui.InfoStyle.Render(i18n.T("hostname_wizard_applying")) + "\n")
	case cloudInitPreserveStepResult:
		if m.resultErr != nil {
			content.WriteString(tui.ErrorStyle.Render(i18n.T("err_operation_failed", m.resultErr)) + "\n")
		} else {
			content.WriteString(tui.SuccessStyle.Render(i18n.T("success")) + "\n")
		}
		content.WriteString("\n" + tui.DimStyle.Render(i18n.T("hostname_wizard_done")) + "\n")
	}

	return tui.BorderStyle.Width(62).Render(content.String())
}

func (m CloudInitPreserveModel) applyCmd() tea.Cmd {
	dryRun := m.cfg != nil && m.cfg.DryRun
	logger := m.logger
	return func() tea.Msg {
		return cloudInitPreserveAppliedMsg{err: hostnameModule.SetPreserveHostname(dryRun, logger)}
	}
}
