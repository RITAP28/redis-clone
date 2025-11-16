package cache

import "fmt"

// complexity in redis sets -> there are two sets: intsets and


func(r *RedisCache) SADD(key string, members []string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		// new set is created
		newSet := make(map[string]struct{})
		added := 0
		for i:=0; i<len(members); i++ {
			if _, ok := newSet[members[i]]; !ok {
				newSet[members[i]] = struct{}{}
				added++
			}
		}
		r.store[key] = &Entry{Type: "set", Value: newSet}
		return added, true
	}

	// see if any one member already exists inside the set
	// if yes, that member is not added into the set again, simply ignored
	set, isSet := entry.Value.(map[string]struct{})
	if !isSet {
		return 0, false
	}

	added := 0
	for i:=0; i<len(members); i++ {
		if _, ok := set[members[i]]; !ok {
			set[members[i]] = struct{}{}
			added++
		}
	}

	entry.Value = set
	r.store[key] = entry

	return added, true
}

func(r *RedisCache) SISMEMBER(key string, member string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return 0, false
	}

	set, isSet := entry.Value.(map[string]struct{})
	if !isSet {
		return 0, false
	}

	// structure of set:
	// set = map[string]struct{}{
	// 		"member1": {},
	// 		"member2": {}
	// }

	_, ok := set[member]
	if !ok {
		return 0, true // member does not exist
	}

	// member exists
	return 1, true
}

func(r *RedisCache) SREM(key string, members []string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return 0, false
	}

	set, isSet := entry.Value.(map[string]struct{})
	if !isSet {
		return 0, false
	}

	for _, v := range members {
		_, ok := set[v]
		if !ok {
			continue
		}

		delete(set, v)
	}

	return 1, true
}

func(r *RedisCache) SCARD(key string) (int, bool) {
	// this command returns the cardinality or number of elements in the set
	// returns the number of elements & returns 1 if the key does not exist
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return 0, false
	}

	set, isSet := entry.Value.(map[string]struct{})
	if !isSet {
		return 0, false
	}

	length := len(set)
	return length, true
}

func(r *RedisCache) SMEMBERS(key string) ([]string, bool) {
	// this command returns all members present in the set
	// return type: Array of strings
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return nil, false
	}

	set, isSet := entry.Value.(map[string]struct{})
	if !isSet {
		return nil, false
	}

	members := []string{}

	// looping with for...range through the set to get all members inside it
	for member := range set {
		members = append(members, member)
	}

	fmt.Print("members present: ", members)
	return members, true
}







