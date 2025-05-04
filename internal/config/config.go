package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"os"
	"time"
)

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
)

type Algorithm string

const (
	RoundRobin       Algorithm = "round-robin"
	LeastConnections Algorithm = "least-connections"
)

type server struct {
	Address    string `mapstructure:"address" validate:"required"`
	HealthPath string `mapstructure:"health_path" validate:"required"`
}

type Config struct {
	Env                 Environment   `mapstructure:"env"`
	Port                int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Servers             []server      `mapstructure:"servers" validate:"required,dive"`
	Algorithm           Algorithm     `mapstructure:"algorithm" validate:"required"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval" validate:"required,gt=0"`
	HealthCheckTimeout  time.Duration `mapstructure:"health_check_timeout" validate:"required,gt=0"`
	DialTimeout         time.Duration `mapstructure:"dial_timeout" validate:"required,gt=0"`
	KeepAlive           time.Duration `mapstructure:"keep_alive" validate:"required,gt=0"`
	MaxIdleConns        int           `mapstructure:"max_idle_conns" validate:"required,gt=0"`
	MaxIdleConnsPerHost int           `mapstructure:"max_idle_conns_per_host" validate:"required,gt=0"`
	IdleConnTimeout     time.Duration `mapstructure:"idle_conn_timeout" validate:"required,gt=0"`
	ShutdownTimeout     time.Duration `mapstructure:"shutdown_timeout" validate:"required,gt=0"`
}

var configFile = "configs/config.yaml"

func Get() (*Config, error) {

	if path, exists := os.LookupEnv("CONFIG_PATH"); exists {
		configFile = path
	}

	config, err := loadConfig(configFile)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func loadConfig(file string) (*Config, error) {

	viper.SetConfigFile(file)
	viper.AutomaticEnv()
	viper.SetDefault("env", string(Development))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &config, nil
}
