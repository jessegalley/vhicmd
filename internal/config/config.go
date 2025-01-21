package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Domain   string `mapstructure:"domain"`
	Project  string `mapstructure:"project"`
	Networks string `mapstructure:"networks"`
	FlavorID string `mapstructure:"flavor_id"`
	ImageID  string `mapstructure:"image_id"`
}

func InitConfig(cfgFile string) (*viper.Viper, error) {
	v := viper.New()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		v.SetConfigFile(filepath.Join(home, ".vhirc"))
		v.SetConfigType("yaml")
	}

	// Environment
	v.SetEnvPrefix("VHI")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return v, nil
		}
		return nil, err
	}

	return v, nil
}

func LoadConfig() (*Config, error) {
	v, err := InitConfig("")
	if err != nil {
		return nil, err
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	v := viper.New()
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".vhirc")
	v.SetConfigFile(configPath)

	v.Set("host", config.Host)
	v.Set("username", config.Username)
	v.Set("password", config.Password)
	v.Set("domain", config.Domain)
	v.Set("project", config.Project)
	v.Set("networks", config.Networks)
	v.Set("flavor_id", config.FlavorID)
	v.Set("image_id", config.ImageID)

	return v.WriteConfig()
}
