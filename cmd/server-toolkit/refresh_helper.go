package main

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akuma-real/server-toolkit/pkg/tui"
)

func initRefreshTickerCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return tui.RefreshMenuTickerCmd()
	}
	return tea.Batch(cmd, tui.RefreshMenuTickerCmd())
}

func keepRefreshTickerCmd(msg tea.Msg, cmd tea.Cmd) tea.Cmd {
	if _, ok := msg.(tui.RefreshMenuMsg); ok {
		if cmd == nil {
			return tui.RefreshMenuTickerCmd()
		}
		return tea.Batch(cmd, tui.RefreshMenuTickerCmd())
	}
	return cmd
}
