package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	// wrapper around sqlx.DB with simple structured logging
	X      *sqlx.DB
	logger *log.Logger
}

func NewDB(db *sqlx.DB, logger *log.Logger) *DB {
	return &DB{X: db, logger: logger}
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	start := time.Now()
	res, err := db.X.Exec(query, args...)
	db.logQuery("Exec", query, args, time.Since(start), err)
	return res, err
}

func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.X.Query(query, args...)
	db.logQuery("Query", query, args, time.Since(start), err)
	return rows, err
}

func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	start := time.Now()
	row := db.X.QueryRow(query, args...)
	db.logQuery("QueryRow", query, args, time.Since(start), nil)
	return row
}

func (db *DB) logQuery(op, query string, args []any, dur time.Duration, err error) {
	if db == nil || db.logger == nil {
		return
	}
	msg := fmt.Sprintf("op=%s duration=%s query=%q args=%v", op, dur, query, args)
	if err != nil {
		db.logger.Printf("%s error=%v", msg, err)
		return
	}
	db.logger.Printf("%s", msg)
}
