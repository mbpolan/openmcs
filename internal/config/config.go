package config

import "github.com/spf13/viper"

// Config is the top-level configuration for the server and world.
type Config struct {
	Store      StoreConfig      `mapstructure:"store"`
	Server     ServerConfig     `mapstructure:"server"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	Interfaces InterfacesConfig `mapstructure:"interfaces"`
}

// ServerConfig contains parameters for the game server.
type ServerConfig struct {
	Host                     string `mapstructure:"host"`
	Port                     int    `mapstructure:"port"`
	WorldID                  int    `mapstructure:"worldId"`
	AssetDir                 string `mapstructure:"assetDir"`
	ScriptsDir               string `mapstructure:"scriptsDir"`
	LogLevel                 string `mapstructure:"logLevel"`
	WelcomeMessage           string `mapstructure:"welcomeMessage"`
	PlayerMaxIdleTimeSeconds int    `mapstructure:"playerMaxIdleTimeSeconds"`
}

// StoreConfig contains parameters for the backend database.
type StoreConfig struct {
	Driver        string                 `mapstructure:"driver"`
	MigrationsDir string                 `mapstructure:"migrationsDir"`
	SQLite3       *SQLite3DatabaseConfig `mapstructure:"sqlite3"`
}

// SQLite3DatabaseConfig contains parameters for a SQLIte3 database.
type SQLite3DatabaseConfig struct {
	URI string `mapstructure:"uri"`
}

// MetricsConfig contains metrics and telemetry configuration options.
type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

// InterfacesConfig contains data for client-side interface.
type InterfacesConfig struct {
	CharacterDesigner SimpleInterfaceConfig       `mapstructure:"characterDesigner"`
	Equipment         EquipmentTabInterfaceConfig `mapstructure:"equipment"`
	Inventory         InventoryTabInterfaceConfig `mapstructure:"inventory"`
}

// SimpleInterfaceConfig contains data for a simple tab interface.
type SimpleInterfaceConfig struct {
	ID int `mapstructure:"id"`
}

// InventoryTabInterfaceConfig contains interface data for the inventory tab interface.
type InventoryTabInterfaceConfig struct {
	ID    int `mapstructure:"id"`
	Slots int `mapstructure:"slots"`
}

// EquipmentTabInterfaceConfig contains interface data for the equipment tab interface.
type EquipmentTabInterfaceConfig struct {
	Slots int `mapstructure:"slots"`
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
