package config

import (
	"fmt"
	"log"
	"os"
	"strings"
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
	m.config.Env[key] = value
	return m.config.envStore.Set(key, value)
}

func (m *Manager) GetEnvVar(key string) (string, error) {
	var value string
	err := m.config.envStore.Get(key, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (m *Manager) RemoveEnvVar(key string) error {
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
	return m.config.envStore.Delete(key)
}
