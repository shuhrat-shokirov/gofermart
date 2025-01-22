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

		splitAfter := strings.SplitAfter(cfg.Server.Address, ":")
		if len(splitAfter) != 2 {
			logger.Fatal("can't parse address", zap.String("address", cfg.Server.Address))
		}

		port, err := strconv.Atoi(splitAfter[1])
		if err != nil {
			logger.Fatal("can't parse port", zap.Error(err))
		}

		api := rest.NewRouter(rest.Config{
			Port:   int64(port),
			Logger: *logger.Sugar(),
		})

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		//TODO: Implement graceful shutdown
		//go func() {
		//	<-stop
		//	if err := newStore.Close(); err != nil {
		//		conf.logger.Errorf("can't close store: %v", err)
		//	}
		//
		//	os.Exit(0)
		//}()

		if err := api.Run(); err != nil {
			logger.Fatal("can't run server", zap.Error(err))
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		// Optionally, os.Exit(1)
	}
}
