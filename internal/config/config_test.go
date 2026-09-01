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

func TestLoadRejectsInvalidTraceSampleRatio(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "1.5")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "OTEL_TRACES_SAMPLER_ARG") {
		t.Fatalf("Load() error = %v, want trace sample rejection", err)
	}
}

func TestLoadRejectsShortSearchServiceToken(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("SEARCH_SERVICE_TOKEN", "short")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "SEARCH_SERVICE_TOKEN") {
		t.Fatalf("Load() error = %v, want search token rejection", err)
	}
}

func TestLoadRejectsShortTrustedProxyToken(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("TRUSTED_PROXY_TOKEN", "short")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "TRUSTED_PROXY_TOKEN") {
		t.Fatalf("Load() error = %v, want trusted proxy token rejection", err)
	}
}
