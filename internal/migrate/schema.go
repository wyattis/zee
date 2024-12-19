package migrate

import (
	"github.com/wyattis/zee/isql/driver"
	"github.com/wyattis/zee/schema"
)

func GetMigrateSchema(driverType driver.Type, schemaName, tableName string) (s *schema.Schema) {
	s = schema.New(driverType, schemaName)
	s.CreateIfNotExists(tableName, func(t *schema.Table) {
		t.Primary("id").Autoincrement()
		t.String("namespace")
		t.Integer("version")
		t.String("hash").Unique()
		t.Boolean("dirty")
		t.Timestamp("started_at").Default(schema.NOW{})
		t.Timestamp("finished_at").Null()
		t.Unique("namespace", "version").IfNotExists()
	})
	return
}
