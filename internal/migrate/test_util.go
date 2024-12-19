package migrate

import (
	"database/sql"

	"github.com/wyattis/zee/isql/driver"
)

var sqliteDbFile = ":memory:"

func setupSqlite() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", sqliteDbFile)
	if err != nil {
		return
	}
	opts := MigrateOptions{}
	opts.Default()
	err = initializeSchema(db, driver.TypeSqlite3, opts)
	return
}
