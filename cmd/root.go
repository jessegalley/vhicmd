/*
Copyright © 2024 jesse galley jesse.galley@gmail.com
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/jessegalley/vhicmd/api"
	"github.com/jessegalley/vhicmd/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var (
	rootCmd = &cobra.Command{
		Use:          "vhicmd",
		Short:        "A command line utility for calling the VHI compute API",
		Long:         ``,
		SilenceUsage: true,
	}

	cfgFile   string
	tok       api.Token
	debugMode bool
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug mode")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	rootCmd.PersistentFlags().StringP("host", "H", "", "VHI host to connect to")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vhirc)")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		debug, _ := cmd.Flags().GetBool("debug")
		viper.Set("debug", debug)
		debugMode = debug
		hostFlag, _ := cmd.Flags().GetString("host")
		host := hostFlag
		if host == "" {
			host = viper.GetString("host")
		}
		if host == "" {
			return fmt.Errorf("no host found in flags or config. Provide --host or set 'host' in .vhirc")
		}

		// If this is the "auth" command, skip the token loading
		if cmd.Name() == "auth" || (cmd.Parent() != nil && cmd.Parent().Name() == "config") {
			return nil
		}

		var err error
		tok, err = api.LoadTokenStruct(host)
		if err != nil {
			if err.Error() == "token for "+host+" is expired" {
				return fmt.Errorf("the auth token for '%s' is expired; re-authenticate using 'vhicmd auth'", host)
			}
			return fmt.Errorf("no valid auth token found on disk for host '%s'; run 'vhicmd auth' first", host)
		}

		return nil
	}
}

func initConfig() {
	viper.AutomaticEnv()
	v, err := config.InitConfig(cfgFile)
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
	}
	// Store viper instance if needed
	viper.Reset()
	*viper.GetViper() = *v
}
