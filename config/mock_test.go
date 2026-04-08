package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToConfig_Defaults(t *testing.T) {
	m := Mock{}
	cfg := m.ToConfig()

	assert.Equal(t, ":8080", cfg.Server.Addr)
	assert.Equal(t, 5*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 5*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 5*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, "info", cfg.Logger.Level)
}

func TestToConfig_PortOverride(t *testing.T) {
	m := Mock{Port: 9090}
	cfg := m.ToConfig()

	assert.Equal(t, ":9090", cfg.Server.Addr)
}

func TestToConfig_AddrOverridesPort(t *testing.T) {
	m := Mock{Port: 9090, Addr: ":3000"}
	cfg := m.ToConfig()

	assert.Equal(t, ":3000", cfg.Server.Addr)
}

func TestToConfig_Timeouts(t *testing.T) {
	m := Mock{
		ReadTimeout:  "10s",
		WriteTimeout: "20s",
		IdleTimeout:  "30s",
	}
	cfg := m.ToConfig()

	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 20*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.IdleTimeout)
}

func TestToConfig_InvalidTimeoutKeepsDefault(t *testing.T) {
	m := Mock{ReadTimeout: "not-a-duration"}
	cfg := m.ToConfig()

	assert.Equal(t, 5*time.Second, cfg.Server.ReadTimeout)
}

func TestToConfig_LogLevel(t *testing.T) {
	m := Mock{LogLevel: "debug"}
	cfg := m.ToConfig()

	assert.Equal(t, "debug", cfg.Logger.Level)
}

func TestApplyOverrides_Port(t *testing.T) {
	m := Mock{}
	cfg := m.ToConfig()
	cfg.ApplyOverrides(CLIOverrides{Port: 9090})

	assert.Equal(t, ":9090", cfg.Server.Addr)
}

func TestApplyOverrides_AddrOverridesPort(t *testing.T) {
	m := Mock{}
	cfg := m.ToConfig()
	cfg.ApplyOverrides(CLIOverrides{Port: 9090, Addr: ":3000"})

	assert.Equal(t, ":3000", cfg.Server.Addr)
}

func TestApplyOverrides_Timeouts(t *testing.T) {
	m := Mock{}
	cfg := m.ToConfig()
	cfg.ApplyOverrides(CLIOverrides{
		ReadTimeout:  "10s",
		WriteTimeout: "20s",
		IdleTimeout:  "30s",
	})

	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 20*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.IdleTimeout)
}

func TestApplyOverrides_InvalidTimeoutKeepsOriginal(t *testing.T) {
	m := Mock{}
	cfg := m.ToConfig()
	cfg.ApplyOverrides(CLIOverrides{ReadTimeout: "bad"})

	assert.Equal(t, 5*time.Second, cfg.Server.ReadTimeout)
}

func TestApplyOverrides_LogLevel(t *testing.T) {
	m := Mock{}
	cfg := m.ToConfig()
	cfg.ApplyOverrides(CLIOverrides{LogLevel: "debug"})

	assert.Equal(t, "debug", cfg.Logger.Level)
}

func TestApplyOverrides_VerboseOverridesLogLevel(t *testing.T) {
	m := Mock{LogLevel: "info"}
	cfg := m.ToConfig()
	cfg.ApplyOverrides(CLIOverrides{LogLevel: "info", Verbose: true})

	assert.Equal(t, "debug", cfg.Logger.Level)
}

func TestApplyOverrides_EmptyKeepsDefaults(t *testing.T) {
	m := Mock{}
	cfg := m.ToConfig()
	cfg.ApplyOverrides(CLIOverrides{})

	assert.Equal(t, ":8080", cfg.Server.Addr)
	assert.Equal(t, 5*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 5*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 5*time.Second, cfg.Server.IdleTimeout)
	assert.Equal(t, "info", cfg.Logger.Level)
}

func TestNewMock_WithServerFields(t *testing.T) {
	content := `{
		"port": 9090,
		"addr": ":3000",
		"readTimeout": "10s",
		"writeTimeout": "20s",
		"idleTimeout": "30s",
		"logLevel": "debug",
		"endpoints": [
			{
				"path": "/test",
				"json": {"ok": true}
			}
		]
	}`

	dir := t.TempDir()
	file := filepath.Join(dir, "mock.json")
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	mck, path, err := NewMock(file)
	require.NoError(t, err)
	assert.Equal(t, dir, path)
	assert.Equal(t, 9090, mck.Port)
	assert.Equal(t, ":3000", mck.Addr)
	assert.Equal(t, "10s", mck.ReadTimeout)
	assert.Equal(t, "20s", mck.WriteTimeout)
	assert.Equal(t, "30s", mck.IdleTimeout)
	assert.Equal(t, "debug", mck.LogLevel)
	assert.Len(t, mck.Endpoints, 1)
	assert.Equal(t, "/test", mck.Endpoints[0].Path)
}
