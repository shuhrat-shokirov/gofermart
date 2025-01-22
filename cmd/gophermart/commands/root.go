package commands

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"gofermart/internal/gophermart/config"
	"gofermart/internal/gophermart/infra/api/rest"
)

var rootCmd = &cobra.Command{
	Use:   "gophermart",
	Short: "GopherMart CLI application",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		loadConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Fatalf("can't initialize logger: %v", err)
		}
		defer func() {
			if err := logger.Sync(); err != nil {
				log.Printf("can't sync logger: %v", err)
			}
		}()

		api := rest.NewRouter(rest.Config{
			Port:   getPortFromAddress(cfg.Server.Address),
			Logger: *logger.Sugar(),
		})

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		//TODO: Implement graceful shutdown

		if err := api.Run(); err != nil {
			logger.Fatal("can't run server", zap.Error(err))
		}
	},
}

func getPortFromAddress(address string) int64 {
	const portSplitLen = 2

	splitAfter := strings.SplitAfter(address, ":")
	if len(splitAfter) != portSplitLen {
		log.Fatalf("can't parse address: %s", address)
	}

	port, err := strconv.Atoi(splitAfter[1])
	if err != nil {
		log.Fatalf("can't parse port: %v", err)
	}

	return int64(port)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		// Optionally, os.Exit(1)
	}
}
