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
	GetUser(ctx context.Context, userId int64) (models.UserData, error)
}

type PoetProvider interface {
	Poets(ctx context.Context, userId int64) ([]models.Poet, error)
	UpdatePoetIsFave(ctx context.Context, userID int64, poetID int64, isFave string) error
}

type ArticleProvider interface {
	Articles(ctx context.Context, userId int64) ([]models.Article, error)
	UpdateArticleIsFave(ctx context.Context, userID int64, articleID int64, isFave string) error
}

type AuthorProvider interface {
	Authors(ctx context.Context, userId int64) ([]models.Author, error)
	UpdateAuthorIsFave(ctx context.Context, userID int64, authorID int64, isFave string) error
}

type ClipProvider interface {
	GetClip(ctx context.Context, clipId int64) (models.Clip, error)
}

type QuizProvider interface {
	GetQuiz(ctx context.Context, quizId int64) (models.Quiz, error)
}

type FeedProvider interface {
	Feed(ctx context.Context, userId int64) (models.Feed, error)
}

type Core struct {
	log             *slog.Logger
	userProvider    UserProvider
	quizProvider    QuizProvider
	clipProvider    ClipProvider
	poetProvider    PoetProvider
	articleProvider ArticleProvider
	authorProvider  AuthorProvider
	feedProvider    FeedProvider
	tokenTTL        time.Duration
}

func New(
	log *slog.Logger,
	userProvider UserProvider,
	quizProvider QuizProvider,
	clipProvider ClipProvider,
	poetProvider PoetProvider,
	articleProvider ArticleProvider,
	authorProvider AuthorProvider,
	feedProvider FeedProvider,
	tokenTTL time.Duration,
) *Core {
	return &Core{
		log:             log,
		userProvider:    userProvider,
		quizProvider:    quizProvider,
		clipProvider:    clipProvider,
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

// UpdateAuthorIsFaveHandler handles the HTTP PATCH request to update the isFave field for an author.
func (c *Core) UpdateAuthorIsFaveHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.UpdateAuthorIsFaveHandler"

	userIDStr := r.URL.Query().Get("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	authorIDStr := r.URL.Query().Get("authorId")
	authorID, err := strconv.ParseInt(authorIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	isFave := r.URL.Query().Get("isFave")

	err = c.authorProvider.UpdateAuthorIsFave(r.Context(), userID, authorID, isFave)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

// UpdateArticleIsFaveHandler handles the HTTP PATCH request to update the isFave field for an author.
func (c *Core) UpdateArticleIsFaveHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.UpdateArticleIsFaveHandler"

	userIDStr := r.URL.Query().Get("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	authorIDStr := r.URL.Query().Get("articleId")
	articleID, err := strconv.ParseInt(authorIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	isFave := r.URL.Query().Get("isFave")

	err = c.articleProvider.UpdateArticleIsFave(r.Context(), userID, articleID, isFave)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

// UpdatePoetIsFaveHandler handles the HTTP PATCH request to update the isFave field for an author.
func (c *Core) UpdatePoetIsFaveHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.UpdatePoetIsFaveHandler"

	userIDStr := r.URL.Query().Get("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	poetIDStr := r.URL.Query().Get("poetId")
	poetID, err := strconv.ParseInt(poetIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	isFave := r.URL.Query().Get("isFave")

	err = c.poetProvider.UpdatePoetIsFave(r.Context(), userID, poetID, isFave)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

func (c *Core) GetClipHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetClipHandler"

	clipIDStr := r.URL.Query().Get("clipId")
	clipID, err := strconv.ParseInt(clipIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	clips, err := c.clipProvider.GetClip(r.Context(), clipID)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clips); err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}
}

func (c *Core) GetQuizHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetQuizHandler"

	quizIDStr := r.URL.Query().Get("quizId")
	quizID, err := strconv.ParseInt(quizIDStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	quiz, err := c.quizProvider.GetQuiz(r.Context(), quizID)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(quiz); err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}
}

func (c *Core) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	const op = "core.GetUserHandler"

	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusBadRequest)
		return
	}

	clips, err := c.userProvider.GetUser(r.Context(), userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clips); err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", op, err), http.StatusInternalServerError)
		return
	}
}
