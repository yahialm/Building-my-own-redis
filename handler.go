package main

import "sync"

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"GET":  get,
	"SET":  set,
}

var SETs = map[string]string{}
var SETsMu = &sync.RWMutex{}


// PING --> PONG
func ping(args []Value) Value {
	return Value{typ: "string", str: "PONG"}
}


// SET command to add a key value pair to the KV DB
func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "ERROR", str: "Expected only 2 args for GET command"}
	}

	key, value := args[0].bulk, args[1].bulk

	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

// GET command to get the value assigned to the key passed as argument
func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: "ERROR", str: "Expected only 1 argument for GET command"}
	}

	key := args[0].bulk
	SETsMu.RLock()	
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}
