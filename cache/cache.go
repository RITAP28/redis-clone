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
	Type 		string			`json:"type"`
	Value 		interface{}		`json:"value"`
	ExpiryTime 	time.Time		`json:"expiryTime"`
}

type Client struct {
	Conn 	net.Conn
	InTransaction 	bool
	Transactions 	[][]interface{}
}

type RedisCache struct {
	mu 		sync.Mutex
	store 	map[string]*Entry // actual structure of a hash map
}

func NewRedisServer() *RedisCache {
	return &RedisCache{
		store: make(map[string]*Entry),
	}
}

func NewClient(Conn net.Conn) *Client {
	return &Client{
		Conn: Conn,
		InTransaction: false,
		Transactions: make([][]interface{}, 0),
	}
}

func(r *RedisCache) StartExpiryCleaner(interval time.Duration) {
	// this function runs continuously in the background to delete the expired keys
	// go func() { ... }() --> runs the cleaner function asynchronously, in a background goroutine
	go func() {
		for {
			time.Sleep(interval)
				
			r.mu.Lock()
			for key, entry := range r.store {
				if !entry.ExpiryTime.IsZero() && time.Now().After(entry.ExpiryTime) {
					delete(r.store, key)
				}
			}
			r.mu.Unlock()
		}
	}()
}

func(r *RedisCache) SET(key string, value interface{}, ttl int) (string, bool) {
	// Atomic SET with expiration: Only the SET command has built-in options to set expiration atomically!
	// command example: SET hello "world" EX 10 or SET hello "world" PX 10000 (both expire in 10 seconds)
	// the above ATOMIC EXPIRATION needs to be implemented!

	r.mu.Lock();
	defer r.mu.Unlock();

	_, exists := r.store[key];
	if exists {
		return "Key with a value already exists", false;
	};

	// creating a new Entryy struct in the map and initialises it's value with value field
	// also providing a expiry time as ttl to the expiryTime field of the Entryy struct
	entry := &Entry{Type: "string", Value: value}

	if ttl > 0 {
		entry.ExpiryTime = time.Now().Add(time.Duration(ttl) * time.Millisecond);
		fmt.Printf("expiry time is %v/n", entry.ExpiryTime.Format("January 2, 2006 at 3:04PM"))
	};

	r.store[key] = entry;
	// fmt.Println("Value successfully set to the given key");
	fmt.Println("Value successfully set to the given key");
	return "Value successfully set to the given key", true;
}

func(r *RedisCache) GET(key string) (interface{}, bool) {
	r.mu.Lock();
	defer r.mu.Unlock();

	entry, exists := r.store[key];
	if !exists {
		// Key does not exist, signalling this as 'false' boolean
		return nil, false
	}

	if !entry.ExpiryTime.IsZero() && time.Now().After(entry.ExpiryTime) {
		// Key has expired, signalling this as 'false' boolean
		delete(r.store, key);
		return nil, false;
	}
	
	fmt.Println("Value successfully obtained for the given key");
	// Key exists and is valid, signalling this as 'true' boolean
	return entry.Value, true
}

func(r *RedisCache) DELETE(key string) bool {
	r.mu.Lock();
	defer r.mu.Unlock();

	_, ok := r.store[key]
	if (!ok) {
		fmt.Println("Key does not exist")
		return false
	}

	delete(r.store, key);
	fmt.Println("Deleted the given key successfully", key);
	return true;
}

func (r *RedisCache) HandleConnection (client *Client) {
	defer client.Conn.Close()

	reader := bufio.NewReader(client.Conn);
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
			client.Conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", err.Error())));
			return
		};

		// commandArray = ["SET", "hello", "world"]
		commandArray, ok := parsedCmd.([]any);
		if !ok {
			fmt.Printf("-ERR command is not a bulk string array\r\n");
			continue
		};

		if len(commandArray) == 0 {
			client.Conn.Write([]byte("-ERR empty command\r\n"))
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

		// validating the command and the arguments
		fmt.Printf("the command is %v\r\n", commandStr)
		fmt.Printf("the arguments are %v\r\n", args)

		// commands for implementing transaction in redis --> MULTI & EXEC
		switch command {
		case "MULTI":
			// length of arguments shall be 0
			if len(args) > 0 {
				client.Conn.Write([]byte("-ERR wrong number of arguments for 'MULTI' command\r\n"))
				continue
			}

			if client.InTransaction {
				client.Conn.Write([]byte("-ERR MULTI calls cannot be nested\r\n"))
				continue
			}

			client.InTransaction = true
			client.Transactions = make([][]interface{}, 0)
			client.Conn.Write([]byte("+OK\r\n"))

		case "EXEC":
			if len(args) > 0 {
				client.Conn.Write([]byte("-ERR wrong number of arguments for 'EXEC' command\r\n"))
				continue
			}

			if !client.InTransaction {
				client.Conn.Write([]byte("-ERR EXEC without MULTI\r\n"))
				continue
			}

			client.InTransaction = false
			results := []string{}

			for _, v := range client.Transactions {
				result := r.ExecuteCommands(client, v)
				results = append(results, result)
			}

			client.Transactions = nil
			fmt.Fprintf(client.Conn, "*%d\r\n", len(results))
			for _, res := range results {
				client.Conn.Write([]byte(res))
			}

		default:
			// in case the client is in InTransaction mode, then all the commands instead of executing normally will go through this code block
			// they will be queued and when EXEC is run, the queued commands will be executed sequentially
			if client.InTransaction && command != "EXEC" && command != "MULTI" {
				ok := r.queueCommands(client, commandArray)
				if !ok {
					client.Conn.Write([]byte("-ERR error while queueing commands\r\n"))
					continue
				}

				client.Conn.Write([]byte("+Queued\r\n"))
				continue
			} else {
				// for all commands when the client is not in transaction mode
				result := r.ExecuteCommands(client, commandArray)
				client.Conn.Write([]byte(result))
			}
		}
	}
}

