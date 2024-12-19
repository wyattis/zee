package isql

import (
	"database/sql"
	"fmt"

	"github.com/wyattis/zee/isql/driver"
	_ "github.com/wyattis/zee/isql/drivers/mysql"
	_ "github.com/wyattis/zee/isql/drivers/postgres"
	_ "github.com/wyattis/zee/isql/drivers/sqlite3"
)

var (
	ErrDriverNotInBuild   = fmt.Errorf("current executable was not built with support for this driver")
	ErrSchemaVersionOlder = fmt.Errorf("schema version is older than the current executable")
	ErrSchemaVersionNewer = fmt.Errorf("schema version is newer than the current executable")
)

func Open(config driver.Config) (db *sql.DB, err error) {
	connector, ok := driver.Connectors[config.Driver]
	if !ok {
		return nil, ErrDriverNotInBuild
	}
	return connector(config)
}

func QueryParams(n int) string {
	p := ""
	for i := 0; i < n; i++ {
		p += "?"
		if i < n-1 {
			p += ", "
		}
	}
	return p
}

func NamedColumns(columns []string) []string {
	c := make([]string, len(columns))
	for i, col := range columns {
		c[i] = fmt.Sprintf(":%s", col)
	}
	return c
}
