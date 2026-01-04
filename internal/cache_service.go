package internal

import (
	"errors"
	"time"
)

var (
	ErrKeyEmpty = errors.New("key cannot be empty")
	ErrKeyNotFound = errors.New("key not found")
)

type CacheService struct {
	storage    *Storage
	defaultTTL time.Duration
}

func NewCacheService(storage *Storage, defaultTTL time.Duration) *CacheService {
	return &CacheService{
		storage: storage,
		defaultTTL: defaultTTL,
	}
}

func (s *CacheService) Set(key, value string, ttl time.Duration) error {
	if key == ""{
		return ErrKeyEmpty
	}

	if ttl == 0 {
		ttl = s.defaultTTL
	}

	s.storage.Set(key, value, ttl)
	return nil
}

func (s *CacheService) Get(key string) (string, bool, error) {
	if key == ""{
		return "", false, ErrKeyEmpty
	}

	val, exists := s.storage.Get(key)
	return val, exists, nil
}

func (s *CacheService) Delete(key string) (bool, error) {
	if key == "" {
		return false, ErrKeyEmpty
	}

	deleted := s.storage.Delete(key)
	return deleted, nil
}

func (s *CacheService) Exists(key string) (bool, error) {
	if key == "" {
		return false, ErrKeyEmpty
	}

	exists := s.storage.Exists(key)
	return exists, nil
}

func (s *CacheService) Expire(key string, ttl time.Duration) error {
	if key == ""{
		return ErrKeyEmpty
	}

	success := s.storage.SetTTL(key, ttl)
	if !success {
		return ErrKeyNotFound
	}
	return nil
}

func (s *CacheService) TTL(key string) (time.Duration, error) {
	if key == "" {
		return 0, ErrKeyEmpty
	}

	ttl, exists := s.storage.GetTTL(key)
	if !exists {
		return -2 * time.Second, nil
	}
	return ttl, nil
}

func (s *CacheService) Keys(pattern string) ([]string, error) {
	if pattern == "" {
		pattern = "*"
	}

	keys := s.storage.Keys(pattern)
	return keys, nil
}