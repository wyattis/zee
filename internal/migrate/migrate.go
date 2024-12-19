package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/wyattis/zee/isql"
	"github.com/wyattis/zee/isql/driver"
	"github.com/wyattis/zee/schema"
)

var logger schema.Printfer = log.New(io.Discard, "", 0)

func SetLogger(l schema.Printfer) {
	logger = l
}

var (
	ErrSchemaVersionHigherThanTarget = fmt.Errorf("schema version is higher than target version")
	ErrSchemaVersionLowerThanTarget  = fmt.Errorf("schema version is lower than target version")
	ErrNoMigrationForVersion         = fmt.Errorf("no migration for version")
	ErrDatabaseIsDirty               = fmt.Errorf("database is dirty")
)

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

func Begin(db isql.IBegin, fn func(tx *sql.Tx) (err error)) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	hasCommitted := false
	defer func() {
		if !hasCommitted {
			fmt.Println("rolling back")
			err2 := tx.Rollback()
			if err2 != nil {
				fmt.Println("rollback err", err2)
			}
		}
	}()
	if err = fn(tx); err != nil {
		return
	}
	if err = tx.Commit(); err != nil {
		return
	}
	hasCommitted = true
	return
}

func initializeSchema(db isql.IBegin, driverType driver.Type, opts MigrateOptions) (err error) {
	return Begin(db, func(tx *sql.Tx) (err error) {
		s := GetMigrateSchema(driverType, opts.SchemaName, opts.MigrationTable)
		fmt.Println("initializing", s.Schema.Statements())
		return s.Schema.Run(tx, logger)
	})
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

func databaseIsClean(db isql.IQueryRow, opts MigrateOptions) bool {
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT count(*) FROM `%s` where dirty", opts.MigrationTable)).Scan(&count)
	return count == 0 && (err == nil || strings.Contains(err.Error(), "no such table"))
}

func validateMigration(migrations []Migration, db isql.IDB, driver driver.Type, opts MigrateOptions, version uint) (schemaVersion uint, err error) {
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
	return currentVersion(db, driver, opts)
}

func migrateDownTo(migrations []Migration, db *sql.DB, driver driver.Type, version uint, opts *MigrateOptions) (err error) {
	if opts == nil {
		opts = &MigrateOptions{}
	}
	opts.Default()
	schemaVersion, err := validateMigration(migrations, db, driver, *opts, version)
	if err != nil {
		return
	}
	if schemaVersion < version {
		err = ErrSchemaVersionLowerThanTarget
		return
	}
	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		if m.Version <= schemaVersion && m.Version > version {
			err = Begin(db, func(tx *sql.Tx) (err error) {
				// mark current migration as dirty before we start
				q := fmt.Sprintf("UPDATE `%s` SET `dirty` = ? WHERE `version` = ? and `namespace` = ?", opts.MigrationTable)
				_, err = tx.Exec(q, true, m.Version, opts.Namespace)
				if err != nil {
					return
				}
				s := schema.New(driver, opts.SchemaName)
				m.Down(s)
				if err = s.Schema.Run(tx, logger); err != nil {
					return
				}
				q = fmt.Sprintf("DELETE FROM `%s` WHERE `version` = ? and `namespace` = ?", opts.MigrationTable)
				_, err = tx.Exec(q, m.Version, opts.Namespace)
				return
			})
			if err != nil {
				return
			}
		}
	}
	return
}
