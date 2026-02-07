package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadReturnsErrorOnInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	oldConfigDir := configDir
	configDir = tmpDir
	t.Cleanup(func() { configDir = oldConfigDir })

	path := filepath.Join(tmpDir, configFile)
	require.NoError(t, os.WriteFile(path, []byte("{invalid-json"), 0644))

	cfg, err := Load()
	require.Error(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "zh_CN", cfg.Language)
}

func TestLoadReturnsDefaultWhenConfigMissing(t *testing.T) {
	tmpDir := t.TempDir()
	oldConfigDir := configDir
	configDir = tmpDir
	t.Cleanup(func() { configDir = oldConfigDir })

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "zh_CN", cfg.Language)
}
