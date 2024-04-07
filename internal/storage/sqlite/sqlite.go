package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"math/rand"
	"sso/internal/domain/models"
	"sso/internal/storage"
	"time"
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

	stmt, err := s.db.Prepare("INSERT INTO users(phone, pass_hash, name, image) VALUES(?, ?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	authorIDs, err := s.getAllIDsFromAuthors(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	articleIDs, err := s.getAllIDsFromArticles(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	poetIDs, err := s.getAllIDsFromPoets(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	authorsName := map[int]string{
		0: "Фёдор Достоевский",
		1: "Антон Чехов",
		2: "Лев Толстой",
	}

	authorsImage := map[int]string{
		0: "https://iili.io/JwdlbqJ.png",
		1: "https://iili.io/Jwdlg0x.png",
		2: "https://iili.io/Jwdl6JV.png",
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomKey := rand.Intn(len(authorsName))

	res, err := stmt.ExecContext(ctx, phone, passHash, authorsName[randomKey], authorsImage[randomKey])
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	for _, authorID := range authorIDs {
		_, err = s.db.ExecContext(ctx, "INSERT INTO authors(userId, authorId, isFave) VALUES(?, ?, 0)", id, authorID)
		if err != nil {
			return id, fmt.Errorf("%s: failed to insert author: %w", op, err)
		}
	}

	for _, articleID := range articleIDs {
		_, err = s.db.ExecContext(ctx, "INSERT INTO articles(userId, articleId, isFave) VALUES(?, ?, 0)", id, articleID)
		if err != nil {
			return id, fmt.Errorf("%s: failed to insert article: %w", op, err)
		}
	}

	for _, poetID := range poetIDs {
		_, err = s.db.ExecContext(ctx, "INSERT INTO poets(userId, poetId, isFave) VALUES(?, ?, 0)", id, poetID)
		if err != nil {
			return id, fmt.Errorf("%s: failed to insert poet: %w", op, err)
		}
	}

	return id, nil
}

