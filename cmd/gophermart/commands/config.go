package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var (
	configFile string
	runAddress string
	dbURI      string
	accrualSys string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to the configuration file")
	rootCmd.PersistentFlags().StringVarP(&runAddress, "a", "a", "", "Address and port to run the service (env: RUN_ADDRESS)")
	rootCmd.PersistentFlags().StringVarP(&dbURI, "d", "d", "", "Database URI (env: DATABASE_URI)")
	rootCmd.PersistentFlags().StringVarP(&accrualSys, "r", "r", "", "Accrual system address (env: ACCRUAL_SYSTEM_ADDRESS)")

	err := viper.BindPFlag("run.address", rootCmd.PersistentFlags().Lookup("a"))
	if err != nil {
		fmt.Printf("Error binding flag: %v\n", err)
	}

	err = viper.BindPFlag("database.uri", rootCmd.PersistentFlags().Lookup("d"))
	if err != nil {
		fmt.Printf("Error binding flag: %v\n", err)
	}

	err = viper.BindPFlag("accrual.system.address", rootCmd.PersistentFlags().Lookup("r"))
	if err != nil {
		fmt.Printf("Error binding flag: %v\n", err)
	}
}

func loadConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config file: %v\n", err)
		} else {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.WatchConfig()
}
