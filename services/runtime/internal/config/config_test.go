package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var configEnvKeys = []string{
	"HTTP_ADDR",
	"CONTROL_PLANE_GRPC_ADDR",
	"INTERNAL_API_KEY",
	"DATABASE_URL",
	"JWT_SECRET",
	"JWT_ACCESS_TTL",
	"JWT_REFRESH_TTL",
	"CONFIG_CACHE_TTL",
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range configEnvKeys {
		t.Setenv(key, "")
	}
}

func TestLoadUsesDefaultsAndRequiredValues(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("INTERNAL_API_KEY", "internal-key")
	t.Setenv("DATABASE_URL", "postgres://example")

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, ":8081", cfg.HTTPAddr)
	require.Equal(t, "localhost:9091", cfg.ControlPlaneGRPCAddr)
	require.Equal(t, "change-me-runtime", cfg.JWTSecret)
	require.Equal(t, 15*time.Minute, cfg.JWTAccessTTL)
	require.Equal(t, 7*24*time.Hour, cfg.JWTRefreshTTL)
	require.Equal(t, 30*time.Second, cfg.ConfigCacheTTL)
}

func TestLoadReadsOverridesAndFallsBackForInvalidDuration(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("INTERNAL_API_KEY", "internal-key")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9000")
	t.Setenv("CONTROL_PLANE_GRPC_ADDR", "control-plane:5000")
	t.Setenv("JWT_SECRET", "custom-secret")
	t.Setenv("JWT_ACCESS_TTL", "45m")
	t.Setenv("JWT_REFRESH_TTL", "invalid")
	t.Setenv("CONFIG_CACHE_TTL", "2m")

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:9000", cfg.HTTPAddr)
	require.Equal(t, "control-plane:5000", cfg.ControlPlaneGRPCAddr)
	require.Equal(t, "custom-secret", cfg.JWTSecret)
	require.Equal(t, 45*time.Minute, cfg.JWTAccessTTL)
	require.Equal(t, 7*24*time.Hour, cfg.JWTRefreshTTL)
	require.Equal(t, 2*time.Minute, cfg.ConfigCacheTTL)
}

func TestLoadRequiresInternalKeyAndDatabaseURL(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("DATABASE_URL", "postgres://example")

	_, err := Load()
	require.EqualError(t, err, "INTERNAL_API_KEY is required")

	t.Setenv("INTERNAL_API_KEY", "internal-key")
	t.Setenv("DATABASE_URL", "")
	_, err = Load()
	require.EqualError(t, err, "DATABASE_URL is required")
}
