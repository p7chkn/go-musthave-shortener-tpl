package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/p7chkn/go-musthave-shortener-tpl/cmd/shortener/configuration"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/models"
)

type PosrgreDataBase struct {
	URI string
}

type TabaleURLS struct {
	user     string
	shortURL string
	longURL  string
}

func New(cfg configuration.ConfigDatabase) *PosrgreDataBase {
	result := &PosrgreDataBase{
		URI: cfg.DataBaseURI,
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
		fmt.Println("Creating table")
		return nil
	}
	fmt.Println("Table already exists")
	return nil
}

func (db *PosrgreDataBase) AddURL(longURL string, shortURL string, user string) error {
	// ctx := context.Background()
	// conn, err := pgx.Connect(ctx, db.URI)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (db *PosrgreDataBase) GetURL(shortURL string) (string, error) {
	// ctx := context.Background()
	// conn, err := pgx.Connect(ctx, db.URI)
	// if err != nil {
	// 	return err
	// }
	return "", nil
}

func (db *PosrgreDataBase) GetUserURL(user string) []models.ResponseGetURL {
	// ctx := context.Background()
	// conn, err := pgx.Connect(ctx, db.URI)
	// if err != nil {
	// 	return err
	// }
	return []models.ResponseGetURL{}
}

func (db *PosrgreDataBase) connect() (*pgx.Conn, context.Context) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, db.URI)
	if err != nil {
		log.Fatal(err)
	}
	return conn, ctx
}
