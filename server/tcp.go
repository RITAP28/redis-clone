package server

import (
	"fmt"
	"net"

	"redis-clone/cache"
)

func StartServer(addr string, r *cache.RedisCache) error {
	ln, err := net.Listen("tcp", addr);
	if err != nil {
		fmt.Println("error occured");
		return err;
	}
	defer ln.Close();

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error());
			continue;
		}
		fmt.Println("Client connected");
		go r.HandleConnection(conn);
	}
}