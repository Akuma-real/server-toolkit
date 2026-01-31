package system

import (
	"testing"

	"github.com/Akuma-real/server-toolkit/internal"
	"github.com/stretchr/testify/assert"
)

func TestDetermineFamily(t *testing.T) {
	tests := []struct {
		id     string
		family DistroFamily
	}{
		{"debian", Debian},
		{"ubuntu", Debian},
		{"linuxmint", Debian},
		{"rhel", RedHat},
		{"centos", RedHat},
		{"fedora", RedHat},
		{"almalinux", RedHat},
		{"rocky", RedHat},
		{"arch", Arch},
		{"alpine", Alpine},
		{"gentoo", Gentoo},
		{"unknown", Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			assert.Equal(t, tt.family, determineFamily(tt.id))
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		level  string
		loglvl int
	}{
		{"DEBUG", 0},
		{"INFO", 1},
		{"WARN", 2},
		{"ERROR", 3},
		{"", 1},
		{"invalid", 1},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := internal.ParseLevel(tt.level)
			assert.Equal(t, tt.loglvl, int(result))
		})
	}
}
