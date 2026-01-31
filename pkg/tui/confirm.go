package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmModel 确认对话框模型
type ConfirmModel struct {
	question  string
	confirmed bool
	cursor    int // 0: No, 1: Yes
	quitting  bool
}

// NewConfirm 创建新确认对话框
func NewConfirm(question string) ConfirmModel {
	return ConfirmModel{
		question:  question,
		confirmed: false,
		cursor:    1, // 默认 Yes
		quitting:  false,
	}
}

// Init 初始化确认对话框
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update 更新确认对话框状态
func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyLeft, tea.KeyShiftTab:
			m.cursor = 0
		case tea.KeyRight, tea.KeyTab:
			m.cursor = 1
		case tea.KeyEnter:
			m.confirmed = (m.cursor == 1)
			m.quitting = true
			return m, nil
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		case 'y', 'Y':
			m.cursor = 1
			m.confirmed = true
			m.quitting = true
			return m, nil
		case 'n', 'N':
			m.cursor = 0
			m.confirmed = false
			m.quitting = true
			return m, nil
		}
	}

	return m, nil
}

// View 渲染确认对话框
func (m ConfirmModel) View() string {
	if m.quitting {
		return ""
	}

	var noStyle, yesStyle lipgloss.Style
	if m.cursor == 0 {
		noStyle = CursorStyle
		yesStyle = NormalStyle
	} else {
		noStyle = NormalStyle
		yesStyle = CursorStyle
	}

	content := NormalStyle.Render(m.question) + "\n\n"
	content += "  " + noStyle.Render("[ No ]") + "  " + yesStyle.Render("[ Yes ]")

	return BorderStyle.Width(len(m.question) + 10).Render(content)
}

// Confirmed 返回是否确认
func (m ConfirmModel) Confirmed() bool {
	return m.confirmed
}

// Quitting 返回是否正在退出
func (m ConfirmModel) Quitting() bool {
	return m.quitting
}

// Reset 重置确认对话框
func (m ConfirmModel) Reset() ConfirmModel {
	m.confirmed = false
	m.cursor = 1
	m.quitting = false
	return m
}
