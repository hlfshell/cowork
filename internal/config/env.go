package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hlfshell/cowork/internal/secure_store"
)

// EnvScope defines the scope for environment variables
type EnvScope string

const (
	// EnvScopeProject stores environment variables in the project .cowork directory
	EnvScopeProject EnvScope = "project"
	// EnvScopeGlobal stores environment variables in the global configuration
	EnvScopeGlobal EnvScope = "global"
)

// GetEnv grabs .env files
func (m *Manager) GetEnv() map[string]string {
	keys, err := m.config.envStore.List("env")
	if err != nil {
		log.Fatalf("Failed to get environment variables: %v", err)
	}

	for _, key := range keys {
		var value string
		err := m.config.envStore.Get(key, &value)
		if err != nil {
			log.Fatalf("Failed to get environment variable: %v", err)
		}
		m.config.Env[key] = value
	}
	return m.config.Env
}

func (m *Manager) SetEnvVar(key, value string) error {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	m.config.Env[key] = value
	return m.config.envStore.Set(key, value)
}

func (m *Manager) GetEnvVar(key string) (string, error) {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return "", fmt.Errorf("failed to load config: %w", err)
		}
	}

	var value string
	err := m.config.envStore.Get(key, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (m *Manager) RemoveEnvVar(key string) error {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	delete(m.config.Env, key)
	return m.config.envStore.Delete(key)
}

func (m *Manager) SetEnvFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// We assume that each line is a key=value pair like that.
	// If it isn't, we error
	envs := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid env file format: %s", line)
		}
		envs[parts[0]] = parts[1]
	}
	for key, value := range envs {
		m.SetEnvVar(key, value)
	}
	return nil
}

func (m *Manager) GetEnvVars() (map[string]string, error) {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	keys, err := m.config.envStore.List("")
	if err != nil {
		return nil, err
	}

	envVars := make(map[string]string)
	for _, key := range keys {
		var value string
		err := m.config.envStore.Get(key, &value)
		if err != nil {
			return nil, err
		}
		envVars[key] = value
	}
	return envVars, nil
}

func (m *Manager) DeleteEnvVar(key string) error {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	return m.config.envStore.Delete(key)
}

// SetEnvVarScoped sets an environment variable with a specific scope
func (m *Manager) SetEnvVarScoped(key, value string, scope EnvScope) error {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	var envStore *secure_store.SecureStore
	var err error

	switch scope {
	case EnvScopeProject:
		if m.config.envStore == nil {
			envStore, err = secure_store.NewSecureStore(".env", m.ProjectConfigPath)
			if err != nil {
				return fmt.Errorf("failed to create project env store: %w", err)
			}
			m.config.envStore = envStore
		} else {
			envStore = m.config.envStore
		}
	case EnvScopeGlobal:
		envStore, err = secure_store.NewSecureStore(".env", m.GlobalConfigPath)
		if err != nil {
			return fmt.Errorf("failed to create global env store: %w", err)
		}
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: project, global", scope)
	}

	return envStore.Set(key, value)
}

// GetEnvVarScoped retrieves an environment variable with scope preference
func (m *Manager) GetEnvVarScoped(key string, scope EnvScope) (string, error) {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return "", fmt.Errorf("failed to load config: %w", err)
		}
	}

	var envStore *secure_store.SecureStore
	var err error

	switch scope {
	case EnvScopeProject:
		if m.config.envStore == nil {
			envStore, err = secure_store.NewSecureStore(".env", m.ProjectConfigPath)
			if err != nil {
				return "", fmt.Errorf("failed to create project env store: %w", err)
			}
			m.config.envStore = envStore
		} else {
			envStore = m.config.envStore
		}
	case EnvScopeGlobal:
		envStore, err = secure_store.NewSecureStore(".env", m.GlobalConfigPath)
		if err != nil {
			return "", fmt.Errorf("failed to create global env store: %w", err)
		}
	default:
		return "", fmt.Errorf("invalid scope: %s. Valid scopes: project, global", scope)
	}

	var value string
	err = envStore.Get(key, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// GetEnvVarsScoped retrieves all environment variables from the specified scope
func (m *Manager) GetEnvVarsScoped(scope EnvScope) (map[string]string, error) {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	var envStore *secure_store.SecureStore
	var err error

	switch scope {
	case EnvScopeProject:
		if m.config.envStore == nil {
			envStore, err = secure_store.NewSecureStore(".env", m.ProjectConfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create project env store: %w", err)
			}
			m.config.envStore = envStore
		} else {
			envStore = m.config.envStore
		}
	case EnvScopeGlobal:
		envStore, err = secure_store.NewSecureStore(".env", m.GlobalConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create global env store: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid scope: %s. Valid scopes: project, global", scope)
	}

	keys, err := envStore.List("")
	if err != nil {
		return nil, err
	}

	envVars := make(map[string]string)
	for _, key := range keys {
		var value string
		err := envStore.Get(key, &value)
		if err != nil {
			return nil, err
		}
		envVars[key] = value
	}
	return envVars, nil
}

// DeleteEnvVarScoped deletes an environment variable from the specified scope
func (m *Manager) DeleteEnvVarScoped(key string, scope EnvScope) error {
	// Ensure config is loaded
	if m.config == nil {
		if _, err := m.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	var envStore *secure_store.SecureStore
	var err error

	switch scope {
	case EnvScopeProject:
		if m.config.envStore == nil {
			envStore, err = secure_store.NewSecureStore(".env", m.ProjectConfigPath)
			if err != nil {
				return fmt.Errorf("failed to create project env store: %w", err)
			}
			m.config.envStore = envStore
		} else {
			envStore = m.config.envStore
		}
	case EnvScopeGlobal:
		envStore, err = secure_store.NewSecureStore(".env", m.GlobalConfigPath)
		if err != nil {
			return fmt.Errorf("failed to create global env store: %w", err)
		}
	default:
		return fmt.Errorf("invalid scope: %s. Valid scopes: project, global", scope)
	}

	return envStore.Delete(key)
}

// GetEnvVarWithFallback retrieves an environment variable, checking project first, then global
func (m *Manager) GetEnvVarWithFallback(key string) (string, EnvScope, error) {
	// Try project scope first
	value, err := m.GetEnvVarScoped(key, EnvScopeProject)
	if err == nil {
		return value, EnvScopeProject, nil
	}

	// If not found in project, try global
	value, err = m.GetEnvVarScoped(key, EnvScopeGlobal)
	if err == nil {
		return value, EnvScopeGlobal, nil
	}

	return "", "", fmt.Errorf("environment variable '%s' not found in project or global scope", key)
}
