package session

import (
	"sync"
	"time"
)

type memoryEntry struct {
	values    map[string]any
	expiresAt time.Time
}

// MemoryStore is a concurrent in-memory session store.
// Its sessions are lost when the application restarts.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]memoryEntry
}

// NewMemoryStore creates an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]memoryEntry)}
}

func (store *MemoryStore) Load(id string, now time.Time) (map[string]any, bool, error) {
	store.mu.RLock()
	entry, exists := store.sessions[id]
	store.mu.RUnlock()
	if !exists {
		return nil, false, nil
	}
	if !entry.expiresAt.After(now) {
		_ = store.Delete(id)
		return nil, false, nil
	}
	return clone(entry.values), true, nil
}

func (store *MemoryStore) Save(id string, values map[string]any, expiresAt time.Time) error {
	store.mu.Lock()
	store.sessions[id] = memoryEntry{
		values:    clone(values),
		expiresAt: expiresAt,
	}
	store.mu.Unlock()
	return nil
}

func (store *MemoryStore) Delete(id string) error {
	store.mu.Lock()
	delete(store.sessions, id)
	store.mu.Unlock()
	return nil
}

func clone(values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
