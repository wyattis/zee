//go:build all || postgres || !nopostgres
// +build all postgres !nopostgres

package postgres

import (
	"database/sql"
	"fmt"

	"github.com/wyattis/zee/isql/driver"

	_ "github.com/lib/pq"
)

func init() {
	driver.Connectors[driver.TypePostgres] = func(config driver.Config) (*sql.DB, error) {
		return connectPostgres(config)
	}
}

func connectPostgres(config driver.Config) (db *sql.DB, err error) {
	if config.Database == "" {
		return nil, fmt.Errorf("database name must be specified to connect to postgres")
	} else if config.User == "" {
		return nil, fmt.Errorf("user must be specified to connect to postgres")
	}
	return sql.Open(driver.TypePostgres.String(), config.PostgresString())
}
