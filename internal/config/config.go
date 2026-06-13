package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Keycloak KeycloakConfig
}

type AppConfig struct {
	Port     string
	Env      string
	LogLevel string
}

type KeycloakConfig struct {
	URL           string // internal base URL used for token/JWKS/admin calls (must be reachable)
	IssuerURL     string // external base URL that appears in token "iss"; defaults to URL
	Realm         string // realm name, e.g. lms
	ClientID      string // confidential client used by the backend
	ClientSecret  string // client secret for admin/token operations
	AdminUser     string // master-realm admin used for user management
	AdminPassword string
}

// Issuer is the expected "iss" claim for tokens of this realm. It is derived
// from IssuerURL, which may differ from URL when Keycloak is reached on a
// different address than the one baked into its tokens (e.g. inside Docker).
func (k KeycloakConfig) Issuer() string {
	return fmt.Sprintf("%s/realms/%s", k.IssuerURL, k.Realm)
}

// CertsURL is the JWKS endpoint serving the realm's public signing keys. It
// uses the internal URL so the backend can always reach it.
func (k KeycloakConfig) CertsURL() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", k.URL, k.Realm)
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Port:     getEnv("APP_PORT", "8080"),
			Env:      getEnv("APP_ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", ""),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "lms"),
			Password: getEnv("POSTGRES_PASSWORD", "lms_password"),
			DBName:   getEnv("POSTGRES_DB", "lms_db"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Keycloak: KeycloakConfig{
			URL:           getEnv("KEYCLOAK_URL", "http://localhost:8081"),
			IssuerURL:     getEnv("KEYCLOAK_ISSUER_URL", ""),
			Realm:         getEnv("KEYCLOAK_REALM", "lms"),
			ClientID:      getEnv("KEYCLOAK_CLIENT_ID", "lms-backend"),
			ClientSecret:  getEnv("KEYCLOAK_CLIENT_SECRET", ""),
			AdminUser:     getEnv("KEYCLOAK_ADMIN_USERNAME", "admin"),
			AdminPassword: getEnv("KEYCLOAK_ADMIN_PASSWORD", "admin"),
		},
	}

	// When the issuer URL is not set explicitly, tokens are issued and
	// validated on the same address (typical for host-run app).
	if cfg.Keycloak.IssuerURL == "" {
		cfg.Keycloak.IssuerURL = cfg.Keycloak.URL
	}

	return cfg, nil
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
