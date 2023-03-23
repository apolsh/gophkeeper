package main

import (
	"context"
	"os"
	"os/signal"

	grpcClient "github.com/apolsh/yapr-gophkeeper/internal/client/backend_client/grpc_client"
	"github.com/apolsh/yapr-gophkeeper/internal/client/controller"
	"github.com/apolsh/yapr-gophkeeper/internal/client/encoder"
	"github.com/apolsh/yapr-gophkeeper/internal/client/storage/database/sqlite"
	"github.com/apolsh/yapr-gophkeeper/internal/client/view"
	"github.com/apolsh/yapr-gophkeeper/internal/config"
	"github.com/apolsh/yapr-gophkeeper/internal/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"

	log = logger.LoggerOfComponent("client-main")
)

// buildVersion - версия сборки
// buildDate - дата сборки
// buildCommit - комментарий сборки.
func main() {
	log.Info("Build version: ", buildVersion)
	log.Info("Build date: ", buildDate)
	log.Info("Build commit: ", buildCommit)

	cfg, err := config.LoadClientConfig()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()

	go func(ctx context.Context) {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}(ctx)

	menu := view.GophkeeperViewInteractiveCLI{}
	serverClient := grpcClient.NewGophkeeperGRPCClient(cfg.SyncServerURL)
	localStorage, err := sqlite.NewGophkeeperLocalStorageSqlite(cfg.BaseDir)
	if err != nil {
		log.Fatal(err)
	}
	controller.NewGophkeeperController(ctx, &menu, &serverClient, localStorage, &encoder.AESGMCEncoder{}, cfg.SyncPeriod)
	err = menu.Show(ctx)
	if err != nil {
		log.Error(err)
	}

}
