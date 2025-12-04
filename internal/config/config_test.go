package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func writeConfigFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0o600)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return path
}

func TestNew_Success(t *testing.T) {
	viper.Reset()
	origWd, err := os.Getwd()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() {
		_ = os.Chdir(origWd)
		viper.Reset()
	}()

	dir, err := os.MkdirTemp("", "config_test_success")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)

	content := `SERVICE_NAME=apps-service
	DEBUG=false
	DEFAULT_TIMEOUT=15
	SERVER_PORT=9090
	DATABASE_URL=postgres://user:pass@localhost/db
	REDIS_HOST=localhost:6379
	`

	cfgPath := writeConfigFile(t, dir, "config.env", content)

	viper.SetConfigFile(cfgPath)

	if !assert.NoError(t, os.Chdir(dir)) {
		t.FailNow()
	}

	v := validator.New()
	cfg, err := New(v)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, "apps-service", cfg.ServiceName)
	assert.Equal(t, false, cfg.Debug)
	assert.Equal(t, 15, cfg.DefaultTimeout)
	assert.Equal(t, 9090, cfg.ServerPort)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseURL)
}

func TestNew_UnmarshalError(t *testing.T) {
	viper.Reset()
	origWd, err := os.Getwd()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer func() {
		_ = os.Chdir(origWd)
		viper.Reset()
	}()

	dir, err := os.MkdirTemp("", "config_test_unmarshal")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer os.RemoveAll(dir)

	content := `SERVICE_NAME=apps-service
	DEBUG=false
	DEFAULT_TIMEOUT=ten
	SERVER_PORT=notanint
	DATABASE_URL=postgres://user:pass@localhost/db
	`
	cfgPath := writeConfigFile(t, dir, "config.env", content)

	viper.SetConfigFile(cfgPath)

	if !assert.NoError(t, os.Chdir(dir)) {
		t.FailNow()
	}

	v := validator.New()
	_, err = New(v)
	if !assert.Error(t, err) {
		t.FailNow()
	}
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}
