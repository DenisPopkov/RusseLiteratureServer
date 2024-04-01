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

type UserProvider interface {
	DeleteUser(ctx context.Context, userId int64) error
}

type PoetProvider interface {
	Poets(ctx context.Context, userId int64) ([]models.Poet, error)
}

type ArticleProvider interface {
	Articles(ctx context.Context, userId int64) ([]models.Article, error)
}

type AuthorProvider interface {
	Authors(ctx context.Context, userId int64) ([]models.Author, error)
}

type FeedProvider interface {
	Feed(ctx context.Context, userId int64) (models.Feed, error)
}

type Core struct {
	log             *slog.Logger
	userProvider    UserProvider
	poetProvider    PoetProvider
	articleProvider ArticleProvider
	authorProvider  AuthorProvider
	feedProvider    FeedProvider
	tokenTTL        time.Duration
}

func New(
	log *slog.Logger,
	userProvider UserProvider,
	poetProvider PoetProvider,
	articleProvider ArticleProvider,
	authorProvider AuthorProvider,
	feedProvider FeedProvider,
	tokenTTL time.Duration,
) *Core {
	return &Core{
		log:             log,
		userProvider:    userProvider,
		poetProvider:    poetProvider,
		articleProvider: articleProvider,
		authorProvider:  authorProvider,
		feedProvider:    feedProvider,
		tokenTTL:        tokenTTL,
	}
}

func (c *Core) GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetFeedHandler"

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

func (c *Core) GetAuthorHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetAuthorHandler"

	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	authors, err := c.authorProvider.Authors(r.Context(), int64(userId))
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

func (c *Core) GetArticlesHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetArticlesHandler"

	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	articles, err := c.articleProvider.Articles(r.Context(), int64(userId))
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(articles); err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}
}

func (c *Core) GetPoetsHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetPoetsHandler"

	// Parse userId from query parameters
	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	poets, err := c.poetProvider.Poets(r.Context(), int64(userId))
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(poets); err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}
}

func (c *Core) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.DeleteUserHandler"

	userIDStr := r.URL.Query().Get("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	err = c.userProvider.DeleteUser(r.Context(), userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
