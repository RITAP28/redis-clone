// package cache
package cache

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"redis-clone/parser"
)

// covalent to interfaces/types in typescript
// explains the structure of value corresponding to any key in the hash map
type Entry struct {
	value interface{}
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

func(r *RedisCache) StartExpiryCleaner(interval time.Duration) {
	// this function runs continuously in the background to delete the expired keys
	// go func() { ... }() --> runs the cleaner function asynchronously, in a background goroutine
	go func() {
		for {
			time.Sleep(interval)
				
			r.mu.Lock()
			for key, entry := range r.store {
				if !entry.expiryTime.IsZero() && time.Now().After(entry.expiryTime) {
					delete(r.store, key)
				}
			}
			r.mu.Unlock()
		}
	}()
}

func(r *RedisCache) SET(key string, value interface{}, ttl int) (string, bool) {
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
		entry.expiryTime = time.Now().Add(time.Duration(ttl) * time.Millisecond);
		fmt.Printf("expiry time is %v/n", entry.expiryTime.Format("January 2, 2006 at 3:04PM"))
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

	if !entry.expiryTime.IsZero() && time.Now().After(entry.expiryTime) {
		// Key has expired, signalling this as 'false' boolean
		delete(r.store, key);
		return nil, false;
	}
	
	fmt.Println("Value successfully obtained for the given key");
	// Key exists and is valid, signalling this as 'true' boolean
	return entry.value, true
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

		// validating the command and the arguments
		fmt.Printf("the command is %v\r\n", commandStr)
		fmt.Printf("the arguments are %v\r\n", args)

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

			// ttl -> time to expiry for the key is set to 10 seconds
			r.SET(key, value, 10000);
			conn.Write([]byte("+OK\r\n"))

		case "GET":
			if len(args) != 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue;
			}

			key, keyOk := args[0].(string);
			if !keyOk {
				conn.Write([]byte("-ERR argument must be string\r\n"));
				continue
			};

			value, ok := r.GET(key);
			if !ok {
				conn.Write([]byte("$-1\r\n"));
				continue;
			}

			stringValue, isString := value.(string);
			if !isString {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			response := fmt.Sprintf("$%d\r\n%s\r\n", len(stringValue), value); // RESP representation of string
			conn.Write([]byte(response))

		case "DELETE":
			if len(args) != 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'delete' command\r\n"));
				continue;
			}

			key, keyOk := args[0].(string);
			if !keyOk {
				conn.Write([]byte("-ERR argument must be string\r\n"));
				continue
			};

			ok := r.DELETE(key);
			if !ok {
				conn.Write([]byte("-ERR Error while performing deletion operation\r\n"));
				continue;
			}

			conn.Write([]byte("+OK\r\n"))

		case "LPUSH":
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'LPUSH' command\r\n"))
				continue
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			// remaining arguments in the commandArray are values corresponding to the key
			values := []string{}
			for _, arg := range args[1:] {
				val, ok := arg.(string)
				if !ok {
					conn.Write([]byte("-ERR all values must be strings\r\n"))
					continue
				}

				values = append(values, val)
			}

			// calling the LPUSH implementation
			listLength, ok := r.LPUSH(key, values...)
			if !ok {
				fmt.Println("something went wrong while processing lists")
				continue
			}

			conn.Write([]byte(fmt.Sprintf(":%d\r\n", listLength)))

		case "RPUSH":
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'RPUSH' command\r\n"))
				continue
			}

			key, ok := args[0].(string)
			if !ok {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			values := []string{}
			for _, arg := range args[1:] {
				val, ok := arg.(string)
				if !ok {
					conn.Write([]byte("-ERR all values must be strings\r\n"))
					continue
				}

				values = append(values, val)
			}

			listLength, ok := r.RPUSH(key, values)
			if !ok {
				conn.Write([]byte("-ERR values could not be pushed\r\n"))
				continue
			}

			conn.Write([]byte(fmt.Sprintf(":%d\r\n", listLength)))

		case "LRANGE":
			// example command: LRANGE myItems 1 2
			// arguments derived from the command: ["myItems", myItems[1], muItems[2]]
			// args[0] = "myItems" --> key
			if len(args) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'LRANGE' command\r\n"))
				continue
			}

			// parsing the key
			key, ok := args[0].(string)
			if !ok {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			// after parsing, the parser reads everything as strings
			// so need to convert strings to integers using strconv.Atoi()
			// converting start & end indices (they come as strings)
			startStr, ok1 := args[1].(string)
			endStr, ok2 := args[2].(string)
			if !ok1 || !ok2 {
				conn.Write([]byte("-ERR indices must be strings\r\n"))
				continue
			}

			start, err1 := strconv.Atoi(startStr)
			end, err2 := strconv.Atoi(endStr)
			if err1 != nil || err2 != nil {
				conn.Write([]byte("-ERR start and end indices must be integers\r\n"))
				continue
			}

			list, ok := r.LRANGE(key, start, end)
			fmt.Println("list: ", list)
			if !ok {
				conn.Write([]byte("-ERR list not found or invalid range\r\n"))
				continue
			}

			// list is of type interface{}
			items, ok := list.([]string)
			fmt.Println("items: ", items)
			if !ok {
				conn.Write([]byte("-ERR internal type error\r\n"))
				continue
			}

			// formatting like Redis output
			var response string
			response = fmt.Sprintf("*%d\r\n", len(items))
			for _,v := range items {
				// response += fmt.Sprintf("$%d) \"%s\"\r\n", i+1, v)
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			}

			conn.Write([]byte(response))

		case "LPOP":
			if len(args) < 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'LPOP' command\r\n"))
				continue
			}

			// parsing the key
			key, ok := args[0].(string)
			if !ok {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			poppedElement, ok := r.LPOP(key)
			if !ok {
				conn.Write([]byte("-ERR list not found or invalid\r\n"))
				continue
			}

			conn.Write([]byte(poppedElement))

		case "RPOP":
			if len(args) < 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'RPOP' command\r\n"))
				continue
			}

			// parsing the key
			key, ok := args[0].(string)
			if !ok {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			poppedElement, ok := r.RPOP(key)
			if !ok {
				conn.Write([]byte("-ERR list not found or invalid\r\n"))
				continue
			}

			conn.Write([]byte(poppedElement))

		case "LLEN":
			// command syntax: LLEN key --> args = [1]
			if len(args) != 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'LLEN' command\r\n"))
				continue
			}

			key, ok := args[0].(string)
			if !ok {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			listLength, ok := r.LLEN(key)
			if !ok {
				conn.Write([]byte(":0\r\n")) // sending 0 if key doesn't exist
				continue
			}

			conn.Write([]byte(fmt.Sprintf(":%d\r\n", listLength)))
		
		case "LINDEX":
			// command syntax: LINDEX key index --> args = [key, index]
			if len(args) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'LINDEX' command\r\n"))
				continue
			}

			key, ok := args[0].(string)
			if !ok {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			index, ok := args[1].(string)
			if !ok {
				conn.Write([]byte("-ERR index is required\r\n"))
				continue
			}

			reqIndex, err := strconv.Atoi(index)
			if err != nil {
				conn.Write([]byte("-ERR index must be integer\r\n"))
				continue
			}

			element, ok := r.LINDEX(key, reqIndex)
			if !ok {
				conn.Write([]byte("$-1\r\n"))
				continue
			}

			conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)))

		case "LSET":
			// command syntax: LSET key index element (LSET myList 0 "four")
			if len(args) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'LSET' command\r\n"))
				continue
			}

			// checking & validating key, index & element
			key, ok := args[0].(string)
			if !ok {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}
			indexStr, ok := args[1].(string)
			if !ok {
				conn.Write([]byte("-ERR index must be provided\r\n"))
				continue
			}
			element, ok := args[2].(string)
			if !ok {
				conn.Write([]byte("-ERR element must be string\r\n"))
				continue
			}

			// type conversion for index, from string to integer
			indexInt, err := strconv.Atoi(indexStr)
			if err != nil {
				conn.Write([]byte("-ERR error converting index to integer\r\n"))
				continue
			}

			// calling the .LSET function
			ok = r.LSET(key, indexInt, element)
			if !ok {
				conn.Write([]byte("-ERR index out of bounds or argument invalid\r\n"))
				continue
			}

			conn.Write([]byte("+OK\r\n"))

		case "LREM":
			// command syntax: LREM key count element (LREM myList -2 "hello")
			if len(args) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for '' command\r\n"))
				continue
			}

			key, keyOk := args[0].(string)
			count, countOk := args[1].(string)
			element, elementOk := args[2].(string)

			if !keyOk || !countOk || !elementOk {
				conn.Write([]byte("-ERR arguments provided are of wrong types\r\n"))
				continue
			}

			countInt, err := strconv.Atoi(count)
			if err != nil {
				conn.Write([]byte("-ERR error converting integer to string"))
				continue
			}

			removed, ok := r.LREM(key, countInt, element)
			if !ok {
				conn.Write([]byte("-ERR something went wrong\r\n"))
				continue
			}

			strRemoved := strconv.Itoa(removed)
			fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(strRemoved), strRemoved)

		case "LTRIM":
			// command syntax: LTRIM key start stop --> args = [key, start, stop]
			if len(args) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'LTRIM' command\r\n"))
				continue
			}

			key, keyOk := args[0].(string)
			start, startOk := args[1].(string)
			stop, stopOk := args[2].(string)

			if !keyOk || !startOk || !stopOk {
				conn.Write([]byte("-ERR arguments are of wrong types\r\n"))
				continue
			}

			startInt, err1 := strconv.Atoi(start)
			stopInt, err2 := strconv.Atoi(stop)

			if err1 != nil || err2 != nil {
				conn.Write([]byte("-ERR start and stop indices must be integers\r\n"))
				continue
			}

			list, ok := r.LTRIM(key, startInt, stopInt)
			if !ok {
				conn.Write([]byte("-ERR list not found or invalid range\r\n"))
				continue
			}

			// list is of type interface{}
			items, ok := list.([]string)
			fmt.Println("items: ", items)
			if !ok {
				conn.Write([]byte("-ERR internal type error\r\n"))
				continue
			}

			var response string
			response = fmt.Sprintf("*%d\r\n", len(items))
			for _, v := range items {
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			}

			conn.Write([]byte(response))

		case "SADD":
			// example command: SADD key member [member ...]
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SADD' command\r\n"))
				continue
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			members := []string{}
			for _, v := range args[1:] {
				member, ok := v.(string)
				if !ok {
					conn.Write([]byte("-ERR members must be string\r\n"))
					continue
				}
				members = append(members, member)
			}

			added, ok := r.SADD(key, members)
			if !ok {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			conn.Write([]byte(fmt.Sprintf(":%d\r\n", added)))

		case "SISMEMBER":
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SISMEMBER' command\r\n"))
				continue
			}

			key, keyOk := args[0].(string)
			member, memberOk := args[1].(string)
			if !keyOk || !memberOk {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			result, ok := r.SISMEMBER(key, member)
			if !ok {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			resultInt := strconv.Itoa(result)

			fmt.Fprintf(conn, ":%v\r\n", resultInt)

		case "SREM":
			// example command: SREM key member [member...]
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SREM' command\r\n"))
				continue
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			members := []string{}
			for _, v := range args[1:] {
				memberStr, ok := v.(string)
				if !ok {
					conn.Write([]byte("-ERR member shall be string\r\n"))
				}
				members = append(members, memberStr)
			}

			result, ok := r.SREM(key, members)
			if !ok {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			fmt.Fprintf(conn, ":%v\r\n", result)

		case "SCARD":
			// command syntax: SCARD key (SCARD list)
			if len(args) != 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SCARD' command\r\n"))
				continue
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			result, ok := r.SCARD(key)
			if !ok {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			fmt.Fprintf(conn, ":%v\r\n", result)

		case "SMEMBERS":
			if len(args) < 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'SMEMBERS' command\r\n"))
				continue
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			members, ok := r.SMEMBERS(key)
			if !ok {
				conn.Write([]byte("*0\r\n")) // returning empty set
				continue
			}

			var response string
			response = fmt.Sprintf("*%d\r\n", len(members))
			for _, v := range members {
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			}

			conn.Write([]byte(response))

		case "HSET":
			if len(args) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'HSET' command\r\n"))
				continue
			}

			key, ok := args[0].(string)
			if !ok {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			fields := []string{}
			for i:=1; i<len(args); i=i+2 {
				fields = append(fields, args[i].(string))
			}

			values := []string{}
			for i:=2; i<len(args); i=i+2 {
				values = append(values, args[i].(string))
			}

			fieldValues := make(map[string]string)
			for idx, v := range fields {
				fieldValues[v] = values[idx]
			}

			result, isOk := r.HSET(key, fieldValues)
			if !isOk {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			fmt.Fprintf(conn, ":%d\r\n", result)

		case "HGET":
			if len(args) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'HGET' command\r\n"))
				continue
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			field, ok2 := args[1].(string)
			if !ok2 {
				conn.Write([]byte("-ERR field must be string\r\n"))
				continue
			}

			result, ok3 := r.HGET(key, field)
			if !ok3 {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			fmt.Fprintf(conn, ":%s\r\n", result)

		case "HGETALL":
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'HGET' command\r\n"))
				continue
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			result, ok2 := r.HGETALL(key)
			if !ok2 {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			var response string
			response = fmt.Sprintf("*%d\r\n", len(result))
			for _,v := range result {
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			}

			conn.Write([]byte(response))

		case "HDEL":
			if len(args) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'HGET' command\r\n"))
				continue
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				conn.Write([]byte("-ERR key and field must be string\r\n"))
				continue
			}

			field := args[1:]
			fields := []string{}
			for _, v := range field {
				if _, ok := v.(string); !ok {
					continue
				}
				vStr := v.(string)
				fields = append(fields, vStr)
			}

			result, ok := r.HDEL(key, fields)
			if !ok {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			resultInt := strconv.Itoa(result)
			fmt.Fprintf(conn, ":%v\r\n", resultInt)

		case "HLEN":
			// command syntax: HLEN key
			if len(args) < 1 {
				conn.Write([]byte("-ERR wrong number of arguments for 'HGET' command\r\n"))
				continue
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			result, ok2 := r.HLEN(key)
			if !ok2 {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			resultInt := strconv.Itoa(result)
			fmt.Fprintf(conn, ":%v\r\n", resultInt)

		case "EXPIRE":
			// command syntax: EXPIRE key seconds
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'EXPIRE' command\r\n"))
				continue
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			seconds, ok2 := args[1].(string)
			if !ok2 {
				conn.Write([]byte("-ERR seconds must be string\r\n"))
				continue
			}

			secondsInt, err := strconv.Atoi(seconds)
			if err != nil {
				conn.Write([]byte("-ERR operation failed while converting seconds to integer\r\n"))
				continue
			}

			result, ok := r.EXPIRE(key, secondsInt)
			if !ok {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			fmt.Fprintf(conn, ":%d\r\n", result)

		case "PEXPIRE":
			// command syntax: PEXPIRE key milliseconds
			if len(args) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'PEXPIRE' command\r\n"))
				continue
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				conn.Write([]byte("-ERR key must be string\r\n"))
				continue
			}

			ms, ok2 := args[1].(string)
			if !ok2 {
				conn.Write([]byte("-ERR milliseconds must be string\r\n"))
				continue
			}

			msInt, err := strconv.Atoi(ms)
			if err != nil {
				conn.Write([]byte("-ERR operation failed while converting milliseconds to integer\r\n"))
				continue
			}

			result, ok := r.EXPIRE(key, msInt)
			if !ok {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				continue
			}

			fmt.Fprintf(conn, ":%d\r\n", result)

		

		default:
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}

