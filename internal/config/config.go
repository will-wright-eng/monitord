package config

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

type Config struct {
    Database DatabaseConfig `json:"database"`
    Monitor  MonitorConfig `json:"monitor"`
    Logging  LogConfig     `json:"logging"`
}

type DatabaseConfig struct {
    Path string `json:"path"`
}

type MonitorConfig struct {
    Endpoints    []Endpoint     `json:"endpoints"`
    ConfigCheck  Duration   `json:"config_check_interval"`
}

type Endpoint struct {
    Name        string        `json:"name"`
    URL         string        `json:"url"`
    Interval    Duration `json:"interval"`
    Timeout     Duration `json:"timeout"`
    Description string        `json:"description,omitempty"`
    Tags        []string      `json:"tags,omitempty"`
    Enabled     bool          `json:"enabled"`
}

type LogConfig struct {
    Path  string `json:"path"`
    Level string `json:"level"`
}

// Add this custom type and methods
type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
    var v interface{}
    if err := json.Unmarshal(b, &v); err != nil {
        return err
    }

    switch value := v.(type) {
    case float64:
        *d = Duration(time.Duration(value) * time.Second)
        return nil
    case string:
        tmp, err := time.ParseDuration(value)
        if err != nil {
            return err
        }
        *d = Duration(tmp)
        return nil
    default:
        return fmt.Errorf("invalid duration type %T", v)
    }
}

// Add this method to convert Duration to time.Duration
func (d Duration) ToDuration() time.Duration {
    return time.Duration(d)
}

// Load reads configuration from the default location
func Load() (*Config, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }

    configPath := filepath.Join(homeDir, ".config/monitord/config.json")
    return LoadFromFile(configPath)
}

// LoadFromFile reads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
    fmt.Printf("Loading config from: %s\n", path)

    data, err := os.ReadFile(path)
    if err != nil {
        fmt.Printf("Error reading config file: %v\n", err)
        return nil, err
    }

    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        fmt.Printf("Error parsing config file: %v\n", err)
        return nil, err
    }

    // Handle relative paths by joining with home directory
    homeDir, err := os.UserHomeDir()
    if err != nil {
        fmt.Printf("Error getting user home directory: %v\n", err)
        return nil, err
    }

    // If database path is relative, make it absolute
    if !filepath.IsAbs(config.Database.Path) {
        config.Database.Path = filepath.Join(homeDir, config.Database.Path)
    }

    // If log path is relative, make it absolute
    if !filepath.IsAbs(config.Logging.Path) {
        config.Logging.Path = filepath.Join(homeDir, config.Logging.Path)
    }

    return &config, nil
}

// Save writes the configuration to the default location
func (c *Config) Save() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }

    configPath := filepath.Join(homeDir, ".config/monitord/config.json")
    return c.SaveToFile(configPath)
}

// SaveToFile writes the configuration to a specific file
func (c *Config) SaveToFile(path string) error {
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }

    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0644)
}
