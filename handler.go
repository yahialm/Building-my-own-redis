package main

import "sync"

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"GET":  get,
	"SET":  set,
	"DEL": rm,
	"HSET": hset,
	"HGET": hget, 
}

// GET & SET command maps
var SETs = map[string]string{}
var SETsMu = &sync.RWMutex{}

// HGET & HSET commands map
var HSETs = map[string]map[string]string{}
var HSETsMux = &sync.RWMutex{}

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

func rm(args []Value) Value {
	keys_len := len(args)

	if keys_len == 0 {
		return Value{typ: "ERROR", str: "Expected at least one argument for DEL command"}
	}

	// Lock
	SETsMu.Lock()
	for k:=0; k<keys_len; k++ {
		delete(SETs, args[k].bulk)
	}
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: "ERROR", str: "Expected only 3 arguments for HSET command"}
	}

	hash := args[0].bulk
	key := args[1].bulk
	value := args[2].bulk

	// Lock
	HSETsMux.Lock()

	// Check if hash exist in the map
	// hash -> users / key -> user1 / value -> yahiaID
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	} 
	
	HSETs[hash][key] = value

	HSETsMux.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: "ERROR", str: "Expected only 2 arguments for HGET command"}
	}

	hash := args[0].bulk
	key := args[1].bulk

	HSETsMux.RLock()
	value, ok := HSETs[hash][key]

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: value}
}
