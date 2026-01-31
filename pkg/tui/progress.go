package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// ProgressModel 进度条模型
type ProgressModel struct {
	progress progress.Model
	message  string
	total    int
	current  int
	quitting bool
}

// NewProgress 创建新进度条
func NewProgress(message string, total int) ProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(50),
		progress.WithoutPercentage(),
	)

	return ProgressModel{
		progress: p,
		message:  message,
		total:    total,
		current:  0,
		quitting: false,
	}
}

// Init 初始化进度条
func (m ProgressModel) Init() tea.Cmd {
	return nil
}

// Update 更新进度条状态
func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressTickMsg:
		if m.current < m.total {
			m.current++
			return m, nil
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	model, cmd := m.progress.Update(msg)
	if progressModel, ok := model.(progress.Model); ok {
		m.progress = progressModel
	}
	return m, cmd
}

// View 渲染进度条
func (m ProgressModel) View() string {
	if m.quitting {
		return ""
	}

	content := NormalStyle.Render(m.message) + "\n\n"
	percent := float64(m.current) / float64(m.total)
	content += m.progress.ViewAs(percent) + "\n\n"
	content += DimStyle.Render(fmt.Sprintf("%d / %d", m.current, m.total))

	return BorderStyle.Width(60).Render(content)
}

// Increment 增加进度
func (m ProgressModel) Increment() tea.Cmd {
	return func() tea.Msg {
		return ProgressTickMsg{}
	}
}

// IsComplete 返回是否完成
func (m ProgressModel) IsComplete() bool {
	return m.current >= m.total
}

// ProgressTickMsg 进度条滴答消息
type ProgressTickMsg struct{}
