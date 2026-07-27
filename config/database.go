package config

import (
	"net"
	"time"

	MySQLDriver "github.com/go-sql-driver/mysql"
)

// DatabaseConfig contains MySQL and connection-pool settings.
type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	Charset         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	AutoMigrate     bool
}

// DSN returns a MySQL driver data source name.
func (config DatabaseConfig) DSN() string {
	return (&MySQLDriver.Config{
		User:      config.User,
		Passwd:    config.Password,
		Net:       "tcp",
		Addr:      net.JoinHostPort(config.Host, config.Port),
		DBName:    config.Name,
		ParseTime: true,
		Loc:       time.Local,
		Params: map[string]string{
			"charset": config.Charset,
		},
	}).FormatDSN()
}
