package main

import (
	"fmt"
	"sync"
	"time"
)

type Entry struct {
	value      string
	expiryTime time.Time
}

type RedisCache struct {
	mu    sync.Mutex
	store map[string]*Entry
}

func NewRedisCache() *RedisCache {
	return &RedisCache{
		store: make(map[string]*Entry),
	}
}

func(r *RedisCache) Set(key, value string, ttl int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry := &Entry{value: value}
	if ttl > 0 {
		entry.expiryTime = time.Now().Add(time.Duration(ttl) * time.Second)
	}

	r.store[key] = entry
	fmt.Println("SET", key, "-->", value)
}

func(r *RedisCache) Get(key string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return "", false
	}

	if !entry.expiryTime.IsZero() && time.Now().After(entry.expiryTime) {
		delete(r.store, key)
		return "", false
	}

	return entry.value, true
}

func(r *RedisCache) Delete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.store, key)
	fmt.Println("DELETE", key)
}