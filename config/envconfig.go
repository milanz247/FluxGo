// Package config loads application configuration from the environment.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// EnvConfig contains environment-driven application settings.
type EnvConfig struct {
	AppName    string
	AppEnv     string
	ServerAddr string
	ViewsRoot  string
}

// Load reads an optional dotenv file and then applies operating-system
// environment variables. OS environment variables always take precedence.
func Load(path string) (EnvConfig, error) {
	fileValues, err := readDotEnv(path)
	if err != nil {
		return EnvConfig{}, err
	}

	return EnvConfig{
		AppName:    value("APP_NAME", "FluxGo", fileValues),
		AppEnv:     value("APP_ENV", "local", fileValues),
		ServerAddr: value("SERVER_ADDR", ":8080", fileValues),
		ViewsRoot:  value("VIEWS_ROOT", "views", fileValues),
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
