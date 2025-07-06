package main

import (
	"fmt"
	"sync"
	"time"
)

// covalent to interfaces/types in typescript
// explains the structure of value corresponding to any key in the hash map
type Entry struct {
	value string
	expiryTime time.Time
}

type RedisCache struct {
	mu sync.Mutex
	store map[string]*Entry // actual structure of a hash map
}

func newRedisServer() *RedisCache {
	return &RedisCache{
		store: make(map[string]*Entry),
	}
}

func(r *RedisCache) SET(key, value string, ttl int) (string, bool) {
	r.mu.Lock();
	defer r.mu.Unlock();

	_, exists := r.store[key];
	if exists {
		return "Key with a value already exists", false;
	};

	// creating a new Entryy struct in the map and initialises it's value with value field
	// also providing a expiry time as ttl to the expiryTime field of the Entryy struct
	entry := &Entry{value: value}
	if ttl > 0 {
		entry.expiryTime = time.Now().Add(time.Duration(ttl) * time.Second);
	};

	r.store[key] = entry;
	// fmt.Println("Value successfully set to the given key");
	fmt.Println("Value successfully set to the given key");
	return "Value successfully set to the given key", true;
}

func(r *RedisCache) GET(key string) (string, bool) {
	r.mu.Lock();
	defer r.mu.Unlock();

	entry, exists := r.store[key];
	if !exists {
		return "No key-value pair exists for the given key", false
	}

	if !entry.expiryTime.IsZero() && time.Now().After(entry.expiryTime) {
		delete(r.store, key);
		return "", false;
	}
	
	fmt.Println("Value successfully obtained for the given key");
	return entry.value, true
}

func(r *RedisCache) DELETE(key string) bool {
	r.mu.Lock();
	defer r.mu.Unlock();

	delete(r.store, key);
	fmt.Println("Deleted the given key successfully", key);
	return true;
}

