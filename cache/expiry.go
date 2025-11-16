package cache

import "time"

// commands needed to be implemented:
// EXPIRE 			--> Done
// PEXPIRE			--> Done
// TTL				--> Done
// PTTL				--> Done
// EXPIRETIME		--> To be implemented
// PEXPIRETIME		--> To be implemented
// PERSIST			--> To be implemented

func(r *RedisCache) EXPIRE(key string, seconds int) (int, bool) {
	// this command sets expiry time in seconds
	r.mu.Lock()
	defer r.mu.Unlock()

	// checking whether the key exists or not
	// if the key does not exist, then 0 is returned
	entry, exists := r.store[key]
	if !exists {
		return 0, false
	}

	// key gets deleted for 0 or negative ttl
	if seconds <= 0 {
		delete(r.store, key)
		return 1, true
	}

	// setting the expiry time for the key in seconds
	entry.ExpiryTime = time.Now().Add(time.Duration(seconds) * time.Second)
	return 1, true
}

func(r *RedisCache) PEXPIRE(key string, ms int) (int, bool) {
	// this command sets expiry time in milliseconds
	r.mu.Lock()
	defer r.mu.Unlock()

	// checking whether the key exists or not
	// if the key does not exist, then 0 is returned
	entry, exists := r.store[key]
	if !exists {
		return 0, false
	}

	// key gets deleted for 0 or negative ttl
	if ms <= 0 {
		delete(r.store, key)
		return 1, true
	}

	// setting the expiry time for the key in milliseconds
	entry.ExpiryTime = time.Now().Add(time.Duration(ms) * time.Millisecond)
	return 1, true
}

func(r *RedisCache) TTL(key string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key];
	if !exists {
		return -2, true // key does not exist
	}

	// checking whether the key has expiration set or not
	if entry.ExpiryTime.IsZero() {
		// key has no expiration set
		return -1, true
	}

	remaining := time.Until(entry.ExpiryTime)
	if remaining <= 0 {
		return -2, true // key has expired but not yet cleaned up
	}

	return int(remaining.Seconds()), true
}

func(r *RedisCache) PTTL(key string) (int, bool) {
	// command syntax: PTTL key --> returns time left for key expiration in milliseconds
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return -2, true
	}

	if entry.ExpiryTime.IsZero() {
		return -1, true
	}

	remaining := time.Until(entry.ExpiryTime)
	if remaining <= 0 {
		return -2, true
	}

	return int(remaining.Milliseconds()), true
}

func(r *RedisCache) PERSIST(key string) int {
	// command syntax: PERSIST key
	// returns 1 --> the expiration was successfully removed
	// returns 0 --> the key did not have an expiration time set/the key did not exist in the database

	r.mu.Lock()
	defer r.mu.Unlock()

	// checking the existence of key in store
	entry, exists := r.store[key]
	if !exists {
		return 0
	}

	// no expiration set to begin with, so 0 is returned
	if entry.ExpiryTime.IsZero() {
		return 0
	}

	// removing expiry for the key
	entry.ExpiryTime = time.Time{}
	return 1
}