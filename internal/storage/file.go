package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/T-Akhma/tg-bot-clean/internal/crypto"
)

type FileStore struct {
	mu    sync.RWMutex
	path  string
	key   []byte
	store map[int64]map[string]string
}

func NewFileStore(path string, key []byte) *FileStore {
	fs := &FileStore{
		path:  path,
		key:   key,
		store: make(map[int64]map[string]string),
	}
	if err := fs.Load(); err != nil {
		slog.Error("Не удалось загрузить данные FileStore", "path", path, "err", err)
	}
	return fs
}

func (f *FileStore) Load() error {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Файл хранилища не найден, будет создан новый", "path", f.path)
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return err
	}
	var decoded map[int64]map[string]string
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	f.mu.Lock()
	f.store = decoded
	f.mu.Unlock()

	return nil
}

func (f *FileStore) Set(chatID int64, key, value string) error {
	plaintext := []byte(value)

	cipherBytes, err := crypto.Encryp(plaintext, f.key)
	if err != nil {
		return fmt.Errorf("encrypt value: %w", err)
	}

	encoded := crypto.EncodeToBase64(cipherBytes)

	f.mu.Lock()

	if _, ok := f.store[chatID]; !ok {
		f.store[chatID] = make(map[string]string)
	}
	f.store[chatID][key] = encoded

	snapshot := f.cloneStoreLocked()

	f.mu.Unlock()

	if err := f.save(snapshot); err != nil {
		slog.Error("Ошибка сохранения FileStore", "err", err)
		return err
	}
	return nil
}

func (f *FileStore) Get(chatID int64, key string) (string, bool) {
	f.mu.RLock()
	inner, ok := f.store[chatID]
	f.mu.RUnlock()
	if !ok {
		return "", false
	}

	encoded, exists := inner[key]
	if !exists {
		return "", false
	}

	cipherBytes, err := crypto.DecodeFromBase64(encoded)
	if err != nil {
		slog.Error("Ошибка декодирования base64 в FileStore.Get", "err", err)
		return "", false
	}

	plaintext, err := crypto.Decryp(cipherBytes, f.key)
	if err != nil {
		slog.Error("Ошибка расшифровки в FileStore.Get", "err", err)
		return "", false
	}

	return string(plaintext), true
}

func (f *FileStore) Delete(chatID int64, key string) error {
	f.mu.Lock()

	inner, ok := f.store[chatID]
	if ok {
		delete(inner, key)
	}
	if len(inner) == 0 {
		delete(f.store, chatID)
	}
	snapshot := f.cloneStoreLocked()

	f.mu.Unlock()

	if err := f.save(snapshot); err != nil {
		slog.Error("Ошибка сохранения FileStore при Delete", "err", err)
		return err
	}
	return nil
}

func (f *FileStore) cloneStoreLocked() map[int64]map[string]string {
	result := make(map[int64]map[string]string, len(f.store))
	for chatID, inner := range f.store {
		innerCopy := make(map[string]string, len(inner))
		for k, v := range inner {
			innerCopy[k] = v
		}
		result[chatID] = innerCopy
	}
	return result
}

func (f *FileStore) save(snapshot map[int64]map[string]string) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(f.path, data, 0o600); err != nil {
		return err
	}

	return nil
}
