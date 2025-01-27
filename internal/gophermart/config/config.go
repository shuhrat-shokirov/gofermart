package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"run"`
	DB        DatabaseConfig  `mapstructure:"database"`
	Migration MigrationConfig `mapstructure:"migration"`
	Accrual   AccrualConfig   `mapstructure:"accrual"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
}

type DatabaseConfig struct {
	URI string `mapstructure:"uri"`
}

type MigrationConfig struct {
	URI string `mapstructure:"uri"`
	Dir string `mapstructure:"dir"`
}

type AccrualConfig struct {
	System struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"system"`
}

func Load() (*Config, error) {
	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal config: %w", err)
	}

	return &cfg, nil
}
