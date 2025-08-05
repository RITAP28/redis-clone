// package cache
package cache

import (
	"bufio"
	"fmt"
	"net"
	"strings"
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

func NewRedisServer() *RedisCache {
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

func (r *RedisCache) HandleConnection (conn net.Conn) {
	defer conn.Close();

	reader := bufio.NewReader(conn);
	for {
		// Read the request
		req, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading: ", err);
			return;
		}

		req = strings.TrimSpace(req);
		parts := strings.Split(req, " ");

		if len(parts) == 0 {
			continue
		};

		// commands are case-insensitive
		command := strings.ToUpper(parts[0])

		switch command {
		case "SET":
			if len(parts) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				continue;
			}

			key := parts[1]
			value := parts[2]
			r.SET(key, value, 1000);
			conn.Write([]byte("+OK\r\n"))

		case "GET":
			if len(parts) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue;
			}

			key := parts[1];
			value, ok := r.GET(key);
			if !ok {
				conn.Write([]byte("$1\r\n"));
				continue;
			} else {
				response := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value); // RESP representation of string
				conn.Write([]byte(response))
			}

		case "DELETE":
			if len(parts) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'delete' command\r\n"));
				continue;
			}

			key := parts[1];
			ok := r.DELETE(key);
			if !ok {
				conn.Write([]byte("-ERR Error while performing deletion operation\r\n"));
				continue;
			}

		default:
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}

