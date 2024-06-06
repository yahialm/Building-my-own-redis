package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	INTEGER = ':'
	ARRAY   = '*'
	ERROR   = '-' // simple errors
	BULK    = '$' // bulk strings
	STRING  = '+' // simple string
	CRLF = "\r\n"
)

type Value struct {
	typ  string
	str  string
	// num  int
	bulk string
	arr  []Value
}

// TODO: Maybe RespMessage is better than Resp
type RespMessage struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *RespMessage {
	return &RespMessage{reader: bufio.NewReader(rd)}
}

// Example: "+OK\r\n" --> "+", "O", "K", "\r", "\n"
// then we slice off the last 2 bytes '\r' '\n'
// finally we return ["+", "O", "K"] with n = 5 (5 reads)

// len(line)>=2 condition is necessary to prevent 
// out of range access

func (r *RespMessage) ReadLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		line = append(line, b)
		n += 1
		if len(line)>=2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, err
}

func (r *RespMessage) ReadInteger() (x int, n int, err error) {
	line, n, err := r.ReadLine()
	if err != nil {
		return -1, 0, err // TODO: x->"0" is not a good option to be returned
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return -1, 0, err
	}

	return int(i64), n, nil
}

/* func (r *Resp) ReadInteger() (x int, n int, err error) {
	line, n, err := r.ReadLine()
	if err != nil {
		return 0, 0, err // TODO: "0" is not a good option to be returned
	}
	i64, err := strconv.ParseInt(string(line[1:]), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return int(i64), n-1, nil
} */

func (r *RespMessage) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// *3\r\n$3\r\nset\r\n$6\r\nplayer\r\n$7\r\nMolinex

func (r *RespMessage) readArray() (Value , error) {
	v := Value{}
	v.typ = "array"

	len, _, err := r.ReadInteger()
	if err != nil {
		return v, err
	}

	v.arr = make([]Value, 0)
	for i:=0; i<len; i++ {
		b, err := r.Read()
		if err != nil {
			return v, err
		}
		v.arr = append(v.arr, b)
	}

	return v, nil
}

func (r *RespMessage) readBulk() (v Value, err error){
	v = Value{}
	v.typ = "bulk"

	len, _, err := r.ReadInteger()
	if err != nil {
		return v, err
	}
	bulk := make([]byte, len)

	_, err = r.reader.Read(bulk)
	if err != nil {
		return v, err
	}

	v.bulk = string(bulk)

	// Read the CRLF, the pointer is left there
	// Without using readLine things will messed up
	r.ReadLine()
	return v, nil
}

//========================================================

// The first part represents the implementation of the parser.
// Our parser is used to parse the received RESP message.
// Each RESP message is represented using "value" stuct.

// Now, we should consider how our program will send RESP compatible response.
// For every type(string, bulk, array,...) a method will be used to represent our message
// to the client in the RESP format (as bytes to send over the conn).

func (v Value) Marshal() []byte {
	resp := []byte{}
	switch v.typ {
	case "string":
		return v.marshalString()
	case "bulk":
		return v.marshalBulk()
	// case "ineteger":
	// 	return v.marshalInteger()
	case "error":
		return v.marshalError()
	case "array":
		return v.marshalArray()
	case "null" :
		return v.marshalNull()
	default:
		return resp
	}
} 

func (v Value) marshalString()  []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, []byte(v.str)...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')
	
	return bytes
}
 

func (v Value) marshalArray() []byte {
	var bytes []byte
	arr_len := len(v.arr)
	bytes = append(bytes, ARRAY) // "*"
	bytes = append(bytes, strconv.Itoa(arr_len)...)
	bytes = append(bytes, '\r', '\n')
	for i:=0; i<arr_len; i++ {
		bytes = append(bytes, v.arr[i].Marshal()...)
		bytes = append(bytes, '\r', '\n')
	}
	return bytes
}

func (v Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, '\r', '\n')
	
	return bytes
}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w Writer) Write(v Value) error {
	var bytes = v.Marshal()
	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}