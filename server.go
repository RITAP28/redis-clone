package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func (r *RedisCache) handleConnection (conn net.Conn) {
	defer conn.Close();

	reader := bufio.NewReader(conn);
	for {
		// Read the request
		req, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading: ", err);
			return;
		}

		req = strings.TrimSpace(req);
		parts := strings.Split(req, " ");

		if len(parts) == 0 {
			continue
		};

		command := strings.ToUpper(parts[0])

		switch command {
		case "SET":
			if len(parts) != 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				continue;
			}

			key := parts[1]
			value := parts[2]
			r.SET(key, value, 1000);
			conn.Write([]byte("+OK\r\n"))

		case "GET":
			if len(parts) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue;
			}

			key := parts[1];
			value, ok := r.GET(key);
			if !ok {
				conn.Write([]byte("$1\r\n"));
				continue;
			} else {
				response := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value); // RESP representation of string
				conn.Write([]byte(response))
			}

		case "DELETE":
			if len(parts) != 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'delete' command\r\n"));
				continue;
			}

			key := parts[1];
			ok := r.DELETE(key);
			if !ok {
				conn.Write([]byte("-ERR Error while performing deletion operation\r\n"));
				continue;
			}

		default:
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}

func (r *RedisCache) handleConnectionTwo (conn net.Conn) {
	response := "+OK\r\n";
	_, err := conn.Write([]byte(response));

	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error());
		return;
	} else {
		fmt.Println("Successful connection");
	}

	conn.Close()
}

func main() {
	fmt.Println("Launching server...");

	ln, err := net.Listen("tcp", ":8080");
	if err != nil {
		panic(err);
	};

	defer ln.Close();

	server := newRedisServer();
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error());
			continue;
		}

		fmt.Println("Client connected");
		// go server.handleConnection(conn);
		go server.handleConnectionTwo(conn);
	}
}