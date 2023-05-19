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
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	WorldID        int    `mapstructure:"worldId"`
	AssetDir       string `mapstructure:"assetDir"`
	LogLevel       string `mapstructure:"logLevel"`
	WelcomeMessage string `mapstructure:"welcomeMessage"`
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
	FriendList        SimpleInterfaceConfig       `mapstructure:"friendList"`
	IgnoreList        SimpleInterfaceConfig       `mapstructure:"ignoreList"`
	Logout            SimpleInterfaceConfig       `mapstructure:"logout"`
	Inventory         InventoryTabInterfaceConfig `mapstructure:"inventory"`
	Skills            SimpleInterfaceConfig       `mapstructure:"skills"`
	Weapon            WeaponTabInterfaceConfig    `mapstructure:"weapon"`
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
	ID      int `mapstructure:"id"`
	Slots   int `mapstructure:"slots"`
	Bonuses struct {
		Attack  EquipmentTabBonusInterfaceConfig `mapstructure:"attack"`
		Defense EquipmentTabBonusInterfaceConfig `mapstructure:"defense"`
	} `mapstructure:"bonuses"`
	Other struct {
		Strength int `mapstructure:"strength"`
		Prayer   int `mapstructure:"prayer"`
	} `mapstructure:"other"`
}

// EquipmentTabBonusInterfaceConfig contains interface data for the equipment bonuses parent interface.
type EquipmentTabBonusInterfaceConfig struct {
	Stab  int `mapstructure:"stab"`
	Slash int `mapstructure:"slash"`
	Crush int `mapstructure:"crush"`
	Magic int `mapstructure:"magic"`
	Range int `mapstructure:"range"`
}

// WeaponTabInterfaceConfig contains interface data for equipped weapon interfaces.
type WeaponTabInterfaceConfig struct {
	TwoHandedSword int `mapstructure:"2hSword"`
	Axe            int `mapstructure:"axe"`
	Bow            int `mapstructure:"bow"`
	Blunt          int `mapstructure:"blunt"`
	Claws          int `mapstructure:"claws"`
	Crossbow       int `mapstructure:"crossbow"`
	Gun            int `mapstructure:"gun"`
	Pickaxe        int `mapstructure:"pickaxe"`
	PoleArm        int `mapstructure:"polearm"`
	PoleStaff      int `mapstructure:"polestaff"`
	Scythe         int `mapstructure:"scythe"`
	SlashSword     int `mapstructure:"slashSword"`
	Spear          int `mapstructure:"spear"`
	Spiked         int `mapstructure:"spiked"`
	StabSword      int `mapstructure:"stabSword"`
	Staff          int `mapstructure:"staff"`
	Thrown         int `mapstructure:"thrown"`
	Whip           int `mapstructure:"whip"`
	Unarmed        int `mapstructure:"unarmed"`
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
