package cache

import "fmt"


func(r *RedisCache) LPUSH(key string, values ...string) (int, bool) {
	fmt.Println("inside LPUSH command function")
	r.mu.Lock()
	defer r.mu.Unlock()

	// checking whether the same key exists in the memory
	entry, exists := r.store[key]

	// CASE 1
	// if it does not exist, then make a new list
	// and append all the values in the right order
	if !exists {
		fmt.Println("key does not exist, new list is created")
		// building the list
		list := []string{}
		for i:=len(values)-1; i>=0; i-- {
			list = append([]string{values[i]}, list...)
			// list = append(list, values[i])
		}

		// then, assigning the list to entry.value
		entry = &Entry{value: list}
		
		// storing the list corresponding to the key and returning the length of the list
		r.store[key] = entry
		return len(list), true
	}

	// CASE 2
	// if the list already exists, then simply append the values to the left of the existing values

	// for example: existing list = ['a', 'b'] & values = ['c', 'd']
	// expected resultant list: lst = ['d', 'c', 'a', 'b']

	lst, isList := entry.value.([]string)
	if !isList {
		fmt.Println("wrong type WRONGTYPE")
		return 0, false
	}

	for i:=len(values)-1; i>=0; i-- {
		lst = append([]string{values[i]}, lst...)
	}

	entry.value = lst
	r.store[key] = entry

	fmt.Println("after creating a new list and prepending values: ", lst)
	return len(lst), true
}

func(r *RedisCache) RPUSH(key string, values []string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		list := append([]string{}, values...)
		entry = &Entry{value: list}
		r.store[key] = entry
		return len(list), true
	}

	lst, isList := entry.value.([]string)
	if !isList {
		return 0, false
	}

	lst = append(lst, values...)
	entry.value = lst
	r.store[key] = entry

	return len(lst), true
}

func(r *RedisCache) LRANGE(key string, start int, end int) (interface{}, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// getting the values from the stored key
	entry, exists := r.store[key]
	if !exists {
		return nil, false
	}

	values, isList := entry.value.([]string)
	if !isList {
		return nil, false
	}

	// handle negative indices
	length := len(values)

	// start --> starting index & end --> ending index
	// suppose a list named "myList" exists with elements like: "item1", "item2", "item3", "item4"
	// cmd: LRANGE myList 0 -1; result: all elements inside the list
	// cmd: LRANGE myList 1 2; result: "item3" "item3"

	// checking for values which are incorrect by nature
	// starting index is negative, then converting it
	if start < 0 {
		start = start + length
		if start < 0 {
			start = 0
		}
	}

	// converting end negative index
	if end < 0 {
		end = length + end
	}

	// bounds checking
	if start > length-1 {
		return []string{}, true // returning empty list if start is out of bounds
	}
	if end >= length {
		end = length - 1
	}

	// returning empty list if range is invalid
	if start > end {
		return []string{}, true
	}

	// for normal cases, simple loop through the list
	// resultant_list := []string{}
	// for i:=start; i<=end; i++ {
	// 	resultant_list = append(resultant_list, values[i])
	// }

	// simple approach as opposed to looping through: slice the list
	resultant_list := values[start:end+1]
	fmt.Println("resultant_list: ", resultant_list)
	return resultant_list, true
}

func(r *RedisCache) LPOP(key string) (string, bool) {
	// syntax: LPOP key [count]
	// [count] can be an integer or null
	// Case A: if count == null, then by default, one element will be popped out | Status: Done
	// Case B: if count >= 1, then [count] number of elements will be pooped out | Status: Not Completed
	r.mu.Lock()
	defer r.mu.Unlock()

	// getting the values stored corresponding to the key
	entry, exists := r.store[key]
	if !exists {
		return "list does not exist", false
	}

	// extracting the list into the values variable
	values, isList := entry.value.([]string)
	if !isList {
		return "incorrect type found", false
	}

	poppedElement := values[0]
	length := len(values)

	// slicing the list by removing the left-most element and storing it in newList
	// replacing the old list with 'newList' i.e., the sliced list
	// returning the popped element
	newList := values[1:length-1]
	entry.value = newList

	return poppedElement, true
}

func(r *RedisCache) RPOP(key string) (string, bool) {
	// every functionality in RPOP is same as LPOP, except here we remove the right-most element in the list i.e., the last element
	// Case A: if count == null, then by default, one element will be popped out | Status: Done
	// Case B: if count >= 1, then [count] number of elements will be pooped out | Status: Not Completed

	r.mu.Lock()
	defer r.mu.Unlock()

	// getting the values stored corresponding to the key
	entry, exists := r.store[key]
	if !exists {
		return "list does not exist", false
	}

	// extracting the list into the values variable
	values, isList := entry.value.([]string)
	if !isList {
		return "incorrect type found", false
	}

	length := len(values)
	poppedElement := values[length-1]
	newList := values[0:length-2]

	entry.value = newList
	return poppedElement, true
}

