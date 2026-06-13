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
	URL           string // base URL, e.g. http://localhost:8081
	Realm         string // realm name, e.g. lms
	ClientID      string // confidential client used by the backend
	ClientSecret  string // client secret for admin/token operations
	AdminUser     string // master-realm admin used for user management
	AdminPassword string
}

// Issuer is the expected "iss" claim and JWKS base for the realm.
func (k KeycloakConfig) Issuer() string {
	return fmt.Sprintf("%s/realms/%s", k.URL, k.Realm)
}

// CertsURL is the JWKS endpoint serving the realm's public signing keys.
func (k KeycloakConfig) CertsURL() string {
	return k.Issuer() + "/protocol/openid-connect/certs"
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
			Realm:         getEnv("KEYCLOAK_REALM", "lms"),
			ClientID:      getEnv("KEYCLOAK_CLIENT_ID", "lms-backend"),
			ClientSecret:  getEnv("KEYCLOAK_CLIENT_SECRET", ""),
			AdminUser:     getEnv("KEYCLOAK_ADMIN_USERNAME", "admin"),
			AdminPassword: getEnv("KEYCLOAK_ADMIN_PASSWORD", "admin"),
		},
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
