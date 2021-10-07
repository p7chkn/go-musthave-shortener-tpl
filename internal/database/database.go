package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"

	"github.com/p7chkn/go-musthave-shortener-tpl/internal/app/handlers"
	"github.com/p7chkn/go-musthave-shortener-tpl/internal/shortener"
)

const numOfWorkers = 10

type GetURLdata struct {
	OriginURL string
	IsDeleted bool
}

type PosrgreDataBase struct {
	conn    *sql.DB
	baseURL string
	ctx     context.Context
}

func NewDatabaseRepository(ctx context.Context, baseURL string, db *sql.DB) handlers.RepositoryInterface {
	return handlers.RepositoryInterface(NewDatabase(ctx, baseURL, db))
}

func NewDatabase(ctx context.Context, baseURL string, db *sql.DB) *PosrgreDataBase {
	result := &PosrgreDataBase{
		conn:    db,
		baseURL: baseURL,
		ctx:     ctx,
	}
	return result
}

func (db *PosrgreDataBase) Ping() error {

	err := db.conn.PingContext(db.ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (db *PosrgreDataBase) AddURL(longURL string, shortURL string, user string) error {

	sqlAddRow := `INSERT INTO urls (user_id, origin_url, short_url)
				  VALUES ($1, $2, $3)`

	_, err := db.conn.ExecContext(db.ctx, sqlAddRow, user, longURL, shortURL)

	if err, ok := err.(*pq.Error); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return handlers.NewErrorWithDB(err, "UniqConstraint")
		}
	}

	return err
}

func (db *PosrgreDataBase) GetURL(shortURL string) (string, error) {

	sqlGetURLRow := `SELECT origin_url, is_deleted FROM urls WHERE short_url=$1 FETCH FIRST ROW ONLY;`
	query := db.conn.QueryRowContext(db.ctx, sqlGetURLRow, shortURL)
	result := GetURLdata{}
	query.Scan(&result.OriginURL, &result.IsDeleted)
	if result.OriginURL == "" {
		return "", handlers.NewErrorWithDB(errors.New("not found"), "Not found")
	}
	if result.IsDeleted {
		return "", handlers.NewErrorWithDB(errors.New("Deleted"), "Deleted")
	}
	return result.OriginURL, nil
}

func (db *PosrgreDataBase) GetUserURL(user string) ([]handlers.ResponseGetURL, error) {

	result := []handlers.ResponseGetURL{}

	sqlGetUserURL := `SELECT origin_url, short_url FROM urls WHERE user_id=$1;`
	rows, err := db.conn.QueryContext(db.ctx, sqlGetUserURL, user)
	if err != nil {
		return result, err
	}
	if rows.Err() != nil {
		return result, rows.Err()
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

	result := []handlers.ManyPostResponse{}
	tx, err := db.conn.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(db.ctx, `INSERT INTO urls (user_id, origin_url, short_url) VALUES ($1, $2, $3)`)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	for _, u := range urls {
		shortURL := shortener.ShorterURL(u.OriginalURL)
		if _, err = stmt.ExecContext(db.ctx, user, u.OriginalURL, shortURL); err != nil {
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

func (db *PosrgreDataBase) DeleteManyURL(urls []string, user string) error {
	// inputCh := make(chan []string)
	// go func() {
	// 	for _, url := range urls {
	// 		inputCh <- []string{url, user}
	// 	}
	// 	close(inputCh)
	// }()

	// setOfurls := []string{}

	// fanOutChs := fanOut(inputCh, numOfWorkers)
	// workerChs := make([]chan string, 0, numOfWorkers)
	// for _, fanOutCh := range fanOutChs {
	// 	newWorker := db.chanIsOwner(fanOutCh)
	// 	workerChs = append(workerChs, newWorker)
	// }

	// for url := range fanIn(workerChs...) {
	// 	setOfurls = append(setOfurls, url)
	// }

	sql := `UPDATE urls SET is_deleted = true WHERE short_url = ANY ($1);`
	_, err := db.conn.ExecContext(db.ctx, sql, pq.Array(urls))
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (db *PosrgreDataBase) IsOwner(url string, user string) bool {
	sqlGetURLRow := `SELECT user_id FROM urls WHERE short_url=$1 FETCH FIRST ROW ONLY;`
	query := db.conn.QueryRowContext(db.ctx, sqlGetURLRow, url)
	result := ""
	query.Scan(&result)
	return result == user
}

func fanOut(inputCh chan []string, n int) []chan []string {
	cs := make([]chan []string, 0, n)
	for i := 0; i < n; i++ {
		cs = append(cs, make(chan []string))
	}

	go func() {
		defer func(cs []chan []string) {
			for _, c := range cs {
				close(c)
			}
		}(cs)

		for i := 0; i < len(cs); i++ {
			if i == len(cs)-1 {
				i = 0
			}

			num, ok := <-inputCh
			if !ok {
				return
			}

			cs[i] <- num
		}
	}()

	return cs
}

func (db *PosrgreDataBase) chanIsOwner(input <-chan []string) (out chan string) {
	out = make(chan string)

	go func() {
		for item := range input {
			isOwner := db.IsOwner(item[0], item[1])
			if isOwner {
				out <- item[0]
			}
		}

		close(out)
	}()

	return out
}

func fanIn(chs ...chan string) (out chan string) {
	out = make(chan string)

	go func() {
		wg := &sync.WaitGroup{}

		for _, ch := range chs {
			wg.Add(1)

			go func(items chan string) {
				defer wg.Done()
				for item := range items {
					out <- item
				}
			}(ch)
		}

		wg.Wait()
		close(out)
	}()

	return out
}
