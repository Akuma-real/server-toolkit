package main

import (
	"errors"
	"os"
	"testing"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/i18n"
	"github.com/Akuma-real/server-toolkit/pkg/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextLogLevel(t *testing.T) {
	assert.Equal(t, "INFO", nextLogLevel("DEBUG"))
	assert.Equal(t, "WARN", nextLogLevel("info"))
	assert.Equal(t, "ERROR", nextLogLevel("WARN"))
	assert.Equal(t, "DEBUG", nextLogLevel("ERROR"))
	assert.Equal(t, "DEBUG", nextLogLevel("invalid"))
}

func TestOnOff(t *testing.T) {
	i18n.Init()
	assert.Equal(t, i18n.T("yes"), onOff(true))
	assert.Equal(t, i18n.T("no"), onOff(false))
}

func TestSettingsApplyBackReturnsParent(t *testing.T) {
	i18n.Init()
	parent := tui.NewMenu("main", "", nil)
	cfg := internal.Default()
	logger := internal.NewLogger(internal.INFO, os.Stdout)

	model := NewSettingsModel(parent, cfg, logger, nil)
	sm, ok := model.(settingsModel)
	require.True(t, ok)

	sm.cursor = len(sm.items) - 1 // back
	next, _ := sm.applySelection()
	_, ok = next.(tui.MenuModel)
	assert.True(t, ok)
}

func TestSettingsApplySelectionDoesNotMutateOnSaveFailure(t *testing.T) {
	i18n.Init()
	originalSave := saveConfig
	t.Cleanup(func() {
		saveConfig = originalSave
	})
	saveConfig = func(*internal.Config) error {
		return errors.New("permission denied")
	}

	cfg := internal.Default()
	logger := internal.NewLogger(internal.INFO, os.Stdout)
	parent := tui.NewMenu("main", "", nil)

	sm, ok := NewSettingsModel(parent, cfg, logger, nil).(settingsModel)
	require.True(t, ok)

	if cfg.Language == "zh_CN" {
		sm.cursor = 0 // lang
	} else {
		sm.cursor = 1 // dryrun
	}

	originalLanguage := cfg.Language
	originalDryRun := cfg.DryRun

	next, _ := sm.applySelection()
	updated, ok := next.(settingsModel)
	require.True(t, ok)

	assert.Equal(t, originalLanguage, cfg.Language)
	assert.Equal(t, originalDryRun, cfg.DryRun)
	assert.NotEmpty(t, updated.status)
}

func TestSettingsApplySelectionSaveSuccessStaysInSettingsWhenRebuildExists(t *testing.T) {
	i18n.Init()
	originalSave := saveConfig
	t.Cleanup(func() {
		saveConfig = originalSave
	})
	saveConfig = func(*internal.Config) error {
		return nil
	}

	cfg := internal.Default()
	logger := internal.NewLogger(internal.INFO, os.Stdout)
	parent := tui.NewMenu("main", "", nil)

	rebuildCalled := false
	rebuild := func() tui.MenuModel {
		rebuildCalled = true
		return tui.NewMenu("main", "", nil)
	}

	sm, ok := NewSettingsModel(parent, cfg, logger, rebuild).(settingsModel)
	require.True(t, ok)

	if cfg.Language == "zh_CN" {
		sm.cursor = 0 // lang
	} else {
		sm.cursor = 1 // dryrun
	}

	next, _ := sm.applySelection()
	updated, ok := next.(settingsModel)
	require.True(t, ok)

	assert.False(t, rebuildCalled)
	assert.Equal(t, i18n.T("settings_saved"), updated.status)
}

func TestSettingsModelRefreshMsgKeepsTicker(t *testing.T) {
	i18n.Init()
	parent := tui.NewMenu("main", "", nil)
	cfg := internal.Default()
	logger := internal.NewLogger(internal.INFO, os.Stdout)

	sm, ok := NewSettingsModel(parent, cfg, logger, nil).(settingsModel)
	require.True(t, ok)

	initCmd := sm.Init()
	require.NotNil(t, initCmd)

	updatedModel, nextCmd := sm.Update(tui.RefreshMenuMsg{})
	_, ok = updatedModel.(settingsModel)
	assert.True(t, ok)
	assert.NotNil(t, nextCmd)
}
