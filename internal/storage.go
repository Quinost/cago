package internal

import (
	"sync"
	"time"
)

type Storage struct {
	mu   sync.RWMutex
	data map[string]StorageItem
}

type StorageItem struct {
	Value     string
	ExpiresAt time.Time
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]StorageItem),
	}
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return "", false
	}

	if checkIfExpired(&item.ExpiresAt, utcNow()) {
		return "", false
	}
	return item.Value, exists
}

func (s *Storage) Set(key, val string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = utcNow().Add(ttl)
	}

	s.data[key] = StorageItem{
		Value:     val,
		ExpiresAt: expiresAt,
	}
}

func (s *Storage) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.data[key]
	delete(s.data, key)
	return exists
}

func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	return !checkIfExpired(&item.ExpiresAt, utcNow())
}

func (s *Storage) GetTTL(key string) (time.Duration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := utcNow()

	item, exists := s.data[key]
	if !exists {
		return 0, false
	}
	if item.ExpiresAt.IsZero() {
		return -1, false
	}
	if checkIfExpired(&item.ExpiresAt, now) {
		return 0, false
	}
	ttl := item.ExpiresAt.Sub(*now)
	return ttl, true
}

func (s *Storage) SetTTL(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	now := utcNow()
	if checkIfExpired(&item.ExpiresAt, now) {
		return false
	}

	if ttl > 0 {
		item.ExpiresAt = now.Add(ttl)
	} else {
		item.ExpiresAt = time.Time{}
	}

	s.data[key] = item
	return true
}

func (s *Storage) Keys(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0)
	now := utcNow()

	for key, item := range s.data {
		if checkIfExpired(&item.ExpiresAt, now){
			continue
		}
		panic("not implemented")
		keys = append(keys, key)
	}

	return keys
}

func (s *Storage) CleanupExired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	now := utcNow()
	for key, item := range s.data {
		if checkIfExpired(&item.ExpiresAt, now) {
			delete(s.data, key)
			count++
		}
	}

	return count
}

func checkIfExpired(expiresAt *time.Time, now *time.Time) bool {
	return !expiresAt.IsZero() && expiresAt.Before(*now)
}

func utcNow() *time.Time {
	now := time.Now().UTC()
	return &now
}
