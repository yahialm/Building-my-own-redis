package main

import (
	"fmt"
	"net"
	"strings"
)

const (
	PORT       = ":6379"
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
		resp := NewResp(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		if value.typ != "array" {
			fmt.Println("Invalid req, array expected")
			continue // next iter, we can't stop the loop
		}

		if len(value.arr) == 0 {
			fmt.Println("Invalid req, array expected with more than 0 args")
			continue
		}

		command := strings.ToUpper(value.arr[0].bulk)
		args := value.arr[1:]

		writer := NewWriter(conn)

		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}

		result := handler(args) // ping(args []Value)
		writer.Write(result)	
	}
}
