package isql

import "database/sql"

func init() {
	var _ IDB = &sql.DB{}
}
