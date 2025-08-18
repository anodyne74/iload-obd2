package config

import (
	"fmt"
	"os"

	"github.com/anodyne74/iload-obd2/internal/transport"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Transport struct {
		Type     string `yaml:"type"`
		Address  string `yaml:"address"`
		BaudRate int    `yaml:"baudRate"`
		Debug    bool   `yaml:"debug"`
	} `yaml:"transport"`

	Testing struct {
		UseMockData bool   `yaml:"useMockData"`
		UseTestTCP  bool   `yaml:"useTestTCP"`
		TCPAddress  string `yaml:"tcpAddress"`
	} `yaml:"testing"`

	Capture struct {
		Enabled  bool   `yaml:"enabled"`
		Filename string `yaml:"filename"`
	} `yaml:"capture"`

	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	Datastore struct {
		SQLite struct {
			Path string `yaml:"path"`
		} `yaml:"sqlite"`
		InfluxDB struct {
			URL    string `yaml:"url"`
			Org    string `yaml:"org"`
			Bucket string `yaml:"bucket"`
			Token  string `yaml:"token"`
		} `yaml:"influxdb"`
	} `yaml:"datastore"`

	Vehicle struct {
		DefaultThresholds struct {
			RPMRedline     float64 `yaml:"rpm_redline"`
			CoolantTempMax float64 `yaml:"coolant_temp_max"`
			EngineLoadMax  float64 `yaml:"engine_load_max"`
		} `yaml:"default_thresholds"`
	} `yaml:"vehicle"`
}

// LoadConfig reads the config file and returns a Config struct
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// GetTransportConfig returns the transport configuration based on test flags and config
func (c *Config) GetTransportConfig() *transport.Config {
	if c.Testing.UseTestTCP {
		return &transport.Config{
			Type:    "tcp",
			Address: c.Testing.TCPAddress,
		}
	}
	if c.Testing.UseMockData {
		return &transport.Config{
			Type: "mock",
		}
	}
	return &transport.Config{
		Type:     c.Transport.Type,
		Address:  c.Transport.Address,
		BaudRate: c.Transport.BaudRate,
	}
}
