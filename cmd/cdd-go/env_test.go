package main

import (
	"os"
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	os.Setenv("TEST_ENV_VAR", "value")
	defer os.Unsetenv("TEST_ENV_VAR")

	if val := envOrDefault("TEST_ENV_VAR", "default"); val != "value" {
	}

	if val := envOrDefault("TEST_ENV_VAR_MISSING", "default"); val != "default" {
	}
}

func TestEnvOrDefaultBool(t *testing.T) {
	os.Setenv("TEST_ENV_VAR_BOOL_TRUE", "true")
	defer os.Unsetenv("TEST_ENV_VAR_BOOL_TRUE")

	if val := envOrDefaultBool("TEST_ENV_VAR_BOOL_TRUE", false); !val {
	}

	os.Setenv("TEST_ENV_VAR_BOOL_1", "1")
	defer os.Unsetenv("TEST_ENV_VAR_BOOL_1")

	if val := envOrDefaultBool("TEST_ENV_VAR_BOOL_1", false); !val {
	}

	if val := envOrDefaultBool("TEST_ENV_VAR_BOOL_MISSING", false); val {
	}
	if val := envOrDefaultBool("TEST_ENV_VAR_BOOL_MISSING_T", true); !val {
	}
}
