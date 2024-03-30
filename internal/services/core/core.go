package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sso/internal/domain/models"
	"time"
)

type AuthorProvider interface {
	Author(ctx context.Context) ([]models.Author, error)
}

type Core struct {
	log            *slog.Logger
	authorProvider AuthorProvider
	tokenTTL       time.Duration
}

func New(
	log *slog.Logger,
	authorProvider AuthorProvider,
	tokenTTL time.Duration,
) *Core {
	return &Core{
		log:            log,
		authorProvider: authorProvider,
		tokenTTL:       tokenTTL,
	}
}

func (c *Core) GetAuthorHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetAuthorHandler"

	// Retrieve author data using AuthorProvider
	authors, err := c.authorProvider.Author(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	// Encode authors data as JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authors); err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}
}
