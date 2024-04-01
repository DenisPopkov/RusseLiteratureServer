package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sso/internal/domain/models"
	"strconv"
	"time"
)

type AuthorProvider interface {
	Author(ctx context.Context) ([]models.Author, error)
}

type FeedProvider interface {
	Feed(ctx context.Context, userId int64) (models.Feed, error)
}

type Core struct {
	log            *slog.Logger
	authorProvider AuthorProvider
	feedProvider   FeedProvider
	tokenTTL       time.Duration
}

func New(
	log *slog.Logger,
	authorProvider AuthorProvider,
	feedProvider FeedProvider,
	tokenTTL time.Duration,
) *Core {
	return &Core{
		log:            log,
		authorProvider: authorProvider,
		feedProvider:   feedProvider,
		tokenTTL:       tokenTTL,
	}
}

func (c *Core) GetAuthorHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetAuthorHandler"

	authors, err := c.authorProvider.Author(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authors); err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}
}

func (c *Core) GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetFeedHandler"

	// Parse userId from query parameters
	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	feed, err := c.feedProvider.Feed(r.Context(), int64(userId))
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(feed); err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}
}
