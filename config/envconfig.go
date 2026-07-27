// Package config loads application configuration from the environment.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// EnvConfig contains environment-driven application settings.
type EnvConfig struct {
	AppName    string
	AppEnv     string
	ServerAddr string
	ViewsRoot  string

	SessionCookie   string
	SessionLifetime time.Duration
	SessionSecure   bool

	Database DatabaseConfig
}

// Load reads an optional dotenv file and then applies operating-system
// environment variables. OS environment variables always take precedence.
func Load(path string) (EnvConfig, error) {
	fileValues, err := readDotEnv(path)
	if err != nil {
		return EnvConfig{}, err
	}

	sessionMinutes, err := integerValue("SESSION_LIFETIME_MINUTES", 120, fileValues)
	if err != nil {
		return EnvConfig{}, err
	}
	sessionSecure, err := booleanValue("SESSION_SECURE", false, fileValues)
	if err != nil {
		return EnvConfig{}, err
	}
	maxIdleConns, err := integerValue("DB_MAX_IDLE_CONNS", 10, fileValues)
	if err != nil {
		return EnvConfig{}, err
	}
	maxOpenConns, err := integerValue("DB_MAX_OPEN_CONNS", 100, fileValues)
	if err != nil {
		return EnvConfig{}, err
	}
	connectionMinutes, err := integerValue("DB_CONN_MAX_LIFETIME_MINUTES", 60, fileValues)
	if err != nil {
		return EnvConfig{}, err
	}
	autoMigrate, err := booleanValue("DB_AUTO_MIGRATE", true, fileValues)
	if err != nil {
		return EnvConfig{}, err
	}

	return EnvConfig{
		AppName:    value("APP_NAME", "FluxGo", fileValues),
		AppEnv:     value("APP_ENV", "local", fileValues),
		ServerAddr: value("SERVER_ADDR", ":8080", fileValues),
		ViewsRoot:  value("VIEWS_ROOT", "views", fileValues),

		SessionCookie:   value("SESSION_COOKIE", "flux_session", fileValues),
		SessionLifetime: time.Duration(sessionMinutes) * time.Minute,
		SessionSecure:   sessionSecure,

		Database: DatabaseConfig{
			Host:            value("DB_HOST", "127.0.0.1", fileValues),
			Port:            value("DB_PORT", "3306", fileValues),
			Name:            value("DB_DATABASE", "fluxgo", fileValues),
			User:            value("DB_USERNAME", "root", fileValues),
			Password:        value("DB_PASSWORD", "", fileValues),
			Charset:         value("DB_CHARSET", "utf8mb4", fileValues),
			MaxIdleConns:    maxIdleConns,
			MaxOpenConns:    maxOpenConns,
			ConnMaxLifetime: time.Duration(connectionMinutes) * time.Minute,
			AutoMigrate:     autoMigrate,
		},
	}, nil
}

func readDotEnv(path string) (map[string]string, error) {
	values := make(map[string]string)
	if strings.TrimSpace(path) == "" {
		return values, nil
	}

	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return values, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open environment file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, rawValue, found := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			return nil, fmt.Errorf("%s:%d: expected KEY=VALUE", path, lineNumber)
		}

		values[key] = unquote(strings.TrimSpace(rawValue))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read environment file %q: %w", path, err)
	}

	return values, nil
}

func value(key, fallback string, fileValues map[string]string) string {
	if environmentValue, exists := os.LookupEnv(key); exists {
		return environmentValue
	}
	if fileValue, exists := fileValues[key]; exists {
		return fileValue
	}
	return fallback
}

func integerValue(key string, fallback int, fileValues map[string]string) (int, error) {
	raw := value(key, strconv.Itoa(fallback), fileValues)
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}

func booleanValue(key string, fallback bool, fileValues map[string]string) (bool, error) {
	raw := value(key, strconv.FormatBool(fallback), fileValues)
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}
	return parsed, nil
}

func unquote(value string) string {
	if len(value) < 2 {
		return value
	}

	first := value[0]
	last := value[len(value)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return value[1 : len(value)-1]
	}
	return value
}
