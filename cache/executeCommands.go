package cache

import (
	"fmt"
	"strconv"
	"strings"
)

func(r *RedisCache) ExecuteCommands(client *Client, cmdArray []any) string {
	if len(cmdArray) == 0 {
		return "-ERR empty command\r\n"
	}

	mainCommand, ok := cmdArray[0].(string)
	if !ok {
		return "-ERR command must be string\r\n"
	}

	args := cmdArray[1:]
	command := strings.ToUpper(mainCommand)

	switch command {
		case "SET":
			if len(args) != 2 {
				return "-ERR wrong number of arguments for 'set' command\r\n"
			}

			key, keyOk := args[0].(string);
			value, valueOk := args[1].(string);

			if !valueOk || !keyOk {
				return "-ERR arguments must be string\r\n"
			};

			// ttl -> time to expiry for the key is set to 10 seconds
			r.SET(key, value, 10000);
			return "+OK\r\n"

		case "GET":
			if len(args) != 1 {
				return "-ERR wrong number of arguments for 'get' command\r\n"
			}

			key, keyOk := args[0].(string);
			if !keyOk {
				return "-ERR argument must be string\r\n"
			};

			value, ok := r.GET(key);
			if !ok {
				return "$-1\r\n"
			}

			stringValue, isString := value.(string);
			if !isString {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			response := fmt.Sprintf("$%d\r\n%s\r\n", len(stringValue), stringValue); // RESP representation of string
			return response

		case "DELETE":
			if len(args) != 1 {
				return "-ERR wrong number of arguments for 'delete' command\r\n"
			}

			key, keyOk := args[0].(string);
			if !keyOk {
				return "-ERR argument must be string\r\n"
			};

			ok := r.DELETE(key);
			if !ok {
				return "-ERR Error while performing deletion operation\r\n"
			}

			return fmt.Sprintf(":%d\r\n", 1)

		case "LPUSH":
			if len(args) < 2 {
				return "-ERR wrong number of arguments for 'LPUSH' command\r\n"
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				return "-ERR key must be string\r\n"
			}

			// remaining arguments in the commandArray are values corresponding to the key
			values := []string{}
			for _, arg := range args[1:] {
				val, ok := arg.(string)
				if !ok {
					return "-ERR all values must be strings\r\n"
				}

				values = append(values, val)
			}

			// calling the LPUSH implementation
			listLength, ok := r.LPUSH(key, values...)
			if !ok {
				fmt.Println("something went wrong while processing lists")
			}

			return fmt.Sprintf(":%d\r\n", listLength)

		case "RPUSH":
			if len(args) < 2 {
				return "-ERR wrong number of arguments for 'LPUSH' command\r\n"
			}

			key, ok := args[0].(string)
			if !ok {
				return "-ERR key must be string\r\n"
			}

			values := []string{}
			for _, arg := range args[1:] {
				val, ok := arg.(string)
				if !ok {
					return "-ERR all values must be strings\r\n"
				}

				values = append(values, val)
			}

			listLength, ok := r.RPUSH(key, values)
			if !ok {
				return "-ERR values could not be pushed\r\n"
			}

			return fmt.Sprintf(":%d\r\n", listLength)

		case "LRANGE":
			// example command: LRANGE myItems 1 2
			// arguments derived from the command: ["myItems", myItems[1], muItems[2]]
			// args[0] = "myItems" --> key
			if len(args) < 3 {
				return "-ERR wrong number of arguments for 'LRANGE' command\r\n"
			}

			// parsing the key
			key, ok := args[0].(string)
			if !ok {
				return "-ERR key must be string\r\n"
			}

			// after parsing, the parser reads everything as strings
			// so need to convert strings to integers using strconv.Atoi()
			// converting start & end indices (they come as strings)
			startStr, ok1 := args[1].(string)
			endStr, ok2 := args[2].(string)
			if !ok1 || !ok2 {
				return "-ERR indices must be strings\r\n"
			}

			start, err1 := strconv.Atoi(startStr)
			end, err2 := strconv.Atoi(endStr)
			if err1 != nil || err2 != nil {
				return "-ERR start and end indices must be integers\r\n"
			}

			list, ok := r.LRANGE(key, start, end)
			fmt.Println("list: ", list)
			if !ok {
				return "-ERR list not found or invalid range\r\n"
			}

			// list is of type interface{}
			items, ok := list.([]string)
			fmt.Println("items: ", items)
			if !ok {
				return "-ERR internal type error\r\n"
			}

			// formatting like Redis output
			var response string
			response = fmt.Sprintf("*%d\r\n", len(items))
			for _,v := range items {
				// response += fmt.Sprintf("$%d) \"%s\"\r\n", i+1, v)
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			}

			return response

		case "LPOP":
			if len(args) < 1 {
				return "-ERR wrong number of arguments for 'LPOP' command\r\n"
			}

			// parsing the key
			key, ok := args[0].(string)
			if !ok {
				return "-ERR key must be string\r\n"
			}

			poppedElement, ok := r.LPOP(key)
			if !ok {
				return "-ERR list not found or invalid\r\n"
			}

			return poppedElement

		case "RPOP":
			if len(args) < 1 {
				return "-ERR wrong number of arguments for 'RPOP' command\r\n"
			}

			// parsing the key
			key, ok := args[0].(string)
			if !ok {
				return "-ERR key must be string\r\n"
			}

			poppedElement, ok := r.RPOP(key)
			if !ok {
				return "-ERR list not found or invalid\r\n"
			}

			return poppedElement

		case "LLEN":
			// command syntax: LLEN key --> args = [1]
			if len(args) != 1 {
				return "-ERR wrong number of arguments for 'LLEN' command\r\n"
			}

			key, ok := args[0].(string)
			if !ok {
				return "-ERR key must be string\r\n"
			}

			listLength, ok := r.LLEN(key)
			if !ok {
				return ":0\r\n" // sending 0 if key doesn't exist
			}

			return fmt.Sprintf(":%d\r\n", listLength)
		
		case "LINDEX":
			// command syntax: LINDEX key index --> args = [key, index]
			if len(args) != 2 {
				return "-ERR wrong number of arguments for 'LINDEX' command\r\n"
			}

			key, ok := args[0].(string)
			if !ok {
				return "-ERR key must be string\r\n"
			}

			index, ok := args[1].(string)
			if !ok {
				return "-ERR index is required\r\n"
			}

			reqIndex, err := strconv.Atoi(index)
			if err != nil {
				return "-ERR index must be integer\r\n"
			}

			element, ok := r.LINDEX(key, reqIndex)
			if !ok {
				return "$-1\r\n"
			}

			return fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)

		case "LSET":
			// command syntax: LSET key index element (LSET myList 0 "four")
			if len(args) != 3 {
				return "-ERR wrong number of arguments for 'LSET' command\r\n"
			}

			// checking & validating key, index & element
			key, ok := args[0].(string)
			if !ok {
				return "-ERR key must be string\r\n"
			}
			indexStr, ok := args[1].(string)
			if !ok {
				return "-ERR index must be provided\r\n"
			}
			element, ok := args[2].(string)
			if !ok {
				return "-ERR element must be string\r\n"
			}

			// type conversion for index, from string to integer
			indexInt, err := strconv.Atoi(indexStr)
			if err != nil {
				return "-ERR error converting index to integer\r\n"
			}

			// calling the .LSET function
			ok = r.LSET(key, indexInt, element)
			if !ok {
				return "-ERR index out of bounds or argument invalid\r\n"
			}

			return "+OK\r\n"

		case "LREM":
			// command syntax: LREM key count element (LREM myList -2 "hello")
			if len(args) != 3 {
				return "-ERR wrong number of arguments for 'LREM' command\r\n"
			}

			key, keyOk := args[0].(string)
			count, countOk := args[1].(string)
			element, elementOk := args[2].(string)

			if !keyOk || !countOk || !elementOk {
				return "-ERR arguments provided are of wrong types\r\n"
			}

			countInt, err := strconv.Atoi(count)
			if err != nil {
				return "-ERR error converting integer to string"
			}

			removed, ok := r.LREM(key, countInt, element)
			if !ok {
				return "-ERR something went wrong\r\n"
			}

			strRemoved := strconv.Itoa(removed)
			return fmt.Sprintf("$%d\r\n%s\r\n", len(strRemoved), strRemoved)

		case "LTRIM":
			// command syntax: LTRIM key start stop --> args = [key, start, stop]
			if len(args) != 3 {
				return "-ERR wrong number of arguments for 'LTRIM' command\r\n"
			}

			key, keyOk := args[0].(string)
			start, startOk := args[1].(string)
			stop, stopOk := args[2].(string)

			if !keyOk || !startOk || !stopOk {
				return "-ERR arguments are of wrong types\r\n"
			}

			startInt, err1 := strconv.Atoi(start)
			stopInt, err2 := strconv.Atoi(stop)

			if err1 != nil || err2 != nil {
				return "-ERR start and stop indices must be integers\r\n"
			}

			list, ok := r.LTRIM(key, startInt, stopInt)
			if !ok {
				return "-ERR list not found or invalid range\r\n"
			}

			// list is of type interface{}
			items, ok := list.([]string)
			fmt.Println("items: ", items)
			if !ok {
				return "-ERR internal type error\r\n"
			}

			var response string
			response = fmt.Sprintf("*%d\r\n", len(items))
			for _, v := range items {
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			}

			return response

		case "SADD":
			// example command: SADD key member [member ...]
			if len(args) < 2 {
				return "-ERR wrong number of arguments for 'SADD' command\r\n"
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				return "-ERR key must be string\r\n"
			}

			members := []string{}
			for _, v := range args[1:] {
				member, ok := v.(string)
				if !ok {
					return "-ERR members must be string\r\n"
				}
				members = append(members, member)
			}

			added, ok := r.SADD(key, members)
			if !ok {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			return fmt.Sprintf(":%d\r\n", added)

		case "SISMEMBER":
			if len(args) < 2 {
				return "-ERR wrong number of arguments for 'SISMEMBER' command\r\n"
			}

			key, keyOk := args[0].(string)
			member, memberOk := args[1].(string)
			if !keyOk || !memberOk {
				return "-ERR key must be string\r\n"
			}

			result, ok := r.SISMEMBER(key, member)
			if !ok {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			resultInt := strconv.Itoa(result)

			return fmt.Sprintf(":%v\r\n", resultInt)

		case "SREM":
			// example command: SREM key member [member...]
			if len(args) < 2 {
				return "-ERR wrong number of arguments for 'SREM' command\r\n"
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				return "-ERR key must be string\r\n"
			}

			members := []string{}
			for _, v := range args[1:] {
				memberStr, ok := v.(string)
				if !ok {
					return "-ERR member shall be string\r\n"
				}
				members = append(members, memberStr)
			}

			result, ok := r.SREM(key, members)
			if !ok {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			return fmt.Sprintf(":%v\r\n", result)

		case "SCARD":
			// command syntax: SCARD key (SCARD list)
			if len(args) != 1 {
				return "-ERR wrong number of arguments for 'SCARD' command\r\n"
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				return "-ERR key must be string\r\n"
			}

			result, ok := r.SCARD(key)
			if !ok {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			return fmt.Sprintf(":%v\r\n", result)

		case "SMEMBERS":
			if len(args) < 1 {
				return "-ERR wrong number of arguments for 'SMEMBERS' command\r\n"
			}

			key, keyOk := args[0].(string)
			if !keyOk {
				return "-ERR key must be string\r\n"
			}

			members, ok := r.SMEMBERS(key)
			if !ok {
				return "*0\r\n" // returning empty set
			}

			var response string
			response = fmt.Sprintf("*%d\r\n", len(members))
			for _, v := range members {
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			}

			return response

		case "HSET":
			if len(args) < 3 {
				return "-ERR wrong number of arguments for 'HSET' command\r\n"
			}

			key, ok := args[0].(string)
			if !ok {
				return "-ERR key must be string\r\n"
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
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			return fmt.Sprintf(":%d\r\n", result)

		case "HGET":
			if len(args) < 3 {
				return "-ERR wrong number of arguments for 'HGET' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key must be string\r\n"
			}

			field, ok2 := args[1].(string)
			if !ok2 {
				return "-ERR field must be string\r\n"
			}

			result, ok3 := r.HGET(key, field)
			if !ok3 {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			return fmt.Sprintf(":%s\r\n", result)

		case "HGETALL":
			if len(args) < 2 {
				return "-ERR wrong number of arguments for 'HGET' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key must be string\r\n"
			}

			result, ok2 := r.HGETALL(key)
			if !ok2 {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			var response string
			response = fmt.Sprintf("*%d\r\n", len(result))
			for _,v := range result {
				response += fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			}

			return response

		case "HDEL":
			if len(args) < 3 {
				return "-ERR wrong number of arguments for 'HGET' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key and field must be string\r\n"
			}

			field := args[1:]
			fields := []string{}
			for _, v := range field {
				if _, ok := v.(string); !ok {
				}
				vStr := v.(string)
				fields = append(fields, vStr)
			}

			result, ok := r.HDEL(key, fields)
			if !ok {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			resultInt := strconv.Itoa(result)
			return fmt.Sprintf(":%v\r\n", resultInt)

		case "HLEN":
			// command syntax: HLEN key
			if len(args) < 1 {
				return "-ERR wrong number of arguments for 'HGET' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key must be string\r\n"
			}

			result, ok2 := r.HLEN(key)
			if !ok2 {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			resultInt := strconv.Itoa(result)
			return fmt.Sprintf(":%v\r\n", resultInt)

		case "EXPIRE":
			// command syntax: EXPIRE key seconds
			if len(args) < 2 {
				return "-ERR wrong number of arguments for 'EXPIRE' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key must be string\r\n"
			}

			seconds, ok2 := args[1].(string)
			if !ok2 {
				return "-ERR seconds must be string\r\n"
			}

			secondsInt, err := strconv.Atoi(seconds)
			if err != nil {
				return "-ERR operation failed while converting seconds to integer\r\n"
			}

			result, ok := r.EXPIRE(key, secondsInt)
			if !ok {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			return fmt.Sprintf(":%d\r\n", result)

		case "PEXPIRE":
			// command syntax: PEXPIRE key milliseconds
			if len(args) < 2 {
				return "-ERR wrong number of arguments for 'PEXPIRE' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key must be string\r\n"
			}

			ms, ok2 := args[1].(string)
			if !ok2 {
				return "-ERR milliseconds must be string\r\n"
			}

			msInt, err := strconv.Atoi(ms)
			if err != nil {
				return "-ERR operation failed while converting milliseconds to integer\r\n"
			}

			result, ok := r.PEXPIRE(key, msInt)
			if !ok {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			return fmt.Sprintf(":%d\r\n", result)

		case "TTL":
			if len(args) < 1 {
				return "-ERR wrong number of arguments for 'TTL' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key must be string\r\n"
			}

			result, ok2 := r.TTL(key)
			if !ok2 {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			resultInt := strconv.Itoa(result)
			return fmt.Sprintf(":%s\r\n", resultInt)

		case "PTTL":
			if len(args) < 1 {
				return "-ERR wrong number of arguments for 'PTTL' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key must be string\r\n"
			}

			result, ok2 := r.PTTL(key)
			if !ok2 {
				return "-WRONGTYPE operation against a key holding the wrong kind of value\r\n"
			}

			resultInt := strconv.Itoa(result)
			return fmt.Sprintf(":%s\r\n", resultInt)

		case "PERSIST":
			if len(args) < 1 {
				return "-ERR wrong number of arguments for 'PTTL' command\r\n"
			}

			key, ok1 := args[0].(string)
			if !ok1 {
				return "-ERR key must be string\r\n"
			}

			result := r.PERSIST(key)
			return fmt.Sprintf(":%d\r\n", result)

		case "SAVE":
			err := r.SaveToDisk("dump.rgb.json")
			if err != nil {
				return "-ERR error while saving file to disc\r\n"
			}

			fmt.Println("data loaded onto disc with saved data")
			return "+OK\r\n"

		default:
			return "-ERR unknown command\r\n"
	}
}