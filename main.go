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

		// Every command is an array of arguments
		if value.typ != "array" {
			fmt.Println("Invalid req, array expected")
			continue // next iter, we can't stop the loop
		}

		// TODO: Is this condition really needed ?
		if len(value.arr) == 0 {
			fmt.Println("Invalid req, array expected with more than 0 args")
			continue
		}

		// In the redis-cli, lowercase chars are authorized, WE NEED TO CONVERT THEM TO UPPERCASE format
		// since our handlers are mapped to uppercase format of the commands GET , SET, PING ...
		command := strings.ToUpper(value.arr[0].bulk)
		args := value.arr[1:]

		// Sending data through the socket
		// `write` is the reponsible object for writing in the connection established with the server
		writer := NewWriter(conn)

		// Getting the correspondant handler for the command
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			// writer.Write(Value{typ: "string", str: command + " is an invalid command"})
			continue
		}
		
		// Execute the logic of the handler
		// passing the args to the handler. 
		result := handler(args) // ping(args []Value)
		
		// Sending the response in the RESP binary format to the client
		// The client is responsible for parsing the response and displaying it.
		writer.Write(result)	
	}
}
