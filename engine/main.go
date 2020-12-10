package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"runtime"
)

// Args COMMENT
type Args struct {
	P     Params
	World [][]byte
}

type Engine struct{}

type AliveCellsReply struct {
	AliveCells     int
	CompletedTurns int
}

type SaveReply struct {
	CompletedTurns int
	World          [][]byte
}

type PauseReply struct {
	CompletedTurns int
	World          [][]byte
}

type IsAlreadyRunningReply struct {
	P                Params
	World            [][]byte
	IsAlreadyRunning bool
}

var WORLD [][]byte
var PARAMS Params
var ALIVECELLS int
var COMPLETEDTURNS = 0
var PAUSECHANNEL = make(chan bool, 1)
var FINISHEDCHANNEL = make(chan [][]byte, 1)
var CANCELCHANNEL = make(chan bool, 1)
var NUMBEROFCONTINUES = 0
var DONECANCELINGCHANNEL = make(chan bool, 1)
var NUMBER_OF_NODES = 2

var NODE_ADDRESS = "3.86.98.70:8031"

func (e *Engine) IsAlreadyRunning(p Params, reply *bool) (err error) {
	if COMPLETEDTURNS-1 > 0 {
		if PARAMS == p {
			*reply = true
			return
		} else {
			//break the already running distributor and then reply false to set up a new one
			CANCELCHANNEL <- true
			<-DONECANCELINGCHANNEL
			*reply = false
			return
		}
	}
	*reply = false
	return
}

// Start function
func (e *Engine) Start(args Args, reply *[][]byte) (err error) {
	PARAMS = args.P

	if NUMBER_OF_NODES == 0 {
		WORLD = distributor(args.P, args.World)
	} else {
		// workerHeight := args.P.ImageHeight / NUMBER_OF_NODES
		// remainderHeight := args.P.ImageHeight % NUMBER_OF_NODES

		// var splitHeight int
		// if remainderHeight > 0 {
		// 	splitHeight = workerHeight + 1
		// } else {
		// 	splitHeight = workerHeight
		// }

		nodeConnection(NODE_ADDRESS)
	}
	*reply = WORLD

	return
}

// Continue function
func (e *Engine) Continue(x int, reply *[][]byte) (err error) {
	NUMBEROFCONTINUES++
	*reply = <-FINISHEDCHANNEL
	return
}

// Save function
func (e *Engine) Save(x int, reply *SaveReply) (err error) {
	saveReply := SaveReply{
		CompletedTurns: COMPLETEDTURNS,
		World:          WORLD,
	}
	*reply = saveReply

	return
}

// Pause function
func (e *Engine) Pause(x int, reply *PauseReply) (err error) {
	PAUSECHANNEL <- true
	pauseReply := PauseReply{
		CompletedTurns: COMPLETEDTURNS,
		World:          WORLD,
	}
	*reply = pauseReply

	return
}

// Execute function
func (e *Engine) Execute(x int, reply *PauseReply) (err error) {
	PAUSECHANNEL <- false
	executeReply := PauseReply{
		CompletedTurns: COMPLETEDTURNS,
		World:          WORLD,
	}
	*reply = executeReply

	return
}

// Quit funtion
func (e *Engine) Quit(x int, reply *int) (err error) {
	*reply = COMPLETEDTURNS

	return
}

// GetAliveCells ...
func (e *Engine) GetAliveCells(x int, reply *AliveCellsReply) (err error) {
	aliveCells := AliveCellsReply{
		AliveCells:     ALIVECELLS,
		CompletedTurns: COMPLETEDTURNS,
	}
	*reply = aliveCells

	return
}

func nodeConnection(address string) {
	node, error := rpc.Dial("tcp", address)

	if error != nil {
		log.Fatal("Unable to connect", error)
	} else {
		fmt.Println("SUCCESS!", node)
	}

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
