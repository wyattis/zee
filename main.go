package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"

	"github.com/wyattis/zee/isql"
	"github.com/wyattis/zee/isql/driver"
	"github.com/wyattis/zee/schema"
)

var (
	ErrSchemaVersionHigherThanTarget = fmt.Errorf("schema version is higher than target version")
	ErrSchemaVersionLowerThanTarget  = fmt.Errorf("schema version is lower than target version")
	ErrNoMigrationForVersion         = fmt.Errorf("no migration for version")
	ErrDatabaseIsDirty               = fmt.Errorf("database is dirty")
)

var logger = slog.New(slog.NewJSONHandler(io.Discard, nil))

func SetLogger(l *slog.Logger) {
	logger = l
}

type MigrateOptions struct {
	Driver         driver.Type
	Namespace      string
	SchemaName     string
	MigrationTable string
}

func (o *MigrateOptions) Default() {
	if o.Driver == "" {
		o.Driver = driver.TypeSqlite3
	}
	if o.Namespace == "" {
		o.Namespace = "default"
	}
	if o.SchemaName == "" {
		o.SchemaName = "default"
	}
	if o.MigrationTable == "" {
		o.MigrationTable = "schema_migrations"
	}
}

type SchemaMutator func(s *schema.Schema)
type Migration struct {
	Version uint
	Hash    []byte
	Up      SchemaMutator
	Down    SchemaMutator
}

var Migrations = []Migration{}

func Add(migration Migration) {
	Migrations = append(Migrations, migration)
}

// MigrateUpTo takes a list of migrations and applies them up to the provided version according to the provided options.
func MigrateUpTo(migrations []Migration, db isql.IDB, version uint, opts *MigrateOptions) (err error) {
	if opts == nil {
		opts = &MigrateOptions{}
	}
	opts.Default()
	mi := NewMigrator(db, *opts)
	mi.Add(migrations...)
	return mi.UpTo(version)
}

// Migrate up to the provided version. Throws an errors if there isn't a migration matching the provided version, if the
// schema version is higher than the provided version or if the database is dirty (failed a previous migration).
func UpTo(db *sql.DB, driverType driver.Type, version uint, opts *MigrateOptions) (err error) {
	if opts == nil {
		opts = &MigrateOptions{}
	}
	opts.Default()
	if len(Migrations) == 0 {
		return errors.New("No migrations registered. Did you forget to import or add your migrations?")
	}
	return MigrateUpTo(Migrations, db, version, opts)
}

func validateMigration(migrations []Migration, db isql.IDB, opts MigrateOptions, version uint) (schemaVersion uint, err error) {
	sort.Slice(Migrations, func(i, j int) bool {
		return Migrations[i].Version < Migrations[j].Version
	})
	if !hasMatchingVersion(migrations, version) {
		err = ErrNoMigrationForVersion
		return
	}
	if !databaseIsClean(db, opts) {
		err = ErrDatabaseIsDirty
		return
	}
	return currentVersion(db, opts.Driver, opts)
}

func databaseIsClean(db isql.IQueryRow, opts MigrateOptions) bool {
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT count(*) FROM `%s` where dirty", opts.MigrationTable)).Scan(&count)
	return count == 0 && (err == nil || strings.Contains(err.Error(), "no such table"))
}

func hasMatchingVersion(migrations []Migration, version uint) bool {
	hasMatchingVersion := false
	for _, m := range migrations {
		if m.Version == version {
			hasMatchingVersion = true
			break
		}
	}
	return hasMatchingVersion
}

func currentVersion(db isql.IDB, driverType driver.Type, opts MigrateOptions) (version uint, err error) {
	if err = initializeSchema(db, driverType, opts); err != nil {
		return
	}
	q := fmt.Sprintf("SELECT `version` FROM `%s` WHERE `namespace` = ? ORDER BY `version` DESC LIMIT 1", opts.MigrationTable)
	err = db.QueryRow(q, opts.Namespace).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return
}

func initializeSchema(db isql.IBegin, driverType driver.Type, opts MigrateOptions) (err error) {
	return isql.Begin(db, func(tx *sql.Tx) (err error) {
		s := GetMigrateSchema(driverType, opts.SchemaName, opts.MigrationTable)
		return s.Schema.Run(tx, logger)
	})
}
