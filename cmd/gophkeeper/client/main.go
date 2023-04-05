package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/apolsh/yapr-gophkeeper/cmd/gophkeeper/tls"
	grpcClient "github.com/apolsh/yapr-gophkeeper/internal/client/backend_client/grpc_client"
	"github.com/apolsh/yapr-gophkeeper/internal/client/controller"
	"github.com/apolsh/yapr-gophkeeper/internal/client/encoder"
	"github.com/apolsh/yapr-gophkeeper/internal/client/storage/database/sqlite"
	"github.com/apolsh/yapr-gophkeeper/internal/client/view"
	"github.com/apolsh/yapr-gophkeeper/internal/config"
	"github.com/apolsh/yapr-gophkeeper/internal/logger"
	"github.com/apolsh/yapr-gophkeeper/internal/misc/scheduler"
	"google.golang.org/grpc/credentials"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"

	log = logger.LoggerOfComponent("client-main")
)

// buildVersion - версия сборки.
// buildDate - дата сборки.
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

	go func(ctx context.Context) {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}(ctx)

	menu := view.GophkeeperViewInteractiveCLI{}
	var serverClient controller.BackendClient
	if cfg.HTTPSEnabled {
		tlsConfig, err := tls.GetTLSConfig()
		if err != nil {
			log.Fatal(fmt.Errorf("could not get TLS configs %v", err))
		}

		serverClient = grpcClient.NewGophkeeperGRPCClientTLS(cfg.SyncServerURL, credentials.NewTLS(tlsConfig))
	} else {
		serverClient = grpcClient.NewGophkeeperGRPCClient(cfg.SyncServerURL)
	}

	localStorage, err := sqlite.NewGophkeeperLocalStorageSqlite(cfg.BaseDir)
	if err != nil {
		log.Fatal(err)
	}
	ctrl := controller.NewGophkeeperController(&menu, serverClient, localStorage, &encoder.AESGMCEncoder{})
	synchronization := scheduler.NewScheduler(ctrl.SynchronizeSecretItems, menu.ShowError)
	synchronization.RunWithInterval(ctx, time.Duration(cfg.SyncPeriod)*time.Second)

	err = menu.Show(ctx)
	if err != nil {
		log.Error(err)
	}

	synchronization.Close()
}
