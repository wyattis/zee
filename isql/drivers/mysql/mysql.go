//go:build all || postgres || !nopostgres
// +build all postgres !nopostgres

package mysql

import (
	"database/sql"

	"github.com/wyattis/zee/isql/driver"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	driver.Connectors[driver.TypeMysql] = func(config driver.Config) (*sql.DB, error) {
		return connectMysql(config)
	}
}

func connectMysql(config driver.Config) (db *sql.DB, err error) {
	return sql.Open(driver.TypeMysql.String(), config.MysqlString())
}
