package main

import (
	"fmt"
	"redis-clone/cache"
	"redis-clone/server"
	"time"
)

func main() {
	fmt.Println("Launching server...");

	redisServer := cache.NewRedisServer()

	// loading saved data --> persistence
	redisServer.LoadData("dump.rgb.json")

	// purging expired keys every 20 seconds in the background
	redisServer.StartExpiryCleaner(20 * time.Second)

	err := server.StartServer(":8080", redisServer);
	if err != nil {
		fmt.Println("Something wrong happened");
		panic(err);
	};
};