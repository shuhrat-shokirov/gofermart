package commands

import (
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gofermart/internal/gophermart/config"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		migrationLoadConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		load, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if err := runMigrations(load.Migration.Uri, load.Migration.Dir); err != nil {
			fmt.Printf("Error running migrations: %v\n", err)
			return
		}

		fmt.Println("Migrations ran successfully")
	},
}

func init() {
	migrateCmd.Flags().StringP("u", "u", "", "Database connection string")
	migrateCmd.Flags().StringP("m", "m", "", "Directory containing migration files")
	rootCmd.AddCommand(migrateCmd)

	err := viper.BindPFlag("migration.uri", migrateCmd.Flags().Lookup("u"))
	if err != nil {
		fmt.Printf("Error binding flag: %v\n", err)
	}

	err = viper.BindPFlag("migration.dsn", migrateCmd.Flags().Lookup("m"))
	if err != nil {
		fmt.Printf("Error binding flag: %v\n", err)
	}
}

func migrationLoadConfig() {
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

func runMigrations(dsn, dir string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", dir),
		dsn)
	if err != nil {
		return fmt.Errorf("err from migration.New: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("err from up migration: %w", err)
	}

	return nil
}
