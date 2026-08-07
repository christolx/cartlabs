package config

import (
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("API_ADDR", ":9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Environment != "test" {
		t.Fatalf("Environment = %q, want test", cfg.Environment)
	}
	if cfg.APIAddress != ":9090" {
		t.Fatalf("APIAddress = %q, want :9090", cfg.APIAddress)
	}
}

func TestLoadRejectsLocalPaymentSecretsOutsideLocalEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ACCESS_TOKEN_SECRET", "01234567890123456789012345678901")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "payment secrets") {
		t.Fatalf("Load() error = %v, want payment secret rejection", err)
	}
}
