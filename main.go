package main

import (
	"fmt"
	"redis-clone/cache"
	"redis-clone/server"
)

func main() {
	fmt.Println("Launching server...");

	redisServer := cache.NewRedisServer();

	err := server.StartServer(":8080", redisServer);
	if err != nil {
		fmt.Println("Something wrong happened");
		panic(err);
	};
};