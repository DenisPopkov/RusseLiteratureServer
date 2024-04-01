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

// Author get authors from db.
func (s *Storage) Author(ctx context.Context) ([]models.Author, error) {
	const op = "storage.sqlite.GetAuthors"

	stmt, err := s.db.Prepare("SELECT * FROM authors")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var authors []models.Author
	for rows.Next() {
		var author models.Author
		err := rows.Scan(&author.ID, &author.Name, &author.Image, &author.Clip)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		authors = append(authors, author)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return authors, nil
}

// Feed get feed from db.
func (s *Storage) Feed(ctx context.Context, userId int64) (models.Feed, error) {
	const op = "storage.sqlite.GetFeed"

	stmt, err := s.db.Prepare(`
		SELECT a.id, a.name, a.image, a.clip
		FROM authors AS a 
		WHERE a.id IN (
			SELECT DISTINCT CAST(json_each.key AS INTEGER) 
			FROM feed 
			CROSS JOIN json_each(feed.authors) AS json_each 
			WHERE feed.id = ?)
	`)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}

	authorsRow, err := stmt.QueryContext(ctx, userId)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}
	defer authorsRow.Close()

	var feed models.Feed
	for authorsRow.Next() {
		var author models.Author
		err := authorsRow.Scan(&author.ID, &author.Name, &author.Image, &author.Clip)
		if err != nil {
			return models.Feed{}, fmt.Errorf("%s: %w", op, err)
		}
		feed.Authors = append(feed.Authors, author)
	}

	if err := authorsRow.Err(); err != nil {
		return feed, fmt.Errorf("%s: %w", op, err)
	}

	// Fetch articles
	articleStmt, err := s.db.Prepare(`
		SELECT id, name, image, clip
		FROM articles 
		WHERE id IN (
			SELECT DISTINCT CAST(json_each.key AS INTEGER) 
			FROM feed 
			CROSS JOIN json_each(feed.articles) AS json_each 
			WHERE feed.id = ?
		)
	`)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}

	articleRows, err := articleStmt.QueryContext(ctx, userId)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}
	defer articleRows.Close()

	for articleRows.Next() {
		var article models.Article
		err := articleRows.Scan(&article.ID, &article.Name, &article.Image, &article.Clip)
		if err != nil {
			return models.Feed{}, fmt.Errorf("%s: %w", op, err)
		}
		feed.Articles = append(feed.Articles, article)
	}

	if err := articleRows.Err(); err != nil {
		return feed, fmt.Errorf("%s: %w", op, err)
	}

	// Fetch poets
	poetStmt, err := s.db.Prepare(`
		SELECT id, name, image, clip
		FROM poets 
		WHERE id IN (
			SELECT DISTINCT CAST(json_each.key AS INTEGER) 
			FROM feed 
			CROSS JOIN json_each(feed.poets) AS json_each 
			WHERE feed.id = ?
		)
	`)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}

	poetRows, err := poetStmt.QueryContext(ctx, userId)
	if err != nil {
		return models.Feed{}, fmt.Errorf("%s: %w", op, err)
	}
	defer poetRows.Close()

	for poetRows.Next() {
		var poet models.Poet
		err := poetRows.Scan(&poet.ID, &poet.Name, &poet.Image, &poet.Clip)
		if err != nil {
			return models.Feed{}, fmt.Errorf("%s: %w", op, err)
		}
		feed.Poets = append(feed.Poets, poet)
	}

	if err := poetRows.Err(); err != nil {
		return feed, fmt.Errorf("%s: %w", op, err)
	}

	feed.ID = userId

	return feed, nil
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
