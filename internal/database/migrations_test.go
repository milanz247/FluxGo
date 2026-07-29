package database_test

import (
	"testing"

	"fluxgo/internal/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDefaultMigrationsAreVersionedAndIdempotent(t *testing.T) {
	connection, err := gorm.Open(sqlite.Open("file:migrations?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.RunMigrations(connection, database.DefaultMigrations()); err != nil {
		t.Fatalf("first migration run: %v", err)
	}
	if err := database.RunMigrations(connection, database.DefaultMigrations()); err != nil {
		t.Fatalf("second migration run: %v", err)
	}

	for _, table := range []string{
		"users",
		"sessions",
		"password_resets",
		"email_verifications",
		"login_attempts",
		"schema_migrations",
	} {
		if !connection.Migrator().HasTable(table) {
			t.Fatalf("expected migrated table %q", table)
		}
	}

	var count int64
	if err := connection.Table("schema_migrations").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != int64(len(database.DefaultMigrations())) {
		t.Fatalf("expected %d migration records, got %d", len(database.DefaultMigrations()), count)
	}
}
