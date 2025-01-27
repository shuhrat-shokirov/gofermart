package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"gofermart/internal/gophermart/config"
	"gofermart/internal/gophermart/core/application"
	"gofermart/internal/gophermart/core/client"
	"gofermart/internal/gophermart/infra/api/rest"
	"gofermart/internal/gophermart/infra/store"
	"gofermart/internal/gophermart/infra/store/memory"
	"gofermart/internal/gophermart/infra/store/postgresql"
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

		var (
			memoryConfig = &memory.Config{}
			dbConfig     *postgresql.Config
		)

		if cfg.DB.URI != "" {
			dbConfig = &postgresql.Config{
				Dsn: cfg.DB.URI,
			}
		}

		newStore, err := store.NewStore(store.Config{
			Memory:     memoryConfig,
			Postgresql: dbConfig,
		})
		if err != nil {
			logger.Fatal("can't create store", zap.Error(err))
		}

		(*logger.Sugar()).Infof("client address: %s", cfg.Accrual.System.Address)

		newClient := client.NewClient(cfg.Accrual.System.Address, cfg.Accrual.System.Limit)

		newApplication := application.NewApplication(application.Config{
			Repo:   newStore,
			Client: newClient,
			Logger: *logger.Sugar(),
			Secret: cfg.Secret,
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go newApplication.RunWorker(ctx)

		api := rest.NewRouter(rest.Config{
			Server: newApplication,
			Port:   getPortFromAddress(cfg.Server.Address),
			Logger: *logger.Sugar(),
		})

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		go func() {
			<-stop
			cancel()
			if err := newStore.Close(); err != nil {
				logger.Error("can't close store", zap.Error(err))
			}

			os.Exit(0)
		}()

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