func(r *RedisCache) LLEN(key string) (int, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return 0, false
	}

	values, isList := entry.value.([]string)
	if !isList {
		return 0, false
	}

	length := len(values)

	return length, true
}

func(r *RedisCache) LINDEX(key string, index int) (string, bool) {
	// command example: LINDEX key index
	// if index > length(entry.value), then (nil) is returned
	// if index is negative, then the index starts from the last --> -1 = last element; -2 = penultimate element

	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return "", false
	}

	values, isList := entry.value.([]string)
	if !isList {
		return "", false
	}

	// handling the negative index case
	if index < 0 {
		index = len(values) + index
	}

	// if the index is still negative from above
	// or if the index is greater than or equal to the length of the list
	if index < 0 || index >= len(values) {
		return "", false
	}

	// getting the element by the index from the list, and returning it
	element := values[index]
	return element, true
}

func(r*RedisCache) LSET(key string, index int, element string) (bool) {
	// command example: LSET key index element
	// error is returned for index out of bounds

	// the usual checking & validation
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return false
	}

	values, isList := entry.value.([]string)
	if !isList {
		return false
	}

	// negative indices need to handled
	length := len(values)
	if index < 0 {
		index = index + length
	}

	// error is returned if index is out of bounds
	if index < 0 || index >= length {
		return false
	}

	values[index] = element
	return true
}

func(r *RedisCache) LREM(key string, count int, element string) (int, bool) {
	// functionality: removes the first count occurences of elements equal to elements from the list stored at key
	// influence of count:
	// 1. count > 0 --> remove elements equal to element moving from head to tail (front to back)
	// 2. count < 0 --> remove elements equal to element moving from tail to head (back to front)
	// 3. count = 0 --> remove all elements equal to element

	// the usual checking & validation
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return 0, true
	}

	values, isList := entry.value.([]string)
	if !isList {
		return 0, false
	}

	// defining some useful variables
	length := len(values) 	// length of the list stored corresponding to the key
	removed := 0			// tracking how many elements have been removed
	newList := []string{}	// empty list

	// my approach: depending upon the value and sign of count, list will be looped, elements will be removed & remaining elements will be appended to the newList
	if count == 0 {
		// remove all occurences of element
		for _, v := range values {
			if v == element {
				removed++
				continue
			}
			newList = append(newList, v)
		}
	} else if count > 0 {
		for _, v := range values {
			if count > removed && v == element {
				removed++
				continue
			}
			newList = append(newList, v)
		}
	} else if count < 0 {
		targetCount := -count // making the targetCount variable positive

		// below is a very inefficient method
		for i:=length-1; i>=0; i-- {
			if values[i] == element && targetCount > removed {
				removed++
				continue
			}

			// prepending the elements to maintain the order of elements in the original list
			newList = append([]string{values[i]}, newList...)
		}

		// more efficient method
		// reverse := make([]string, 0, length)

		// // traverse from tail to head
		// for i:=length-1; i>=0; i++ {
		// 	v := values[i]
		// 	if v == element && targetCount > removed {
		// 		removed++
		// 		continue
		// 	}
		// 	reverse = append(reverse, v)
		// }

		// // reverse back to original order
		// for i:=len(reverse)-1; i>=0; i-- {
		// 	newList = append(newList, reverse[i])
		// }
	}

	// assigning the newList to the key
	entry.value = newList
	return removed, true
}

func(r *RedisCache) LTRIM(key string, start int, stop int) (interface{}, bool) {
	// the usual checking & validation
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.store[key]
	if !exists {
		return nil, false
	}

	values, isList := entry.value.([]string)
	if !isList {
		return nil, false
	}

	length := len(values)

	// checking negative indices
	// negative starting index
	if start < 0 {
		start = start + length
		if start < 0 {
			start = 0
		}
	}

	// negative ending index
	if stop < 0 {
		stop = stop + length
	}

	// if the 'stop' is larger than the length of the list
	// then, it will be treated as the last index of the list
	if stop > length-1 {
		stop = length - 1
	}

	if start > stop {
		// empty list will be returned
		entry.value = []string{}
		return entry.value, true
	}

	newList := []string{}
	for i:=start; i<=stop; i++ {
		newList = append(newList, values[i])
	}

	entry.value = newList
	return newList, true
}

