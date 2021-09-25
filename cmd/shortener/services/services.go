package services

import (
	"context"
	"database/sql"
	"log"
)

func SetUpDataBase(db *sql.DB) error {

	ctx := context.Background()

	var extention string
	query := db.QueryRowContext(ctx, "SELECT 'exists' FROM pg_extension WHERE extname='uuid-ossp';")
	query.Scan(&extention)
	if extention != "exists" {
		_, err := db.ExecContext(ctx, `CREATE EXTENSION "uuid-ossp";`)
		if err != nil {
			return err
		}
		log.Println("Create EXTENSION")
	}
	sqlCreateDB := `CREATE TABLE IF NOT EXISTS urls (
								id serial PRIMARY KEY,
								user_id uuid DEFAULT uuid_generate_v4 (), 	
								origin_url VARCHAR NOT NULL, 
								short_url VARCHAR NOT NULL UNIQUE
					);`
	res, err := db.ExecContext(ctx, sqlCreateDB)
	log.Println("Create table", err, res)
	return nil
}
