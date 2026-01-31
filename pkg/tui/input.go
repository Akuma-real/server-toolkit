package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel 输入框模型
type InputModel struct {
	prompt   string
	input    textinput.Model
	focus    bool
	quitting bool
}

// NewInput 创建新输入框
func NewInput(prompt, placeholder string) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 156
	ti.Width = 50

	return InputModel{
		prompt:   prompt,
		input:    ti,
		focus:    false,
		quitting: false,
	}
}

// Init 初始化输入框
func (m InputModel) Init() tea.Cmd {
	return nil
}

// Update 更新输入框状态
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.quitting = true
			return m, nil
		case tea.KeyCtrlC, tea.KeyEsc:
			m.input.Reset()
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View 渲染输入框
func (m InputModel) View() string {
	if m.quitting {
		return ""
	}

	var promptStyle lipgloss.Style
	if m.focus {
		promptStyle = SelectedStyle
	} else {
		promptStyle = NormalStyle
	}

	return promptStyle.Render(m.prompt) + "\n" + m.input.View()
}

// Value 获取输入值
func (m InputModel) Value() string {
	return m.input.Value()
}

// Focus 设置焦点
func (m InputModel) Focus() InputModel {
	m.focus = true
	m.input.Focus()
	return m
}

// Blur 移除焦点
func (m InputModel) Blur() InputModel {
	m.focus = false
	m.input.Blur()
	return m
}

// Reset 重置输入框
func (m InputModel) Reset() InputModel {
	m.input.Reset()
	m.quitting = false
	return m
}

// Quitting 返回是否正在退出
func (m InputModel) Quitting() bool {
	return m.quitting
}
