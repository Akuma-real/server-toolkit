package main

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/i18n"
	"github.com/Akuma-real/server-toolkit/pkg/tui"
	"github.com/stretchr/testify/require"
)

func TestWizardModelsKeepRefreshTicker(t *testing.T) {
	i18n.Init()
	parent := tui.NewMenu("main", "", nil)
	cfg := internal.Default()
	logger := internal.NewLogger(internal.INFO, os.Stdout)

	models := []interface {
		Init() tea.Cmd
		Update(tea.Msg) (tea.Model, tea.Cmd)
	}{
		NewHostnameWizard(parent, cfg, logger, true, true),
		NewCloudInitPreserveModel(parent, cfg, logger),
		NewSSHInstallKeysWizard(parent, cfg, logger),
		NewSSHListKeysModel(parent, cfg, logger),
		NewSSHDisablePasswordModel(parent, cfg, logger),
	}

	for _, model := range models {
		require.NotNil(t, model.Init())
		_, cmd := model.Update(tui.RefreshMenuMsg{})
		require.NotNil(t, cmd)
	}
}
