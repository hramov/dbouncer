package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/hramov/dbouncer/internal"
	"github.com/hramov/dbouncer/pkg/jsonify"
	"github.com/hramov/dbouncer/pkg/storage/drivers"
)

type storage struct {
	db *sql.DB
}

func (s *storage) QueryTx(ctx context.Context, query string, args ...interface{}) ([]string, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	defer tx.Commit()

	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return jsonify.Jsonify(rows), nil
}

func (s *storage) QueryRowTx(ctx context.Context, query string, args ...interface{}) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (s *storage) ExecTx(ctx context.Context, query string, args ...interface{}) ([]string, error) {
	return nil, errors.New("not implemented")
}

func New(name string, dsn string) (internal.Storage, error) {
	switch name {
	case "postgres":
		db, err := drivers.NewPostgres(dsn)
		if err != nil {
			return nil, err
		}
		return &storage{
			db: db,
		}, nil
	default:
		return nil, fmt.Errorf("unknown storage: %s", name)
	}
}
