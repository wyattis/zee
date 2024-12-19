package main

import (
	"database/sql"
	"log/slog"
	"os"
)

var sqliteDbFile = ":memory:"

func setupSqlite() (db *sql.DB, err error) {
	SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	db, err = sql.Open("sqlite3", sqliteDbFile)
	if err != nil {
		return
	}
	return
}
