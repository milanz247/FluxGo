package database

import (
	"fmt"
	"sort"
	"time"

	"fluxgo/app/models"
	"fluxgo/internal/auth"
	"fluxgo/internal/session"
	"gorm.io/gorm"
)

type schemaMigration struct {
	Version   int       `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	AppliedAt time.Time `gorm:"not null"`
}

func (schemaMigration) TableName() string { return "schema_migrations" }

type Migration struct {
	Version int
	Name    string
	Up      func(*gorm.DB) error
}

func DefaultMigrations() []Migration {
	return []Migration{
		{Version: 1, Name: "create_users", Up: func(tx *gorm.DB) error {
			return ensureTable(tx, &models.User{}, []string{
				"Name", "Email", "PasswordHash", "EmailVerifiedAt", "CreatedAt", "UpdatedAt",
			})
		}},
		{Version: 2, Name: "create_auth_and_session_tables", Up: func(tx *gorm.DB) error {
			for _, model := range []any{
				&session.DatabaseRecord{},
				&models.PasswordReset{},
				&models.EmailVerification{},
				&auth.LoginAttempt{},
			} {
				if !tx.Migrator().HasTable(model) {
					if err := tx.Migrator().CreateTable(model); err != nil {
						return err
					}
				}
			}
			return nil
		}},
		{Version: 3, Name: "remove_legacy_user_phone", Up: func(tx *gorm.DB) error {
			if tx.Migrator().HasColumn(&models.User{}, "phone") {
				return tx.Migrator().DropColumn(&models.User{}, "phone")
			}
			return nil
		}},
		{Version: 4, Name: "index_sessions_by_user", Up: func(tx *gorm.DB) error {
			if !tx.Migrator().HasColumn(&session.DatabaseRecord{}, "UserID") {
				if err := tx.Migrator().AddColumn(&session.DatabaseRecord{}, "UserID"); err != nil {
					return err
				}
			}
			if !tx.Migrator().HasIndex(&session.DatabaseRecord{}, "UserID") {
				return tx.Migrator().CreateIndex(&session.DatabaseRecord{}, "UserID")
			}
			return nil
		}},
	}
}

func RunMigrations(database *gorm.DB, migrations []Migration) error {
	if !database.Migrator().HasTable(&schemaMigration{}) {
		if err := database.Migrator().CreateTable(&schemaMigration{}); err != nil {
			return fmt.Errorf("create migration table: %w", err)
		}
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	for _, migration := range migrations {
		var count int64
		if err := database.Model(&schemaMigration{}).
			Where("version = ?", migration.Version).Count(&count).Error; err != nil {
			return fmt.Errorf("check migration %d: %w", migration.Version, err)
		}
		if count > 0 {
			continue
		}
		err := database.Transaction(func(tx *gorm.DB) error {
			if err := migration.Up(tx); err != nil {
				return err
			}
			return tx.Create(&schemaMigration{
				Version: migration.Version, Name: migration.Name, AppliedAt: time.Now(),
			}).Error
		})
		if err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}
	}
	return nil
}

func ensureTable(database *gorm.DB, model any, fields []string) error {
	if !database.Migrator().HasTable(model) {
		return database.Migrator().CreateTable(model)
	}
	for _, field := range fields {
		if !database.Migrator().HasColumn(model, field) {
			if err := database.Migrator().AddColumn(model, field); err != nil {
				return err
			}
		}
	}
	if !database.Migrator().HasIndex(model, "Email") {
		return database.Migrator().CreateIndex(model, "Email")
	}
	return nil
}
