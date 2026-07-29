package session

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DatabaseRecord is the persisted representation of a session.
type DatabaseRecord struct {
	ID        string    `gorm:"primaryKey;size:64"`
	Payload   []byte    `gorm:"type:longblob;not null"`
	UserID    *uint     `gorm:"index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	UpdatedAt time.Time
}

func (DatabaseRecord) TableName() string { return "sessions" }

// DatabaseStore persists sessions through GORM.
type DatabaseStore struct{ database *gorm.DB }

func NewDatabaseStore(database *gorm.DB) *DatabaseStore {
	return &DatabaseStore{database: database}
}

func (store *DatabaseStore) Load(id string, now time.Time) (map[string]any, bool, error) {
	var record DatabaseRecord
	err := store.database.First(&record, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("load session: %w", err)
	}
	if !record.ExpiresAt.After(now) {
		_ = store.Delete(id)
		return nil, false, nil
	}
	values := make(map[string]any)
	if err := gob.NewDecoder(bytes.NewReader(record.Payload)).Decode(&values); err != nil {
		return nil, false, fmt.Errorf("decode session: %w", err)
	}
	return values, true, nil
}

func (store *DatabaseStore) Save(id string, values map[string]any, expiresAt time.Time) error {
	var payload bytes.Buffer
	if err := gob.NewEncoder(&payload).Encode(values); err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	record := DatabaseRecord{
		ID: id, Payload: payload.Bytes(), UserID: sessionUserID(values), ExpiresAt: expiresAt,
	}
	err := store.database.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"payload", "user_id", "expires_at", "updated_at"}),
	}).Create(&record).Error
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func sessionUserID(values map[string]any) *uint {
	value, exists := values["user_id"]
	if !exists {
		return nil
	}
	switch id := value.(type) {
	case uint:
		return &id
	case uint64:
		converted := uint(id)
		return &converted
	case int:
		if id > 0 {
			converted := uint(id)
			return &converted
		}
	}
	return nil
}

func (store *DatabaseStore) Delete(id string) error {
	if err := store.database.Delete(&DatabaseRecord{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (store *DatabaseStore) DeleteExpired(now time.Time) error {
	return store.database.Where("expires_at <= ?", now).Delete(&DatabaseRecord{}).Error
}
