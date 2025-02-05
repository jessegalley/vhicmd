package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
        v.SetConfigType("yaml")
    } else {
        // Get original user's home directory via SUDO_USER env var
        var home string
        if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
            out, err := exec.Command("getent", "passwd", sudoUser).Output()
            if err == nil {
                home = strings.Split(string(out), ":")[5]
            }
        }

        // Fallback to regular UserHomeDir if not running with sudo
        if home == "" {
            var err error
            home, err = os.UserHomeDir()
            if err != nil {
                return nil, err
            }
        }

        v.SetConfigFile(filepath.Join(home, ".vhirc"))
        v.SetConfigType("yaml")
    }

	// touch the file if it doesn't exist, chmod 600
	if _, err := os.Stat(v.ConfigFileUsed()); os.IsNotExist(err) {
		f, err := os.Create(v.ConfigFileUsed())
		if err != nil {
			return nil, err
		}
		f.Chmod(0600)
		f.Close()
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
