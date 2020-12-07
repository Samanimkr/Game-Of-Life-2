package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"runtime"
)

// Args COMMENT
type Args struct {
	P     Params
	World [][]byte
	Turn  int
}

type Engine struct{}

var WORLD [][]byte
var TURNS int
var PARAMS Params

// Start COMMENT
func (e *Engine) Start(args Args, reply *[][]byte) (err error) {
	TURNS = args.Turn
	WORLD = distributor(args.P, args.World)
	*reply = WORLD

	return
}

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread() // not sure what this does but was in skeleton

	// Port for connection to controller
	portPtr := flag.String("port", ":8030", "listening on this port")
	flag.Parse()                             // call after all flags are defined to parse command line into flags
	rpc.Register(&Engine{})                  // WHAT DOES THIS DO?
	ln, error := net.Listen("tcp", *portPtr) // listens for connections
	if error != nil {                        // produces error message if fails to connect
		log.Fatal("Unable to connect:", error)
	}
	defer ln.Close() // stops execution until surrounding functions return
	rpc.Accept(ln)   // accepts connections on ln and serves requests to server for each incoming connection
}
