// Package database - пакет для взаимодействия с базой данных Postgres.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/responses"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/usecases"
	custom_errors "github.com/p7chkn/go-musthave-shortener-tpl/internal/errors"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

// GetURLData - структура для возвращения данных о URL.
type GetURLData struct {
	OriginURL string
	IsDeleted bool
}

// PostgresDataBase - структура для взаимодейтсивя с базой данных.
type PostgresDataBase struct {
	conn    *sql.DB
	baseURL string
}

// NewDatabaseRepository - создание нового интерфейства для репозитория.
func NewDatabaseRepository(baseURL string, db *sql.DB) usecases.UserRepositoryInterface {
	return usecases.UserRepositoryInterface(NewDatabase(baseURL, db))
}

// NewDatabase - создание новой структуры взаимодействия с базой данных.
func NewDatabase(baseURL string, db *sql.DB) *PostgresDataBase {
	result := &PostgresDataBase{
		conn:    db,
		baseURL: baseURL,
	}
	return result
}

// Ping - проверка подключения к базе данных.
func (db *PostgresDataBase) Ping(ctx context.Context) error {

	err := db.conn.PingContext(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// AddURL - добавление записи о новой сокращенной URL.
func (db *PostgresDataBase) AddURL(ctx context.Context, longURL string, shortURL string, user string) error {

	sqlAddRow := `INSERT INTO urls (user_id, origin_url, short_url)
				  VALUES ($1, $2, $3)`

	_, err := db.conn.ExecContext(ctx, sqlAddRow, user, longURL, shortURL)

	if err, ok := err.(*pq.Error); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return custom_errors.NewErrorWithDB(err, "UniqConstraint")
		}
	}

	return err
}

// GetURL - получение данных о изначальном URL по сокращенному URL.
func (db *PostgresDataBase) GetURL(ctx context.Context, shortURL string) (string, error) {

	sqlGetURLRow := `SELECT origin_url, is_deleted FROM urls WHERE short_url=$1 FETCH FIRST ROW ONLY;`
	query := db.conn.QueryRowContext(ctx, sqlGetURLRow, shortURL)
	result := GetURLData{}
	if err := query.Scan(&result.OriginURL, &result.IsDeleted); err != nil {
		return "", nil
	}
	if result.OriginURL == "" {
		return "", custom_errors.NewErrorWithDB(errors.New("not found"), "Not found")
	}
	if result.IsDeleted {
		return "", custom_errors.NewErrorWithDB(errors.New("deleted"), "deleted")
	}
	return result.OriginURL, nil
}

// GetUserURL - получение всех URL пользователя.
func (db *PostgresDataBase) GetUserURL(ctx context.Context, user string) ([]responses.GetURL, error) {

	var result []responses.GetURL

	sqlGetUserURL := `SELECT origin_url, short_url FROM urls WHERE user_id=$1 AND is_deleted=false;`
	rows, err := db.conn.QueryContext(ctx, sqlGetUserURL, user)
	if err != nil {
		return result, err
	}
	if rows.Err() != nil {
		return result, rows.Err()
	}
	defer rows.Close()

	for rows.Next() {
		var u responses.GetURL
		err = rows.Scan(&u.OriginalURL, &u.ShortURL)
		if err != nil {
			return result, err
		}
		u.ShortURL = db.baseURL + u.ShortURL
		result = append(result, u)
	}

	return result, nil
}

// AddManyURL - добавление многих URL сразу.
func (db *PostgresDataBase) AddManyURL(ctx context.Context, urls []responses.ManyPostURL, user string) ([]responses.ManyPostResponse, error) {

	var result []responses.ManyPostResponse
	tx, err := db.conn.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO urls (user_id, origin_url, short_url) VALUES ($1, $2, $3)`)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	for _, u := range urls {
		shortURL := shortener.ShorterURL(u.OriginalURL)
		if _, err = stmt.ExecContext(ctx, user, u.OriginalURL, shortURL); err != nil {
			return nil, err
		}
		result = append(result, responses.ManyPostResponse{
			CorrelationID: u.CorrelationID,
			ShortURL:      db.baseURL + shortURL,
		})
	}

	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	return result, err
}

// DeleteManyURL - удаление многиз URL по id.
func (db *PostgresDataBase) DeleteManyURL(ctx context.Context, urls []string, user string) error {

	sqlDeleteURL := `UPDATE urls SET is_deleted = true WHERE short_url = ANY ($1);`
	var urlsToDelete []string
	for _, url := range urls {
		if db.isOwner(ctx, url, user) {
			urlsToDelete = append(urlsToDelete, url)
		}
	}
	_, err := db.conn.ExecContext(ctx, sqlDeleteURL, pq.Array(urlsToDelete))
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgresDataBase) GetStats(ctx context.Context) (responses.StatResponse, error) {
	sqlGetStats := `SELECT COUNT(DISTINCT user_id), COUNT (DISTINCT origin_url) FROM urls;`
	query := db.conn.QueryRowContext(ctx, sqlGetStats)
	result := responses.StatResponse{}

	err := query.Scan(&result.CountUser, &result.CountURL)
	return result, err

}

// isOwner - вспомогательная функция, которая определняет владелец ли переданный
// пользователь, указанной записи сокращенного URL.
func (db *PostgresDataBase) isOwner(ctx context.Context, url string, user string) bool {
	sqlGetURLRow := `SELECT user_id FROM urls WHERE short_url=$1 FETCH FIRST ROW ONLY;`
	query := db.conn.QueryRowContext(ctx, sqlGetURLRow, url)
	result := ""
	query.Scan(&result)
	return result == user
}
