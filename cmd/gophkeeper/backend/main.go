package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc "github.com/apolsh/yapr-gophkeeper/internal/backend/server"
	"github.com/apolsh/yapr-gophkeeper/internal/backend/service"
	"github.com/apolsh/yapr-gophkeeper/internal/backend/storage/database/postgres"
	token "github.com/apolsh/yapr-gophkeeper/internal/backend/token_manager"
	"github.com/apolsh/yapr-gophkeeper/internal/config"
	"github.com/apolsh/yapr-gophkeeper/internal/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"

	log = logger.LoggerOfComponent("backend-main")
)

// Server базовый интерфейс для серверов различного типа.
type Server interface {
	Start() error
	Stop(ctx context.Context) error
}

var _ Server = (*grpc.GRPCGophkeeperServer)(nil)

// buildVersion - версия сборки
// buildDate - дата сборки
// buildCommit - комментарий сборки.
func main() {
	log.Info("Build version: ", buildVersion)
	log.Info("Build date: ", buildDate)
	log.Info("Build commit: ", buildCommit)

	cfg := config.LoadServerConfig()

	logger.SetGlobalLevel(cfg.LogLevel)

	var userStorage service.UserStorage
	var secretStorage service.SecretStorage

	switch cfg.Storage {
	case config.PostgresStorageType:
		var err error //??? rights userStorage is unused if :=
		userStorage, err = postgres.NewGophkeeperStoragePG(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		secretStorage, err = postgres.NewGophkeeperStoragePG(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal(errors.New("unknown storage type"))
	}

	tokenManger := token.NewJWTTokenManager(cfg.TokenSecretKey)
	gophkeeperService := service.NewGophkeeperService(tokenManger, userStorage, secretStorage)
	grpcServer := grpc.NewGRPCGophkeeperServer(cfg.ServerAddr, gophkeeperService, tokenManger)

	done := make(chan bool)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-quit
		log.Info("server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := grpcServer.Stop(ctx)
		if err != nil {
			log.Fatal(fmt.Errorf("could not gracefully shutdown the grpc server: %v", err))
		}

		userStorage.Close()
		secretStorage.Close()
		close(done)
	}()

	err := grpcServer.Start()
	if err != nil {
		log.Fatal(fmt.Errorf("could not start grpc server: %v", err))
	}

	<-done
	log.Info("Server stopped")

}
