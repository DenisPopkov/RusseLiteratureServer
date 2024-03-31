package restapp

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"sso/internal/services/core"
)

type App struct {
	log         *slog.Logger
	httpServer  *http.Server
	port        int
	coreService *core.Core
}

func New(
	log *slog.Logger,
	coreService *core.Core,
	port int,
) *App {
	return &App{
		log:         log,
		port:        port,
		coreService: coreService,
	}
}

// MustRun runs the HTTP server and panics if any error occurs.
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Run runs the HTTP server.
func (a *App) Run() error {
	const op = "restapp.Run"

	router := mux.NewRouter()
	router.HandleFunc("/author", a.coreService.GetAuthorHandler).Methods("GET")

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: router,
	}

	if err := a.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
