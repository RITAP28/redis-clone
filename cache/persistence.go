package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func(r *RedisCache) SaveToDisk(filename string) error {
	// locking the store
	r.mu.Lock()
	defer r.mu.Unlock()

	// json.MarshalIndent() is used to encode Go data structures (like structs, maps or slices) into JSON formatted data, with human-readable indentation
	// advantages: Debugging and logging data; Generating configuration files & Displaying readable API responses to the user
	// json.MarshalIndent() makes the response more human-readable than json.Marshal()

	// converting the data inside r.store into a human-readable format
	data, err := json.MarshalIndent(r.store, "", " ")
	if err != nil {
		return err
	}

	// now saving the JSON snapshot in disk
	// A FileMode represents a file's mode and permission bits.
	// 0644 (a standard LINUX permission): Owner can read/write , Group and Others can only read
	permissions := os.FileMode(0644)

	// need to check whether the dump file exists or not in the directory
	// if yes, then the content shall be updated
	// if not, then file will be created anyways
	return os.WriteFile(filename, data, permissions)
}

// for loading the data on startup of the redis server
func(r *RedisCache) LoadData(filename string) error {
	content, err := os.ReadFile(filename)

	// handling edge case -> if the loaded file is not found in the disc, then starting fresh is the only option
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("No persistence file found - starting fresh.")
		return nil
	}

	if err != nil {
		return err
	}

	var store map[string]*Entry
	if loadErr := json.Unmarshal(content, &store); loadErr != nil {
		return loadErr
	}

	for key, entry := range store {
		switch entry.Type {
		case "string":
			rawString, ok := entry.Value.(string)
			if ok {
				// strValue := rawString.(string)
				entry.Value = rawString
			}
		case "list":
			rawList, ok := entry.Value.([]interface{})
			if ok {
				strList := make([]string, len(rawList))
				for i, v := range rawList {
					strList[i] = fmt.Sprintf("%v", v)
				}
				entry.Value = strList
			}
		case "set":
			rawSet, ok := entry.Value.(map[string]interface{})
			if ok {
				strSet := make(map[string]struct{}, len(rawSet))
				for member := range rawSet {
					strSet[member] = struct{}{}
				}
				entry.Value = strSet
			}
		case "hash":
			rawMap, ok := entry.Value.(map[string]interface{})
			if ok {
				strMap := make(map[string]string, len(rawMap))
				for k, v := range rawMap {
					strMap[k] = fmt.Sprintf("%v", v)
				}
				entry.Value = strMap
			}
		}

		r.store[key] = entry
	}

	r.mu.Lock()
	r.store = store
	r.mu.Unlock()
	return nil
}
