package drivers

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
	"time"
)

func NewMssql(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error = errors.New("trying to connect to MSSQL")

	counter := 0

	for err != nil {
		db, err = sql.Open("sqlserver", dsn)
		if err != nil {
			log.Printf("cannot connect to mssql: %v\n", err)
			counter++
			if counter >= 5 {
				return nil, fmt.Errorf("cannot connect to mssql")
			}
			time.Sleep(5 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			log.Printf("cannot ping postgres: %v\n", err)
			counter++
			if counter >= 5 {
				return nil, fmt.Errorf("cannot pind mssql")
			}
			time.Sleep(5 * time.Second)
			continue
		}
	}
	return db, nil
}
