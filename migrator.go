package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/wyattis/zee/isql"
	"github.com/wyattis/zee/schema"
)

func NopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func NewMigrator(db isql.IDB, opts MigrateOptions) *Migrator {
	return &Migrator{
		migrations: []Migration{},
		db:         db,
		logger:     NopLogger(),
		opts:       opts,
	}
}

type Migrator struct {
	migrations []Migration
	db         isql.IDB
	logger     *slog.Logger
	opts       MigrateOptions
}

func (mi *Migrator) WithLogger(logger *slog.Logger) *Migrator {
	mi.logger = logger
	return mi
}

func (mi *Migrator) Add(migrations ...Migration) {
	mi.migrations = append(mi.migrations, migrations...)
}

// UpTo migrates the database up to the given version. It will not run migrations that have already been run and will
// fail if the target version is lower than the current schema version. In other words, it will not roll back.
func (mi *Migrator) UpTo(version uint) (err error) {
	schemaVersion, err := validateMigration(mi.migrations, mi.db, mi.opts, version)
	if err != nil {
		return
	}
	if schemaVersion > version {
		err = ErrSchemaVersionHigherThanTarget
		return
	}
	for _, m := range mi.migrations {
		if m.Version > schemaVersion && m.Version <= version {
			err = isql.Begin(mi.db, func(tx *sql.Tx) (err error) {
				s := schema.New(mi.opts.Driver, mi.opts.SchemaName)
				m.Up(s)
				hash, err := s.Schema.Hash()
				if err != nil {
					return
				}
				// mark current migration as dirty before we start
				q := fmt.Sprintf("INSERT INTO `%s` (`namespace`, `version`, `hash`, `dirty`) VALUES (?, ?, ?, ?)", mi.opts.MigrationTable)
				_, err = tx.Exec(q, mi.opts.Namespace, m.Version, hash, true)
				if err != nil {
					return
				}

				if err = s.Schema.Run(tx, mi.logger); err != nil {
					return
				}
				q = fmt.Sprintf("UPDATE `%s` SET `dirty` = ?, `finished_at` = %s WHERE `version` = ? and `namespace` = ?", mi.opts.MigrationTable, schema.NOW{}.Constant(mi.opts.Driver))
				_, err = tx.Exec(q, false, m.Version, mi.opts.Namespace)
				return
			})
			if err != nil {
				return
			}
		}
	}
	mi.logger.Info("Database is up to date with version", "version", version)
	return
}

// DownTo migrates the database down to the given version. It will fail if the target version is higher than the current
// schema version.
func (mi *Migrator) DownTo(version uint) (err error) {
	err = errors.New("not implemented")
	return
}

// To migrates the database to the given version. It will run migrations either up or down depending on the relationship
// between the current schema version and the target version.
func (mi *Migrator) To(version uint) (err error) {
	err = errors.New("not implemented")
	return
}
