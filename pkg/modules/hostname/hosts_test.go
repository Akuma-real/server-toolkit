package hostname

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateHostsInsertAfterNoPanicAndInsert(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	content := "127.0.0.1 localhost\n::1 localhost ip6-localhost ip6-loopback\n"
	require.NoError(t, os.WriteFile(hostsPath, []byte(content), 0644))

	oldHostsFile := hostsFile
	oldBackupFn := backupFileFn
	oldSafeWriteFn := safeWriteFn
	hostsFile = hostsPath
	backupFileFn = func(string) (string, error) { return "", nil }
	safeWriteFn = func(path string, data []byte, perm os.FileMode) error {
		return os.WriteFile(path, data, perm)
	}
	t.Cleanup(func() {
		hostsFile = oldHostsFile
		backupFileFn = oldBackupFn
		safeWriteFn = oldSafeWriteFn
	})

	logger := internal.NewLogger(internal.INFO, os.Stdout)
	assert.NotPanics(t, func() {
		err := UpdateHosts("old", "new-host", "", InsertAfter, false, logger)
		require.NoError(t, err)
	})

	out, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	assert.Contains(t, string(out), "127.0.1.1 new-host")
	assert.Contains(t, string(out), "127.0.0.1 localhost\n127.0.1.1 new-host")
}

func TestUpdateHostsReplace127Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	content := "127.0.0.1 localhost\n127.0.1.1 old old.example.com\n"
	require.NoError(t, os.WriteFile(hostsPath, []byte(content), 0644))

	oldHostsFile := hostsFile
	oldBackupFn := backupFileFn
	oldSafeWriteFn := safeWriteFn
	hostsFile = hostsPath
	backupFileFn = func(string) (string, error) { return "", nil }
	safeWriteFn = func(path string, data []byte, perm os.FileMode) error {
		return os.WriteFile(path, data, perm)
	}
	t.Cleanup(func() {
		hostsFile = oldHostsFile
		backupFileFn = oldBackupFn
		safeWriteFn = oldSafeWriteFn
	})

	logger := internal.NewLogger(internal.INFO, os.Stdout)
	require.NoError(t, UpdateHosts("old", "new-host", "new-host.example.com", Replace127, false, logger))
	require.NoError(t, UpdateHosts("old", "new-host", "new-host.example.com", Replace127, false, logger))

	out, err := os.ReadFile(hostsPath)
	require.NoError(t, err)
	text := string(out)
	assert.Equal(t, 1, countLinesWithPrefix(text, "127.0.1.1"))
	assert.Contains(t, text, "127.0.1.1 new-host.example.com new-host")
}

func countLinesWithPrefix(content, prefix string) int {
	count := 0
	start := 0
	for start < len(content) {
		end := start
		for end < len(content) && content[end] != '\n' {
			end++
		}
		line := content[start:end]
		if len(line) >= len(prefix) && line[:len(prefix)] == prefix {
			count++
		}
		if end == len(content) {
			break
		}
		start = end + 1
	}
	return count
}
