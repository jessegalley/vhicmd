package cmd

import (
	"fmt"
	"strings"

	"github.com/jessegalley/vhicmd/internal/config"
	"github.com/spf13/cobra"
)

// Available config keys
var configKeys = []string{
	"host",
	"username",
	"password",
	"domain",
	"project",
	"networks",
	"flavor_id",
	"image_id",
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage vhicmd configuration",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all config values",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config file flag from parent
		configFile, _ := cmd.Flags().GetString("config")
		v, err := config.InitConfig(configFile)
		if err != nil {
			return err
		}

		settings := v.AllSettings()
		for _, key := range configKeys {
			value := settings[key]
			if value == nil || value == "" {
				fmt.Printf("%s: UNSET\n", key)
			} else if key == "password" && value != "" {
				fmt.Printf("%s: ********\n", key)
			} else {
				fmt.Printf("%s: %v\n", key, value)
			}
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set key value",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		key := strings.ToLower(args[0])
		value := args[1]

		// Validate key exists
		valid := false
		for _, validKey := range configKeys {
			if validKey == key {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid config key: %s", key)
		}

		v, err := config.InitConfig(configFile)
		if err != nil {
			return err
		}

		v.Set(key, value)
		if err := v.WriteConfig(); err != nil {
			if err := v.SafeWriteConfig(); err != nil {
				return fmt.Errorf("error writing config: %v", err)
			}
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get key",
	Short: "Get a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		key := strings.ToLower(args[0])

		// Validate key exists
		valid := false
		for _, validKey := range configKeys {
			if validKey == key {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid config key: %s", key)
		}

		v, err := config.InitConfig(configFile)
		if err != nil {
			return err
		}

		value := v.Get(key)
		if value == nil || value == "" {
			fmt.Printf("%s: UNSET\n", key)
		} else if key == "password" {
			fmt.Printf("%s: ********\n", key)
		} else {
			fmt.Printf("%s: %v\n", key, value)
		}
		return nil
	},
}

func init() {
	configCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.vhirc)")

	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}
