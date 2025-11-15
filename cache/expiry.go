package cache

import "time"

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
	entry.expiryTime = time.Now().Add(time.Duration(seconds) * time.Second)
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
	entry.expiryTime = time.Now().Add(time.Duration(ms) * time.Millisecond)
	return 1, true
}