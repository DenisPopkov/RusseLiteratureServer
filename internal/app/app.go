package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	restapp "sso/internal/app/rest"
	"sso/internal/services/auth"
	"sso/internal/services/core"
	"sso/internal/storage/sqlite"
	"time"
)

type App struct {
	log        *slog.Logger
	GRPCServer *grpcapp.App
	RestServer *restapp.App
	port       int
}

func NewGrpc(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)
	grpcApp := grpcapp.New(log, authService, port)

	return &App{
		log:        log,
		GRPCServer: grpcApp,
		port:       port,
	}
}

func NewRest(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	coreService := core.New(log, storage, storage, storage, storage, tokenTTL)
	restApp := restapp.New(log, coreService, port)

	return &App{
		log:        log,
		RestServer: restApp,
		port:       port,
	}
}
