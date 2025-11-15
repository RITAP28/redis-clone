package main

import (
	"fmt"
	"redis-clone/cache"
	"redis-clone/server"
	"time"
)

func main() {
	fmt.Println("Launching server...");

	redisServer := cache.NewRedisServer();

	// purging expired keys every 10 seconds in the background
	redisServer.StartExpiryCleaner(10 * time.Second)

	err := server.StartServer(":8080", redisServer);
	if err != nil {
		fmt.Println("Something wrong happened");
		panic(err);
	};
};