func (s *Storage) getAllIDsFromAuthors(ctx context.Context) ([]int, error) {
	query := fmt.Sprintf("SELECT id FROM author")
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *Storage) getAllIDsFromArticles(ctx context.Context) ([]int, error) {
	query := fmt.Sprintf("SELECT id FROM article")
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *Storage) getAllIDsFromPoets(ctx context.Context) ([]int, error) {
	query := fmt.Sprintf("SELECT id FROM poet")
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

// User returns user by phone.
func (s *Storage) User(ctx context.Context, phone string) (models.User, error) {
	const op = "storage.sqlite.User"

	stmt, err := s.db.Prepare("SELECT id, phone, pass_hash, name, image FROM users WHERE phone = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, phone)

	var user models.User
	err = row.Scan(&user.ID, &user.Phone, &user.PassHash, &user.Name, &user.Image)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// GetUser returns user by id.
func (s *Storage) GetUser(ctx context.Context, userId int64) (models.UserData, error) {
	const op = "storage.sqlite.GetUser"

	stmt, err := s.db.Prepare("SELECT name, image FROM users WHERE id = ?")
	if err != nil {
		return models.UserData{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, userId)

	var user models.UserData
	err = row.Scan(&user.Name, &user.Image)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserData{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.UserData{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// Authors retrieves authors from the database for a given user ID.
func (s *Storage) Authors(ctx context.Context, userId int64) ([]models.Author, error) {
	const op = "storage.sqlite.GetAuthors"

	stmt, err := s.db.Prepare(`
		SELECT a.id, a.name, a.image, a.clip, ua.isFave
		FROM author AS a
		INNER JOIN authors AS ua ON a.id = ua.authorId
		WHERE ua.userId = ?
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
		SELECT a.id, a.name, a.image, a.clip, a.description, ua.isFave
		FROM article AS a
		INNER JOIN articles AS ua ON a.id = ua.articleId
		WHERE ua.userId = ?
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
		err := rows.Scan(&article.ID, &article.Name, &article.Image, &article.Clip, &article.Description, &article.IsFave)
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
		SELECT p.id, p.name, p.image, p.clip, ua.isFave
		FROM poet AS p
		INNER JOIN poets AS ua ON p.id = ua.poetId
		WHERE ua.userId = ?
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

// DeleteUser deletes a user by their ID.
func (s *Storage) DeleteUser(ctx context.Context, userID int64) error {
	const op = "storage.sqlite.DeleteUser"

	stmt, err := s.db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmtAuthors, err := s.db.Prepare("DELETE FROM authors WHERE userId = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmtArticles, err := s.db.Prepare("DELETE FROM articles WHERE userId = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmtPoets, err := s.db.Prepare("DELETE FROM poets WHERE userId = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, userID)
	_, err = stmtAuthors.ExecContext(ctx, userID)
	_, err = stmtArticles.ExecContext(ctx, userID)
	_, err = stmtPoets.ExecContext(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// UpdateAuthorIsFave updates the isFave field for an author in the authors table by their userID and authorID.
func (s *Storage) UpdateAuthorIsFave(ctx context.Context, userId int64, authorId int64) error {
	const op = "storage.sqlite.UpdateAuthorIsFave"

	var currentIsFave int
	err := s.db.QueryRowContext(ctx, "SELECT isFave FROM authors WHERE userId = ? AND authorId = ?", userId, authorId).Scan(&currentIsFave)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: author with userId %d and authorId %d not found", op, userId, authorId)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	newIsFave := 1 - currentIsFave

	_, err = s.db.ExecContext(ctx, "UPDATE authors SET isFave = ? WHERE userId = ? AND authorId = ?", newIsFave, userId, authorId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// UpdateArticleIsFave updates the isFave field for an article
func (s *Storage) UpdateArticleIsFave(ctx context.Context, userId int64, articleId int64) error {
	const op = "storage.sqlite.UpdateAuthorIsFave"

	var currentIsFave int
	err := s.db.QueryRowContext(ctx, "SELECT isFave FROM articles WHERE userId = ? AND articleId = ?", userId, articleId).Scan(&currentIsFave)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: article with userId %d and articleId %d not found", op, userId, articleId)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	newIsFave := 1 - currentIsFave

	_, err = s.db.ExecContext(ctx, "UPDATE articles SET isFave = ? WHERE userId = ? AND articleId = ?", newIsFave, userId, articleId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// UpdatePoetIsFave updates the isFave field for a poet
func (s *Storage) UpdatePoetIsFave(ctx context.Context, userId int64, poetId int64) error {
	const op = "storage.sqlite.UpdateAuthorIsFave"

	var currentIsFave int
	err := s.db.QueryRowContext(ctx, "SELECT isFave FROM poets WHERE userId = ? AND poetId = ?", userId, poetId).Scan(&currentIsFave)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: poet with userId %d and poetId %d not found", op, userId, poetId)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	newIsFave := 1 - currentIsFave

	_, err = s.db.ExecContext(ctx, "UPDATE poets SET isFave = ? WHERE userId = ? AND poetId = ?", newIsFave, userId, poetId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetClip retrieves a clip from the database by ID.
func (s *Storage) GetClip(ctx context.Context, clipID int64) (models.Clip, error) {
	const op = "storage.sqlite.GetClip"

	stmtClipText, err := s.db.Prepare("SELECT text, title FROM clipText WHERE clipTextId = ?")
	if err != nil {
		return models.Clip{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmtClipText.Close()

	rows, err := stmtClipText.QueryContext(ctx, clipID)
	if err != nil {
		return models.Clip{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var clipTexts []models.ClipText

	for rows.Next() {
		var clipText models.ClipText
		if err := rows.Scan(&clipText.Text, &clipText.Title); err != nil {
			return models.Clip{}, fmt.Errorf("%s: %w", op, err)
		}
		clipTexts = append(clipTexts, clipText)
	}
	if err := rows.Err(); err != nil {
		return models.Clip{}, fmt.Errorf("%s: %w", op, err)
	}

	stmtClip, err := s.db.Prepare("SELECT id, image FROM clip WHERE id = ?")
	if err != nil {
		return models.Clip{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmtClip.Close()

	var clip models.Clip
	err = stmtClip.QueryRowContext(ctx, clipID).Scan(&clip.ID, &clip.Image)
	if err != nil {
		return models.Clip{}, fmt.Errorf("%s: %w", op, err)
	}

	quiz, err := s.GetQuiz(ctx, clipID)
	if err != nil {
		return models.Clip{}, fmt.Errorf("%s: %w", op, err)
	}

	clip.Text = clipTexts
	clip.Quiz = quiz

	return clip, nil
}

// GetQuiz retrieves a quiz from the database by ID.
func (s *Storage) GetQuiz(ctx context.Context, quizID int64) (models.Quiz, error) {
	const op = "storage.sqlite.GetQuiz"

	stmtAnswers, err := s.db.Prepare("SELECT id, text, isRight FROM answers WHERE id = ?")
	if err != nil {
		return models.Quiz{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmtAnswers.Close()

	rows, err := stmtAnswers.QueryContext(ctx, quizID)
	if err != nil {
		return models.Quiz{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var answers []models.Answer

	for rows.Next() {
		var answer models.Answer
		if err := rows.Scan(&answer.ID, &answer.Text, &answer.IsRight); err != nil {
			return models.Quiz{}, fmt.Errorf("%s: %w", op, err)
		}
		answers = append(answers, answer)
	}
	if err := rows.Err(); err != nil {
		return models.Quiz{}, fmt.Errorf("%s: %w", op, err)
	}

	stmtQuiz, err := s.db.Prepare("SELECT id, question, description, image FROM quiz WHERE id = ?")
	if err != nil {
		return models.Quiz{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmtQuiz.Close()

	var quiz models.Quiz
	err = stmtQuiz.QueryRowContext(ctx, quizID).Scan(&quiz.ID, &quiz.Question, &quiz.Description, &quiz.Image)
	if err != nil {
		return models.Quiz{}, fmt.Errorf("%s: %w", op, err)
	}

	quiz.Answers = answers

	return quiz, nil
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
