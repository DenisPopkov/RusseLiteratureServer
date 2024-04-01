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
	"strconv"
	"strings"
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

	authors := map[string]string{
		"1001": "false",
		"1002": "false",
		"1003": "false",
		"1004": "false",
	}
	authorsJson, err := json.Marshal(authors)
	if err != nil {
		panic(err)
	}

	articles := map[string]string{
		"2001": "false",
		"2002": "false",
	}
	articlesJson, err := json.Marshal(articles)
	if err != nil {
		panic(err)
	}

	poets := map[string]string{
		"3001": "false",
		"3002": "false",
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

// UpdateAuthorIsFave updates the isFave field for an author in the feed JSON by their ID.
func (s *Storage) UpdateAuthorIsFave(ctx context.Context, userID int64, authorID int64, isFave string) ([]models.Author, error) {
	const op = "storage.sqlite.UpdateAuthorIsFave"

	var authorsJSON string
	err := s.db.QueryRowContext(ctx, "SELECT authors FROM feed WHERE id = ?", userID).Scan(&authorsJSON)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	strBool, err := strconv.ParseBool(isFave)
	old := !strBool

	updatedAuthorsJSON := strings.Replace(authorsJSON, fmt.Sprintf(`"%d":"%s"`, authorID, strconv.FormatBool(old)), fmt.Sprintf(`"%d":"%s"`, authorID, isFave), 1)

	_, err = s.db.ExecContext(ctx, "UPDATE feed SET authors = ? WHERE id = ?", updatedAuthorsJSON, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.db.QueryContext(ctx, "SELECT * FROM authors")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var updatedAuthors []models.Author
	for rows.Next() {
		var author models.Author
		if err := rows.Scan(&author.ID, &author.Name, &author.Image, &author.Clip, &author.IsFave); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if author.ID == authorID {
			author.IsFave = isFave
		}
		updatedAuthors = append(updatedAuthors, author)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return updatedAuthors, nil
}

// UpdateArticleIsFave updates the isFave field for an article in the feed JSON by their ID.
func (s *Storage) UpdateArticleIsFave(ctx context.Context, userID int64, articleID int64, isFave string) ([]models.Article, error) {
	const op = "storage.sqlite.UpdateAuthorIsFave"

	var articleJSON string
	err := s.db.QueryRowContext(ctx, "SELECT articles FROM feed WHERE id = ?", userID).Scan(&articleJSON)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	strBool, err := strconv.ParseBool(isFave)
	old := !strBool

	updatedArticleJSON := strings.Replace(articleJSON, fmt.Sprintf(`"%d":"%s"`, articleID, strconv.FormatBool(old)), fmt.Sprintf(`"%d":"%s"`, articleID, isFave), 1)

	_, err = s.db.ExecContext(ctx, "UPDATE feed SET articles = ? WHERE id = ?", updatedArticleJSON, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.db.QueryContext(ctx, "SELECT * FROM articles")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var updatedArticles []models.Article
	for rows.Next() {
		var article models.Article
		if err := rows.Scan(&article.ID, &article.Name, &article.Image, &article.Clip, &article.IsFave); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if article.ID == articleID {
			article.IsFave = isFave
		}
		updatedArticles = append(updatedArticles, article)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return updatedArticles, nil
}

// UpdatePoetIsFave updates the isFave field for a poet in the feed JSON by their ID.
func (s *Storage) UpdatePoetIsFave(ctx context.Context, userID int64, poetID int64, isFave string) ([]models.Poet, error) {
	const op = "storage.sqlite.UpdateAuthorIsFave"

	var poetJSON string
	err := s.db.QueryRowContext(ctx, "SELECT poets FROM feed WHERE id = ?", userID).Scan(&poetJSON)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	strBool, err := strconv.ParseBool(isFave)
	old := !strBool

	updatedPoetJSON := strings.Replace(poetJSON, fmt.Sprintf(`"%d":"%s"`, poetID, strconv.FormatBool(old)), fmt.Sprintf(`"%d":"%s"`, poetID, isFave), 1)

	_, err = s.db.ExecContext(ctx, "UPDATE feed SET poets = ? WHERE id = ?", updatedPoetJSON, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := s.db.QueryContext(ctx, "SELECT * FROM poets")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var updatedPoets []models.Poet
	for rows.Next() {
		var poet models.Poet
		if err := rows.Scan(&poet.ID, &poet.Name, &poet.Image, &poet.Clip, &poet.IsFave); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if poet.ID == poetID {
			poet.IsFave = isFave
		}
		updatedPoets = append(updatedPoets, poet)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return updatedPoets, nil
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
