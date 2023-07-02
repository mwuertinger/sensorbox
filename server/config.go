package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents a config file
type Config struct {
	Http                      HttpConfig     `yaml:"http"`
	Influx                    InfluxConfig   `yaml:"influx"`
	Devices                   map[int]Device `yaml:"devices"`
	Ntfy                      NtfyConfig     `yaml:"notification"`
	Battery                   BatteryConfig  `yaml:"battery"`
	TemperatureAlertLocations [][2]string    `yaml:"temperatureAlertLocations"`
}

type HttpConfig struct {
	Listen           string        `yaml:"listen"`
	ReadWriteTimeout time.Duration `yaml:"readWriteTimeout"`
	ShutdownTimeout  time.Duration `yaml:"shutdownTimeout"`
	CertFile         string        `yaml:"cert"`
	KeyFile          string        `yaml:"key"`
}

type InfluxConfig struct {
	Server  string        `yaml:"server"`
	Token   string        `yaml:"token"`
	Timeout time.Duration `yaml:"timeout"`
}

type Device struct {
	Location  string `yaml:"location"`
	AuthToken string `yaml:"authToken"`
}

type NtfyConfig struct {
	Url     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

type BatteryConfig struct {
	ThresholdVoltage float32 `yaml:"thresholdVoltage"`
}

// parseConfig reads config file at path and returns the content or an error
func parseConfig(path string) (*Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
