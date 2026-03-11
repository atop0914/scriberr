package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Server   ServerConfig   `yaml:"server"`
	Log      LogConfig      `yaml:"log"`
	Model    ModelConfig    `yaml:"model"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

type ModelConfig struct {
	Size       string `yaml:"size"`       // tiny, base, small, medium, large, large-v2, large-v3
	CacheDir   string `yaml:"cacheDir"`   // Custom cache directory
	MaxRetries int    `yaml:"maxRetries"` // Max download retries
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"certFile"`
	KeyFile  string `yaml:"keyFile"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // json, text
}

func Default() *Config {
	return &Config{
		App: AppConfig{
			Name:        "scriberr",
			Version:     "1.0.0",
			Environment: "development",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "",
			Database: "scriberr",
			SSLMode:  "disable",
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
		Model: ModelConfig{
			Size:       "base",
			MaxRetries: 3,
		},
	}
}

func Load() (*Config, error) {
	// Check for config file path from environment
	configPath := os.Getenv("SCIBERR_CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Try to read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default if no config file
			return Default(), nil
		}
		return nil, err
	}

	// Parse YAML
	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
