package main

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func setEnv(t *testing.T, values map[string]string) func() {
	t.Helper()
	keys := []string{"KVETCH_LOG_LEVEL", "IN_MEMORY", "DATASTORE", "PORT", "PROMETHEUS_PORT", "GARBAGE_COLLECTION_INTERVAL"}
	previous := map[string]*string{}
	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			previous[key] = &value
		} else {
			previous[key] = nil
		}
		os.Unsetenv(key)
	}
	for key, value := range values {
		os.Setenv(key, value)
	}

	return func() {
		for key, value := range previous {
			if value == nil {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, *value)
			}
		}
	}
}

func Test_LogLevel_DefaultsToWarn(t *testing.T) {
	defer setEnv(t, map[string]string{"DATASTORE": "/tmp/kvetch"})()

	s, err := getSettingsFromEnv()
	assert.NilError(t, err)
	assert.Equal(t, s.LogLevel, "warn")
}

func Test_LogLevel_IsNormalized(t *testing.T) {
	cases := map[string]string{
		"debug":   "debug",
		" INFO ":  "info",
		"Warning": "warn",
		"warn":    "warn",
		"ERROR":   "error",
	}

	for input, expected := range cases {
		func() {
			defer setEnv(t, map[string]string{"DATASTORE": "/tmp/kvetch", "KVETCH_LOG_LEVEL": input})()

			s, err := getSettingsFromEnv()
			assert.NilError(t, err, input)
			assert.Equal(t, s.LogLevel, expected, input)
		}()
	}
}

func Test_LogLevel_RejectsUnknownLevel(t *testing.T) {
	defer setEnv(t, map[string]string{"DATASTORE": "/tmp/kvetch", "KVETCH_LOG_LEVEL": "verbose"})()

	_, err := getSettingsFromEnv()
	assert.ErrorContains(t, err, "KVETCH_LOG_LEVEL")
}
