package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// MenuItem 菜单项
type MenuItem struct {
	ID      string
	Label   string
	Submenu *MenuModel
	Action  func() tea.Cmd
}

// MenuModel 菜单模型
type MenuModel struct {
	title    string
	subtitle string
	choices  []MenuItem
	cursor   int
	selected string
	parent   *MenuModel
	quitting bool
}

// NewMenu 创建新菜单
func NewMenu(title, subtitle string, choices []MenuItem) MenuModel {
	return MenuModel{
		title:    title,
		subtitle: subtitle,
		choices:  choices,
		cursor:   0,
		selected: "",
		quitting: false,
	}
}

// Init 初始化菜单
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update 更新菜单状态
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ParentMenuMsg:
		if m.parent != nil {
			return *m.parent, nil
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.parent != nil {
				return m, func() tea.Msg {
					return ParentMenuMsg{}
				}
			}
			m.quitting = true
			return m, tea.Quit

		case tea.KeyUp, tea.KeyShiftTab:
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.choices) - 1
			}

		case tea.KeyDown, tea.KeyTab:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		case tea.KeyEnter:
			if m.cursor < len(m.choices) {
				choice := &m.choices[m.cursor]
				m.selected = choice.ID
				if choice.Submenu != nil {
					choice.Submenu.parent = &m
					return *choice.Submenu, nil
				}
				if choice.Action != nil {
					return m, choice.Action()
				}
			}
		}
	}

	return m, nil
}

// View 渲染菜单
func (m MenuModel) View() string {
	if m.quitting {
		return ""
	}

	// 构建菜单内容
	var content string

	// 标题
	content += TitleStyle.Width(60).Render(m.title) + "\n\n"

	// 系统信息（如果有）
	if m.subtitle != "" {
		content += SubtitleStyle.Render(m.subtitle) + "\n\n"
	}

	// 菜单项
	for i, choice := range m.choices {
		choiceLabel := choice.Label
		if i == m.cursor {
			choiceLabel = CursorStyle.Render("> " + choiceLabel)
		} else {
			choiceLabel = NormalStyle.Render("  " + choiceLabel)
		}
		content += choiceLabel + "\n"
	}

	return BorderStyle.Width(62).Render(content)
}

// SetParent 设置父菜单
func (m MenuModel) SetParent(parent *MenuModel) MenuModel {
	m.parent = parent
	return m
}

// ParentMenuMsg 父菜单消息
type ParentMenuMsg struct{}

// SelectedMenuMsg 选中菜单消息
type SelectedMenuMsg struct {
	ID string
}
