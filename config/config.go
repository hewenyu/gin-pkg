package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Security SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslMode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	AccessTokenSecret      string        `mapstructure:"accessTokenSecret"`
	RefreshTokenSecret     string        `mapstructure:"refreshTokenSecret"`
	AccessTokenDuration    time.Duration `mapstructure:"accessTokenDuration"`
	RefreshTokenDuration   time.Duration `mapstructure:"refreshTokenDuration"`
	EnableRegistration     bool          `mapstructure:"enableRegistration"`
	DefaultAccessTokenExp  int64         `mapstructure:"defaultAccessTokenExp"`
	DefaultRefreshTokenExp int64         `mapstructure:"defaultRefreshTokenExp"`
}

type SecurityConfig struct {
	TimestampValidityWindow time.Duration `mapstructure:"timestampValidityWindow"`
	NonceValidityDuration   time.Duration `mapstructure:"nonceValidityDuration"`
	SignatureSecret         string        `mapstructure:"signatureSecret"`
}

// Load reads configuration from file or environment variables
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults if not specified
	if config.Auth.AccessTokenDuration == 0 {
		config.Auth.AccessTokenDuration = 24 * time.Hour
	}
	if config.Auth.RefreshTokenDuration == 0 {
		config.Auth.RefreshTokenDuration = 30 * 24 * time.Hour
	}
	if config.Security.TimestampValidityWindow == 0 {
		config.Security.TimestampValidityWindow = 60 * time.Second
	}
	if config.Security.NonceValidityDuration == 0 {
		config.Security.NonceValidityDuration = 2 * time.Minute
	}
	if config.Auth.DefaultAccessTokenExp == 0 {
		config.Auth.DefaultAccessTokenExp = 86400 // 24 hours in seconds
	}
	if config.Auth.DefaultRefreshTokenExp == 0 {
		config.Auth.DefaultRefreshTokenExp = 2592000 // 30 days in seconds
	}

	return &config, nil
}
