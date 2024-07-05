// app/db.go
package main

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

// getDB initializes and returns a singleton database connection pool
func getDB() (*sql.DB, error) {
	var err error
	once.Do(func() {
		db, err = sql.Open("sqlite3", "/app/data/data.db")
		if err == nil {
			// Set up connection pooling parameters if needed
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(5)
		}
	})
	return db, err
}
