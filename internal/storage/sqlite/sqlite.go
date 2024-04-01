package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"sso/internal/domain/models"
	"sso/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}

func (s *Storage) SaveUser(ctx context.Context, phone string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(phone, pass_hash, feed) VALUES(?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, phone, passHash, 0)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	_, err = s.db.ExecContext(ctx, "UPDATE users SET feed = ? WHERE phone = ?", id, phone)
	if err != nil {
		return id, fmt.Errorf("%s: failed to update feed: %w", op, err)
	}

	authors := map[string]bool{
		"1001": false,
		"1002": false,
		"1003": false,
		"1004": false,
	}
	authorsJson, err := json.Marshal(authors)
	if err != nil {
		panic(err)
	}

	articles := map[string]bool{
		"2001": false,
		"2002": false,
	}
	articlesJson, err := json.Marshal(articles)
	if err != nil {
		panic(err)
	}

	poets := map[string]bool{
		"3001": false,
		"3002": false,
	}
	poetsJson, err := json.Marshal(poets)
	if err != nil {
		panic(err)
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO feed(id, authors, articles, poets) VALUES(?, ?, ?, ?)", id, string(authorsJson), string(articlesJson), string(poetsJson))
	if err != nil {
		return id, fmt.Errorf("%s: failed to add feed: %w", op, err)
	}

	return id, nil
}

// User returns user by phone.
func (s *Storage) User(ctx context.Context, phone string) (models.User, error) {
	const op = "storage.sqlite.User"

	stmt, err := s.db.Prepare("SELECT id, phone, pass_hash, feed FROM users WHERE phone = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, phone)

	var user models.User
	err = row.Scan(&user.ID, &user.Phone, &user.PassHash, &user.Feed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// Authors retrieves authors from the database for a given user ID.
func (s *Storage) Authors(ctx context.Context, userId int64) ([]models.Author, error) {
	const op = "storage.sqlite.GetAuthors"

	stmt, err := s.db.Prepare(`
		SELECT id, name, image, clip, isFave
		FROM authors 
		WHERE id IN (
			SELECT DISTINCT CAST(json_each.key AS INTEGER) 
			FROM feed 
			CROSS JOIN json_each(feed.authors) AS json_each 
			WHERE feed.id = ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var authors []models.Author
	for rows.Next() {
		var author models.Author
		err := rows.Scan(&author.ID, &author.Name, &author.Image, &author.Clip, &author.IsFave)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		return authors, fmt.Errorf("%s: %w", op, err)
	}

	return authors, nil
}

// Articles retrieves articles from the database for a given user ID.
func (s *Storage) Articles(ctx context.Context, userId int64) ([]models.Article, error) {
	const op = "storage.sqlite.GetArticles"

	stmt, err := s.db.Prepare(`
		SELECT id, name, image, clip, isFave
		FROM articles 
		WHERE id IN (
			SELECT DISTINCT CAST(json_each.key AS INTEGER) 
			FROM feed 
			CROSS JOIN json_each(feed.articles) AS json_each 
			WHERE feed.id = ?
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.Name, &article.Image, &article.Clip, &article.IsFave)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return articles, fmt.Errorf("%s: %w", op, err)
	}

	return articles, nil
}

// Poets retrieves poets from the database for a given user ID.
func (s *Storage) Poets(ctx context.Context, userId int64) ([]models.Poet, error) {
	const op = "storage.sqlite.GetPoets"

	stmt, err := s.db.Prepare(`
		SELECT id, name, image, clip, isFave
		FROM poets 
		WHERE id IN (
			SELECT DISTINCT CAST(json_each.key AS INTEGER) 
			FROM feed 
			CROSS JOIN json_each(feed.poets) AS json_each 
			WHERE feed.id = ?
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var poets []models.Poet
	for rows.Next() {
		var poet models.Poet
		err := rows.Scan(&poet.ID, &poet.Name, &poet.Image, &poet.Clip, &poet.IsFave)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		poets = append(poets, poet)
	}

	if err := rows.Err(); err != nil {
		return poets, fmt.Errorf("%s: %w", op, err)
	}

	return poets, nil
}

// Feed retrieves the feed containing authors, articles, and poets for a given user ID.
func (s *Storage) Feed(ctx context.Context, userId int64) (models.Feed, error) {
	const op = "storage.sqlite.GetFeed"

	// Get authors
	authors, err := s.Authors(ctx, userId)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}

	// Get articles
	articles, err := s.Articles(ctx, userId)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}

	// Get poets
	poets, err := s.Poets(ctx, userId)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}

	feed := models.Feed{
		ID:       userId,
		Authors:  authors,
		Articles: articles,
		Poets:    poets,
	}

	return feed, nil
}

// DeleteUser deletes a user by their ID.
func (s *Storage) DeleteUser(ctx context.Context, userID int64) error {
	const op = "storage.sqlite.DeleteUser"

	stmt, err := s.db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) App(ctx context.Context) (models.App, error) {
	const op = "storage.sqlite.App"

	stmt, err := s.db.Prepare("SELECT name, secret FROM apps")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx)

	var app models.App
	err = row.Scan(&app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
