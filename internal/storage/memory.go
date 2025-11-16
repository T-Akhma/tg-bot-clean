package storage

import "sync"

type MemoryStore struct {
	mu    sync.RWMutex
	store map[int64]map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		store: make(map[int64]map[string]string),
	}
}

func (m *MemoryStore) Set(chatID int64, key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.store[chatID]; !ok {
		m.store[chatID] = make(map[string]string)
	}

	m.store[chatID][key] = value
}

func (m *MemoryStore) Get(chatID int64, key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if userData, ok := m.store[chatID]; ok {
		val, exists := userData[key]
		return val, exists
	}
	return "", false
}

func (m *MemoryStore) Delete(chatID int64, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if userData, ok := m.store[chatID]; ok {
		delete(userData, key)
	}
}
