package cache

func(r *RedisCache) HSET(key string, fieldValues map[string]string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		// creating a hash table and inserting field-value pairs into the hashtable
		hashTable := make(map[string]string)
		for field, value := range fieldValues {
			hashTable[field] = value
		}
		r.store[key] = &Entry{Type: "hash", Value: hashTable}
		return len(hashTable), true
	}

	hashTable, ok := entry.Value.(map[string]string)
	if !ok {
		return -1, false // WRONGTYPE error in real Redis
	}

	// updating existing hashtable and counting new fields
	newCount := 0
	for field, value := range fieldValues {
		if _, exists := hashTable[field]; !exists {
			newCount++
		}

		hashTable[field] = value
	}

	return newCount, true
}

func(r *RedisCache) HGET(key string, field string) (string, bool) {
	// command syntax: HGET key field
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return "", false
	}

	hash := entry.Value.(map[string]string)
	return hash[field], true
}

func(r *RedisCache) HGETALL(key string) (map[string]string, bool) {
	// command syntax: HGETALL key
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return nil, false
	}

	hash, ok := entry.Value.(map[string]string)
	if !ok {
		return nil, false
	}

	return hash, true
}

func(r *RedisCache) HDEL(key string, fields []string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return -1, false
	}

	// getting the hashtable
	hashTable := entry.Value.(map[string]string)

	// getting the length of the hashtable
	// then checking if there are no elements inside hashtable, then the whole hashtable shall be deleted
	length := len(hashTable)
	if length == 0 {}

	deleteCount := 0
	for _, v := range fields {
		// if a key does not exist, then it is simply ignored
		if _, exists := hashTable[v]; !exists {
			continue
		}

		delete(hashTable, v)
		deleteCount++
	}

	return deleteCount, true
}

func(r *RedisCache) HLEN(key string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return 0, false
	}

	hashTable := entry.Value.(map[string]string)

	return len(hashTable), true
}