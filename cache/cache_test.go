package cache

import (
	"sync"
	"testing"
	"time"
)

// testing basic SET and GET operations
func TestSetAndGet(t *testing.T) {
	cache := NewRedisServer();

	// testing a simple SET command
	cache.SET("hello", "world", 0);

	val, ok := cache.GET("hello");
	if !ok {
		t.Fatalf("GET failed for key1: expected to find key, but it was not found");
	}

	if val != "world" {
		t.Errorf("GET failed value for key1: expected value 'value1', but got '%s'", val);
	}
}

// testing GET method for getting a value for a key that doesn't exist
func TestGetNotExistentKey(t *testing.T) {
	cache := NewRedisServer();
	_, ok := cache.GET("nonexistent");

	if ok {
		t.Fatal("GET succeeded for a non-existent key, but it should have failed");
	};
};

// testing the Time-To-Live (ttl) expiry logic
func TestKeyExpiry(t *testing.T) {
	cache := NewRedisServer();

	// setting a key with a very short TTL of 1 second
	cache.SET("hello", "world", 1);

	// then immediately checking whether the key exists
	_, ok := cache.GET("hello");
	if !ok {
		t.Fatal("GET failed for 'short-lived' immediately after setting, but it should exist");
	};
	
	// checking after a little longer
	time.Sleep(1100 * time.Millisecond);

	// the key shall be gone
	_, okk := cache.GET("hello");
	if okk {
		t.Fatal("GET succeeded for 'short-lived' after TTL expired, but it should have been deleted");
	}
}

// testing DELETE method
func TestDelete(t *testing.T) {
	cache := NewRedisServer();

	cache.SET("hello", "world", 0);

	_, ok := cache.GET("hello");
	if !ok {
		t.Fatal("Setup for TestDelete failed: key 'hello' was not set correctly");
	};

	deleted := cache.DELETE("hello");
	if !deleted {
		t.Fatal("DELETE operation returned false, expected true");
	};

	_, okk := cache.GET("hello");
	if okk {
		t.Fatal("Key 'hello' still exists after being deleted");
	};
};

// test for concurrency safety using the mutex
func TestConcurrency(t *testing.T) {
	cache := NewRedisServer();

	// what is the sync package and WaitGroup method?
	var wg sync.WaitGroup
	numGoRoutines := 100;

	// we will have 100 goroutines all trying to SET and GET at the same time
	// if the mutex is working, this will not crash with a "concurrent map write" error
	for i := 0; i < numGoRoutines; i++ {
		go func (i int) {
			defer wg.Done();
			key := "concurrent_key"
			value := "concurrent_value"

			// Mix of writes and reads
			cache.SET(key, value, 0);
			cache.GET(key);
		}(i);
	};

	// wait for all goroutines to finish
	wg.Wait()

	val, ok := cache.GET("concurrent_key")
	if !ok || val != "concurrent_value" {
		t.Errorf("Final check after concurrent access failed. Got val: '%s', ok : %t", val, ok);
	}
} 