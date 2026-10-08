package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	ConfigType        = "yml"
	DefaultConfigPath = "/run/secrets"
	DefaultConfigName = "cluer"
	EnvPrefix         = "cluer"
)

type (
	Config struct {
		ServerConfig   *ServerConfig   `mapstructure:"server" validate:"required"`
		PostgresConfig *PostgresConfig `mapstructure:"postgres" validate:"omitempty"`
		Logger         LoggerConfig    `mapstructure:"logger" validate:"omitempty"`
	}

	ServerConfig struct {
		Address string `mapstructure:"address" validate:"required,hostname_port"`
	}

	PostgresConfig struct {
		Address  string `mapstructure:"address" validate:"required,hostname_port"`
		User     string `mapstructure:"user" validate:"required"`
		Password string `mapstructure:"password" validate:"required"`
		DBName   string `mapstructure:"dbname" validate:"required"`
		SSLMode  string `mapstructure:"sslmode" validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full"`
	}

	LoggerConfig struct {
		Default ComponentConfig            `mapstructure:"default"`
		Modules map[string]ComponentConfig `mapstructure:"modules"`
	}

	ComponentConfig struct {
		Level string `mapstructure:"level" validate:"required,oneof=trace debug info warn error fatal panic"`
	}
)

func (c *Config) GetLoggerConfig(module string) ComponentConfig {
	if loggerCfg, ok := c.Logger.Modules[module]; ok {
		return loggerCfg
	}
	log.Warn().Str("module", module).Msg("Not found logger config for module")
	return c.Logger.Default
}

func (pc *PostgresConfig) GetDSN() string {
	sslMode := pc.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(pc.User, pc.Password),
		Host:     pc.Address,
		Path:     "/" + pc.DBName,
		RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
	}

	return dsn.String()
}

func (pc *PostgresConfig) SafeDSN() string {
	return fmt.Sprintf("postgres://%s:***@%s/%s", pc.User, pc.Address, pc.DBName)
}

func Load() *Config {
	yamlConfig := viper.New()

	yamlConfig.AutomaticEnv()
	yamlConfig.SetEnvPrefix(EnvPrefix)
	yamlConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	yamlConfig.SetDefault("config.name", DefaultConfigName)
	yamlConfig.SetDefault("config.path", DefaultConfigPath)
	yamlConfig.SetDefault("config.type", ConfigType)

	yamlConfig.AddConfigPath(yamlConfig.GetString("config.path"))
	yamlConfig.SetConfigType(yamlConfig.GetString("config.type"))
	yamlConfig.SetConfigName(yamlConfig.GetString("config.name"))

	if err := yamlConfig.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("Failed to read config")
		return nil
	}

	config := new(Config)
	if err := yamlConfig.Unmarshal(&config); err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal config")
		return nil
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		log.Fatal().Err(err).Msg("Failed to validate config")
		return nil
	}

	return config
}
