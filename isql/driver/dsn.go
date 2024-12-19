package driver

import (
	"net/url"
	"strings"
)

func ConfigFromConnectionString(src string) (config Config, err error) {
	config.Environment = EnvironmentLocal
	// Parse the DSN
	u, err := url.Parse(src)
	if err != nil {
		return
	}
	// Set the driver
	config.Driver, err = ParseType(u.Scheme)
	if err != nil {
		return
	}
	// Set the host
	config.Host = u.Hostname()
	// Set the port
	config.Port = u.Port()
	// Set the user
	if u.User != nil {
		config.User = u.User.Username()
		config.Password, _ = u.User.Password()
	}
	// Set the database
	config.Database = strings.TrimLeft(u.Path, "/")
	// Set the query
	query := u.Query()
	config.SslMode = query.Get("sslmode")
	// Set the environment
	if strings.Contains(config.Host, "/") {
		config.Environment = EnvironmentCloudRun
		config.SocketDir = strings.TrimRight(config.Host, config.Host)
		config.Host = strings.TrimLeft(config.Host, config.SocketDir)
	}
	return
}
