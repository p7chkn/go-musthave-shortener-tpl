package database

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

type PosrgreDataBase struct {
	URI     string
	baseURL string
}

func NewDatabaseRepository(cfg *configuration.Config) handlers.RepositoryInterface {
	return handlers.RepositoryInterface(NewDatabase(cfg))
}

func NewDatabase(cfg *configuration.Config) *PosrgreDataBase {
	result := &PosrgreDataBase{
		URI:     cfg.DataBase.DataBaseURI,
		baseURL: cfg.BaseURL,
	}
	result.setUp()
	return result
}

func (db *PosrgreDataBase) Ping() error {
	conn, ctx := db.connect()
	defer conn.Close(ctx)

	err := conn.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (db *PosrgreDataBase) setUp() error {
	conn, ctx := db.connect()
	defer conn.Close(ctx)

	query := conn.QueryRow(ctx, "SELECT 'exists' FROM pg_tables WHERE tablename='urls';")
	var result string
	query.Scan(&result)

	if result != "exists" {
		var extention string
		query := conn.QueryRow(ctx, "SELECT 'exists' FROM pg_extension WHERE extname='uuid-ossp';")
		query.Scan(&extention)
		if extention != "exists" {
			_, err := conn.Exec(ctx, `CREATE EXTENSION "uuid-ossp";`)
			if err != nil {
				return err
			}
			log.Println("Create EXTENSION")
		}
		sqlCreateDB := `CREATE TABLE urls (
									id serial PRIMARY KEY,
									user_id uuid DEFAULT uuid_generate_v4 (), 	
									origin_url VARCHAR NOT NULL, 
									short_url VARCHAR NOT NULL
						);`
		_, err := conn.Exec(ctx, sqlCreateDB)
		log.Println("Create table", err)
		return err
	}
	log.Println("Table already exists")
	return nil
}

func (db *PosrgreDataBase) AddURL(longURL string, shortURL string, user string) error {
	conn, ctx := db.connect()
	defer conn.Close(ctx)

	sqlAddRow := `INSERT INTO urls (user_id, origin_url, short_url)
				  VALUES ($1, $2, $3)`

	_, err := conn.Exec(ctx, sqlAddRow, user, longURL, shortURL)

	return err
}

func (db *PosrgreDataBase) GetURL(shortURL string) (string, error) {
	conn, ctx := db.connect()
	defer conn.Close(ctx)
	sqlGetURLRow := `SELECT origin_url FROM urls WHERE short_url=$1 FETCH FIRST ROW ONLY;`
	query := conn.QueryRow(ctx, sqlGetURLRow, shortURL)
	result := ""
	query.Scan(&result)
	if result == "" {
		return "", errors.New("not found")
	}
	return result, nil
}

func (db *PosrgreDataBase) GetUserURL(user string) ([]handlers.ResponseGetURL, error) {
	conn, ctx := db.connect()
	defer conn.Close(ctx)
	result := []handlers.ResponseGetURL{}

	sqlGetUserURL := `SELECT origin_url, short_url FROM urls WHERE user_id=$1;`
	rows, err := conn.Query(ctx, sqlGetUserURL, user)
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
	conn, ctx := db.connect()
	defer conn.Close(ctx)

	result := []handlers.ManyPostResponse{}
	tx, err := conn.Begin(ctx)
	defer tx.Rollback(ctx)

	sqlAddRow := `INSERT INTO urls (user_id, origin_url, short_url)
	VALUES ($1, $2, $3)`

	for _, u := range urls {
		shortURL := shortener.ShorterURL(u.OriginalURL)
		if _, err = tx.Exec(ctx, sqlAddRow, user, u.OriginalURL, shortURL); err != nil {
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
	tx.Commit(ctx)
	return result, nil
}

func (db *PosrgreDataBase) connect() (*pgx.Conn, context.Context) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, db.URI)
	if err != nil {
		log.Fatal(err)
	}
	return conn, ctx
}
