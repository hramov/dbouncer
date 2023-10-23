package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

const AppVersion = "0.0.1"

type Storage struct {
	Id          int           `yaml:"id"`
	Name        string        `yaml:"name"`
	Dsn         string        `yaml:"dsn"`
	PoolMax     int           `yaml:"pool_max"`
	IdleMax     int           `yaml:"idle_max"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	LifeTime    time.Duration `yaml:"life_time"`
	Workers     int           `yaml:"workers"`
}

type Config struct {
	Version     string
	Port        int           `yaml:"port"`
	MetricsPort int           `yaml:"metrics_port"`
	Timeout     time.Duration `yaml:"timeout"`
	Storages    []Storage     `yaml:"storages"`
}

func LoadConfig(configPath string, cfg *Config) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, cfg)
	cfg.Version = AppVersion
	return err
}
