package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MessageType 消息类型
type MessageType int

const (
	InfoMessage MessageType = iota
	SuccessMessage
	WarningMessage
	ErrorMessage
)

// MessageModel 消息提示模型
type MessageModel struct {
	message  string
	mtype    MessageType
	quitting bool
}

// NewMessage 创建新消息提示
func NewMessage(message string, mtype MessageType) MessageModel {
	return MessageModel{
		message:  message,
		mtype:    mtype,
		quitting: false,
	}
}

// Init 初始化消息提示
func (m MessageModel) Init() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tea.Quit
	})
}

// Update 更新消息提示状态
func (m MessageModel) Update(msg tea.Msg) (MessageModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyEsc, tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View 渲染消息提示
func (m MessageModel) View() string {
	if m.quitting {
		return ""
	}

	var style lipgloss.Style
	switch m.mtype {
	case InfoMessage:
		style = InfoStyle
	case SuccessMessage:
		style = SuccessStyle
	case WarningMessage:
		style = WarningStyle
	case ErrorMessage:
		style = ErrorStyle
	}

	content := style.Render(m.message)

	return BorderStyle.Width(len(m.message) + 10).Render(content)
}
