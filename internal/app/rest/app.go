package restapp

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"sso/internal/services/core"
	"strings"
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

type MiddlewareFunc func(http.Handler) http.Handler

// AuthMiddleware middleware function for JWT token validation and UID extraction
func (a *App) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse JWT token: %v", err), http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid JWT token claims", http.StatusUnauthorized)
			return
		}

		uidFloat64, ok := claims["uid"].(float64)
		if !ok {
			http.Error(w, "UID not found in JWT token claims", http.StatusUnauthorized)
			return
		}

		uid := int64(uidFloat64)

		ctx := context.WithValue(r.Context(), "uid", uid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Run runs the HTTP server.
func (a *App) Run() error {
	const op = "restapp.Run"

	router := mux.NewRouter()

	authMiddleware := func(next http.Handler) http.Handler {
		return a.AuthMiddleware(next)
	}

	authRouter := router.PathPrefix("").Subrouter()
	authRouter.Use(authMiddleware)

	authRouter.HandleFunc("/poets", a.coreService.UpdatePoetIsFaveHandler).Methods("PATCH")
	authRouter.HandleFunc("/articles", a.coreService.UpdateArticleIsFaveHandler).Methods("PATCH")
	authRouter.HandleFunc("/authors", a.coreService.UpdateAuthorIsFaveHandler).Methods("PATCH")
	authRouter.HandleFunc("/user", a.coreService.DeleteUserHandler).Methods("DELETE")
	authRouter.HandleFunc("/user", a.coreService.GetUserHandler).Methods("GET")
	authRouter.HandleFunc("/quiz", a.coreService.GetQuizHandler).Methods("GET")
	authRouter.HandleFunc("/clip", a.coreService.GetClipHandler).Methods("GET")
	authRouter.HandleFunc("/articles", a.coreService.GetArticlesHandler).Methods("GET")
	authRouter.HandleFunc("/poets", a.coreService.GetPoetsHandler).Methods("GET")
	authRouter.HandleFunc("/authors", a.coreService.GetAuthorHandler).Methods("GET")

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: router,
	}

	if err := a.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
