package database

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type DataBase struct {
	URI string
}

func Ping(URI string) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, URI)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	err = conn.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}
