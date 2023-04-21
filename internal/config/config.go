package config

import "github.com/spf13/viper"

// Config is the top-level configuration for the server and world.
type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

// ServerConfig contains parameters for the game server.
type ServerConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	WorldID        int    `mapstructure:"worldId"`
	AssetDir       string `mapstructure:"assetDir"`
	LogLevel       string `mapstructure:"logLevel"`
	WelcomeMessage string `mapstructure:"welcomeMessage"`
}

// Load reads the game server configuration file from the given path.
func Load(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
