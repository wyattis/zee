package schema

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wyattis/zee/isql/driver"
)

type testStatement struct {
	Table        string
	Create       TableMutator
	Alter        TableMutator
	SqliteResult string
}

var createStatements = []testStatement{
	{
		Table: "test",
		Create: func(t *Table) {
			t.Integer("id")
		},
		SqliteResult: "CREATE TABLE `test` (\n'id' INTEGER NOT NULL\n);",
	},
	{
		Table: "test",
		Create: func(t *Table) {
			t.Integer("id")
			t.VarChar("name", 255)
		},
		SqliteResult: "CREATE TABLE `test` (\n'id' INTEGER NOT NULL,\n'name' TEXT NOT NULL\n);",
	},
	{
		Table: "single_primary",
		Create: func(t *Table) {
			t.Primary("id").Autoincrement()
			t.NVarChar("name", 255).Null()
			t.DateTime("created_at").Default(NOW{})
		},
		SqliteResult: "CREATE TABLE `single_primary` (\n'id' INTEGER PRIMARY KEY AUTOINCREMENT,\n'name' TEXT NULL,\n'created_at' TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP\n);",
	},
	{
		Table: "multiple_primary",
		Create: func(t *Table) {
			t.String("id").Primary()
			t.String("user_id").Primary()
		},
		SqliteResult: "CREATE TABLE `multiple_primary` (\n'id' TEXT NOT NULL,\n'user_id' TEXT NOT NULL,\nPRIMARY KEY ('id', 'user_id')\n);",
	},
	{
		Table: "single_unique",
		Create: func(t *Table) {
			t.Primary("id")
			t.String("username").Unique()
		},
		SqliteResult: "CREATE TABLE `single_unique` (\n'id' INTEGER PRIMARY KEY,\n'username' TEXT NOT NULL UNIQUE\n);",
	},
	{
		Table: "multiple_unique_columns",
		Create: func(t *Table) {
			t.Primary("id")
			t.String("username").Unique()
			t.String("email").Unique()
		},
		SqliteResult: "CREATE TABLE `multiple_unique_columns` (\n'id' INTEGER PRIMARY KEY,\n'username' TEXT NOT NULL UNIQUE,\n'email' TEXT NOT NULL UNIQUE\n);",
	},
	{
		Table: "compound_unique",
		Create: func(t *Table) {
			t.String("username")
			t.String("email")
			t.Unique("username", "email")
		},
		SqliteResult: "CREATE TABLE `compound_unique` (\n'username' TEXT NOT NULL,\n'email' TEXT NOT NULL);CREATE UNIQUE INDEX 'unq_compound_unique_username_email' ON `compound_unique`('username', 'email');",
	},
	{
		Table: "single_index",
		Create: func(t *Table) {
			t.Primary("id")
			t.String("username").Index("idx_username")
		},
		SqliteResult: "CREATE TABLE `single_index` (\n'id' INTEGER PRIMARY KEY,\n'username' TEXT NOT NULL);CREATE INDEX `idx_username` ON `single_index`('username');\n",
	},
	{
		Table: "single_foreign_key",
		Create: func(t *Table) {
			t.Primary("id")
			t.Integer("user_id").References("users", "id")
			t.String("username")
		},
		SqliteResult: "CREATE TABLE `single_foreign_key` (\n'id' INTEGER PRIMARY KEY,\n'user_id' INTEGER NOT NULL,\n'username' TEXT NOT NULL,\nFOREIGN KEY ('user_id') REFERENCES `users`('id'));\n",
	},
	{
		Table: "multiple_foreign_keys",
		Create: func(t *Table) {
			t.Integer("user_id").References("user", "id")
			t.Integer("study_id").References("study", "id")
			t.Unique("user_id", "study_id")
		},
		SqliteResult: "CREATE TABLE `multiple_foreign_keys` (\n'user_id' INTEGER NOT NULL,\n'study_id' INTEGER NOT NULL,\nFOREIGN KEY ('user_id') REFERENCES `user`('id'),\nFOREIGN KEY ('study_id') REFERENCES `study`('id'));CREATE UNIQUE INDEX 'unq_multiple_foreign_keys_user_id_study_id' ON `multiple_foreign_keys`('user_id', 'study_id');\n",
	},
}

func TestSqliteCreate(t *testing.T) {
	for i, s := range createStatements {
		t.Run(fmt.Sprintf("Create '%s' - %d", s.Table, i), func(t *testing.T) {
			schema := New(driver.TypeSqlite3, "test")
			var table *Table
			schema.Create(s.Table, func(t *Table) {
				table = t
				s.Create(t)
			})
			t.Logf("Create '%s' - %d", s.Table, i)
			statements := table.tableDef.Statements()
			sql := strings.Join(statements, ";") + ";"
			if !sqlStatementsAreEqual(s.SqliteResult, sql) {
				t.Errorf("Expected \n%s\n but got \n%s\n", strings.TrimSpace(s.SqliteResult), strings.TrimSpace(sql))
			}
		})
	}
}

var alterStatements = []testStatement{
	{
		Table: "column_rename",
		Alter: func(t *Table) {
			t.Column("id").Name("new_id")
		},
		SqliteResult: "ALTER TABLE `column_rename` RENAME COLUMN `id` TO `new_id`;",
	},
}

func TestSqliteAlter(t *testing.T) {
	for i, s := range alterStatements {
		schema := New(driver.TypeSqlite3, "test")
		var table *Table
		schema.Table(s.Table, func(t *Table) {
			table = t
			s.Alter(t)
		})
		t.Logf("Alter '%s' - %d", s.Table, i)
		statements := table.tableDef.Statements()
		sql := strings.Join(statements, ";") + ";"
		if !sqlStatementsAreEqual(s.SqliteResult, sql) {
			t.Errorf("Expected \n%s\n but got \n%s\n", strings.TrimSpace(s.SqliteResult), strings.TrimSpace(sql))
		}
	}
}
