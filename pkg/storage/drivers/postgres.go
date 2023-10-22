package drivers

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func NewPostgres(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error = errors.New("trying to connect to database")

	counter := 0

	for err != nil {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("cannot connect to postgres: %v\n", err)
			counter++
			if counter >= 5 {
				return nil, fmt.Errorf("cannot pind postgres")
			}
			time.Sleep(5 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			log.Printf("cannot ping postgres: %v\n", err)
			counter++
			if counter >= 5 {
				return nil, fmt.Errorf("cannot pind postgres")
			}
			time.Sleep(5 * time.Second)
			continue
		}
	}
	return db, nil
}
