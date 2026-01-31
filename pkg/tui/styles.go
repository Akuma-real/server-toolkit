package tui

import "github.com/charmbracelet/lipgloss"

// Styles 定义了 TUI 界面的样式
var (
	// TitleStyle 标题样式
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	// BorderStyle 边框样式
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	// CursorStyle 光标样式
	CursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)

	// NormalStyle 普通文本样式
	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	// SelectedStyle 选中项样式
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	// ErrorStyle 错误样式
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F")).
			Bold(true)

	// SuccessStyle 成功样式
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	// WarningStyle 警告样式
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true)

	// InfoStyle 信息样式
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Bold(true)

	// DimStyle 暗淡样式
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	// SubtitleStyle 副标题样式
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")).
			Bold(true)
)
