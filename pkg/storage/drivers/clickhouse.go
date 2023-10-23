package drivers

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mailru/go-clickhouse/v2"
	"log"
	"time"
)

func NewClickhouse(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error = errors.New("trying to connect to MSSQL")

	counter := 0

	for err != nil {
		db, err = sql.Open("chhttp", dsn)
		if err != nil {
			log.Printf("cannot connect to clickhouse: %v\n", err)
			counter++
			if counter >= 5 {
				return nil, fmt.Errorf("cannot connect to clickhouse")
			}
			time.Sleep(5 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			log.Printf("cannot ping clickhouse: %v\n", err)
			counter++
			if counter >= 5 {
				return nil, fmt.Errorf("cannot pind clickhouse")
			}
			time.Sleep(5 * time.Second)
			continue
		}
	}
	return db, nil
}
