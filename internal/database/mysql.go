// Package database creates application database connections.
package database

import (
	"fmt"

	"fluxgo/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ConnectMySQL opens and verifies a pooled GORM MySQL connection.
func ConnectMySQL(config config.DatabaseConfig) (*gorm.DB, error) {
	connection, err := gorm.Open(mysql.Open(config.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open MySQL connection: %w", err)
	}

	sqlDB, err := connection.DB()
	if err != nil {
		return nil, fmt.Errorf("access MySQL connection pool: %w", err)
	}
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping MySQL: %w", err)
	}
	return connection, nil
}

// Migrate creates or updates tables for the supplied models.
func Migrate(connection *gorm.DB, models ...any) error {
	if err := connection.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate database: %w", err)
	}
	return nil
}
