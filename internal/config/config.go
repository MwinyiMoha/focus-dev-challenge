package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/mwinyimoha/commons/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceName    string `mapstructure:"SERVICE_NAME" validate:"required"`
	AppTier        string `mapstructure:"APP_TIER" validate:"required,oneof=web worker"`
	Debug          bool   `mapstructure:"DEBUG"`
	DefaultTimeout int    `mapstructure:"DEFAULT_TIMEOUT" validate:"required,gt=0"`
	ServerPort     int    `mapstructure:"SERVER_PORT" validate:"required"`
	DatabaseURL    string `mapstructure:"DATABASE_URL" validate:"required"`
	RedisHost      string `mapstructure:"REDIS_HOST" validate:"required"`
	RedisDB        int    `mapstructure:"REDIS_DB"`
	DefaultQueue   string `mapstructure:"DEFAULT_QUEUE"`
}

func New(val *validator.Validate) (*Config, error) {
	v := viper.New()
	v.SetConfigType("env")

	v.SetDefault("SERVICE_NAME", "")
	v.SetDefault("APP_TIER", "web")
	v.SetDefault("DEBUG", true)
	v.SetDefault("DEFAULT_TIMEOUT", 10)
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("DATABASE_URL", "")
	v.SetDefault("REDIS_HOST", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("DEFAULT_QUEUE", "tasks")

	v.AutomaticEnv()

	v.AddConfigPath("./")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.WrapError(err, errors.Internal, "failed to load configuration file")
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.WrapError(err, errors.Internal, "failed to unmarshal config")
	}

	if err := cfg.validate(val); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return errors.WrapError(err, errors.InvalidArgument, "invalid config")
	}

	return nil
}
