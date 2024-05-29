package main

import (
	"fmt"
	"net"
)

const (
	PORT = ":6379"
	CONN_PROTO = "tcp"
)

func main() {

	fmt.Println("Listening on port" + PORT)

	lis, err := net.Listen(CONN_PROTO, PORT)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	conn, err := lis.Accept()
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	defer conn.Close()

	for {
		resp := NewRsep(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		_ = value

		// Writing "+OK\r\n" to the connection
		writer := NewWriter(conn)
		writer.Write(Value{typ: "string", str: "OK"})
	}
}