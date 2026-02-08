package main

import (
	"io"
	"os"
	"testing"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/Akuma-real/server-toolkit/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateStatusSetAndGet(t *testing.T) {
	setUpdateStatus(updateStatus{Available: true, Latest: "v1.2.3"})
	got := getUpdateStatus()
	assert.True(t, got.Available)
	assert.Equal(t, "v1.2.3", got.Latest)
}

func TestCancelAsyncUpdateCheckClearsStatus(t *testing.T) {
	setUpdateStatus(updateStatus{Available: true, Latest: "v9.9.9"})
	cancelAsyncUpdateCheck()
	got := getUpdateStatus()
	assert.False(t, got.Available)
	assert.Empty(t, got.Latest)
	assert.False(t, got.CheckFailed)
}

func TestSetUpdateStatusIfGeneration(t *testing.T) {
	updateStateMu.Lock()
	currentUpdateSnapshot = updateStateSnapshot{generation: 7, status: updateStatus{}}
	updateStateMu.Unlock()

	ok := setUpdateStatusIfGeneration(6, updateStatus{Available: true, Latest: "v1"})
	assert.False(t, ok)

	got := getUpdateStatus()
	assert.False(t, got.Available)

	ok = setUpdateStatusIfGeneration(7, updateStatus{Available: true, Latest: "v2"})
	assert.True(t, ok)

	got = getUpdateStatus()
	assert.True(t, got.Available)
	assert.Equal(t, "v2", got.Latest)
}

func TestBuildMainMenuHasInitRefreshCmd(t *testing.T) {
	i18n.Init()
	cfg := internal.Default()
	logger := internal.NewLogger(internal.INFO, os.Stdout)

	menu := buildMainMenu(cfg, logger)
	cmd := menu.Init()
	require.NotNil(t, cmd)
}

func TestNewUpdateCheckLoggerIsSilentForInfoWarn(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = r.Close()
	})

	oldStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	logger := newUpdateCheckLogger()
	require.NotNil(t, logger)

	logger.Info("info log should be silent")
	logger.Warn("warn log should be silent")

	require.NoError(t, w.Close())
	content, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Empty(t, content)
}
