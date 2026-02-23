package storage

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	X      *sqlx.DB
	logger *slog.Logger
}

func NewDB(db *sqlx.DB, logger *slog.Logger) *DB {
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

func (db *DB) Get(dest any, query string, args ...any) error {
	start := time.Now()
	err := db.X.Get(dest, query, args...)
	db.logQuery("Get", query, args, time.Since(start), err)
	return err
}

func (db *DB) Select(dest any, query string, args ...any) error {
	start := time.Now()
	err := db.X.Select(dest, query, args...)
	db.logQuery("Select", query, args, time.Since(start), err)
	return err
}

func (db *DB) logQuery(op, query string, args []any, dur time.Duration, err error) {
	if db == nil || db.logger == nil {
		return
	}
	query = fmt.Sprintf(query, args...)

	logger := db.logger.With(
		slog.Group(
			"sql_query",
			slog.String("op", op),
			slog.Duration("duration", dur),
			slog.String("query", query),
		),
	)

	if err != nil {
		logger.Error(fmt.Sprintf("Error executing sql query: %v", err))
		return
	}

	if dur > (time.Millisecond * 500) {
		logger.Warn("SQL query executed too long")
		return
	}

	db.logger.Debug("SQL query executed")
}
