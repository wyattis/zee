package driver

import "testing"

type c struct {
	s string
	c Config
}

var cases = []c{{
	s: "postgres://user:xxxxxxxx@xxxxxxxxx:5432/database?sslmode=disable",
	c: Config{
		Environment: EnvironmentLocal,
		Driver:      TypePostgres,
		SslMode:     "disable",
		Host:        "xxxxxxxxx",
		Port:        "5432",
		User:        "user",
		Password:    "xxxxxxxx",
		Database:    "database",
	},
}, {
	s: "postgres://user:xxxxxxxx@xxxxxxxxxxxxxxxxxxxxxxxxx",
	c: Config{
		Environment: EnvironmentLocal,
		Driver:      TypePostgres,
		Host:        "xxxxxxxxxxxxxxxxxxxxxxxxx",
		User:        "user",
		Password:    "xxxxxxxx",
	},
}, {
	s: "postgres://xxxxxxxxxx?sslmode=disable",
	c: Config{
		Environment: EnvironmentLocal,
		Driver:      TypePostgres,
		SslMode:     "disable",
		Host:        "xxxxxxxxxx",
	},
}}

func TestConfigFromConnectionString(t *testing.T) {
	for _, c := range cases {
		config, err := ConfigFromConnectionString(c.s)
		if err != nil {
			t.Error(err)
		}
		if config != c.c {
			t.Errorf("expected %v, got %v", c.c, config)
		}
	}
}
