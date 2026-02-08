package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuItem 菜单项
type MenuItem struct {
	ID      string
	Label   string
	Submenu *MenuModel
	Next    func(parent MenuModel) tea.Model
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
	status   string
	initCmd  tea.Cmd
	// subtitleProvider 用于动态生成副标题（例如异步状态提示）
	subtitleProvider func() string
	// unimplementedMsg 用于当菜单项既没有 Submenu 也没有 Action 时的提示文本
	//（由调用方注入，避免在 tui 包内硬编码 i18n 文案）
	unimplementedMsg string
	quitting         bool
}

// NewMenu 创建新菜单
func NewMenu(title, subtitle string, choices []MenuItem) MenuModel {
	return MenuModel{
		title:    title,
		subtitle: subtitle,
		choices:  choices,
		cursor:   0,
		selected: "",
		status:   "",
		quitting: false,
	}
}

// Init 初始化菜单
func (m MenuModel) Init() tea.Cmd {
	return m.initCmd
}

// Update 更新菜单状态
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ParentMenuMsg:
		if m.parent != nil {
			return *m.parent, nil
		}
		return m, nil

	case RefreshMenuMsg:
		// 用于触发重新渲染（如异步状态更新后）
		return m, refreshMenuTickerCmd()

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
			m.status = ""
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.choices) - 1
			}

		case tea.KeyDown, tea.KeyTab:
			m.status = ""
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
					m.status = ""
					choice.Submenu.parent = &m
					return *choice.Submenu, nil
				}
				if choice.Next != nil {
					m.status = ""
					return choice.Next(m), nil
				}
				if choice.Action != nil {
					m.status = ""
					return m, choice.Action()
				}
				if m.unimplementedMsg != "" {
					m.status = m.unimplementedMsg
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

	if m.subtitleProvider != nil {
		m.subtitle = m.subtitleProvider()
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
			if choice.Submenu == nil && choice.Next == nil && choice.Action == nil {
				choiceLabel = DimStyle.Render("  " + choiceLabel)
			} else {
				choiceLabel = NormalStyle.Render("  " + choiceLabel)
			}
		}
		content += choiceLabel + "\n"
	}

	// 状态提示（例如：未实现功能）
	if m.status != "" {
		content += "\n" + InfoStyle.Render(m.status) + "\n"
	}

	return BorderStyle.Width(62).Render(content)
}

// SetInitCmd 设置菜单 Init 时返回的命令
func (m MenuModel) SetInitCmd(cmd tea.Cmd) MenuModel {
	m.initCmd = cmd
	return m
}

// SetSubtitleProvider 设置副标题动态提供器
func (m MenuModel) SetSubtitleProvider(provider func() string) MenuModel {
	m.subtitleProvider = provider
	return m
}

// SetParent 设置父菜单
func (m MenuModel) SetParent(parent *MenuModel) MenuModel {
	m.parent = parent
	return m
}

// SetUnimplementedMessage 设置当选择“叶子菜单项”（无 Submenu/Action）时的提示文案
func (m MenuModel) SetUnimplementedMessage(msg string) MenuModel {
	m.unimplementedMsg = msg
	return m
}

// ParentMenuMsg 父菜单消息
type ParentMenuMsg struct{}

// RefreshMenuMsg 触发菜单重新渲染的消息
type RefreshMenuMsg struct{}

func refreshMenuTickerCmd() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(_ time.Time) tea.Msg {
		return RefreshMenuMsg{}
	})
}

// RefreshMenuTickerCmd 导出刷新 ticker 命令，用于菜单初始化时启动自动刷新循环。
func RefreshMenuTickerCmd() tea.Cmd {
	return refreshMenuTickerCmd()
}

// SelectedMenuMsg 选中菜单消息
type SelectedMenuMsg struct {
	ID string
}
