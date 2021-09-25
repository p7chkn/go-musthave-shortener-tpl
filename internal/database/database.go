package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

type PosrgreDataBase struct {
	conn    *sql.DB
	baseURL string
}

func NewDatabaseRepository(baseURL string, db *sql.DB) handlers.RepositoryInterface {
	return handlers.RepositoryInterface(NewDatabase(baseURL, db))
}

func NewDatabase(baseURL string, db *sql.DB) *PosrgreDataBase {
	result := &PosrgreDataBase{
		conn:    db,
		baseURL: baseURL,
	}
	return result
}

func (db *PosrgreDataBase) Ping() error {

	ctx := context.Background()

	err := db.conn.PingContext(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (db *PosrgreDataBase) AddURL(longURL string, shortURL string, user string) error {

	ctx := context.Background()

	sqlAddRow := `INSERT INTO urls (user_id, origin_url, short_url)
				  VALUES ($1, $2, $3)`

	_, err := db.conn.ExecContext(ctx, sqlAddRow, user, longURL, shortURL)

	if err, ok := err.(*pq.Error); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return handlers.NewUniqueConstraintError(err)
		}
	}

	return err
}

func (db *PosrgreDataBase) GetURL(shortURL string) (string, error) {

	ctx := context.Background()
	sqlGetURLRow := `SELECT origin_url FROM urls WHERE short_url=$1 FETCH FIRST ROW ONLY;`
	query := db.conn.QueryRowContext(ctx, sqlGetURLRow, shortURL)
	result := ""
	query.Scan(&result)
	if result == "" {
		return "", errors.New("not found")
	}
	return result, nil
}

func (db *PosrgreDataBase) GetUserURL(user string) ([]handlers.ResponseGetURL, error) {

	ctx := context.Background()
	result := []handlers.ResponseGetURL{}

	sqlGetUserURL := `SELECT origin_url, short_url FROM urls WHERE user_id=$1;`
	rows, err := db.conn.QueryContext(ctx, sqlGetUserURL, user)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var u handlers.ResponseGetURL
		err = rows.Scan(&u.OriginalURL, &u.ShortURL)
		if err != nil {
			return result, err
		}
		u.ShortURL = db.baseURL + u.ShortURL
		result = append(result, u)
	}

	return result, nil
}

func (db *PosrgreDataBase) AddManyURL(urls []handlers.ManyPostURL, user string) ([]handlers.ManyPostResponse, error) {

	ctx := context.Background()

	result := []handlers.ManyPostResponse{}
	tx, err := db.conn.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO urls (user_id, origin_url, short_url) VALUES ($1, $2, $3)`)
	// _ = ` INSERT INTO urls (user_id, origin_url, short_url) VALUES ('a72c8923-3220-e8b9-0357-da73b5e3373c', 'http://iloverestaurant.ru/','98fv58Wr3hGGIzm2-aH2zA628Ng=')
	// ON CONFLICT (short_url)
	// DO SELECT * FROM urls;`

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	for _, u := range urls {
		shortURL := shortener.ShorterURL(u.OriginalURL)
		if _, err = stmt.ExecContext(ctx, user, u.OriginalURL, shortURL); err != nil {
			return nil, err
		}
		result = append(result, handlers.ManyPostResponse{
			CorrelationID: u.CorrelationID,
			ShortURL:      db.baseURL + shortURL,
		})
	}

	if err != nil {
		return nil, err
	}
	tx.Commit()
	return result, nil
}
