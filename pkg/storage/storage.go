package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/hramov/dbouncer/internal"
	"github.com/hramov/dbouncer/pkg/jsonify"
	"github.com/hramov/dbouncer/pkg/storage/drivers"
	"log"
	"sync"
	"time"
)

type storage struct {
	db *sql.DB
}

type storageMap = map[string]internal.Storage

var storages storageMap
var mu = &sync.RWMutex{}

func (s *storage) QueryTx(ctx context.Context, query string, args ...interface{}) ([]string, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	defer tx.Commit()

	if err != nil {
		log.Println("db error: ", err.Error())
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

func New(name string, dsn string, maxOpenConns int, maxIdleConns int, idleTimeout time.Duration, lifeTime time.Duration) (internal.Storage, error) {

	var db *sql.DB
	var err error

	if storages == nil {
		storages = make(storageMap)
	}

	mu.RLock()
	st, ok := storages[name]
	mu.RUnlock()

	if ok {
		return st, nil
	}

	switch name {
	case "postgres":
		db, err = drivers.NewPostgres(dsn)
		if err != nil {
			return nil, err
		}
	case "mssql":
		db, err = drivers.NewMssql(dsn)
		if err != nil {
			return nil, err
		}
	case "clickhouse":
		db, err = drivers.NewClickhouse(dsn)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown storage: %s", name)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(lifeTime)
	db.SetConnMaxIdleTime(idleTimeout)

	st = &storage{
		db: db,
	}

	mu.Lock()
	storages[name] = st
	mu.Unlock()

	return st, nil
}

func GetStorage(name string) internal.Storage {
	mu.RLock()
	defer mu.RUnlock()

	st, ok := storages[name]
	if !ok {
		return nil
	}
	return st
}
