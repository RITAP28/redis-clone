package server

import (
	"fmt"
	"net"
	"os"
	"redis-clone/cache"
)


func StartServer(addr string, r *cache.RedisCache) error {
	ln, err := net.Listen("tcp", addr);
	if err != nil {
		fmt.Println("error occured while starting the server: ", err);
		os.Exit(1);
		return err;
	}
	defer ln.Close();

	for {
		conn, err := ln.Accept()
		client := cache.NewClient(conn)

		if err != nil {
			fmt.Println("Error while accepting requests: ", err.Error());
			continue;
		}
		fmt.Println("Client connected");
		go r.HandleConnection(client);
	}
}