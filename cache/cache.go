// package cache
package cache

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"redis-clone/parser"
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
		// let the command coming from the client to the redis server
		// pass through the resp parser


		// for the example command: SET hello world from client
		// parsedCmd: []interface{"SET", "hello", "world"}
		parsedCmd, err := parser.HandleRESP(reader);
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Printf("Client disconnected");
				return;
			};

			fmt.Println("Error reading command: ", err);
			conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", err.Error())));
			return
		};

		// commandArray = ["SET", "hello", "world"]
		commandArray, ok := parsedCmd.([]any);
		if !ok {
			fmt.Printf("-ERR command is not a bulk string array\r\n");
			continue
		};

		if len(commandArray) == 0 {
			continue
		}

		// commandStr = "SET"
		commandStr, ok := commandArray[0].(string);
		if !ok {
			fmt.Printf("-ERR command name is not a string\r\n");
			continue
		};

		command := strings.ToUpper(commandStr);
		// args = ["hello", "world"]
		args := commandArray[1:];

		switch command {
		case "SET":
			if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				continue;
			}

			key, keyOk := args[0].(string);
			value, valueOk := args[1].(string);

			if !valueOk || !keyOk {
				conn.Write([]byte("-ERR arguments must be string\r\n"));
				continue;
			};

			// ttl -> time to expiry for the key is set to 1 second
			r.SET(key, value, 1000);
			conn.Write([]byte("+OK\r\n"))

		case "GET":
			if len(args) != 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue;
			}

			key, keyOk := args[1].(string);
			if !keyOk {
				conn.Write([]byte("-ERR argument must be string\r\n"));
				continue
			};

			value, ok := r.GET(key);
			if !ok {
				conn.Write([]byte("$1\r\n"));
				continue;
			} else {
				response := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value); // RESP representation of string
				conn.Write([]byte(response))
			}

		case "DELETE":
			if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'delete' command\r\n"));
				continue;
			}

			key, keyOk := args[1].(string);
			if !keyOk {
				conn.Write([]byte("-ERR argument must be string\r\n"));
				continue
			};

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

