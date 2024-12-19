//go:build all || sqlite3 || !nosqlite3
// +build all sqlite3 !nosqlite3

package sqlite3

import (
	"database/sql"

	"github.com/wyattis/zee/isql/driver"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	driver.Connectors[driver.TypeSqlite3] = func(config driver.Config) (db *sql.DB, err error) {
		db, err = sql.Open(driver.TypeSqlite3.String(), config.Database)
		if err != nil {
			return
		}
		_, err = db.Exec("PRAGMA foreign_keys = ON")
		return
	}
}
