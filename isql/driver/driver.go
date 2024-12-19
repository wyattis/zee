package driver

import (
	"database/sql"
	"net/url"
	"strings"
)

//go:generate go-enum --marshal --flag

// ENUM(postgres, mysql, sqlite3)
type Type string

// ENUM(local, cloud_run)
type Environment string

type Config struct {
	Environment Environment `default:"local"`
	Driver      Type        `default:"sqlite3"`
	SslMode     string      `default:"disable"`
	Host        string      `default:"127.0.0.1"`
	Port        string      `default:"5432"`
	User        string
	Password    string
	Database    string
	SocketDir   string
}

func (c Config) String() (dsn string) {
	switch c.Driver {
	case TypePostgres:
		return c.PostgresString()
	case TypeMysql:
		return c.MysqlString()
	case TypeSqlite3:
		return c.Sqlite3String()
	}
	return
}

func (c Config) PostgresString() string {
	pairs := map[string]string{
		"host":     c.Host,
		"port":     c.Port,
		"user":     c.User,
		"password": c.Password,
		"dbname":   c.Database,
		"sslmode":  c.SslMode,
	}
	if c.Environment == EnvironmentCloudRun {
		pairs["host"] = c.SocketDir + "/" + c.Host
	}

	vals := []string{}
	for k, v := range pairs {
		if v != "" {
			if strings.Contains(v, " ") {
				v = "'" + v + "'"
			}
			vals = append(vals, k+"="+v)
		}
	}
	return strings.Join(vals, " ")
}

func (c Config) MysqlString() (dsn string) {
	dsn = c.Driver.String() + "://"
	if c.User != "" {
		dsn += c.User
		if c.Password != "" {
			dsn += ":" + c.Password
		}
		dsn += "@"
	}
	dsn += c.Host
	if c.Port != "" {
		dsn += ":" + c.Port
	}
	if c.Environment == EnvironmentCloudRun {
		dsn += c.SocketDir
	}
	dsn += "/" + c.Database

	query := url.Values{}

	if len(query) > 0 {
		dsn += "?" + query.Encode()
	}
	return
}

func (c Config) Sqlite3String() string {
	return c.Database
}

type ConnectionFactory func(config Config) (*sql.DB, error)

var Connectors = map[Type]ConnectionFactory{}

// type SchemaMutatorFactory = func() SchemaMutator

// var Mutators = map[DriverType]SchemaMutatorFactory{}
